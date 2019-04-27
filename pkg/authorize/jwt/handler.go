package jwt

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"github.com/openshift/telemeter/pkg/authorize"
)

type authorizeClusterHandler struct {
	partitionKey	string
	labels		map[string]string
	expireInSeconds	int64
	signer		*Signer
	clusterAuth	authorize.ClusterAuthorizer
}

func NewAuthorizeClusterHandler(partitionKey string, expireInSeconds int64, signer *Signer, labels map[string]string, ca authorize.ClusterAuthorizer) *authorizeClusterHandler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &authorizeClusterHandler{partitionKey: partitionKey, expireInSeconds: expireInSeconds, signer: signer, labels: labels, clusterAuth: ca}
}
func (a *authorizeClusterHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if req.Method != "POST" {
		http.Error(w, "Only POST is allowed to this endpoint", http.StatusMethodNotAllowed)
		return
	}
	req.Body = http.MaxBytesReader(w, req.Body, 4*1024)
	defer req.Body.Close()
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	uniqueIDKey := "id"
	cluster := req.Form.Get(uniqueIDKey)
	if len(cluster) == 0 {
		http.Error(w, fmt.Sprintf("The '%s' parameter must be specified via URL or url-encoded form body", uniqueIDKey), http.StatusBadRequest)
		return
	}
	auth := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if strings.ToLower(auth[0]) != "bearer" {
		http.Error(w, "Only bearer authorization allowed", http.StatusUnauthorized)
		return
	}
	if len(auth) != 2 || len(strings.TrimSpace(auth[1])) == 0 {
		http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
		return
	}
	clientToken := auth[1]
	subject, err := a.clusterAuth.AuthorizeCluster(clientToken, cluster)
	if err != nil {
		type statusCodeErr interface {
			Error() string
			HTTPStatusCode() int
		}
		if scerr, ok := err.(statusCodeErr); ok {
			if scerr.HTTPStatusCode() >= http.StatusInternalServerError {
				log.Printf("error: unable to authorize request: %v", scerr)
			}
			if scerr.HTTPStatusCode() == http.StatusTooManyRequests {
				w.Header().Set("Retry-After", "300")
			}
			http.Error(w, scerr.Error(), scerr.HTTPStatusCode())
			return
		}
		uid := rand.Int63()
		log.Printf("error: unable to authorize request %d: %v", uid, err)
		http.Error(w, fmt.Sprintf("Internal server error, requestid=%d", uid), http.StatusInternalServerError)
		return
	}
	labels := map[string]string{a.partitionKey: cluster}
	for k, v := range a.labels {
		labels[k] = v
	}
	authToken, err := a.signer.GenerateToken(Claims(subject, labels, a.expireInSeconds, []string{"federate"}))
	if err != nil {
		log.Printf("error: unable to generate token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(authorize.TokenResponse{Version: 1, Token: authToken, ExpiresInSeconds: a.expireInSeconds, Labels: labels})
	if err != nil {
		log.Printf("error: unable to marshal token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(data); err != nil {
		log.Printf("writing auth token failed: %v", err)
	}
}
