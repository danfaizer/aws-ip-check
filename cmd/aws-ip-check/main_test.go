package main

import (
	"flag"
	"net"
	"os"
	"reflect"
	"testing"

	"gotest.tools/assert"
)

const (
	wrongIPv4   = "a.a.a.a"
	wrongIPv6   = "2001:cdba::3257:9652:XX:XX:XX:XX"
	correctIPv4 = "192.168.0.1"
	correctIPv6 = "2001:cdba::3257:9652"
	awsHostname = "console.aws.amazon.com"
)

func resetFlagsForTesting() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func setEnvVarsForTesting(ev map[string]string) {
	for k, v := range ev {
		os.Setenv(k, v)
	}
}

func resetEnvVarsForTesting(ev map[string]string) {
	for k := range ev {
		os.Unsetenv(k)
	}
}

func mustGetAWSIPv4Address() string {
	addr, err := net.LookupIP(awsHostname)
	if err != nil {
		return "never should reach here"
	}
	return addr[0].String()
}

func TestMainMustReadCfg(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		args        []string
		expectedCfg config
	}{
		{
			name:        "NoIPProvidedNoEnvVarsProvided",
			args:        []string{"aws-ip-check"},
			envVars:     map[string]string{},
			expectedCfg: config{},
		},
		{
			name:    "AWSIPProvidedNoEnvVarsProvided",
			args:    []string{"aws-ip-check", mustGetAWSIPv4Address()},
			envVars: map[string]string{},
			expectedCfg: config{
				ip:            mustGetAWSIPv4Address(),
				cacheFilePath: "",
				cacheTTL:      0,
				extraInfo:     false,
				extraFormat:   "text",
			},
		},
		{
			name:    "AWSIPProvidedExtraInfoNoEnvVarsProvided",
			args:    []string{"aws-ip-check", "-extra", "-format", "text", mustGetAWSIPv4Address()},
			envVars: map[string]string{},
			expectedCfg: config{
				ip:            mustGetAWSIPv4Address(),
				cacheFilePath: "",
				cacheTTL:      0,
				extraInfo:     true,
				extraFormat:   "text",
			},
		},
		{
			name: "IPProvidedEnvVarsProvided",
			args: []string{"aws-ip-check", mustGetAWSIPv4Address()},
			envVars: map[string]string{
				envCachefilePath: "/tmp/test.json",
				envCacheTTL:      "3600",
				envExtraInfo:     "TRUE",
				envExtraFormat:   "text",
			},
			expectedCfg: config{
				ip:            mustGetAWSIPv4Address(),
				cacheFilePath: "/tmp/test.json",
				cacheTTL:      3600,
				extraInfo:     true,
				extraFormat:   "text",
			},
		},
		{
			name: "IPProvidedWrongEnvVarsProvided",
			args: []string{"aws-ip-check", mustGetAWSIPv4Address()},
			envVars: map[string]string{
				envCacheTTL:  "string",
				envExtraInfo: "wrong",
			},
			expectedCfg: config{
				ip:          mustGetAWSIPv4Address(),
				cacheTTL:    0,
				extraInfo:   false,
				extraFormat: "text",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			if tt.args != nil {
				os.Args = tt.args
			}
			setEnvVarsForTesting(tt.envVars)
			defer resetEnvVarsForTesting(tt.envVars)
			resetFlagsForTesting()
			flag.Parse()
			cfg := mustReadCfg()
			if !reflect.DeepEqual(tt.expectedCfg, *cfg) {
				t.Logf("\nexpected cfg: %+v\n  actual cfg: %+v\n", tt.expectedCfg, *cfg)
			}
			assert.Equal(t, true, reflect.DeepEqual(tt.expectedCfg, *cfg))
		})
	}
}

func TestMainRealMain(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config
		expectedStatus int
	}{
		{
			name:           "NoIPProvided",
			cfg:            &config{},
			expectedStatus: 2,
		},
		{
			name:           "WrongIPv4Provided",
			cfg:            &config{ip: wrongIPv4},
			expectedStatus: 2,
		},
		{
			name:           "WrongIPv4Provided",
			cfg:            &config{ip: wrongIPv6},
			expectedStatus: 2,
		},
		{
			name:           "NotAWSIPv4Provided",
			cfg:            &config{ip: correctIPv4},
			expectedStatus: 1,
		},
		{
			name:           "NotAWSIPv6Provided",
			cfg:            &config{ip: correctIPv4},
			expectedStatus: 1,
		},
		{
			name:           "AWSIPProvided",
			cfg:            &config{ip: mustGetAWSIPv4Address()},
			expectedStatus: 0,
		},
		{
			name: "AWSIPProvidedWithExtraInfoText",
			cfg: &config{
				ip:          mustGetAWSIPv4Address(),
				extraInfo:   true,
				extraFormat: "text",
			},
			expectedStatus: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := realMain(tt.cfg)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}
