package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/Sembit26/ms-reportes/apps/api/internal/service"
)

// casoBasico refleja la estructura del archivo testdata/reporte-basico.json.
type casoBasico struct {
	Plantilla service.Plantilla `json:"plantilla"`
	Datos     []map[string]any  `json:"datos"`
}

// cargarCasoBasico lee el archivo de prueba y lo convierte a estructuras de Go.
func cargarCasoBasico(t *testing.T) (service.Plantilla, []map[string]any) {
	t.Helper()

	contenido, err := os.ReadFile("testdata/reporte-basico.json")
	if err != nil {
		t.Fatalf("no se pudo leer el archivo de prueba: %v", err)
	}

	var caso casoBasico
	if err := json.Unmarshal(contenido, &caso); err != nil {
		t.Fatalf("no se pudo interpretar el archivo de prueba: %v", err)
	}
	return caso.Plantilla, caso.Datos
}

// TestGenerarPDF_ConCasoBasico verifica que el PDF generado no esté vacío y
// comience con la firma "%PDF".
func TestGenerarPDF_ConCasoBasico(t *testing.T) {
	generador := service.NuevoGeneradorReportes()
	plantilla, datos := cargarCasoBasico(t)

	archivo, err := generador.Generar(context.Background(), plantilla, datos, service.FormatoPDF)
	if err != nil {
		t.Fatalf("no se esperaba error generando el PDF: %v", err)
	}
	if len(archivo) == 0 {
		t.Fatal("el PDF generado está vacío")
	}
	if !bytes.HasPrefix(archivo, []byte("%PDF")) {
		t.Fatalf("el PDF no comienza con la firma %%PDF")
	}
}

// TestGenerarXLSX_ConCasoBasico verifica que el XLSX generado comience con la
// firma "PK" y que, al abrirlo con la librería de Excel, contenga exactamente
// los encabezados y los valores del caso básico.
func TestGenerarXLSX_ConCasoBasico(t *testing.T) {
	generador := service.NuevoGeneradorReportes()
	plantilla, datos := cargarCasoBasico(t)

	archivo, err := generador.Generar(context.Background(), plantilla, datos, service.FormatoXLSX)
	if err != nil {
		t.Fatalf("no se esperaba error generando el XLSX: %v", err)
	}
	if len(archivo) == 0 {
		t.Fatal("el XLSX generado está vacío")
	}
	if !bytes.HasPrefix(archivo, []byte("PK")) {
		t.Fatalf("el XLSX no comienza con la firma PK")
	}

	// Abrimos el archivo con la misma librería para validar su contenido.
	libro, err := excelize.OpenReader(bytes.NewReader(archivo))
	if err != nil {
		t.Fatalf("no se pudo abrir el XLSX generado: %v", err)
	}
	defer func() { _ = libro.Close() }()

	// Usamos la primera hoja del libro (la única que crea el motor).
	hojas := libro.GetSheetList()
	if len(hojas) == 0 {
		t.Fatal("el XLSX generado no tiene hojas")
	}
	hoja := hojas[0]

	// Verificamos el título, los encabezados y los valores esperados.
	esperado := []struct {
		celda string
		valor string
	}{
		{"A1", "Resumen de documentos"},
		{"A2", "Código"},
		{"B2", "Estado"},
		{"A3", "DOC-001"},
		{"B3", "Aprobado"},
		{"A4", "DOC-002"},
		{"B4", "Pendiente"},
	}

	for _, caso := range esperado {
		obtenido, err := libro.GetCellValue(hoja, caso.celda)
		if err != nil {
			t.Fatalf("no se pudo leer la celda %s: %v", caso.celda, err)
		}
		if obtenido != caso.valor {
			t.Errorf("celda %s: se esperaba %q, se obtuvo %q", caso.celda, caso.valor, obtenido)
		}
	}
}

// TestGenerar_FormatoInvalido verifica que un formato no admitido devuelva el
// error controlado ErrFormatoNoSoportado.
func TestGenerar_FormatoInvalido(t *testing.T) {
	generador := service.NuevoGeneradorReportes()
	plantilla, datos := cargarCasoBasico(t)

	for _, formato := range []string{"docx", ""} {
		_, err := generador.Generar(context.Background(), plantilla, datos, formato)
		if !errors.Is(err, service.ErrFormatoNoSoportado) {
			t.Errorf("formato %q: se esperaba ErrFormatoNoSoportado, se obtuvo %v", formato, err)
		}
	}
}

// TestGenerar_PlantillaInvalida verifica que una plantilla incompleta devuelva
// el error controlado ErrPlantillaInvalida.
func TestGenerar_PlantillaInvalida(t *testing.T) {
	generador := service.NuevoGeneradorReportes()
	datos := []map[string]any{{"codigo": "DOC-001", "estado": "Aprobado"}}

	casos := []struct {
		nombre    string
		plantilla service.Plantilla
	}{
		{
			nombre: "sin columnas",
			plantilla: service.Plantilla{
				ID:     "550e8400-e29b-41d4-a716-446655440000",
				Titulo: "Título",
			},
		},
		{
			nombre: "sin título",
			plantilla: service.Plantilla{
				ID:       "550e8400-e29b-41d4-a716-446655440000",
				Columnas: []service.Columna{{Campo: "codigo", Encabezado: "Código"}},
			},
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			_, err := generador.Generar(context.Background(), caso.plantilla, datos, service.FormatoPDF)
			if !errors.Is(err, service.ErrPlantillaInvalida) {
				t.Errorf("se esperaba ErrPlantillaInvalida, se obtuvo %v", err)
			}
		})
	}
}

// TestGenerar_FilaConCampoFaltante verifica que una fila a la que le falta una
// clave definida en las columnas no provoque un panic y genere el archivo.
func TestGenerar_FilaConCampoFaltante(t *testing.T) {
	generador := service.NuevoGeneradorReportes()

	for _, formato := range []string{service.FormatoPDF, service.FormatoXLSX} {
		// Cargamos de nuevo el caso en cada iteración para no reutilizar el
		// mapa mutado de la iteración anterior.
		plantilla, datos := cargarCasoBasico(t)
		datos[1] = map[string]any{"codigo": "DOC-002"} // falta la clave "estado"

		archivo, err := generador.Generar(context.Background(), plantilla, datos, formato)
		if err != nil {
			t.Fatalf("formato %q: no se esperaba error con una fila incompleta: %v", formato, err)
		}
		if len(archivo) == 0 {
			t.Fatalf("formato %q: el archivo generado está vacío", formato)
		}
	}
}

// TestGenerar_ContextoCancelado verifica que el motor respete la cancelación
// del contexto y devuelva un error sin generar el archivo.
func TestGenerar_ContextoCancelado(t *testing.T) {
	generador := service.NuevoGeneradorReportes()
	plantilla, datos := cargarCasoBasico(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelamos antes de generar

	_, err := generador.Generar(ctx, plantilla, datos, service.FormatoPDF)
	if err == nil {
		t.Fatal("se esperaba un error por contexto cancelado")
	}
}
