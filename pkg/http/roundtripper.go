package http

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"unicode/utf8"
)

type bearerRoundTripper struct {
	token	string
	wrapper	http.RoundTripper
}

func NewBearerRoundTripper(token string, rt http.RoundTripper) http.RoundTripper {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &bearerRoundTripper{token: token, wrapper: rt}
}
func (rt *bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", rt.token))
	return rt.wrapper.RoundTrip(req)
}

type debugRoundTripper struct{ next http.RoundTripper }

func NewDebugRoundTripper(next http.RoundTripper) *debugRoundTripper {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &debugRoundTripper{next}
}
func (rt *debugRoundTripper) RoundTrip(req *http.Request) (res *http.Response, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	reqd, _ := httputil.DumpRequest(req, false)
	reqBody := bodyToString(&req.Body)
	res, err = rt.next.RoundTrip(req)
	if err != nil {
		log.Println(err)
		return
	}
	resd, _ := httputil.DumpResponse(res, false)
	resBody := bodyToString(&res.Body)
	log.Printf("request url %v\n%s%s\n------ response\n"+"%s%s\n======\n", req.URL, string(reqd), reqBody, string(resd), resBody)
	return
}
func bodyToString(body *io.ReadCloser) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if *body == nil {
		return "<nil>"
	}
	var b bytes.Buffer
	_, err := b.ReadFrom(*body)
	if err != nil {
		panic(err)
	}
	if err = (*body).Close(); err != nil {
		panic(err)
	}
	*body = ioutil.NopCloser(&b)
	s := b.String()
	if utf8.ValidString(s) {
		return s
	}
	return hex.Dump(b.Bytes())
}
