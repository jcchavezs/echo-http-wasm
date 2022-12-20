package httpwasm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/http-wasm/http-wasm-host-go/tck"
	"github.com/labstack/echo/v4"
)

// BackendHandler is a http.Handler implementing the logic expected by the TCK.
// It serves to echo back information from the request to the response for
// checking expectations.
func backendHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("x-httpwasm-next-method", c.Request().Method)
		c.Response().Header().Set("x-httpwasm-next-uri", c.Request().RequestURI)
		for k, vs := range c.Request().Header {
			for i, v := range vs {
				c.Response().Header().Add(fmt.Sprintf("x-httpwasm-next-header-%s-%d", k, i), v)
			}
		}
		return nil
	}
}

const testPort = 1323

func TestTCK(t *testing.T) {
	// Initialize the TCK guest wasm module.
	mw, err := NewMiddleware(context.Background(), tck.GuestWASM)
	if err != nil {
		t.Fatal(err)
	}

	// Set the delegate handler of the middleware to the backend.
	h := mw.NewHandler(context.Background(), backendHandler())

	// Start the server.
	server := echo.New()
	server.Any("/*", h)
	go func() {
		server.Start(fmt.Sprintf(":%d", testPort))
	}()

	for {
		if server.ListenerAddr() == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		break
	}

	// Run tests, issuing HTTP requests to server.
	tck.Run(t, fmt.Sprintf("http://%s", server.Listener.Addr().String()))
}
