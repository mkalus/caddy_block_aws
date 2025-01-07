package caddy_block_aws

import (
	"context"
	"crypto/rand"
	"net"
	"testing"
)

func Test_createMatcherAndMatch(t *testing.T) {
	// reset matcher
	matcher = nil

	data := &AWSData{
		Prefixes: []AWSIPRange{
			{IPPrefix: "198.51.100.0/24"},
			{IPPrefix: "203.0.113.0/24"},
		},
	}

	err := createMatcher(data)
	if err != nil {
		t.Error(err)
	}

	if matcher == nil {
		t.Error("matcher is nil")
	}

	if !matcher.Match(net.ParseIP("198.51.100.1")) {
		t.Error("expected match for 198.51.100.1")
	}
}

func Test_Matches(t *testing.T) {
	// reset matcher
	matcher = nil

	if Matches("192.51.100.1") { // match vs. nil matcher should work
		t.Error("expected no match for 192.51.100.1")
	}

	data := &AWSData{
		Prefixes: []AWSIPRange{
			{IPPrefix: "198.51.100.0/24"},
			{IPPrefix: "203.0.113.0/24"},
		},
	}

	err := createMatcher(data)
	if err != nil {
		t.Error(err)
	}

	if !Matches("198.51.100.1") {
		t.Error("expected match for 198.51.100.1")
	}

	if Matches("203.5.113.1") {
		t.Error("expected no match for 203.5.113.1")
	}

	if Matches("65") { // invalid address
		t.Error("expected no match for invalid address")
	}
}

func Benchmark_Matches(b *testing.B) {
	ips, err := populateTestIPs()
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.Run("Matches", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, ip := range ips {
				Matches(ip)
			}
		}
	})

	b.Run("MatchesWithCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, ip := range ips {
				MatchesWithCache(ctx, ip)
			}
		}
	})
}

func populateTestIPs() ([]string, error) {
	// load AWS ranges
	data, err := loadAWSData()
	if err != nil {
		return nil, err
	}

	// populate matcher
	if err = createMatcher(data); err != nil {
		return nil, err
	}

	// create random ips to test against
	prefixes := data.GetPrefixes()
	ips := make([]string, 5000+len(prefixes))
	for i := 0; i < 5000; i++ {
		ips[i] = makeRandomIP() // to test, we will only create IPv4 here
	}

	// now from the block ranges
	for i, prefix := range prefixes {
		ips[5000+i] = randomIPFromRange(prefix)
	}

	return ips, nil
}

// create a random IP address
// inspired by https://gist.github.com/porjo/f1e6b79af77893ee71e857dfba2f8e9a
func makeRandomIP() string {
	r := make([]byte, 4)
	_, _ = rand.Read(r)

	return net.IP(r).String()
}

// create a random IP address within a given CIDR range
// inspired by https://gist.github.com/dontlaugh/bec85f9792613a0898eb4d558c146d02
// but with major changes to work with IPv6
func randomIPFromRange(cidr string) string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return ""
	}

	// The number of leading 1s in the mask
	// and the mask size in total
	ones, maskSize := ipnet.Mask.Size()

	// single address range
	if ones == maskSize {
		return ip.String()
	}

	// create random addressSized byte slice
	addressSize := maskSize / 8
	r := make([]byte, addressSize)
	_, err = rand.Read(r)
	if err != nil {
		return ""
	}

	quotient := ones / 8
	remainder := ones % 8

	for i := 0; i <= quotient; i++ {
		if i == quotient {
			shifted := r[i] >> remainder
			r[i] = ^ipnet.IP[i] & shifted
		} else {
			r[i] = ipnet.IP[i]
		}
	}

	return net.IP(r).String()
}
