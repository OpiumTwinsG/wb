package httpserver

import ("net/http"
		 "time"
		"context"
		"orderservice/internal/pkg/logger"
		)

type OrderServiceServer struct {
	srv *http.Server
}

func NewOrderServiceServer(addr string, handler http.Handler) *OrderServiceServer{
	return &OrderServiceServer{
		srv: &http.Server{
			Addr: addr,
			Handler: handler,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
	}
}

// Start запускает HTTP‑сервер (блокирующий вызов).
func (s *OrderServiceServer) Start() error {
	// logger.Log.Infow("http ListenAndServe", "addr", s.srv.Addr)
	err := s.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		// logger.Log.Errorw("http server stopped with error", "err", err)
	} else {
		// logger.Log.Infow("http server stopped")
	}
	return err
}

// Stop завершает работу сервера с таймаутом 5 секунд.
func (s *OrderServiceServer) Stop() {
	logger.Log.Infow("shutting down http server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		logger.Log.Errorw("server shutdown error", "err", err)
	}
}