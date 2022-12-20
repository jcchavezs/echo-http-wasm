package httpwasm

import (
	"context"
	"net/http"

	handlerapi "github.com/http-wasm/http-wasm-host-go/api/handler"
	"github.com/http-wasm/http-wasm-host-go/handler"
	"github.com/labstack/echo/v4"
)

type Middleware handlerapi.Middleware[echo.HandlerFunc]

type middleware struct {
	m handler.Middleware
}

func (w *middleware) NewHandler(_ context.Context, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// The guest Wasm actually handles the request. As it may call host
		// functions, we add context parameters of the current request.
		s := newRequestState(c, next, w.m.Features())
		ctx := context.WithValue(c.Request().Context(), requestStateKey{}, s)
		outCtx, ctxNext, requestErr := w.m.HandleRequest(ctx)
		if requestErr != nil {
			handleErr(c.Response(), requestErr)
		}

		// Returning zero means the guest wants to break the handler chain, and
		// handle the response directly.
		if uint32(ctxNext) == 0 {
			return nil
		}

		// Otherwise, the host calls the next handler.
		err := s.handleNext()

		// Finally, call the guest with the response or error
		if err = w.m.HandleResponse(outCtx, uint32(ctxNext>>32), err); err != nil {
			return err
		}

		return nil
	}
}

func (w *middleware) Close(ctx context.Context) error {
	return w.m.Close(ctx)
}

func NewMiddleware(ctx context.Context, guest []byte, options ...handler.Option) (Middleware, error) {
	m, err := handler.NewMiddleware(ctx, guest, host{}, options...)
	if err != nil {
		return nil, err
	}
	return &middleware{m: m}, nil
}

func handleErr(w http.ResponseWriter, requestErr error) {
	// TODO: after testing, shouldn't send errors into the HTTP response.
	w.WriteHeader(500)
	w.Write([]byte(requestErr.Error())) // nolint
}
