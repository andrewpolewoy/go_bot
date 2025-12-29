package app

import (
	"context"
	"net/http"
	"time"
)

func (a *App) Run() error {
	a.log.Info("starting http server", "addr", a.server.Addr)

	raw := a.cfg.Raw
	if raw.Server.TLS.Enabled {
		if err := a.server.ListenAndServeTLS(raw.Server.TLS.CertFile, raw.Server.TLS.KeyFile); err != nil {
			if err == http.ErrServerClosed {
				return nil
			}
			a.log.Error("http server error", "err", err)
			return err
		}
		return nil
	}

	if err := a.server.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		a.log.Error("http server error", "err", err)
		return err
	}
	return nil
}

func (a *App) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			return err
		}
	}

	if a.db != nil {
		a.db.Close()
	}

	return nil
}
