package profiling

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"
)

type ProfileServer interface {
	Run(ctx context.Context) error
}

type profileServer struct {
	httpServer      *http.Server
	onShutdownError func(err error)
}

func NewServer(onShutdownError func(err error)) ProfileServer {
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	httpServer := &http.Server{Addr: ":6060", Handler: mux}
	return &profileServer{
		httpServer:      httpServer,
		onShutdownError: onShutdownError,
	}
}

func (ps *profileServer) Run(ctx context.Context) error {
	go func() { // shutdown goroutine blocked by ctx
		<-ctx.Done()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if err := ps.httpServer.Shutdown(timeoutCtx); err != nil {
			ps.onShutdownError(err)
		}
	}()
	err := ps.httpServer.ListenAndServe()
	if ctx.Err() != nil && err == http.ErrServerClosed {
		return nil // ctx got canceled
	}
	return err
}
