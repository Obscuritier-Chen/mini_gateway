package gateway

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"mini_gateway/internal/lua"
)

type Gateway struct {
	defaultUpstream string
	upstreamHost    string
	luaEngine       *lua.Engine
	proxy           *httputil.ReverseProxy
}

func New(defaultUpstream string, luaEngine *lua.Engine) (*Gateway, error) {
	remote, err := url.Parse(defaultUpstream)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upstream path: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	return &Gateway{
		defaultUpstream: defaultUpstream,
		upstreamHost:    remote.Host,
		luaEngine:       luaEngine,
		proxy:           proxy,
	}, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = generateRequestID()
	}
	r.Header.Set("X-Request-ID", reqID)
	w.Header().Set("X-Request-ID", reqID)

	log.Printf("[gateway][%s] request received: %s %s", reqID, r.Method, r.URL.Path)

	statusCode, msg, err := g.luaEngine.ExecuteRule(r, reqID)
	if err != nil {
		log.Printf("[gateway] lua rule executed abnormaly: %v", err)
		http.Error(w, "Internal Gateway Error", http.StatusInternalServerError)
		return
	}

	if statusCode != http.StatusOK {
		log.Printf("[gateway] request is intercepted by lua rule: [%d] %s", statusCode, msg)
		w.WriteHeader(statusCode)
		w.Write([]byte(msg))
		return
	}

	r.Header.Set("X-Gateway-By", "Mini-Gateway-Go")
	r.Host = g.upstreamHost

	g.proxy.ServeHTTP(w, r)
}

func generateRequestID() string {
	bytes := make([]byte, 6)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%d-%s", time.Now().UnixNano()/1e6, hex.EncodeToString(bytes))
}
