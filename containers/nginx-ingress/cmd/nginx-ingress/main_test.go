package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestValidatePort(t *testing.T) {
	badPorts := []int{80, 443, 1, 1022, 65536}
	for _, badPort := range badPorts {
		err := validatePort(badPort)
		if err == nil {
			t.Errorf("Expected error for port %v\n", badPort)
		}
	}

	goodPorts := []int{8080, 8081, 8082, 1023, 65535}
	for _, goodPort := range goodPorts {
		err := validatePort(goodPort)
		if err != nil {
			t.Errorf("Error for valid port:  %v err: %v\n", goodPort, err)
		}
	}

}

func TestParseNginxStatusAllowCIDRs(t *testing.T) {
	var badCIDRs = []struct {
		input         string
		expectedError error
	}{
		{
			"earth, ,,furball",
			errors.New("invalid IP address: earth"),
		},
		{
			"127.0.0.1,10.0.1.0/24, ,,furball",
			errors.New("invalid CIDR address: an empty string is an invalid CIDR block or IP address"),
		},
		{
			"false",
			errors.New("invalid IP address: false"),
		},
	}
	for _, badCIDR := range badCIDRs {
		_, err := parseNginxStatusAllowCIDRs(badCIDR.input)
		if err == nil {
			t.Errorf("parseNginxStatusAllowCIDRs(%q) returned no error when it should have returned error %q", badCIDR.input, badCIDR.expectedError)
		} else if err.Error() != badCIDR.expectedError.Error() {
			t.Errorf("parseNginxStatusAllowCIDRs(%q) returned error %q when it should have returned error %q", badCIDR.input, err, badCIDR.expectedError)
		}
	}

	var goodCIDRs = []struct {
		input    string
		expected []string
	}{
		{
			"127.0.0.1",
			[]string{"127.0.0.1"},
		},
		{
			"10.0.1.0/24",
			[]string{"10.0.1.0/24"},
		},
		{
			"127.0.0.1,10.0.1.0/24,68.121.233.214 , 24.24.24.24/32",
			[]string{"127.0.0.1", "10.0.1.0/24", "68.121.233.214", "24.24.24.24/32"},
		},
	}
	for _, goodCIDR := range goodCIDRs {
		result, err := parseNginxStatusAllowCIDRs(goodCIDR.input)
		if err != nil {
			t.Errorf("parseNginxStatusAllowCIDRs(%q) returned an error when it should have returned no error: %q", goodCIDR.input, err)
		}

		if !reflect.DeepEqual(result, goodCIDR.expected) {
			t.Errorf("parseNginxStatusAllowCIDRs(%q) returned %v expected %v: ", goodCIDR.input, result, goodCIDR.expected)
		}
	}
}

func TestValidateCIDRorIP(t *testing.T) {
	badCIDRs := []string{"localhost", "thing", "~", "!!!", "", " ", "-1"}
	for _, badCIDR := range badCIDRs {
		err := validateCIDRorIP(badCIDR)
		if err == nil {
			t.Errorf(`Expected error for invalid CIDR "%v"\n`, badCIDR)
		}
	}

	goodCIDRs := []string{"0.0.0.0/32", "0.0.0.0/0", "127.0.0.1/32", "127.0.0.0/24", "23.232.65.42"}
	for _, goodCIDR := range goodCIDRs {
		err := validateCIDRorIP(goodCIDR)
		if err != nil {
			t.Errorf("Error for valid CIDR: %v err: %v\n", goodCIDR, err)
		}
	}
}
