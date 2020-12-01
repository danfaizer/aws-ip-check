[![Build Status](https://travis-ci.org/danfaizer/aws-ip-check.svg?branch=master)](https://travis-ci.org/danfaizer/aws-ip-check)
[![codecov](https://codecov.io/gh/danfaizer/aws-ip-check/branch/master/graph/badge.svg)](https://codecov.io/gh/danfaizer/aws-ip-check)
[![Go Report Card](https://goreportcard.com/badge/github.com/danfaizer/aws-ip-check)](https://goreportcard.com/report/github.com/danfaizer/aws-ip-check)
[![GoDoc](https://godoc.org/github.com/danfaizer/aws-ip-check?status.svg)](https://godoc.org/github.com/danfaizer/aws-ip-check)

# aws-ip-check
`aws-ip-check` provides a go library and a command line tool to check if an IP address belongs to Amazon Web Service (AWS) infrastructure or not.

This tool relies on public [AWS IP range file](https://ip-ranges.amazonaws.com/ip-ranges.json).

## Command line tool
Make sure you have a working `go` environment.

To install `aws-ip-check` cli run:
```bash
go get -x github.com/danfaizer/aws-ip-check/cmd/aws-ip-check
```

### Usage

```bash
$ ~/go/bin/aws-ip-check -help

usage: aws-ip-check [flags] ip
  -cache string
      File path where to store cache.
  -extra
      Provide extra info of the IP if belongs to AWS.
  -format string
      Output format for the IP extra info. (default "text")
  -ttl int
      AWS soruce IP data cache TTL in seconds. 0 means infinite.
```

`aws-ip-check` will return the following exit statuses:

| status | description |
| ---: | --- |
| 0 | IP belongs to AWS IP range |
| 1 | IP does NOT belong to AWS IP range |
| 2 | aws-ip-check execution error |

### Example
```bash
aws-ip-check 192.168.0.1
echo $?
1

aws-ip-check $(host console.aws.amazon.com | tail -1 | awk '{print $NF}')
echo $?
0
```

### Library
```go
import (
    "github.com/danfaizer/aws-ip-check/pkg/check"
)

func CheckAWSIP(ip string) (bool, error) {
  c, err := check.NewClient(&check.Options{})
  if err != nil {
    return false, err
  }

  found, _, err := c.Check("192.168.0.1")
  return found, err
}
```

### Docker
Build the docker image:
```bash
docker build . -t aws-check-ip
```

Run the docker image:
```bash
docker run --rm aws-check-ip 192.168.0.1
echo $?
1

docker run --rm aws-check-ip $(host console.aws.amazon.com | tail -1 | awk '{print $NF}')
echo $?
0
```
