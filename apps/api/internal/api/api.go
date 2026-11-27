// Package api expone los handlers HTTP y la definición OpenAPI 3.1 con Huma.
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/PINGESO-CIVIL/ms-reportes/apps/api/internal/service"
)

const nombreServicio = "ms-reportes"

// Dependencias agrupa las dependencias necesarias para construir el router HTTP.
type Dependencias struct {
	Version  string
	Ambiente string
	Salud    *service.ServicioSalud
	Logger   *slog.Logger
}

// RespuestaSalud representa el estado de vida del proceso.
type RespuestaSalud struct {
	Body struct {
		Estado   string `json:"estado" example:"ok" doc:"Estado del proceso."`
		Servicio string `json:"servicio" example:"ms-reportes" doc:"Nombre del microservicio."`
		Version  string `json:"version" example:"dev" doc:"Versión compilada del binario."`
	}
}

// RespuestaDisponibilidad representa el estado de conectividad con PostgreSQL.
type RespuestaDisponibilidad struct {
	Body struct {
		Estado    string `json:"estado" example:"listo" doc:"Disponibilidad del servicio."`
		BaseDatos string `json:"base_datos" example:"ok" doc:"Estado de PostgreSQL."`
	}
}

// NuevoRouter configura Huma, los endpoints de salud y el middleware de observabilidad.
func NuevoRouter(deps Dependencias) (http.Handler, huma.API) {
	version := deps.Version
	if version == "" {
		version = "dev"
	}
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	mux := http.NewServeMux()
	cfg := huma.DefaultConfig("MS Reportes", version)
	cfg.Info.Description = "API del microservicio MS Reportes (PINGESO Civil, Grupo 1)."
	if deps.Ambiente == "prod" {
		cfg.DocsPath = ""
		cfg.OpenAPIPath = ""
		cfg.SchemasPath = ""
	} else {
		cfg.DocsPath = "/api/docs"
		cfg.OpenAPIPath = "/api/openapi"
		cfg.SchemasPath = "/api/schemas"
	}

	api := humago.New(mux, cfg)

	huma.Register(api, huma.Operation{
		OperationID: "obtener-salud-proceso",
		Method:      http.MethodGet,
		Path:        "/healthz",
		Summary:     "Verifica si el proceso está vivo",
		Tags:        []string{"Salud"},
	}, func(_ context.Context, _ *struct{}) (*RespuestaSalud, error) {
		resp := &RespuestaSalud{}
		resp.Body.Estado = "ok"
		resp.Body.Servicio = nombreServicio
		resp.Body.Version = version
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "obtener-disponibilidad",
		Method:      http.MethodGet,
		Path:        "/readyz",
		Summary:     "Verifica si el servicio y PostgreSQL responden",
		Tags:        []string{"Salud"},
	}, func(ctx context.Context, _ *struct{}) (*RespuestaDisponibilidad, error) {
		if deps.Salud == nil {
			return nil, huma.Error503ServiceUnavailable("base de datos no disponible")
		}
		if err := deps.Salud.VerificarDisponibilidad(ctx); err != nil {
			return nil, huma.Error503ServiceUnavailable("base de datos no disponible")
		}
		resp := &RespuestaDisponibilidad{}
		resp.Body.Estado = "listo"
		resp.Body.BaseDatos = "ok"
		return resp, nil
	})

	return conObservabilidad(version, logger, mux), api
}

type registradorEstado struct {
	http.ResponseWriter
	estado int
}

func (r *registradorEstado) WriteHeader(codigo int) {
	r.estado = codigo
	r.ResponseWriter.WriteHeader(codigo)
}

func conObservabilidad(version string, logger *slog.Logger, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = nuevoRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)

		reg := &registradorEstado{
			ResponseWriter: w,
			estado:         http.StatusOK,
		}
		siguiente.ServeHTTP(reg, r)

		logger.Info("solicitud_http",
			slog.String("servicio", nombreServicio),
			slog.String("version", version),
			slog.String("request_id", requestID),
			slog.String("metodo", r.Method),
			slog.String("ruta", r.URL.Path),
			slog.Int("estado", reg.estado),
			slog.Int64("duracion_ms", time.Since(inicio).Milliseconds()),
		)
	})
}

func nuevoRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	var out [36]byte
	hex.Encode(out[0:8], b[0:4])
	out[8] = '-'
	hex.Encode(out[9:13], b[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], b[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], b[8:10])
	out[23] = '-'
	hex.Encode(out[24:36], b[10:16])
	return string(out[:])
}
