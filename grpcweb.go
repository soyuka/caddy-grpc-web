package grpcweb

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/soyuka/grpcweb"
)

func init() {
	caddy.RegisterModule(Handler{})
	httpcaddyfile.RegisterHandlerDirective("grpc_web", parseCaddyfile)
}

// Handler is an HTTP handler that bridges gRPC-Web <--> gRPC requests.
type Handler struct {
}

// CaddyModule returns the Caddy module information.
func (Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.grpc_web",
		New: func() caddy.Module { return new(Handler) },
	}
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if grpcweb.IsGRPCWebRequest(r) {
		grpcServerAdapter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = next.ServeHTTP(w, r)
		})

		webHandler := &grpcweb.Handler{GRPCServer: grpcServerAdapter}
		webHandler.ServeHTTP(w, r)
		return nil // The request has been handled.
	}

	// Pass-thru for all other requests.
	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile sets up h from Caddyfile tokens.
func (h *Handler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			return d.Errf("unknown subdirective '%s'", d.Val())
		}
	}
	return nil
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var handler Handler
	err := handler.UnmarshalCaddyfile(h.Dispenser)
	return handler, err
}

// Interface guards
var (
	_ caddyhttp.MiddlewareHandler = (*Handler)(nil)
	_ caddyfile.Unmarshaler       = (*Handler)(nil)
)
