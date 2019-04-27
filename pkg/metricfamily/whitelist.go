package metricfamily

import (
	clientmodel "github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

type whitelist [][]*labels.Matcher

func NewWhitelist(rules []string) (Transformer, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var ms [][]*labels.Matcher
	for i := range rules {
		matchers, err := promql.ParseMetricSelector(rules[i])
		if err != nil {
			return nil, err
		}
		ms = append(ms, matchers)
	}
	return whitelist(ms), nil
}
func (t whitelist) Transform(family *clientmodel.MetricFamily) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	var ok bool
Metric:
	for i, m := range family.Metric {
		if m == nil {
			continue
		}
		for _, matchset := range t {
			if match(family.GetName(), m, matchset...) {
				ok = true
				continue Metric
			}
		}
		family.Metric[i] = nil
	}
	return ok, nil
}
func match(name string, metric *clientmodel.Metric, matchers ...*labels.Matcher) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
Matcher:
	for _, m := range matchers {
		if m.Name == "__name__" && m.Matches(name) {
			continue
		}
		for _, label := range metric.Label {
			if label == nil || m.Name != label.GetName() || !m.Matches(label.GetValue()) {
				continue
			}
			continue Matcher
		}
		return false
	}
	return true
}
