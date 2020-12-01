package check

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

// Range defines an AWS IP range.
type Range struct {
	IPPrefix string `json:"ip_prefix"`
	Region   string `json:"region"`
	Service  string `json:"service"`
}

// AWSIPRange is read from rawURL const.
type AWSIPRange struct {
	CreateDate string  `json:"createDate"`
	SyncToken  string  `json:"syncToken"`
	Prefixes   []Range `json:"prefixes"`
}

// Client is an aws-ip-check client.
type Client struct {
	opts *Options
	ips  *AWSIPRange
	last int64
}

// Options define the Client options.
type Options struct {
	CacheTimeout  int64
	CacheFilePath string
	IPRangeURL    string
}

const (
	rawURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"
)

func fetchAWSIPRangeFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []byte{}, fmt.Errorf("wrong server response from [%s]", url)
	}

	return ioutil.ReadAll(resp.Body)
}

func updateCacheFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, os.FileMode(0644))
}

func fetchAWSIPRangeFromCacheFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return []byte{}, os.ErrNotExist
	}
	if info.IsDir() {
		return []byte{}, fmt.Errorf("cache file path is a directory [%s]", path)
	}

	return ioutil.ReadFile(path)
}

func (c *Client) updateIPs() error {
	// Lifetime cache already loaded.
	if c.last != 0 && c.opts.CacheTimeout == 0 {
		return nil
	}
	// Cache has not expired yet.
	if c.last != 0 && (c.last+c.opts.CacheTimeout > time.Now().Unix()) {
		return nil
	}
	data := []byte{}
	var err error
	// Fetch the client local cache file.
	if c.opts.CacheFilePath != "" {
		data, err = fetchAWSIPRangeFromCacheFile(c.opts.CacheFilePath)
		if err != nil && err != os.ErrNotExist {
			return err
		}
	}

	// Not using the cache file.
	if len(data) == 0 ||
		// Or cache from file has expired.
		(c.last+c.opts.CacheTimeout < time.Now().Unix() && c.ips != nil) {
		data, err = fetchAWSIPRangeFromURL(c.opts.IPRangeURL)
		if err != nil {
			return err
		}
		// Update cache file data if cache file path is specified in the options.
		if c.opts.CacheFilePath != "" {
			err = updateCacheFile(c.opts.CacheFilePath, data)
			if err != nil {
				return err
			}
		}
	}

	var awsIPRange AWSIPRange
	err = json.Unmarshal(data, &awsIPRange)
	if err != nil {
		return err
	}

	c.ips = &awsIPRange
	c.last = time.Now().Unix()
	return nil
}

// NewClient returns an aws-check-ip Client with specified Options.
// AWSIPRange data is loaded when the Client is being created.
func NewClient(opts *Options) (*Client, error) {
	c := Client{
		opts: opts,
	}

	if c.opts.IPRangeURL == "" {
		c.opts.IPRangeURL = rawURL
	}
	err := c.updateIPs()

	return &c, err
}

// Check checks if the provided IP address belongs to AWS public IP ranges.
func (c *Client) Check(ipStr string) (bool, Range, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, Range{}, fmt.Errorf("malformed IP provided [%s]", ipStr)
	}

	err := c.updateIPs()
	if err != nil {
		return false, Range{}, err
	}

	// TODO: Very inefficient, implement something like this: https://blog.sqreen.com/demystifying-radix-trees/
	for _, p := range c.ips.Prefixes {
		_, cidrnet, err := net.ParseCIDR(p.IPPrefix)
		if err != nil {
			// TODO: Evaluate what to do here.
			continue
		}
		if cidrnet.Contains(ip) {
			return true, p, nil
		}
	}
	return false, Range{}, nil
}
