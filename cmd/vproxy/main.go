package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/bytefmt"
	vproxy "github.com/ieee0824/valve-proxy"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})

	var port string
	flag.StringVar(&port, "port", "3128", "Proxy listen port")
	var upRateStr string
	flag.StringVar(&upRateStr, "u", "10GB", "Proxy upstream bandwidth ratelimit")
	var downRateStr string
	flag.StringVar(&downRateStr, "d", "10GB", "Proxy downstream bandwidth ratelimit")
	flag.Parse()

	upRate, err := bytefmt.ToBytes(upRateStr)
	if err != nil {
		log.Fatal(err)
	}
	downRate, err := bytefmt.ToBytes(downRateStr)
	if err != nil {
		log.Fatal(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	go vproxy.Listen(port, upRate, downRate, shutdown)
	<-shutdown
}
