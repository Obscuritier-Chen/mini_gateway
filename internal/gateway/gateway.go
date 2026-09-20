package gateway

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"mini_gateway/internal/lua"
)

type Gateway struct {
	defaultUpstream string
	luaEngine *lua.Engine
}

func New(defaultUpstream string, luaEngine *lua.Engine) *Gateway {
	return &Gateway{
		defaultUpstream: defaultUpstream,
		luaEngine: luaEngine,
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[gateway] request received: %s %s", r.Method, r.URL.Path)

	statusCode, msg, err := g.luaEngine.ExecuteRule(r.URL.Path, r.Header.Get("User-Agent"))
	if err != nil {
		log.Printf("[gateway] lua rule executed abnormaly: %v", err)
		http.Error(w, "Intenal Gateway Error", http.StatusInternalServerError)
		return
	}

	if statusCode != http.StatusOK {
		log.Printf("[gateway] request is intercepted by lua rule: [%d] %s", statusCode, msg)
		w.WriteHeader(statusCode)
		w.Write([]byte(msg))
		return
	}

	remote, err := url.Parse(g.defaultUpstream)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	r.Header.Set("X-Gateway-By", "Mini-Gateway-Go")
	r.Host = remote.Host

	proxy.ServeHTTP(w, r)
}
