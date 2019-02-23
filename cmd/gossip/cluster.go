package gossip

import (
	"github.com/hashicorp/memberlist"
	"sync"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/pkg/errors"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log"
	"github.com/oklog/ulid"
	"net"
	"strconv"
	"strings"
	"context"
	"math/rand"
	"io/ioutil"
	"github.com/go/src/fmt"
	"github.com/improbable-eng/thanos/pkg/runutil"
)

// Peer is a single peer in a gossip cluster.
type Peer struct {
	logger               log.Logger
	mlistMtx             sync.RWMutex
	mlist                *memberlist.Memberlist
	stopc                chan struct{}
	cfg                  *memberlist.Config
	knownPeers           []string
	advertiseAddr        string
	refreshInterval      time.Duration
	data                 *data
	gossipMsgsReceived   prometheus.Counter
	gossipClusterMembers prometheus.Gauge
}

const (
	DefaultPushPullInterval = 5 * time.Second
	DefaultGossipInterval   = 5 * time.Second
	DefaultRefreshInterval  = 60 * time.Second
)

// PeerType describes a peer's role in the cluster.
type PeerType string

// Constants holding valid PeerType values.
const (
	// PeerTypeMaster is for peers that implements WriteApi and are used for writing data.
	PeerTypeMaster = "master"
	// PeerTypeSlave is for peers that implements  QueryAPI and are used for read data.
	PeerTypeSlave = "slave"
)

// PeerMetadata are the information that can change in runtime of the peer.
type PeerMetadata struct {
	Version      int64
	LastSyncTime time.Time
}

func (this PeerMetadata) ToString() string {
	return fmt.Sprintf("Version=%s,LastSyncTime=%s", this.Version, this.LastSyncTime.Local().String())
}

// PeerState contains state for the peer.
type PeerState struct {
	// Type represents type of the peer holding the state.
	Type PeerType

	// StoreAPIAddr is a host:port address of gRPC StoreAPI of the peer holding the state. Required for PeerTypeSource and PeerTypeStore.
	StoreAPIAddr string
	// QueryAPIAddr is a host:port address of HTTP QueryAPI of the peer holding the state. Required for PeerTypeQuery type only.
	QueryAPIAddr string

	// Metadata holds metadata of the peer holding the state.
	Metadata PeerMetadata
}

// Join joins to the memberlist gossip cluster using knownPeers and given peerType and initialMetadata.
func (p *Peer) Join(peerType PeerType, initialMetadata PeerMetadata) error {

	if p.hasJoined() {
		return errors.New("peer already joined; close it first to rejoin")
	}

	var ml *memberlist.Memberlist
	d := newDelegate(p.logger, ml.NumMembers, p.data, p.gossipMsgsReceived, p.gossipClusterMembers)
	p.cfg.Delegate = d
	p.cfg.Events = d

	ml, err := memberlist.Create(p.cfg)
	if err != nil {
		return errors.Wrap(err, "create memberlist")
	}

	n, _ := ml.Join(p.knownPeers)
	level.Debug(p.logger).Log("msg", "joined cluster", "peerType", peerType)

	if n > 0 {
		go warnIfAlone(p.logger, 10*time.Second, p.stopc, ml.NumMembers)
	}

	p.mlistMtx.Lock()
	p.mlist = ml
	p.mlistMtx.Unlock()

	// Initialize state with ourselves.
	p.data.Set(p.Name(), PeerState{
		Type:     peerType,
		Metadata: initialMetadata,
	})

	if p.refreshInterval != 0 {
		go p.periodicallyRefresh()
	}

	return nil
}

func (p *Peer) periodicallyRefresh() {
	tick := time.NewTicker(p.refreshInterval)
	defer tick.Stop()

	for {
		select {
		case <-p.stopc:
			return
		case <-tick.C:
			if err := p.Refresh(); err != nil {
				level.Error(p.logger).Log("msg", "Refreshing memberlist", "err", err)
			}
		}
	}
}

func warnIfAlone(logger log.Logger, d time.Duration, stopc chan struct{}, numNodes func() int) {
	tick := time.NewTicker(d)
	defer tick.Stop()

	for {
		select {
		case <-stopc:
			return
		case <-tick.C:
			if n := numNodes(); n <= 1 {
				level.Warn(logger).Log("NumMembers", n, "msg", "I appear to be alone in the cluster")
			}
		}
	}
}

