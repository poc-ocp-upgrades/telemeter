package metricfamily

import (
	prom "github.com/prometheus/client_model/go"
)

type elide struct{ labelSet map[string]struct{} }

func NewElide(labels ...string) *elide {
	_logClusterCodePath()
	defer _logClusterCodePath()
	labelSet := make(map[string]struct{})
	for i := range labels {
		labelSet[labels[i]] = struct{}{}
	}
	return &elide{labelSet}
}
func (t *elide) Transform(family *prom.MetricFamily) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if family == nil || len(family.Metric) == 0 {
		return true, nil
	}
	for i := range family.Metric {
		var filtered []*prom.LabelPair
		for j := range family.Metric[i].Label {
			if _, elide := t.labelSet[family.Metric[i].Label[j].GetName()]; elide {
				continue
			}
			filtered = append(filtered, family.Metric[i].Label[j])
		}
		family.Metric[i].Label = filtered
	}
	return true, nil
}
