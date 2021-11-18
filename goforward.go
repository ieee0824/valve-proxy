package goforward

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/juju/ratelimit"
	log "github.com/sirupsen/logrus"
)

type RateLimitedConn struct {
	net.Conn
	upBucket   *ratelimit.Bucket
	downBucket *ratelimit.Bucket
}

func (wrap RateLimitedConn) Read(b []byte) (n int, err error) {
	written, err := wrap.Conn.Read(b)

	// wrap.Bucket.Wait(int64(written))
	wrap.downBucket.Wait(int64(written))

	return written, err
}

func (wrap RateLimitedConn) Write(b []byte) (n int, err error) {
	wrap.upBucket.Wait(int64(len(b)))

	return wrap.Conn.Write(b)
}

func handleTunneling(w http.ResponseWriter, r *http.Request, upBucket *ratelimit.Bucket, downBucket *ratelimit.Bucket) {
	conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	destConn := RateLimitedConn{conn, upBucket, downBucket}

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	// up
	go transfer(destConn, clientConn)
	// down
	go transfer(clientConn, destConn)
}

func copyWithLog(destination io.Writer, source io.Reader, description string) {
	start := time.Now()

	if bytes, err := io.Copy(destination, source); err != nil {
		log.Warn(err)
	} else {
		size := bytefmt.ByteSize(uint64(bytes))
		duration := time.Since(start)
		rate := bytefmt.ByteSize(uint64(float64(bytes)/duration.Seconds())) + "/s"

		log.WithFields(log.Fields{"Size": size, "Duration": duration, "Rate": rate}).Info(description)
	}
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()

	copyWithLog(destination, source, "Transferred")
}

func handleHTTP(w http.ResponseWriter, req *http.Request, upBucket *ratelimit.Bucket, downBucket *ratelimit.Bucket) {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 3 * time.Second,
				DualStack: true,
			}
			conn, err := dialer.DialContext(ctx, network, addr)

			return RateLimitedConn{conn, upBucket, downBucket}, err
		},
	}

	resp, err := transport.RoundTrip(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	copyWithLog(w, resp.Body, req.RequestURI)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func Listen(port string, upRate uint64, downRate uint64, shutdown <-chan os.Signal) {
	log.Info("listening on :" + port + ", upload band :" + bytefmt.ByteSize(uint64(upRate)) + "bps, download band :" + bytefmt.ByteSize(uint64(downRate)) + "bps")
	upRate /= 8
	downRate /= 8

	upBucket := ratelimit.NewBucketWithRate(float64(upRate), int64(upRate))
	downBucket := ratelimit.NewBucketWithRate(float64(downRate), int64(downRate))

	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.WithFields(log.Fields{"Method": r.Method, "RemoteAddr": r.RemoteAddr}).Info(r.RequestURI)

			if r.Method == http.MethodConnect {
				handleTunneling(w, r, upBucket, downBucket)
			} else {
				handleHTTP(w, r, upBucket, downBucket)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	go server.ListenAndServe()

	<-shutdown

	log.Info("Goforward Exiting")

	server.Shutdown(context.Background())
}