// Refresh renews membership cluster, this will refresh DNS names and join newly added members
func (p *Peer) Refresh() error {
	p.mlistMtx.Lock()
	defer p.mlistMtx.Unlock()

	if p.mlist == nil {
		return nil
	}

	resolvedPeers, err := resolvePeers(context.Background(), p.knownPeers, p.advertiseAddr, *net.DefaultResolver, false)
	if err != nil {
		return errors.Wrapf(err, "refresh cluster could not resolve peers: %v", resolvedPeers)
	}

	currMembers := p.mlist.Members()
	var notConnected []string
	for _, peer := range resolvedPeers {
		var isPeerFound bool

		for _, mem := range currMembers {
			if mem.Address() == peer {
				isPeerFound = true
				break
			}
		}

		if !isPeerFound {
			notConnected = append(notConnected, peer)
		}
	}

	if len(notConnected) == 0 {
		level.Debug(p.logger).Log("msg", "refresh cluster done", "peers", strings.Join(p.knownPeers, ","), "resolvedPeers", strings.Join(resolvedPeers, ","))
		return nil
	}

	curr, err := p.mlist.Join(notConnected)
	if err != nil {
		level.Error(p.logger).Log("msg", "refresh cluster could not join peers", "peers", strings.Join(notConnected, ","))
	}

	level.Debug(p.logger).Log("msg", "refresh cluster done, peers joined", "peers", strings.Join(notConnected, ","), "before", len(currMembers), "after", curr)

	return nil
}

func (p *Peer) hasJoined() bool {
	p.mlistMtx.RLock()
	defer p.mlistMtx.RUnlock()
	return p.mlist != nil
}

// SetTimestamps updates internal metadata's timestamps stored in PeerState for this peer.
// Note that this data will be propagated based on gossipInterval we set.
func (p *Peer) SetTimestamps(mint int64, maxt int64) {
	if !p.hasJoined() {
		return
	}
	s, _ := p.data.Get(p.Name())
	s.Metadata.Version += 1
	s.Metadata.LastSyncTime = time.Now()
	p.data.Set(p.Name(), s)
}

// Close leaves the cluster waiting up to timeout and shutdowns peer if cluster left.
// TODO(bplotka): Add this method into run.Group closing logic for each command. This will improve graceful shutdown.
func (p *Peer) Close(timeout time.Duration) {
	if !p.hasJoined() {
		return
	}

	if err := p.mlist.Leave(timeout); err != nil {
		level.Error(p.logger).Log("msg", "memberlist leave failed", "err", err)
	}
	close(p.stopc)
	if err := p.mlist.Shutdown(); err != nil {
		level.Error(p.logger).Log("msg", "memberlist shutdown failed", "err", err)
	}
	p.mlist = nil
}

// Name returns the unique ID of this peer in the cluster.
func (p *Peer) Name() string {
	if !p.hasJoined() {
		return ""
	}

	return p.mlist.LocalNode().Name
}

