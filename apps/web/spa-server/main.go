package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const directorioEstaticos = "/srv"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	puerto := os.Getenv("API_PORT")
	if puerto == "" {
		puerto = "8080"
	}
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		logger.Error("falta_backend_url")
		os.Exit(1)
	}
	destino, err := url.Parse(backendURL)
	if err != nil {
		logger.Error("backend_url_invalida", slog.String("error", err.Error()))
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(destino)
	manejador := construirManejador(directorioEstaticos, proxy)

	servidor := &http.Server{
		Addr:              ":" + puerto,
		Handler:           manejador,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	go func() {
		<-ctx.Done()
		ctxApagado, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelar()
		_ = servidor.Shutdown(ctxApagado)
	}()

	logger.Info("spa_server_iniciado", slog.String("puerto", puerto))
	if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("error_spa_server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func construirManejador(raiz string, proxy *httputil.ReverseProxy) http.Handler {
	archivoBase := filepath.Join(raiz, "index.html")
	servidorArchivos := http.FileServer(http.Dir(raiz))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if r.URL.Path == "/healthz" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"estado":"ok","servicio":"ms-reportes-web"}`))
			return
		}

		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			proxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		rutaRelativa := strings.TrimPrefix(filepath.Clean("/"+r.URL.Path), string(filepath.Separator))
		rutaCompleta := filepath.Join(raiz, rutaRelativa)
		if rutaRelativa != "" && strings.HasPrefix(rutaCompleta, raiz+string(filepath.Separator)) {
			if info, err := os.Stat(rutaCompleta); err == nil && !info.IsDir() {
				servidorArchivos.ServeHTTP(w, r)
				return
			}
		}

		http.ServeFile(w, r, archivoBase)
	})
}
