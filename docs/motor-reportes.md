# Motor de Generación de Reportes (PDF / XLSX) — Hito 1

> Documento de referencia para el equipo de **MS Reportes** (PINGESO Civil, Grupo 1).
> Explica qué se implementó en el Hito 1, cómo está organizado el código y cómo
> consumir el motor desde la API. Pensado para que Nelson (API) y Seba (backend)
> entiendan lo realizado sin necesidad de revisar el código línea por línea.

Este documento acompaña al **Plan Definitivo de Actividades** (sección *Hito 1 ·
A. Motor de Generación*). Ante cualquier duda, el Plan y el DAS §3 son la fuente
de verdad.

---

## 1. Objetivo del Hito 1

Construir el **motor de generación de archivos** del microservicio: una pieza de
código que recibe una plantilla y unos datos, y devuelve un archivo **PDF** o
**XLSX** listo para descargar.

Por ahora el motor trabaja de forma **independiente**: recibe todo por parámetros
y devuelve bytes. No depende de la base de datos ni de los servicios externos
(G2 Gestor de Información, G3 Plantillas). Esas integraciones se harán en hitos
posteriores.

---

## 2. Decisiones técnicas

| Decisión | Elección | Motivo |
| :--- | :--- | :--- |
| Módulo Go | `github.com/Sembit26/ms-reportes/apps/api` | Remoto de trabajo actual. Se renombrará al nombre definitivo del servicio bajo `github.com/PINGESO-CIVIL/<servicio>/...` al traspasar el repo (tarea de Seba, Hito 0). |
| Versión de Go | `go 1.26.0` | La librería de Excel (`excelize v2.11.1`) exige Go ≥ 1.26. El DAS §3 fija la imagen `golang:1.26.5-alpine3.23` (toolchain `go1.26.5`) para el Dockerfile. |
| Librería PDF | `github.com/go-pdf/fpdf v0.9.0` | Estándar, ligera y bien mantenida para generar PDF simples. |
| Librería Excel | `github.com/xuri/excelize/v2 v2.11.1` | La librería de referencia para XLSX en Go. |

---

## 3. Estructura de archivos

```
apps/api/
├── go.mod                              # módulo Go + dependencias
├── go.sum                              # checksums de dependencias
└── internal/
    └── service/
        ├── reporte.go                  # tipos, interfaz, errores y validación
        ├── generador_pdf.go            # genera el PDF
        ├── generador_xlsx.go           # genera el XLSX
        ├── reporte_test.go             # pruebas unitarias
        └── testdata/
            └── reporte-basico.json     # caso de prueba (entrada de ejemplo)
```

> `reporte-basico.json` vive en `internal/service/testdata/` como **fixture de
> prueba** (no suelto en la raíz del repo, tal como exige el plan). En el Hito 2,
> los archivos *golden* de las pruebas E2E irán en `apps/api/e2e/testdata/`.

### Qué hace cada archivo

- **`reporte.go`** — Define las estructuras de datos (`Plantilla`, `Columna`),
  la interfaz pública `GeneradorReportes`, los errores controlados, el
  constructor `NuevoGeneradorReportes`, la validación de la plantilla y el
  despacho al generador según el formato.
- **`generador_pdf.go`** — Toma la plantilla y los datos y arma un PDF con el
  título y una tabla (encabezados + filas).
- **`generador_xlsx.go`** — Arma un archivo Excel con el título, una fila de
  encabezados y una fila por cada registro.
- **`reporte_test.go`** — Verifica con el caso `reporte-basico.json` que el motor
  produce los archivos correctos y que maneja bien los errores.

---

## 4. Contrato del motor (lo que usa la API)

El motor se consume a través de una **interfaz**. Es el único punto de entrada
que necesita conocer el handler de la API.

> Esta es la **firma acordada** entre `internal/api` (Nelson) e
> `internal/service` (Jorge) según la Decisión Previa #3 del plan: el handler
> llama al motor pasando la plantilla, las filas y el formato, y recibe los bytes
> del archivo. Ni el handler toca el store ni el motor toca HTTP, SQL o archivos.

```go
// GeneradorReportes es el contrato del motor de generación.
type GeneradorReportes interface {
    Generar(ctx context.Context, plantilla Plantilla, datos []map[string]any, formato string) ([]byte, error)
}
```

Para obtener una instancia:

```go
generador := service.NuevoGeneradorReportes()
```

### Estructuras de entrada

Calzan exactamente con la estructura de `reporte-basico.json`:

```go
// Plantilla describe la estructura del reporte.
type Plantilla struct {
    ID       string    `json:"id"`
    Titulo   string    `json:"titulo"`
    Columnas []Columna `json:"columnas"`
}

// Columna representa una columna del reporte.
type Columna struct {
    Campo      string `json:"campo"`
    Encabezado string `json:"encabezado"`
}
```

- `plantilla.id`: identificador (UUID) de la plantilla.
- `plantilla.titulo`: título que se muestra en el reporte.
- `plantilla.columnas`: lista de columnas. Cada una tiene un `campo` (la clave
  que se busca en cada fila de `datos`) y un `encabezado` (el texto visible).
