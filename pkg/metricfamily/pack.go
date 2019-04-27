package metricfamily

import clientmodel "github.com/prometheus/client_model/go"

func PackMetrics(family *clientmodel.MetricFamily) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	metrics := family.Metric
	j := len(metrics)
	next := 0
Found:
	for i := 0; i < j; i++ {
		if metrics[i] != nil {
			continue
		}
		if next <= i {
			next = i + 1
		}
		for k := next; k < j; k++ {
			if metrics[k] == nil {
				continue
			}
			metrics[i], metrics[k] = metrics[k], nil
			next = k + 1
			continue Found
		}
		family.Metric = metrics[:i]
		break
	}
	return len(family.Metric) > 0, nil
}
func Pack(families []*clientmodel.MetricFamily) []*clientmodel.MetricFamily {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	j := len(families)
	next := 0
Found:
	for i := 0; i < j; i++ {
		if families[i] != nil && len(families[i].Metric) > 0 {
			continue
		}
		if next <= i {
			next = i + 1
		}
		for k := next; k < j; k++ {
			if families[k] == nil || len(families[k].Metric) == 0 {
				continue
			}
			families[i], families[k] = families[k], nil
			next = k + 1
			continue Found
		}
		return families[:i]
	}
	return families
}
