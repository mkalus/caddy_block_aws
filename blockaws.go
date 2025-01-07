package caddy_block_aws

import (
	"context"
	"encoding/json"
	"github.com/paralleltree/ipfilter"
	"github.com/viccon/sturdyc"
	"go.uber.org/zap"
	"net"
	"net/http"
	"time"
)

var loadingAWSData bool

// the matcher itself is global and read only (changed only in createMatcher function)
var matcher *ipfilter.IPMatcher

type AWSIPRange struct {
	IPPrefix string `json:"ip_prefix"`
	// rest is not relevant for our use case
}

type AWSIPv6Range struct {
	IPPrefix string `json:"ipv6_prefix"`
	// rest is not relevant for our use case
}

type AWSData struct {
	SyncToken    string         `json:"syncToken"`
	CreateDate   string         `json:"createDate"`
	Prefixes     []AWSIPRange   `json:"prefixes"`
	IPv6Prefixes []AWSIPv6Range `json:"ipv6_prefixes"`
}

// GetPrefixes returns all prefixes (both IPv4 and IPv6) in the AWSData struct as a slice of strings
func (aws AWSData) GetPrefixes() []string {
	ips := make([]string, len(aws.Prefixes)+len(aws.IPv6Prefixes))
	for i, prefix := range aws.Prefixes {
		ips[i] = prefix.IPPrefix
	}
	for i, prefix := range aws.IPv6Prefixes {
		ips[len(aws.Prefixes)+i] = prefix.IPPrefix
	}
	return ips
}

// loadAWSData loads ip ranges from https://ip-ranges.amazonaws.com/ip-ranges.json and parses them into AWSData struct
func loadAWSData() (*AWSData, error) {
	resp, err := http.Get("https://ip-ranges.amazonaws.com/ip-ranges.json")
	if err != nil {
		return nil, err
	}
	var data AWSData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func createMatcher(data *AWSData) error {
	// create ip matcher from list of ip ranges
	// matcher is global...
	var err error
	matcher, err = ipfilter.NewIPMatcher(data.GetPrefixes())
	if err != nil {
		return err
	}
	return nil
}

// LoadInitialAWSData loads ip ranges from https://ip-ranges.amazonaws.com/ip-ranges.json
func LoadInitialAWSData(logger *zap.Logger) {
	// try to load data only once
	if loadingAWSData {
		return
	}
	loadingAWSData = true
	logger.Info("Updating AWS blocker module")

	data, err := loadAWSData()
	if err != nil {
		logger.Error("Failed to load AWS IP ranges", zap.Error(err))
		return
	}

	if err = createMatcher(data); err != nil {
		logger.Error("Failed to create IP matcher", zap.Error(err))
		return
	}

	logger.Info("Loaded AWS IP ranges", zap.Int("ranges ipv4", len(data.Prefixes)), zap.Int("ranges ipv6", len(data.IPv6Prefixes)), zap.String("createDate", data.CreateDate))
}

// Matches checks if given IP address is in the list of blocked AWS IP addresses
func Matches(ip string) bool {
	if matcher == nil {
		return false // matcher not initialized or some other error occurred - we do not want Caddy to crash
	}
	return matcher.Match(net.ParseIP(ip))
}

// sturdyc.Client is used for caching IP ranges to reduce the number of API calls
var cacheClient *sturdyc.Client[bool]

func init() {
	cacheClient = sturdyc.New[bool](10000, 10, 2*time.Hour, 10)
}

// MatchesWithCache checks if given IP address is in the list of blocked AWS IP addresses, using caching
func MatchesWithCache(ctx context.Context, ip string) bool {
	if matcher == nil {
		return false // matcher not initialized or some other error occurred - we do not want Caddy to crash
	}

	fetchFunc := func(ctx context.Context) (bool, error) {
		return matcher.Match(net.ParseIP(ip)), nil
	}

	result, _ := cacheClient.GetOrFetch(ctx, ip, fetchFunc)
	return result
}
