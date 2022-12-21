package httpwasm

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/http-wasm/http-wasm-host-go/api/handler"
	"github.com/http-wasm/http-wasm-host-go/testing/handlertest"
	"github.com/labstack/echo/v4"
)

var testCtx = context.Background()

func Test_host(t *testing.T) {
	e := echo.New()
	defer e.Close()

	newCtx := func(_ handler.Features) (context.Context, handler.Features) {
		// TODO(jcchavezs): adds support for buffered response body
		r, _ := http.NewRequest("GET", "", bytes.NewReader(nil))
		w := &httptest.ResponseRecorder{HeaderMap: map[string][]string{}}
		c := e.NewContext(r, w)
		return context.WithValue(testCtx, requestStateKey{}, &requestState{c: c}), 0
	}

	if err := handlertest.HostTest(t, host{}, newCtx); err != nil {
		t.Fatal(err)
	}
}

// Test_host_GetProtocolVersion ensures HTTP/2.0 is readable
func Test_host_GetProtocolVersion(t *testing.T) {
	e := echo.New()
	defer e.Close()

	tests := []string{"HTTP/1.1", "HTTP/2.0"}

	h := host{}
	for _, tt := range tests {
		tc := tt
		t.Run(tc, func(t *testing.T) {
			r := &http.Request{Proto: tc}
			c := e.NewContext(r, nil)
			ctx := context.WithValue(testCtx, requestStateKey{}, &requestState{c: c})

			if want, have := tc, h.GetProtocolVersion(ctx); want != have {
				t.Errorf("unexpected protocolVersion, want: %v, have: %v", want, have)
			}
		})
	}
}
