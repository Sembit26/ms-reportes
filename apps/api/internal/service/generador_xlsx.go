package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// nombreHoja es el nombre de la hoja de cálculo donde se escriben los datos.
const nombreHoja = "Reporte"

// generarXLSX construye un archivo Excel con el título del reporte, una fila de
// encabezados y una fila por cada registro de datos. Devuelve los bytes del
// archivo (.xlsx).
//
// La distribución del contenido es la siguiente:
//   - Fila 1: título del reporte.
//   - Fila 2: encabezados de las columnas.
//   - Filas 3 en adelante: un registro de datos por fila.
func generarXLSX(ctx context.Context, plantilla Plantilla, datos []map[string]any) ([]byte, error) {
	// Creamos un archivo Excel vacío en memoria.
	archivo := excelize.NewFile()
	defer func() {
		// Cerramos el archivo para liberar los recursos internos.
		_ = archivo.Close()
	}()

	// Renombramos la hoja inicial "Sheet1" por un nombre descriptivo.
	if err := archivo.SetSheetName("Sheet1", nombreHoja); err != nil {
		return nil, fmt.Errorf("no se pudo preparar la hoja de cálculo: %w", err)
	}

	// Si el contexto se cancela, no seguimos generando.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Fila 1: título del reporte.
	if err := archivo.SetCellValue(nombreHoja, "A1", plantilla.Titulo); err != nil {
		return nil, fmt.Errorf("no se pudo escribir el título: %w", err)
	}

	// Fila 2: encabezados de las columnas, uno por columna.
	for indice, columna := range plantilla.Columnas {
		celda, err := excelize.CoordinatesToCellName(indice+1, 2)
		if err != nil {
			return nil, fmt.Errorf("no se pudo calcular la celda del encabezado: %w", err)
		}
		if err := archivo.SetCellValue(nombreHoja, celda, columna.Encabezado); err != nil {
			return nil, fmt.Errorf("no se pudo escribir el encabezado: %w", err)
		}
	}

	// Filas de datos a partir de la fila 3. Si falta una clave en el registro,
	// valorACelda escribe una celda vacía sin provocar error.
	for fila, registro := range datos {
		// Respetamos la cancelación del contexto en cada fila.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for indice, columna := range plantilla.Columnas {
			celda, err := excelize.CoordinatesToCellName(indice+1, fila+3)
			if err != nil {
				return nil, fmt.Errorf("no se pudo calcular la celda de datos: %w", err)
			}
			if err := archivo.SetCellValue(nombreHoja, celda, valorACelda(registro[columna.Campo])); err != nil {
				return nil, fmt.Errorf("no se pudo escribir un dato: %w", err)
			}
		}
	}

	// Escribimos el archivo completo en un buffer de memoria.
	var buffer bytes.Buffer
	if err := archivo.Write(&buffer); err != nil {
		return nil, fmt.Errorf("no se pudo generar el XLSX: %w", err)
	}
	return buffer.Bytes(), nil
}
