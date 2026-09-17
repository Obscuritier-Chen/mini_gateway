package mock

import (
	"fmt"
	"log"
	"net/http"
)

func Start(addr string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := fmt.Sprintf(`{"status": "ok", "source": "local-upstream-%s", "path": "%s", "header_test": "%s"}`,
			addr, r.URL.Path, r.Header.Get("X-Gateway-By"))
		w.Write([]byte(resp))
	})

	log.Printf("[upstream] mock upstream service started, listening at port %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[upsteam] server failed to start: %v", err)
	}
}
