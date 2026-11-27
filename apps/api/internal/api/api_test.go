package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PINGESO-CIVIL/ms-reportes/apps/api/internal/api"
	"github.com/PINGESO-CIVIL/ms-reportes/apps/api/internal/service"
)

type verificadorSimulado struct {
	err error
}

func (v verificadorSimulado) VerificarConexion(_ context.Context) error {
	return v.err
}

func TestCargarConfiguracion(t *testing.T) {
	base := map[string]string{
		"API_PORT":     "8080",
		"API_ENV":      "dev",
		"DATABASE_URL": "postgres://reportes:clave@postgres:5432/reportes",
		"FRONTEND_URL": "http://localhost:5110",
		"TZ":           "America/Santiago",
	}

	cfg, err := api.CargarConfiguracion(func(k string) string { return base[k] })
	if err != nil {
		t.Fatalf("no se esperaba error con configuración válida: %v", err)
	}
	if cfg.Puerto != "8080" || cfg.Ambiente != "dev" {
		t.Fatalf("configuración inesperada: %+v", cfg)
	}

	if _, err := api.CargarConfiguracion(func(k string) string {
		if k == "DATABASE_URL" {
			return ""
		}
		return base[k]
	}); err == nil {
		t.Fatal("se esperaba error cuando falta DATABASE_URL")
	}

	if _, err := api.CargarConfiguracion(func(k string) string {
		if k == "API_ENV" {
			return "staging"
		}
		return base[k]
	}); err == nil {
		t.Fatal("se esperaba error con API_ENV inválido")
	}
}

func TestEndpointsSaludYDocs(t *testing.T) {
	routerDev, _ := api.NuevoRouter(api.Dependencias{
		Version:  "test-1",
		Ambiente: "dev",
		Salud:    service.NuevoServicioSalud(verificadorSimulado{}),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	routerDev.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/healthz devolvió %d, se esperaba 200", rec.Code)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("faltó el header X-Request-Id en la respuesta")
	}

	recReady := httptest.NewRecorder()
	reqReady := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	routerDev.ServeHTTP(recReady, reqReady)
	if recReady.Code != http.StatusOK {
		t.Fatalf("/readyz con BD ok devolvió %d, se esperaba 200", recReady.Code)
	}

	recDocs := httptest.NewRecorder()
	reqDocs := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	routerDev.ServeHTTP(recDocs, reqDocs)
	if recDocs.Code != http.StatusOK {
		t.Fatalf("/api/docs en dev devolvió %d, se esperaba 200", recDocs.Code)
	}

	routerProd, _ := api.NuevoRouter(api.Dependencias{
		Version:  "test-1",
		Ambiente: "prod",
		Salud:    service.NuevoServicioSalud(verificadorSimulado{err: errors.New("caída")}),
	})

	recReadyDown := httptest.NewRecorder()
	routerProd.ServeHTTP(recReadyDown, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if recReadyDown.Code != http.StatusServiceUnavailable {
		t.Fatalf("/readyz con BD caída devolvió %d, se esperaba 503", recReadyDown.Code)
	}

	recDocsProd := httptest.NewRecorder()
	routerProd.ServeHTTP(recDocsProd, httptest.NewRequest(http.MethodGet, "/api/docs", nil))
	if recDocsProd.Code != http.StatusNotFound {
		t.Fatalf("/api/docs en prod devolvió %d, se esperaba 404", recDocsProd.Code)
	}
}
