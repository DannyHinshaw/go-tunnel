package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/proxy"
)

func main() {

	// Docker container running Tor
	proxyAddress := "127.0.0.1:9050"
	proxyUrlStr := fmt.Sprintf("http://%s", proxyAddress)
	proxyUrl, err := url.Parse(proxyUrlStr)
	if err != nil {
		log.Fatal(err)
	}

	// Setup connection to forward requests to Tor via SOCKS5
	s5, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
	if err != nil {
		log.Fatal(err)
	}
	dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		return s5.Dial(network, address)
	}

	// Create reverse proxy to forward requests Tor container.
	reverseProxy := httputil.NewSingleHostReverseProxy(proxyUrl)
	reverseProxy.Transport = &http.Transport{DialContext: dialContext}
	reverseProxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Println("error serving request: ", err)
	}
	log.Printf("configured server: %s", proxyUrl)

	// Create reverse proxy server (tunnel)
	server := http.Server{
		Addr: fmt.Sprintf(":%d", 8080),
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			log.Println("forwarding request to proxy")
			reverseProxy.ServeHTTP(writer, request)
		}),
	}

	// TODO: Try SOCKS5 server
	// Run server
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
