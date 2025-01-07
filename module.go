package caddy_block_aws

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"net/http"
)

func init() {
	caddy.RegisterModule(BlockAWS{})
	httpcaddyfile.RegisterHandlerDirective("blockaws", parseCaddyfileForAWS)
}

type BlockAWS struct {
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (BlockAWS) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.blockaws",
		New: func() caddy.Module { return new(BlockAWS) },
	}
}

func (m *BlockAWS) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger()

	LoadInitialAWSData(m.logger)

	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m BlockAWS) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if MatchesWithCache(r.Context(), r.RemoteAddr) {
		m.logger.Info("Blocking AWS IP address", zap.String("ip", r.RemoteAddr))
		http.Error(w, "IP address is blocked", http.StatusForbidden)
		return nil
	}
	return next.ServeHTTP(w, r)
}

func (m *BlockAWS) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name
	return nil
}

// parseCaddyfileForAWS unmarshals tokens from h into a new Middleware.
func parseCaddyfileForAWS(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m BlockAWS
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*BlockAWS)(nil)
	_ caddyhttp.MiddlewareHandler = (*BlockAWS)(nil)
	_ caddyfile.Unmarshaler       = (*BlockAWS)(nil)
)
