
A proxy that can change the bandwidth of the network.

forked: https://github.com/jamesmoriarty/goforward

# install
```
$ go get github.com/ieee0824/valve-proxy
$ go install github.com/ieee0824/valve-proxy/cmd/vproxy
```

# options
```
$ vproxy -h
Usage of vproxy:
  -d string
        Proxy downstream bandwidth ratelimit (default "10GB")
  -port string
        Proxy listen port (default "3128")
  -u string
        Proxy upstream bandwidth ratelimit (default "10GB")
```

# docker
```
$ sudo docker build -t vproxy .
$ sudo docker run -it -p 3128:3128 vproxy -d 1M -u 1M
INFO[0000] listening on :3128, upload band :1Mbps, download band :1Mbps 
```