package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mini_gateway/internal/lua"
	"mini_gateway/internal/gateway"
	"mini_gateway/internal/mock"
)

func main() {
	go mock.Start(":8082")

	luaEngine, err := lua.New("rules/rule.lua")
	if err != nil {
		log.Fatalf("[main] failed to initialise lua engineL %v", err)
	}

	gw := gateway.New("http://127.0.0.1:8082", luaEngine)

	server := &http.Server{
		Addr:         ":8081",
		Handler:      gw,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[gateway] gateway ready, listening at: 8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[gateway] abnormal exit :%v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[gateway] closing the gateway service")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[gateway] shutdown err %v", err)
	}
	log.Println("[gateway] server closed smoothly")

}
