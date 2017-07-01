package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"

	"github.com/jarcoal/httpmock"
)

const (
	wrongPath   = "/wrong/file.json"
	correctPath = "/tmp/file.json"
	wrongRawURL = "https://ip-ranges.amazonaws.com/ip-ranges.json.wrong"
	wrongIPv4   = "a.a.a.a"
	wrongIPv6   = "2001:cdba::3257:9652:XX:XX:XX:XX"
	correctIPv4 = "192.168.1.1"
	correctIPv6 = "2001:cdba::3257:9652"
)

func getAWSIPAddress() (net.IP, error) {
	addr, err := net.LookupIP("www.amazon.com")
	if err != nil {
		return net.IPv6loopback, errors.New("Unable to resolve www.amazon.com")
	}
	return addr[0], nil
}

func TestDownloadFileWrongPath(t *testing.T) {
	var err error

	err = downloadFile(correctPath)
	if err != nil {
		t.Errorf("Test failed, expected error due to wrong path: '%s'", wrongPath)
	}
}

func TestDownloadFileError(t *testing.T) {
	var err error

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", rawURL,
		httpmock.NewStringResponder(404, `NOT FOUND`))

	err = downloadFile(correctPath)
	if err == nil {
		t.Errorf("Test failed, expected error due to rawURL request response: '%s'", err)
	}
}

func TestDownloadFile(t *testing.T) {
	var err error

	err = downloadFile(correctPath)
	if err != nil {
		t.Errorf("Test failed, not error expected: '%s'", err)
	}
}

func TestValidIPWrongIPv4(t *testing.T) {
	var b bool
	b = validIP(wrongIPv4)
	if b {
		t.Errorf("Test failed, expecting FALSE ip: '%s'", wrongIPv4)
	}
}

func TestValidIPCorrectIPv4(t *testing.T) {
	var b bool
	b = validIP(correctIPv4)
	if !b {
		t.Errorf("Test failed, expecting TRUE ip: '%s'", correctIPv4)
	}
}

func TestValidIPWrongIPv6(t *testing.T) {
	var b bool
	b = validIP(wrongIPv6)
	if b {
		t.Errorf("Test failed, expecting FALSE ip: '%s'", wrongIPv6)
	}
}

func TestValidIPCorrectIPv6(t *testing.T) {
	var b bool
	b = validIP(correctIPv6)
	if !b {
		t.Errorf("Test failed, expecting TRUE ip: '%s'", correctIPv6)
	}
}

func TestRealMainWithWrongNoArgs(t *testing.T) {
	var ret int
	ret = realMain()
	if ret != 2 {
		t.Errorf("Test failed, no arguments provided to cmd")
	}
}

func TestRealMain(t *testing.T) {
	t.Run("NonAWSIPAddress", func(t *testing.T) {
		var outbuf bytes.Buffer
		args := []string{"run", "main.go", "-ip=192.168.1.1"}
		cmd := exec.Command("go", args...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = &outbuf
		cmd.Stderr = os.Stderr
		cmd.Run()
		res := cmd.ProcessState
		if res.String() != "exit status 1" {
			t.Errorf("Test failed, unexpected exist status 1 for non AWS range IP address")
		}
		if outbuf.String() != "IP 192.168.1.1 not found in AWS ip ranges\n" {
			t.Errorf("Test failed, unexpected output message: %s", outbuf.String())
		}
	})

	t.Run("AWSIPAddress", func(t *testing.T) {
		var outbuf bytes.Buffer
		awsIP, err := getAWSIPAddress()
		if err != nil {
			t.Error("Test failed, unable to get a valid AWS IP adddress")
		}
		ipArg := fmt.Sprintf("-ip=%s", awsIP)
		args := []string{"run", "main.go", ipArg}
		cmd := exec.Command("go", args...)
		cmd.Env = os.Environ()
		cmd.Stdin = os.Stdin
		cmd.Stdout = &outbuf
		cmd.Stderr = os.Stderr
		cmd.Run()
		res := cmd.ProcessState
		if res.String() != "exit status 0" {
			t.Errorf("Test failed, unexpected exist status 0 for AWS range IP address")
		}
		awsIPoutput := fmt.Sprintf("IP %s found in AWS ip range\n", awsIP)
		if outbuf.String() != awsIPoutput {
			t.Errorf("Test failed, unexpected output message: %s", outbuf.String())
		}
	})
}
