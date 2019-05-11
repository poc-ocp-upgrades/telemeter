package metricfamily

import clientmodel "github.com/prometheus/client_model/go"

func DropTimestamp(family *clientmodel.MetricFamily) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if family == nil {
		return true, nil
	}
	for _, m := range family.Metric {
		if m == nil {
			continue
		}
		m.TimestampMs = nil
	}
	return true, nil
}
