package metricfamily

import clientmodel "github.com/prometheus/client_model/go"

type Count struct {
	families	int
	metrics		int
}

func (t *Count) Metrics() int {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return t.metrics
}
func (t *Count) Transform(family *clientmodel.MetricFamily) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	t.families++
	t.metrics += len(family.Metric)
	return true, nil
}
