package app

import (
	"context"
	"net/http"
	"time"
)

func (a *App) Run() {
	a.log.Info("starting http server", "addr", a.server.Addr)

	var err error
	raw := a.cfg.Raw
	if raw.Server.TLS.Enabled {
		err = a.server.ListenAndServeTLS(raw.Server.TLS.CertFile, raw.Server.TLS.KeyFile)
	} else {
		err = a.server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		a.log.Error("http server error", "err", err)
	}
}

func (a *App) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	a.log.Info("shutting down http server")
	return a.server.Shutdown(ctx)
}
