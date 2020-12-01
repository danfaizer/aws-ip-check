package check

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"gotest.tools/assert"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func TestCheck(t *testing.T) {
	tests := []struct {
		name         string
		ip           string
		options      *Options
		found        bool
		wantError    bool
		errorMessage string
	}{
		{
			name:      "HappyPathLocalCacheFoundIP",
			options:   &Options{CacheFilePath: "testdata/ok-ip-ranges.json"},
			ip:        "192.168.0.1",
			found:     true,
			wantError: false,
		},
		{
			name:      "HappyPathLocalCacheNotFoundIP",
			options:   &Options{CacheFilePath: "testdata/ok-ip-ranges.json"},
			ip:        "10.0.0.1",
			found:     false,
			wantError: false,
		},
		{
			name:         "MalformedLocalCacheFile",
			options:      &Options{CacheFilePath: "testdata/malformed-ip-ranges.json"},
			ip:           "10.0.0.1",
			found:        false,
			wantError:    true,
			errorMessage: "unexpected end of JSON input",
		},
		{
			name:         "LocalCacheIsADirectory",
			options:      &Options{CacheFilePath: "testdata/"},
			ip:           "10.0.0.1",
			found:        false,
			wantError:    true,
			errorMessage: "cache file path is a directory [testdata/]",
		},
		{
			name:         "MalformedIPv4",
			ip:           "a.a.a.a",
			found:        false,
			wantError:    true,
			errorMessage: "malformed IP provided [a.a.a.a]",
		},
		{
			name:         "MalformedIPv6",
			ip:           "2001:cdba::3257:9652:XX:XX:XX:XX",
			found:        false,
			wantError:    true,
			errorMessage: "malformed IP provided [2001:cdba::3257:9652:XX:XX:XX:XX]",
		},
		{
			name:      "NonExistingCacheFile",
			options:   &Options{CacheFilePath: "/tmp/random-" + strconv.Itoa(seededRand.Int()) + ".json"},
			ip:        "192.168.0.1",
			found:     false,
			wantError: false,
		},
		{
			name:         "WrongAWSIPRangeURL",
			options:      &Options{IPRangeURL: "https://ip-ranges.amazonaws.com/ip-ranges.json.wrong"},
			ip:           "192.168.0.1",
			found:        false,
			wantError:    true,
			errorMessage: "wrong server response from [https://ip-ranges.amazonaws.com/ip-ranges.json.wrong]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errorDuringTest error
			if tt.options == nil {
				tt.options = &Options{}
			}
			c, err := NewClient(tt.options)
			if err != nil {
				t.Logf("%s", err)
				errorDuringTest = err
			}
			found, _, err := c.Check(tt.ip)
			if err != nil {
				t.Logf("%s", err)
				errorDuringTest = err
			}
			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.wantError, errorDuringTest != nil)
			if errorDuringTest != nil {
				assert.Equal(t, tt.errorMessage, errorDuringTest.Error())
			}
		})
	}
}

func TestCheckCacheTTL(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		options   *Options
		found     bool
		wantError bool
		expired   bool
		wait      time.Duration
	}{
		{
			name: "LocalCacheNotExpired",
			options: &Options{
				CacheFilePath: "testdata/ip-ranges.json",
				CacheTimeout:  5,
			},
			ip:        "13.32.91.219", // AWS IP
			found:     true,
			wantError: false,
			expired:   false,
			wait:      0 * time.Second,
		},
		{
			name: "LocalCacheExpired",
			options: &Options{
				CacheFilePath: "testdata/ip-ranges.json",
				CacheTimeout:  1,
			},
			ip:        "13.32.91.219", // AWS IP
			found:     true,
			wantError: false,
			expired:   true,
			wait:      1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errorDuringTest error
			if tt.options == nil {
				tt.options = &Options{}
			}
			c, err := NewClient(tt.options)
			if err != nil {
				t.Logf("%s", err)
				errorDuringTest = err
			}
			last := c.last
			found, _, err := c.Check(tt.ip)
			if err != nil {
				t.Logf("%s", err)
				errorDuringTest = err
			}
			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.wantError, errorDuringTest != nil)

			time.Sleep(tt.wait)
			found, _, err = c.Check(tt.ip)
			if err != nil {
				t.Logf("%s", err)
				errorDuringTest = err
			}
			assert.Equal(t, tt.expired, last != c.last)
		})
	}
}

func TestCheckExtraInfo(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		options   *Options
		found     bool
		extra     Range
		wantError bool
	}{
		{
			name: "CheckFoundExtraInfo",
			options: &Options{
				CacheFilePath: "testdata/ip-ranges.json",
			},
			ip:        "13.32.91.219", // AWS IP
			found:     true,
			wantError: false,
			extra: Range{
				IPPrefix: "13.32.0.0/15",
				Region:   "GLOBAL",
				Service:  "AMAZON",
			},
		},
		{
			name: "CheckNotFoundExtraInfo",
			options: &Options{
				CacheFilePath: "testdata/ip-ranges.json",
			},
			ip:        "192.168.0.1", // AWS IP
			found:     false,
			wantError: false,
			extra:     Range{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errorDuringTest error
			if tt.options == nil {
				tt.options = &Options{}
			}
			c, err := NewClient(tt.options)
			if err != nil {
				t.Logf("%s", err)
				errorDuringTest = err
			}
			found, extra, err := c.Check(tt.ip)
			if err != nil {
				t.Logf("%s", err)
				errorDuringTest = err
			}
			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.wantError, errorDuringTest != nil)
			assert.Equal(t, tt.extra.IPPrefix, extra.IPPrefix)
			assert.Equal(t, tt.extra.Region, extra.Region)
			assert.Equal(t, tt.extra.Service, extra.Service)
		})
	}
}
