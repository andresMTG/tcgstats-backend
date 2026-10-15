package controllers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/andresMTG/tcgstats-backend/internal/api/middleware"
	"github.com/go-chi/chi/v5"
)

func newRequest(method, target, body string) *http.Request {
	return httptest.NewRequest(method, target, strings.NewReader(body))
}

func withUserID(request *http.Request, id string) *http.Request {
	ctx := context.WithValue(request.Context(), middleware.UserIDKey, id)
	return request.WithContext(ctx)
}

func withRouteParam(request *http.Request, name, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(name, value)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}

func recorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}
