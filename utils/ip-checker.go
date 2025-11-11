package utils

type IPCheckerStruct struct {
	apiKey string
}

// IPChecker is a factory method that acts as a static member
func IPChecker() *IPCheckerStruct {
	return &IPCheckerStruct{
		apiKey: "A804D17F1EE16FBE269FE00610B95C97",
	}
}

/**


package common

import (
	"bitbucket.org/shieldiot/pulse/pulse-iprep/config"
	"bitbucket.org/shieldiot/pulse/pulse-iprep/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-yaaf/yaaf-common/logger"
	"github.com/go-yaaf/yaaf-common/utils/cache"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// inFlightTrackerData to track in-flight requests for IP address reputation.
type inFlightTrackerData struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type IpReputationService struct {
	ipRepSvcUri string

	//local in-memory, 1st level cache for IPStatus data
	localCache *cache.Cache[string, model.IPStatus]

	//in-flight request tracker
	//use standard map and manual locking,
	//since there are operation that need to be performed
	//as "atomic", inside critical section
	tracker     map[string]inFlightTrackerData
	trackerLock sync.RWMutex

	blackIpsCounter,
	whiteIpsCounter int64

	//for u-test purposes only
	TestCallback func(ip string)
}

func NewIpReputationService(ipRepSvcUri string) *IpReputationService {

	c := cache.NewTtlCache[string, model.IPStatus]()

	// Cloudflare DNS
	c.Set("1.0.0.1", whiteIp("1.0.0.1"))
	c.Set("1.1.1.1", whiteIp("1.1.1.1"))

	// Google Public DNS
	c.Set("8.8.4.4", whiteIp("8.8.4.4"))
	c.Set("8.8.8.8", whiteIp("8.8.8.8"))

	// Quad9 DNS
	c.Set("9.9.9.9", whiteIp("9.9.9.9"))

	// Comodo Secure DNS
	c.Set("8.26.56.26", whiteIp("8.26.56.26"))
	c.Set("8.20.247.20", whiteIp("8.20.247.20"))

	// OpenDNS
	c.Set("208.67.222.222", whiteIp("208.67.222.222"))
	c.Set("208.67.220.220", whiteIp("208.67.220.220"))

	// Root DNS Servers
	c.Set("198.41.0.4", whiteIp("198.41.0.4"))         // Root DNS Server A)
	c.Set("192.228.79.201", whiteIp("192.228.79.201")) // Root DNS Server B)
	c.Set("192.33.4.12", whiteIp("192.33.4.12"))       // Root DNS Server C)
	c.Set("199.7.91.13", whiteIp("199.7.91.13"))       // Root DNS Server D)
	c.Set("192.203.230.10", whiteIp("192.203.230.10")) // Root DNS Server E)
	c.Set("192.5.5.241", whiteIp("192.5.5.241"))       // Root DNS Server F)
	c.Set("192.112.36.4", whiteIp("192.112.36.4"))     // Root DNS Server G)
	c.Set("198.97.190.53", whiteIp("198.97.190.53"))   // Root DNS Server H)
	c.Set("192.36.148.17", whiteIp("192.36.148.17"))   // Root DNS Server I)
	c.Set("192.58.128.30", whiteIp("192.58.128.30"))   // Root DNS Server J)
	c.Set("193.0.14.129", whiteIp("193.0.14.129"))     // Root DNS Server K)
	c.Set("199.7.83.42", whiteIp("199.7.83.42"))       // Root DNS Server L)
	c.Set("202.12.27.33", whiteIp("202.12.27.33"))     // Root DNS Server M)

	c.Set("92.68.2.1", whiteIp("92.68.2.1"))

	return &IpReputationService{
		localCache:  c,
		tracker:     make(map[string]inFlightTrackerData),
		trackerLock: sync.RWMutex{},
		ipRepSvcUri: ipRepSvcUri,
	}
}

// GetSingle retrieves the reputation status of a single IP address.
//
// Parameters:
// - ip: The IP address to check.
//
// Returns:
// - model.IPStatus: The status of the IP address, defaulting to WHITE_IP if any errors occur.
func (s *IpReputationService) GetSingle(ip string) (val model.IPStatus) {

	var (
		err   error
		found bool
	)

	// set default as WHITE_IP
	val = whiteIp(ip)

	// hit in-memory cache first
	if val, found = s.localCache.Get(ip); found {
		return
	}

	//test if IP is private
	if IsPrivateIP(ip) {
		return val
	}
	//test if belongs to CenSys
	if IsCenSysIP(ip) {
		return val
	}

	//check for in-flight request for the same IP address
	s.trackerLock.Lock()
	if tracker, found := s.tracker[ip]; !found {
		//add tracker. This operation - create value AND set it to map -  MUST be atomic
		ctx, cancel := context.WithCancel(context.Background())
		s.tracker[ip] = inFlightTrackerData{
			ctx:    ctx,
			cancel: cancel,
		}
		s.trackerLock.Unlock()
	} else {
		//if tracker found, we have its private copy, so unlock map
		s.trackerLock.Unlock()

		//wait for a signal of completion
		<-tracker.ctx.Done()

		//check local cache again
		val, _ = s.localCache.Get(ip)
		return val
	}

	// If cache miss, and there is no in-flight request for the IP,
	// fetch the IP status from the IP reputation service
	// In case of any error, return WHITE_IP but do not cache it, allowing for future attempts
	if val, err = s.fetchSingleIpStatusFromUri(ip); err == nil {
		var toInc *int64
		s.localCache.Set(ip, val)

		if val.IsBlack() {
			toInc = &s.blackIpsCounter
		} else {
			toInc = &s.whiteIpsCounter
		}
		atomic.AddInt64(toInc, 1)
	}

	s.trackerLock.Lock()
	if tracker, found := s.tracker[ip]; found {
		//broadcast completion to all waiters
		tracker.cancel()
		//delete tracker from map, all waiters wait on their local copy
		//(which holds pointer to completion channel )
		delete(s.tracker, ip)
	}
	s.trackerLock.Unlock()

	return
}

// IsInLocalCache checks if the IP address is present in the local cache.
//
// Parameters:
// - ip: The IP address to check.
//
// Returns:
// - model.IPStatus: The status of the IP address if found in the cache, otherwise default to WHITE_IP.
// - bool: A boolean indicating whether the IP address was found in the cache.
func (s *IpReputationService) IsInLocalCache(ip string) (val model.IPStatus, found bool) {

	// assume value not found
	found = false

	// set default as WHITE_IP
	val = whiteIp(ip)
	val, found = s.localCache.Get(ip)

	return
}

func (s *IpReputationService) GetCounters() (whiteIpsCounter, blackIpsCounter int64) {
	return atomic.LoadInt64(&s.whiteIpsCounter), atomic.LoadInt64(&s.blackIpsCounter)
}

// fetchSingleIpStatusFromUri retrieves the reputation status of a single IP address from a remote service.
// It constructs a request to the remote IP reputation service, handles the response, and parses the IP status.
//
// Parameters:
// - ip: The IP address to check.
//
// Returns:
// - model.IPStatus: The status of the IP address.
// - error: Any error that occurred during the process.
func (s *IpReputationService) fetchSingleIpStatusFromUri(ip string) (model.IPStatus, error) {

	var (
		err  error
		resp *http.Response
	)

	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	uri := fmt.Sprintf("%s/check-ip?ip=%s", s.ipRepSvcUri, ip)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Add("X-API-KEY", config.GetConfig().IpRepApiKey())

	if s.TestCallback != nil {
		s.TestCallback(ip)
	}
	if resp, err = client.Do(req); err != nil {
		logger.Error("error when requesting ip status from %s: %s", uri, err)
		return whiteIp(ip), err
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("error when requesting ip status from %s. HTTP response code: %d", uri, resp.StatusCode)
		logger.Error("%s", err)
		return whiteIp(ip), err
	}

	ips := model.IPStatus{}
	if err = json.NewDecoder(resp.Body).Decode(&ips); err != nil {
		logger.Error("error decoding ip status from %s: %s", uri, err)
		return whiteIp(ip), err
	}

	return ips, err
}

func whiteIp(ip string) model.IPStatus {
	return model.IPStatus{
		IP:     ip,
		Status: model.WHITE_IP,
	}
}

*/
