package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/go-pdf/fpdf"
)

// generarPDF construye un documento PDF con el título del reporte y una tabla
// con los encabezados y las filas de datos. Devuelve los bytes del archivo.
func generarPDF(ctx context.Context, plantilla Plantilla, datos []map[string]any) ([]byte, error) {
	// Creamos el documento en orientación vertical (P), en milímetros y hoja A4.
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Si el contexto se cancela, no seguimos generando.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Título del reporte, centrado y en negrita.
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, plantilla.Titulo, "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Calculamos el ancho de cada columna repartiendo el ancho útil de la
	// página (ancho total menos los márgenes) de forma equitativa.
	anchoPagina, _ := pdf.GetPageSize()
	margenIzquierdo, _, margenDerecho, _ := pdf.GetMargins()
	anchoUtil := anchoPagina - margenIzquierdo - margenDerecho
	anchoColumna := anchoUtil / float64(len(plantilla.Columnas))

	// Fila de encabezados, centrados y en negrita.
	pdf.SetFont("Helvetica", "B", 11)
	for _, columna := range plantilla.Columnas {
		pdf.CellFormat(anchoColumna, 8, columna.Encabezado, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	// Una fila por cada registro de datos. Si falta una clave, valorACelda
	// devuelve una celda vacía sin provocar error.
	pdf.SetFont("Helvetica", "", 10)
	for _, fila := range datos {
		// Respetamos la cancelación del contexto en cada fila.
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, columna := range plantilla.Columnas {
			pdf.CellFormat(anchoColumna, 8, valorACelda(fila[columna.Campo]), "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	// Escribimos el documento completo en un buffer de memoria.
	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, fmt.Errorf("no se pudo generar el PDF: %w", err)
	}
	return buffer.Bytes(), nil
}