func New(
	l log.Logger,
	reg *prometheus.Registry,
	bindAddr string,
	advertiseAddr string,
	knownPeers []string,
	waitIfEmpty bool,
	pushPullInterval time.Duration,
	gossipInterval time.Duration,
	refreshInterval time.Duration,

) (*Peer, error) {

	l = log.With(l, "component", "cluster")
	bindHost, bindPortStr, err := net.SplitHostPort(bindAddr)
	if err != nil {
		return nil, err
	}
	bindPort, err := strconv.Atoi(bindPortStr)
	if err != nil {
		return nil, errors.Wrap(err, "invalid listen address")
	}

	// Best-effort deduction of advertise address.
	advertiseHost, advertisePort, err := CalculateAdvertiseAddress(bindAddr, advertiseAddr)
	if err != nil {
		level.Warn(l).Log("err", "couldn't deduce an advertise address: "+err.Error())
	}

	if IsUnroutable(advertiseHost) {
		level.Warn(l).Log("err", "this node advertises itself on an unroutable address", "host", advertiseHost, "port", advertisePort)
		level.Warn(l).Log("err", "this node will be unreachable in the cluster")
		level.Warn(l).Log("err", "provide --cluster.advertise-address as a routable IP address or hostname")
	}

	resolvedPeers, err := resolvePeers(context.Background(), knownPeers, advertiseAddr, *net.DefaultResolver, waitIfEmpty)
	if err != nil {
		return nil, errors.Wrap(err, "resolve peers")
	}
	level.Debug(l).Log("msg", "resolved peers to following addresses", "peers", strings.Join(resolvedPeers, ","))

	// TODO(fabxc): generate human-readable but random names?
	name, err := ulid.New(ulid.Now(), rand.New(rand.NewSource(time.Now().UnixNano())))
	if err != nil {
		return nil, err
	}

	cfg := memberlist.DefaultLANConfig()
	cfg.Name = name.String()
	cfg.BindAddr = bindHost
	cfg.BindPort = bindPort
	cfg.GossipInterval = gossipInterval
	cfg.PushPullInterval = pushPullInterval
	cfg.LogOutput = ioutil.Discard
	if advertiseAddr != "" {
		cfg.AdvertiseAddr = advertiseHost
		cfg.AdvertisePort = advertisePort
	}

	gossipMsgsReceived := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "thanos_gossip_messages_received_total",
		Help: "Total gossip NotifyMsg calls.",
	})
	gossipClusterMembers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "thanos_cluster_members",
		Help: "Number indicating current number of members in cluster.",
	})

	reg.MustRegister(gossipMsgsReceived)
	reg.MustRegister(gossipClusterMembers)

	return &Peer{
		logger:               l,
		knownPeers:           knownPeers,
		cfg:                  cfg,
		refreshInterval:      refreshInterval,
		gossipMsgsReceived:   gossipMsgsReceived,
		gossipClusterMembers: gossipClusterMembers,
		stopc:                make(chan struct{}),
		data:                 &data{data: map[string]PeerState{}},
		advertiseAddr:        advertiseAddr,
	}, nil

}

func IsUnroutable(host string) bool {
	if ip := net.ParseIP(host); ip != nil && (ip.IsUnspecified() || ip.IsLoopback()) {
		return true // typically 0.0.0.0 or localhost
	} else if ip == nil && strings.ToLower(host) == "localhost" {
		return true
	}
	return false
}

func removeMyAddr(ips []net.IPAddr, targetPort string, myAddr string) []net.IPAddr {
	var result []net.IPAddr

	for _, ip := range ips {
		if net.JoinHostPort(ip.String(), targetPort) == myAddr {
			continue
		}
		result = append(result, ip)
	}
	return result
}

func resolvePeers(ctx context.Context, peers []string, myAddress string, res net.Resolver, waitIfEmpty bool) ([]string, error) {
	var resolvedPeers []string

	for _, peer := range peers {
		host, port, err := net.SplitHostPort(peer)
		if err != nil {
			return nil, errors.Wrapf(err, "split host/port for peer %s", peer)
		}

		ips, err := res.LookupIPAddr(ctx, host)
		if err != nil {
			// Assume direct address.
			resolvedPeers = append(resolvedPeers, peer)
			continue
		}

		if len(ips) == 0 {
			var lookupErrSpotted bool
			retryCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			err := runutil.Retry(2*time.Second, retryCtx.Done(), func() error {
				if lookupErrSpotted {
					// We need to invoke cancel in next run of retry when lookupErrSpotted to preserve LookupIPAddr error.
					cancel()
				}

				ips, err = res.LookupIPAddr(retryCtx, host)
				if err != nil {
					lookupErrSpotted = true
					return errors.Wrapf(err, "IP Addr lookup for peer %s", peer)
				}

				ips = removeMyAddr(ips, port, myAddress)
				if len(ips) == 0 {
					if !waitIfEmpty {
						return nil
					}
					return errors.New("empty IPAddr result. Retrying")
				}

				return nil
			})
			if err != nil {
				return nil, err
			}
		}

		for _, ip := range ips {
			resolvedPeers = append(resolvedPeers, net.JoinHostPort(ip.String(), port))
		}
	}

	return resolvedPeers, nil
}
