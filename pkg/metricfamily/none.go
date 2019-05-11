package metricfamily

import clientmodel "github.com/prometheus/client_model/go"

func None(*clientmodel.MetricFamily) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return true, nil
}
