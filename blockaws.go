package caddy_block_aws

import (
	"encoding/json"
	"github.com/paralleltree/ipfilter"
	"go.uber.org/zap"
	"net"
	"net/http"
)

var matcher *ipfilter.IPMatcher

type AWSIPRange struct {
	IPPrefix string `json:"ip_prefix"`
	// rest is not relevant for our use case
}

type AWSData struct {
	SyncToken  string       `json:"syncToken"`
	CreateDate string       `json:"createDate"`
	Prefixes   []AWSIPRange `json:"prefixes"`
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
	// create list of ip ranges
	ips := make([]string, len(data.Prefixes))
	for i, prefix := range data.Prefixes {
		ips[i] = prefix.IPPrefix
	}

	// create ip matcher from list of ip ranges
	// matcher is global...
	var err error
	matcher, err = ipfilter.NewIPMatcher(ips)
	if err != nil {
		return err
	}
	return nil
}

// LoadInitialAWSData loads ip ranges from https://ip-ranges.amazonaws.com/ip-ranges.json
func LoadInitialAWSData(logger *zap.Logger) {
	data, err := loadAWSData()
	if err != nil {
		logger.Error("Failed to load AWS IP ranges", zap.Error(err))
		return
	}

	if err = createMatcher(data); err != nil {
		logger.Error("Failed to create IP matcher", zap.Error(err))
		return
	}

	logger.Info("Loaded AWS IP ranges", zap.Int("ip_count", len(data.Prefixes)))
}

// Matches checks if given IP address is in the list of blocked AWS IP addresses
func Matches(ip string) bool {
	if matcher == nil {
		return false // matcher not initialized or some other error occurred - we do not want Caddy to crash
	}
	return matcher.Match(net.ParseIP(ip))
}
