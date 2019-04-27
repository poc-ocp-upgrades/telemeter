package authorize

type ClusterAuthorizerFunc func(token, cluster string) (subject string, err error)

func (f ClusterAuthorizerFunc) AuthorizeCluster(token, cluster string) (subject string, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return f(token, cluster)
}

type ClusterAuthorizer interface {
	AuthorizeCluster(token, cluster string) (subject string, err error)
}
