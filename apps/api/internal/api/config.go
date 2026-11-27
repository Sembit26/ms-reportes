package api

import (
	"fmt"
	"slices"
	"strings"
)

// Configuracion reúne las variables de entorno requeridas por la API.
type Configuracion struct {
	Puerto               string
	Ambiente             string
	BaseDatosURL         string
	FrontendURL          string
	ZonaHoraria          string
	AuthJWKSURL          string
	GestorInformacionURL string
	PlantillasURL        string
}

// CargarConfiguracion lee y valida las variables de entorno obligatorias al arranque.
func CargarConfiguracion(obtenerEnv func(string) string) (Configuracion, error) {
	obligatorias := []string{
		"API_PORT",
		"API_ENV",
		"DATABASE_URL",
		"FRONTEND_URL",
		"TZ",
	}

	var faltantes []string
	for _, clave := range obligatorias {
		if strings.TrimSpace(obtenerEnv(clave)) == "" {
			faltantes = append(faltantes, clave)
		}
	}
	if len(faltantes) > 0 {
		return Configuracion{}, fmt.Errorf("faltan variables de entorno obligatorias: %s", strings.Join(faltantes, ", "))
	}

	ambiente := strings.TrimSpace(obtenerEnv("API_ENV"))
	if !slices.Contains([]string{"dev", "test", "prod"}, ambiente) {
		return Configuracion{}, fmt.Errorf("API_ENV inválido %q: debe ser dev, test o prod", ambiente)
	}

	return Configuracion{
		Puerto:               strings.TrimSpace(obtenerEnv("API_PORT")),
		Ambiente:             ambiente,
		BaseDatosURL:         strings.TrimSpace(obtenerEnv("DATABASE_URL")),
		FrontendURL:          strings.TrimSpace(obtenerEnv("FRONTEND_URL")),
		ZonaHoraria:          strings.TrimSpace(obtenerEnv("TZ")),
		AuthJWKSURL:          strings.TrimSpace(obtenerEnv("AUTH_JWKS_URL")),
		GestorInformacionURL: strings.TrimSpace(obtenerEnv("GESTOR_INFORMACION_URL")),
		PlantillasURL:        strings.TrimSpace(obtenerEnv("PLANTILLAS_URL")),
	}, nil
}
