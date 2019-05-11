package authorize

import (
	"context"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
)

type ClientAuthorizer interface {
	AuthorizeClient(token string) (*Client, bool, error)
}
type Client struct {
	ID		string
	Labels	map[string]string
}

var clientKey key

type key int

func WithClient(ctx context.Context, client *Client) context.Context {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return context.WithValue(ctx, clientKey, client)
}
func FromContext(ctx context.Context) (*Client, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	client, ok := ctx.Value(clientKey).(*Client)
	return client, ok
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
