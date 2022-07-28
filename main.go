package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dsh/px/app"
	"dsh/px/router"
)

const (
	defaultListenAddr = ":5000"
	listenAddrUsage   = "[addr]:port to listen on"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	app.LoadDotEnv(os.Getenv("PX_ENV")) // load env vars

	// Process command line args and assign env variables
	var listenAddr string
	flag.StringVar(&listenAddr, "l", defaultListenAddr, listenAddrUsage)
	flag.StringVar(&listenAddr, "listen", defaultListenAddr, listenAddrUsage)

	gotFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		gotFlags[f.Name] = true
	})

	if os.Getenv("HOST_ADDR") == "" || gotFlags["l"] || gotFlags["listen"] {
		os.Setenv("HOST_ADDR", listenAddr)
	}

	flag.Parse()
}

func main() {
	// The HTTP Server
	server := &http.Server{
		Addr:    os.Getenv("HOST_ADDR"),
		Handler: router.New(app.New()),
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancelCtx := context.WithTimeout(
			serverCtx, 30*time.Second)
		defer shutdownCancelCtx()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	log.Printf("Ready to serve on %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
