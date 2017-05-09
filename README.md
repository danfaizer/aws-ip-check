# aws-ip-check
`aws-ip-check` is a dummy command line tool to check if an IP address belongs to Amazon Web Service (AWS) infrastructure or not.

This tool relies on public [AWS IP range file](https://ip-ranges.amazonaws.com/ip-ranges.json).

## Installation
Make sure you have a working Go environment.

To install `aws-ip-check` cli, simply run:
```
$ go get github.com/danfaizer/aws-ip-check/cmd/aws-ip-check
```
## Usage
```
$ aws-ip-check help
usage: aws-ip-check [flags]
  -ip string
    	IP address to check if belongs to AWS
  -path string
    	File path to store AWS ip-ranges.json (default "/tmp/aws-ip-ranges.json")
```

`aws-ip-check` will return the following exit statuses:

| status | description |
| ---: | --- |
| 0 | IP belongs to AWS IP range |
| 1 | IP does NOT belong to AWS IP range |
| 2 | aws-ip-check execution error |

## Example
```
$ aws-ip-check -ip 192.168.1.1
IP 192.168.1.1 not found in AWS ip range


$ aws-ip-check -ip $(nslookup www.amazon.com  | grep "Address" | tail -1 | awk '{print $NF}')
IP X.X.X.X found in AWS ip range
```
