package caddy_block_aws

import (
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
