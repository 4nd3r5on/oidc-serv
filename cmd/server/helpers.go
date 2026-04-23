package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

func getEnv(name, fallback string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		return fallback
	}
	return val
}

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

func shutdownServ(ctx context.Context, server Shutdowner) error {
	shutdownCtx, release := context.WithTimeout(ctx, 10*time.Second)
	defer release()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}
	return nil
}

func listenAndServe(ctx context.Context, server *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return shutdownServ(context.Background(), server)
	}
}
