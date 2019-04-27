package forwarder

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	from, err := url.Parse("https://redhat.com")
	if err != nil {
		t.Fatalf("failed to parse `from` URL: %v", err)
	}
	toAuthorize, err := url.Parse("https://openshift.com")
	if err != nil {
		t.Fatalf("failed to parse `toAuthorize` URL: %v", err)
	}
	toUpload, err := url.Parse("https://k8s.io")
	if err != nil {
		t.Fatalf("failed to parse `toUpload` URL: %v", err)
	}
	tc := []struct {
		c	Config
		err	bool
	}{{c: Config{}, err: true}, {c: Config{From: from}, err: false}, {c: Config{From: from, ToUpload: toUpload}, err: false}, {c: Config{From: from, ToAuthorize: toAuthorize}, err: true}, {c: Config{From: from, ToToken: "foo"}, err: true}, {c: Config{From: from, ToAuthorize: toAuthorize, ToToken: "foo"}, err: false}, {c: Config{From: from, FromTokenFile: "/this/path/does/not/exist"}, err: true}, {c: Config{From: from, ToTokenFile: "/this/path/does/not/exist"}, err: true}, {c: Config{From: from, AnonymizeSalt: "1"}, err: false}, {c: Config{From: from, AnonymizeLabels: []string{"foo"}}, err: true}, {c: Config{From: from, AnonymizeLabels: []string{"foo"}, AnonymizeSalt: "1"}, err: false}, {c: Config{From: from, AnonymizeLabels: []string{"foo"}, AnonymizeSaltFile: "/this/path/does/not/exist"}, err: true}, {c: Config{From: from, AnonymizeLabels: []string{"foo"}, AnonymizeSalt: "1", AnonymizeSaltFile: "/this/path/does/not/exist"}, err: false}, {c: Config{From: from, FromCAFile: "/this/path/does/not/exist"}, err: true}}
	for i := range tc {
		if _, err := New(tc[i].c); (err != nil) != tc[i].err {
			no := "no"
			if tc[i].err {
				no = "an"
			}
			t.Errorf("test case %d: got '%v', expected %s error", i, err, no)
		}
	}
}
func TestReconfigure(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	from, err := url.Parse("https://redhat.com")
	if err != nil {
		t.Fatalf("failed to parse `from` URL: %v", err)
	}
	c := Config{From: from}
	w, err := New(c)
	if err != nil {
		t.Fatalf("failed to create new worker: %v", err)
	}
	from2, err := url.Parse("https://redhat.com")
	if err != nil {
		t.Fatalf("failed to parse `from2` URL: %v", err)
	}
	tc := []struct {
		c	Config
		err	bool
	}{{c: Config{}, err: true}, {c: Config{From: from2}, err: false}, {c: Config{From: from, FromTokenFile: "/this/path/does/not/exist"}, err: true}}
	for i := range tc {
		if err := w.Reconfigure(tc[i].c); (err != nil) != tc[i].err {
			no := "no"
			if tc[i].err {
				no = "an"
			}
			t.Errorf("test case %d: got %q, expected %s error", i, err, no)
		}
	}
}
func TestRun(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	c := Config{From: &url.URL{}}
	w, err := New(c)
	if err != nil {
		t.Fatalf("failed to create new worker: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var once sync.Once
	var wg sync.WaitGroup
	wg.Add(1)
	ts2 := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		cancel()
		once.Do(wg.Done)
	}))
	defer ts2.Close()
	ts1 := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		go func() {
			from, err := url.Parse(ts2.URL)
			if err != nil {
				t.Fatalf("failed to parse second test server URL: %v", err)
			}
			if err := w.Reconfigure(Config{From: from}); err != nil {
				t.Fatalf("failed to reconfigure worker with second test server url: %v", err)
			}
		}()
	}))
	defer ts1.Close()
	from, err := url.Parse(ts1.URL)
	if err != nil {
		t.Fatalf("failed to parse first test server URL: %v", err)
	}
	if err := w.Reconfigure(Config{From: from}); err != nil {
		t.Fatalf("failed to reconfigure worker with first test server url: %v", err)
	}
	wg.Add(1)
	go func() {
		w.Run(ctx)
		wg.Done()
	}()
	wg.Wait()
}

type fakeRoundTripper struct{ fn func(*http.Request) }

func (frt *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	frt.fn(req)
	return &http.Response{Body: ioutil.NopCloser(bytes.NewBuffer(nil)), StatusCode: http.StatusOK}, nil
}
