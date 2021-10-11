package caddyrtaes

import (
	
	"net/http"
	"strings"
	"io"
	"io/ioutil"
	"os/exec"
	//"encoding/json"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("rtaes", parseCaddyfile)
}

// Middleware implements an HTTP handler that writes the
// visitor's IP address to a file or stream.
type Middleware struct {
	// The file or stream to write to. Can be "stdout"
	// or "stderr".
	// Output string `json:"output,omitempty"`

	// w io.Writer
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.rtaes",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger(m)
	// switch m.Output {
	// case "stdout":
	// 	m.w = os.Stdout
	// case "stderr":
	// 	m.w = os.Stderr
	// default:
	// 	return fmt.Errorf("an output stream is required")
	// }
	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	// if m.w == nil {
	// 	return fmt.Errorf("no writer")
	// }
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	logger := m.logger.With(
		zap.Object("request", caddyhttp.LoggableHTTPRequest{Request: r}),
	)

	changed := m.handle(r, logger)

	if changed {
		logger.Debug("rewrote request",
			zap.String("uri", r.RequestURI),
		)
	}

	return next.ServeHTTP(w, r)
}

// rewrite performs the rewrites on r using repl, which should
// have been obtained from r, but is passed in for efficiency.
// It returns true if any changes were made to r.
func (m Middleware) handle(r *http.Request, logger *zap.Logger) bool {
	oldURI := r.RequestURI

	// r.URL.RawQuery = "lala=sdsd"

	// update the encoded copy of the URI
	r.RequestURI = r.URL.RequestURI()

	buf := new(strings.Builder)
	io.Copy(buf, r.Body)
	rawBody := buf.String()

	// rawBody = rawBody + "jeje=asasdasd"
	app := "php"

    arg0 := "/opt/scripts/rtaes.php"
    arg1 := string(r.URL.RawQuery)
    arg2 := string(rawBody)
	arg3 := string(r.Header.Get("Authorization"))
    // arg3 := "golang"

    cmd := exec.Command(app, arg0, arg1, arg2, arg3)
    stdout, err := cmd.Output()

    if err != nil {
		return false
    }

	// var result map[string]interface{}
	// json.Unmarshal(stdout, &result)

	// r.URL.RawQuery = result["uri"].(string)
	// rawBody = result["body"].(string)

	// // update the encoded copy of the URI
	// r.RequestURI = r.URL.RequestURI()

    // Print the output
	rawBody = string(stdout)

	r.Body = ioutil.NopCloser(strings.NewReader(rawBody))
	r.ContentLength = int64(len(rawBody))

	// return true if anything changed
	return r.RequestURI != oldURI
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)