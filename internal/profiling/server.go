package profiling

import (
	"context"
	"errors"
	"net/http"
	"net/http/pprof"
	"time"
)

type ProfileServer struct {
	httpServer      *http.Server
	onShutdownError func(err error)
}

func NewServer(onShutdownError func(err error)) *ProfileServer {
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	const readTimeout = 10 * time.Minute
	httpServer := &http.Server{
		Addr:              ":6060",
		Handler:           mux,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: time.Second,
	}
	return &ProfileServer{
		httpServer:      httpServer,
		onShutdownError: onShutdownError,
	}
}

func (ps *ProfileServer) Run(ctx context.Context) error {
	go func() { // shutdown goroutine blocked by ctx
		<-ctx.Done()
		const timeoutDuration = 10 * time.Millisecond
		timeoutCtx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
		defer cancel()
		if err := ps.httpServer.Shutdown(timeoutCtx); err != nil { //nolint:contextcheck
			ps.onShutdownError(err)
		}
	}()
	err := ps.httpServer.ListenAndServe()
	if ctx.Err() != nil && errors.Is(err, http.ErrServerClosed) {
		return nil // ctx got canceled
	}
	return err
}
