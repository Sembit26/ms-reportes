// Package service contiene el motor de generación de reportes en PDF y XLSX.
//
// El motor recibe todos los datos en memoria y devuelve el archivo generado
// como una secuencia de bytes. No realiza llamadas HTTP, ni consultas SQL, ni
// lee archivos locales en tiempo de ejecución: toda la información llega por
// los parámetros de la función.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// FormatoPDF identifica el formato de salida PDF.
const FormatoPDF = "pdf"

// FormatoXLSX identifica el formato de salida Excel (XLSX).
const FormatoXLSX = "xlsx"

// Columna representa una columna de la plantilla del reporte.
//
// Campo es la clave que se busca en cada fila de datos y Encabezado es el
// texto que se muestra como título de la columna en el archivo generado.
type Columna struct {
	Campo      string `json:"campo"`
	Encabezado string `json:"encabezado"`
}

// Plantilla describe la estructura de un reporte: su identificador, su título
// y la lista de columnas que debe mostrar.
type Plantilla struct {
	ID       string    `json:"id"`
	Titulo   string    `json:"titulo"`
	Columnas []Columna `json:"columnas"`
}

// Errores controlados que devuelve el motor. El handler de la API puede
// distinguirlos con errors.Is para responder el código HTTP adecuado.
var (
	// ErrFormatoNoSoportado se devuelve cuando el formato pedido no es "pdf"
	// ni "xlsx".
	ErrFormatoNoSoportado = errors.New("formato no soportado: debe ser 'pdf' o 'xlsx'")

	// ErrPlantillaInvalida se devuelve cuando la plantilla no es válida: le
	// falta el identificador, el título o no define columnas válidas.
	ErrPlantillaInvalida = errors.New("plantilla inválida")
)

// GeneradorReportes es el contrato que debe cumplir el motor de generación.
//
// Recibe el contexto, la plantilla, las filas de datos y el formato deseado;
// devuelve los bytes del archivo generado o un error controlado.
type GeneradorReportes interface {
	Generar(ctx context.Context, plantilla Plantilla, datos []map[string]any, formato string) ([]byte, error)
}

// generadorReportes es la implementación concreta del motor de generación.
// No guarda estado, por lo que puede reutilizarse de forma segura.
type generadorReportes struct{}

// NuevoGeneradorReportes crea una instancia lista para usar del motor.
func NuevoGeneradorReportes() GeneradorReportes {
	return &generadorReportes{}
}

// Generar valida la entrada, elige el generador según el formato y devuelve el
// archivo resultante.
//
// El formato se acepta sin distinguir mayúsculas de minúsculas (por ejemplo
// "PDF" o "Pdf" equivalen a "pdf").
func (g *generadorReportes) Generar(ctx context.Context, plantilla Plantilla, datos []map[string]any, formato string) ([]byte, error) {
	// Si el contexto ya fue cancelado o expiró, se corta la ejecución de
	// inmediato para no generar un archivo que nadie va a recibir.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Normalizamos el formato a minúsculas y sin espacios para aceptar "PDF",
	// "Pdf", " pdf ", etc.
	switch strings.ToLower(strings.TrimSpace(formato)) {
	case FormatoPDF:
		if err := validarPlantilla(plantilla); err != nil {
			return nil, err
		}
		return generarPDF(ctx, plantilla, datos)
	case FormatoXLSX:
		if err := validarPlantilla(plantilla); err != nil {
			return nil, err
		}
		return generarXLSX(ctx, plantilla, datos)
	default:
		return nil, ErrFormatoNoSoportado
	}
}

// validarPlantilla comprueba que la plantilla tenga lo mínimo para generar un
// reporte: un identificador, un título y al menos una columna con campo y
// encabezado definidos.
func validarPlantilla(plantilla Plantilla) error {
	if strings.TrimSpace(plantilla.ID) == "" {
		return fmt.Errorf("%w: falta el identificador de la plantilla", ErrPlantillaInvalida)
	}
	if strings.TrimSpace(plantilla.Titulo) == "" {
		return fmt.Errorf("%w: falta el título de la plantilla", ErrPlantillaInvalida)
	}
	if len(plantilla.Columnas) == 0 {
		return fmt.Errorf("%w: la plantilla no define columnas", ErrPlantillaInvalida)
	}
	for i, columna := range plantilla.Columnas {
		if strings.TrimSpace(columna.Campo) == "" {
			return fmt.Errorf("%w: la columna %d no tiene campo", ErrPlantillaInvalida, i+1)
		}
		if strings.TrimSpace(columna.Encabezado) == "" {
			return fmt.Errorf("%w: la columna %d no tiene encabezado", ErrPlantillaInvalida, i+1)
		}
	}
	return nil
}

// valorACelda convierte un valor de una fila de datos a su representación en
// texto. Si el valor es nil (por ejemplo, porque la clave no existe en el
// mapa), devuelve una cadena vacía, evitando así errores de ejecución.
func valorACelda(valor any) string {
	if valor == nil {
		return ""
	}
	switch v := valor.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
