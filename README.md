# tracetcp-go 
[![Build Status](https://travis-ci.org/0xcafed00d/tracetcp-go.svg?branch=master)](https://travis-ci.org/0xcafed00d/tracetcp-go)

Reimplementation of tracetcp (http://simulatedsimian.github.io/tracetcp.html) in Go.

## Installation:
```bash
$ go get github.com/0xcafed00d/tracetcp-go/cmd/tracetcp
```
Installs tracetcp executable into $GOPATH/bin


## Configuration:
As tracetcp uses raw sockets it needs to be run as root, using sudo. 
To avoid running as root, issue the following command: 

```bash
sudo setcap cap_net_raw=ep tracetcp 
```

If tracetcp is rebuilt, setcap will need to be run again. 

## Usage:
```bash
➤ ./tracetcp www.news.com
Tracing route to 64.30.224.82 (phx1-rb-gtm3-tron-xw-lb.cnet.com) on port 80 over a maximum of 30 hops:

1       4ms     3ms     3ms     Wintermute (192.168.1.1)
2      10ms    10ms     9ms     10.239.152.1
3      11ms     9ms    11ms     perr-core-2a-ae9-609.network.virginmedia.net (62.252.175.129)
4         *       *       *
5         *       *       *
6         *       *       *
7      25ms    16ms    13ms     brhm-bb-1c-ae0-0.network.virginmedia.net (62.254.42.110)
8      18ms    17ms    17ms     213.161.65.149
9         *       *       *
10    194ms   161ms   162ms     ae-1-8.bar1.Phoenix1.Level3.net (4.69.133.29)
11    157ms   155ms   156ms     CBS-CORPORA.bar1.Phoenix1.Level3.net (4.53.106.166)
12    158ms   157ms   158ms     ae2-0.io-phx1-ex8216-1.cnet.com (64.30.227.54)
13 Connected to 64.30.224.82 on port 80
```



