
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