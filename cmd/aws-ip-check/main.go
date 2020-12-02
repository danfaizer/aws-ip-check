package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/danfaizer/aws-ip-check/pkg/check"
)

const (
	// Exit codes
	codeFound = iota
	codeNotFound
	codeError

	// Environment variable configuration
	envCachefilePath = "AWS_IP_CHECK_CACHE_FILE_PATH"
	envCacheTTL      = "AWS_IP_CHECK_CACHE_TTL"
	envExtraInfo     = "AWS_IP_CHECK_EXTRA_INFO"
	envExtraFormat   = "AWS_IP_CHECK_EXTRA_FORMAT"
)

type config struct {
	ip            string
	cacheFilePath string
	cacheTTL      int64
	extraInfo     bool
	extraFormat   string
}

func mustReadCfg() *config {
	var cfg config

	flag.Usage = usage
	cacheFilePath := flag.String("cache", "", "File path where to store cache.")
	cacheTTL := flag.Int64("ttl", 0, "AWS soruce IP data cache TTL in seconds. 0 means infinite.")
	extraInfo := flag.Bool("extra", false, "Provide extra info of the IP if belongs to AWS.")
	extraFormat := flag.String("format", "text", "Output format for the IP extra info.")
	flag.Parse()

	if len(flag.Args()) < 1 {
		return &config{}
	}

	// Set configuration from flags and defaults.
	cfg.ip = flag.Arg(0)
	cfg.cacheFilePath = *cacheFilePath
	cfg.cacheTTL = *cacheTTL
	cfg.extraInfo = *extraInfo
	cfg.extraFormat = *extraFormat

	// If any configuration is defined by env var, overwrite.
	if os.Getenv(envCachefilePath) != "" {
		cfg.cacheFilePath = os.Getenv(envCachefilePath)
	}
	if os.Getenv(envCacheTTL) != "" {
		ttl, err := strconv.ParseInt(os.Getenv(envCacheTTL), 10, 64)
		// If parse rise an error, leave default.
		if err == nil {
			cfg.cacheTTL = ttl
		}
	}
	if os.Getenv(envExtraInfo) != "" {
		extraInfo, err := strconv.ParseBool(os.Getenv(envExtraInfo))
		// If parse rise an error, leave default.
		if err == nil {
			cfg.extraInfo = extraInfo
		}
	}
	if os.Getenv(envExtraFormat) != "" {
		cfg.extraFormat = os.Getenv(envExtraFormat)
	}

	return &cfg
}

func main() {
	os.Exit(realMain(mustReadCfg()))
}

func usage() {
	fmt.Println("usage: aws-ip-check [flags] ip")
	flag.PrintDefaults()
}

func printExtra(format string, ip string, awsRange *check.Range) {
	switch format {
	default:
		fmt.Printf("ip range region service\n")
		fmt.Printf("%s %s %s %s\n", ip, awsRange.IPPrefix, awsRange.Region, awsRange.Service)
	}
}

func realMain(cfg *config) int {
	if cfg.ip == "" {
		usage()
		fmt.Println("ip argument is required")
		return codeError
	}

	c, err := check.NewClient(&check.Options{
		CacheTimeout:  cfg.cacheTTL,
		CacheFilePath: cfg.cacheFilePath,
	})
	if err != nil {
		fmt.Printf("error creating client: %s", err)
		return codeError
	}

	found, awsRange, err := c.Check(cfg.ip)
	if err != nil {
		fmt.Printf("ip [%s] check caused error: %s", cfg.ip, err)
		return codeError
	}
	if !found {
		return codeNotFound
	}
	if cfg.extraInfo {
		printExtra(cfg.extraFormat, cfg.ip, &awsRange)
	}

	return codeFound
}