- `datos`: lista de filas (`[]map[string]any`), donde cada mapa tiene las claves
  definidas en `campo`. En el plan a esta lista se la llama "filas"; aquí el
  parámetro se llama `datos` porque coincide con la clave `datos` del JSON.

---

## 5. Reglas de validación y manejo de errores

### Formato de salida

- Solo se aceptan `"pdf"` y `"xlsx"`.
- La comparación **ignora mayúsculas y espacios** (por ejemplo `"PDF"`, `"Pdf"`
  o `" pdf "` equivalen a `"pdf"`).

### Errores controlados (exportados)

El handler puede distinguirlos con `errors.Is` para devolver el código HTTP
adecuado:

```go
service.ErrFormatoNoSoportado // formato distinto de "pdf" / "xlsx"  → HTTP 400
service.ErrPlantillaInvalida  // plantilla incompleta o inválida     → HTTP 422
```

### Validación de la plantilla

Se rechaza (con `ErrPlantillaInvalida`) cuando:

- Falta el `id` (vacío).
- Falta el `titulo` (vacío o solo espacios).
- No hay `columnas` (lista vacía).
- Alguna columna no tiene `campo` o `encabezado`.

### Datos con claves faltantes

Si una fila de `datos` **no tiene** una clave definida en `columnas`, el motor
escribe una **celda vacía (`""`)** en esa posición. Nunca produce un *panic*
ni un error.

### Cancelación / timeout

El motor respeta `context.Context`: si el contexto se cancela o expira, corta la
ejecución y devuelve el error del contexto. Se verifica antes de empezar y en
cada fila.

---

## 6. Formato de los archivos generados

### PDF

- Hoja A4, orientación vertical, márgenes de 10 mm.
- Primera línea: **título** centrado y en negrita.
- Debajo: una tabla con los **encabezados** (negrita, centrados) y una fila por
  cada registro (alineados a la izquierda). Las columnas reparten el ancho útil
  de forma equitativa.
- El archivo comienza con la firma `%PDF` (comprobado en las pruebas).

### XLSX

El contenido se distribuye así en la hoja `Reporte`:

| Fila | Contenido |
| :--- | :--- |
| 1 | Título del reporte |
| 2 | Encabezados de las columnas |
| 3 en adelante | Un registro de datos por fila |

Ejemplo con `reporte-basico.json`:

| A | B |
| :--- | :--- |
| Resumen de documentos | |
| Código | Estado |
| DOC-001 | Aprobado |
| DOC-002 | Pendiente |

- El archivo es un ZIP (firma `PK`), como todo `.xlsx` válido.

---

## 7. Pruebas unitarias

Ubicadas en `apps/api/internal/service/reporte_test.go`. Usan el caso
`testdata/reporte-basico.json` (título "Resumen de documentos", columnas
`Código` y `Estado`, filas `DOC-001/Aprobado` y `DOC-002/Pendiente`).

| Prueba | Qué verifica |
| :--- | :--- |
| `TestGenerarPDF_ConCasoBasico` | El PDF no está vacío y comienza con `%PDF`. |
| `TestGenerarXLSX_ConCasoBasico` | El XLSX comienza con `PK` y, al abrirlo con `excelize`, contiene exactamente título, encabezados y valores. |
| `TestGenerar_FormatoInvalido` | `"docx"` y `""` devuelven `ErrFormatoNoSoportado`. |
| `TestGenerar_PlantillaInvalida` | Plantillas sin columnas o sin título devuelven `ErrPlantillaInvalida`. |
| `TestGenerar_FilaConCampoFaltante` | Una fila incompleta no provoca *panic* y genera el archivo. |
| `TestGenerar_ContextoCancelado` | Un contexto cancelado detiene la ejecución con error. |

---

## 8. Cómo ejecutar

> Nota: la máquina de desarrollo de Jorge no tiene Go instalado, por eso en el
> Hito 1 se validó usando un contenedor Docker con la imagen oficial `golang:1.26`.
> El Dockerfile definitivo usará la imagen fijada por el DAS §3
> (`golang:1.26.5-alpine3.23`); tarea de Seba.

### Con Docker (funciona en cualquier máquina)

```bash
cd apps/api
docker run --rm -v "$PWD":/work -w /work -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomodcache golang:1.26 go test -v ./internal/service/...
```

### Con Go 1.26+ instalado localmente

```bash
cd apps/api
gofmt -l .                            # debe imprimir nada (todo formateado)
go vet ./...
go mod verify                         # verifica la integridad de las dependencias
go test -v ./internal/service/...
```

Resultado de la última ejecución: **6 pruebas, todas en verde** (`ok`).

---

## 9. Notas para la integración (próximos pasos)

- **Nelson (API):** consumir `service.NuevoGeneradorReportes()` y llamar
  `Generar(...)` desde el handler. Distinguir los errores con `errors.Is` para
  mapear a HTTP 400/422. Devolver el `[]byte` con el `Content-Type` correcto:
  `application/pdf` para PDF y
  `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` para XLSX.
- **Seba (backend):** el motor es una capa pura de dominio (sin HTTP, sin SQL,
  sin lectura de archivos). Cuando G2 (datos) y G3 (plantillas) publiquen sus
  contratos OpenAPI, se agregará un cliente HTTP que transforme sus respuestas a
  las estructuras `Plantilla` y `[]map[string]any` de este motor.
