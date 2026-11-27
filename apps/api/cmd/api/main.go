package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PINGESO-CIVIL/ms-reportes/apps/api/internal/api"
	"github.com/PINGESO-CIVIL/ms-reportes/apps/api/internal/service"
	"github.com/PINGESO-CIVIL/ms-reportes/apps/api/internal/store"
)

var version = "dev"

func main() {
	rutaOpenAPI := flag.String("export-openapi", "", "Exporta el contrato OpenAPI 3.1 en YAML y termina.")
	flag.Parse()

	if *rutaOpenAPI != "" {
		if err := exportarOpenAPI(*rutaOpenAPI); err != nil {
			fmt.Fprintf(os.Stderr, "error al exportar OpenAPI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := api.CargarConfiguracion(os.Getenv)
	if err != nil {
		logger.Error("configuracion_invalida", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	pool, err := pgxpool.New(ctx, cfg.BaseDatosURL)
	if err != nil {
		logger.Error("error_pool_postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	almacenSalud := store.NuevoAlmacenSalud(pool)
	servicioSalud := service.NuevoServicioSalud(almacenSalud)

	manejador, _ := api.NuevoRouter(api.Dependencias{
		Version:  version,
		Ambiente: cfg.Ambiente,
		Salud:    servicioSalud,
		Logger:   logger,
	})

	servidor := &http.Server{
		Addr:              ":" + cfg.Puerto,
		Handler:           manejador,
		ReadHeaderTimeout: 5 * time.Second,
	}

	erroresServidor := make(chan error, 1)
	go func() {
		logger.Info("servidor_iniciado",
			slog.String("servicio", "ms-reportes"),
			slog.String("version", version),
			slog.String("puerto", cfg.Puerto),
			slog.String("ambiente", cfg.Ambiente),
		)
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			erroresServidor <- err
		}
		close(erroresServidor)
	}()

	select {
	case err := <-erroresServidor:
		if err != nil {
			logger.Error("error_servidor", slog.String("error", err.Error()))
			os.Exit(1)
		}
	case <-ctx.Done():
		ctxApagado, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelar()
		if err := servidor.Shutdown(ctxApagado); err != nil {
			logger.Error("error_apagado", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("servidor_detenido", slog.String("servicio", "ms-reportes"))
	}
}

func exportarOpenAPI(ruta string) error {
	_, humaAPI := api.NuevoRouter(api.Dependencias{
		Version:  version,
		Ambiente: "dev",
	})
	contenido, err := humaAPI.OpenAPI().YAML()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	return os.WriteFile(ruta, contenido, 0o644)
}
