package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Busenuryurdakul/cli-calculator/internal/bookmark"
	"github.com/Busenuryurdakul/cli-calculator/internal/server"
)

func main() {
	addr, err := resolveAddr(os.Args[1:], os.Getenv("API_ADDR"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := run(addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func resolveAddr(args []string, env string) (string, error) {
	addr := "127.0.0.1:8081"
	if env != "" {
		addr = env
	}
	switch len(args) {
	case 0:
		return addr, nil
	case 1:
		return args[0], nil
	default:
		return "", fmt.Errorf("usage: api [addr]")
	}
}

func run(addr string) error {
	store := bookmark.NewMemoryStore()
	h := server.New(store)

	srv := &http.Server{Addr: addr, Handler: h}
	errCh := make(chan error, 1)
	go func() {
		fmt.Fprintf(os.Stderr, "listening on http://%s\n", addr)
		errCh <- srv.ListenAndServe()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen %s: %w", addr, err)
		}
		return nil
	case <-sig:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		err := <-errCh
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen %s: %w", addr, err)
		}
		return nil
	}
}
