package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

// Prefix defines an AWS IP range
type Prefix struct {
	IPPrefix string `json:"ip_prefix"`
	Region   string `json:"region"`
	Service  string `json:"service"`
}

// AWSIPRange is read from https://ip-ranges.amazonaws.com/ip-ranges.json
type AWSIPRange struct {
	CreateDate string `json:"createDate"`
	SyncToken  string `json:"syncToken"`
	Prefixes   []Prefix
}

const (
	rawURL = "https://ip-ranges.amazonaws.com/ip-ranges.json"
)

func downloadFile(path string) (err error) {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(rawURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func validIP(ip string) bool {
	ipv4 := false
	ipv6 := false

	testIPv4 := net.ParseIP(ip)
	if testIPv4.To4() != nil {
		ipv4 = true
	}

	testIPv6 := net.ParseIP(ip)
	if testIPv6.To16() != nil {
		ipv6 = true
	}
	return ipv4 || ipv6
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aws-ip-check [flags]")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	ipAddress := flag.String("ip", "", "IP address to check if belongs to AWS")
	awsIPRangeFilePath := flag.String("path", "/tmp/aws-ip-ranges.json", "File path to store AWS ip-ranges.json")
	extraInfo := flag.Bool("extra", false, "Print extra info of the IP if pertains to AWS")
	flag.Parse()
	if *ipAddress == "" {
		usage()
		os.Exit(2)
	}

	if valid := validIP(*ipAddress); valid == false {
		fmt.Printf("invalid IP address provided: %s\n", *ipAddress)
		os.Exit(2)
	}

	if _, err := os.Stat(*awsIPRangeFilePath); os.IsNotExist(err) {
		downloadFile(*awsIPRangeFilePath)
	}

	file, err := ioutil.ReadFile(*awsIPRangeFilePath)
	if err != nil {
		fmt.Printf("file error: %v\n", err)
		os.Exit(2)
	}

	var ipRange AWSIPRange
	json.Unmarshal(file, &ipRange)

	contains := false
	var extra string

	for _, r := range ipRange.Prefixes {
		_, cidrnet, err := net.ParseCIDR(r.IPPrefix)
		if err != nil {
			fmt.Printf("error parsing CIDR %s: %v\n", r.IPPrefix, err)
			os.Exit(2)
		}
		ip := net.ParseIP(*ipAddress)
		if cidrnet.Contains(ip) {
			contains = true

			if *extraInfo {
				extra += fmt.Sprintf(",%v;%v;%v", cidrnet.String(), r.Region, r.Service)
			} else {
				break
			}
		}
	}

	if contains {
		fmt.Printf("IP %s found in AWS ip range%s\n", *ipAddress, extra)
		os.Exit(0)
	}

	fmt.Printf("IP %s not found in AWS ip ranges\n", *ipAddress)
	os.Exit(1)
}
