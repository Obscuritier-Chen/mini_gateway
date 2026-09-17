package gateway

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Gateway struct {
	defaultUpstream string
}

func New(defaultUpstream string) *Gateway {
	return &Gateway{
		defaultUpstream: defaultUpstream,
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[gateway] request received: %s %s", r.Method, r.URL.Path)

	targetURL := g.defaultUpstream

	remote, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	r.Header.Set("X-Gateway-By", "Mini-Gateway-Go")
	r.Host = remote.Host

	proxy.ServeHTTP(w, r)
}
