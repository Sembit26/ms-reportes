# Modos de Falla — MS Reportes

Este documento detalla los posibles modos de falla del microservicio **MS Reportes** (PINGESO Civil, Grupo 1), las medidas implementadas para impedirlos o mitigarlos y la forma en que se verifica cada caso.

---

## 1. Validación de Entrada y Contrato (`POST /api/reportes`)

| Modo de falla | Causa / Escenario | Qué la impide (Mitigación) | Cómo se verifica |
| :--- | :--- | :--- | :--- |
| **Formato de salida inválido** | El cliente solicita un formato distinto a `pdf` o `xlsx` (ej. `csv`, `docx` o vacío). | Validación estricta en el contrato OpenAPI/Huma (enum cerrado `pdf` \| `xlsx`) y validación defensiva en `internal/service` que retorna error controlado (HTTP 400 / 422 en formato RFC 9457). | Prueba unitaria del motor con formato inválido y prueba de integración contra `POST /api/reportes` confirmando rechazo sin generar archivo. |
| **Plantilla faltante o incompleta** | El request omite `plantilla`, envía un `id` que no es UUID válido, o envía `columnas` vacías / sin `campo` ni `encabezado`. | Validación de esquema en Huma (campos requeridos, formato UUID, longitud mínima de arreglos) y validación previa en el motor antes de renderizar. | Pruebas unitarias y de API enviando payloads sin `plantilla`, con UUID mal formado y con `columnas: []`, verificando respuesta de error controlado. |
| **Datos faltantes o vacíos** | El arreglo `datos` viene ausente (`null`) o vacío (`[]`). | Validación de entrada según regla de negocio: si `datos` es nulo se rechaza la solicitud (400/422); si está vacío, se controla explícitamente (rechazo controlado o reporte vacío solo con encabezados, sin provocar *panic*). | Caso de prueba unitario con `datos: nil` y `datos: []`, verificando que el motor no aborte ni genere un archivo corrupto. |
| **Desajuste entre columnas y datos** | Una columna definida en `plantilla.columnas` (ej. `codigo`, `estado`) no existe en alguna fila de `datos`. | El motor accede de forma segura al mapa de cada fila: si falta una clave, renderiza celda vacía (`""`) o retorna error de validación según la regla acordada, evitando *nil pointer dereference*. | Prueba unitaria con una fila incompleta (ej. `DOC-002` sin el campo `estado`) verificando que el PDF y el XLSX se generen de forma determinista y sin errores de ejecución. |

---

## 2. Motor de Generación PDF / XLSX (`internal/service`)

| Modo de falla | Causa / Escenario | Qué la impide (Mitigación) | Cómo se verifica |
| :--- | :--- | :--- | :--- |
| **Archivo corrupto o ilegible** | Error al escribir el buffer, cierre incompleto del documento o `Content-Type` incorrecto en la respuesta HTTP. | Generación completa en memoria/buffer validando cada error de la librería PDF/XLSX antes de responder; definición explícita de `application/pdf` y `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` en el contrato. | Prueba automatizada que valida la firma de bytes (*magic bytes*: `%PDF` para PDF, `PK` para XLSX) y prueba manual abriendo ambos archivos generados con `reporte-basico.json` (`Código`, `Estado`, `DOC-001`, `DOC-002`). |
| **Tamaño excesivo de datos / Memoria** | Solicitud con un volumen desproporcionado de filas que podría agotar la memoria del contenedor (OOM). | El límite de tamaño de payload se aplica en la capa de API/contrato (Huma), rechazando con HTTP 413 (*Payload Too Large*) antes de llegar al motor. El motor es una función pura en memoria; como defensa adicional puede sumar un tope máximo de filas configurable. | Prueba de integración enviando payloads por sobre el límite (HTTP 413) y prueba unitaria/de carga si se activa el tope de filas en el motor. |
| **Timeout en generación** | Un reporte complejo tarda demasiado en construirse y bloquea la conexión HTTP. | Propagación de `context.Context` con timeout explícito hacia el motor de generación; si el contexto expira, se cancela el trabajo y se responde con error controlado (HTTP 504 / 408). | Prueba con contexto cancelado o con timeout muy corto verificando que el motor detiene la ejecución y retorna error sin dejar goroutines colgadas. |

---

## 3. Dependencias Externas (G2: Gestor de Información y G3: Plantillas)

| Modo de falla | Causa / Escenario | Qué la impide (Mitigación) | Cómo se verifica |
| :--- | :--- | :--- | :--- |
| **Proveedor caído o inalcanzable** | MS Gestor de Información (G2) o MS Plantillas (G3) no responde o devuelve error 5xx. | Clientes HTTP generados desde `contracts/externos/` con manejo explícito de errores de red y estado HTTP; MS Reportes devuelve error claro en formato RFC 9457 (HTTP 502 / 503) sin exponer trazas internas ni colgar la request. | En Hito 1: pruebas contra mocks (Prism/MSW) simulando 500/503. En Hito 2: pruebas E2E con servidores HTTP mock locales simulando caída de G2/G3 y verificando que no se registre auditoría de éxito. |
| **Timeout de proveedor externo** | G2 o G3 demora más de lo tolerable en responder. | Timeout explícito configurado en todo cliente HTTP saliente (5 segundos por defecto) usando `context.Context`. | Prueba E2E con servidor mock local que retrasa la respuesta por sobre el timeout configurado, verificando respuesta 504 controlada. |
| **Recurso no encontrado en proveedor** | El identificador de plantilla no existe en G3 o los datos solicitados no existen en G2 (HTTP 404). | Propagación controlada del error indicando al cliente que el recurso referenciado no existe (HTTP 404 / 422 en formato RFC 9457) sin intentar generar un archivo vacío. | Prueba de integración/E2E simulando respuesta 404 desde G2 o G3. |

---

## 4. Seguridad y Persistencia (Alcance Hito 2)

> *Nota: Estos modos de falla se dejan identificados desde el Hito 1 y su verificación automatizada completa (E2E contra PostgreSQL real y matriz de roles) se cerrará en el Hito 2.*

| Modo de falla | Causa / Escenario | Qué la impide (Mitigación) | Cómo se verifica |
| :--- | :--- | :--- | :--- |
| **Token JWT inválido o expirado** | Petición sin cookie de sesión o con JWT rechazado por `AUTH_JWKS_URL` (G6). | Middleware de autenticación validando firma (EdDSA/RS256) y vigencia antes de llegar al handler; respuesta HTTP 401. | Pruebas en `acceso_test.go` y E2E sin token o con token expirado/alterado. |
| **Permiso insuficiente** | Usuario autenticado sin el permiso requerido (ej. `reportes:generar`). | Validación de autorización por permiso en el endpoint; respuesta HTTP 403 sin ejecutar el motor. | Matriz de acceso por rol en `acceso_test.go` y prueba E2E verificando rechazo 403. |
| **Base de datos PostgreSQL no disponible** | Falla de conexión a PostgreSQL al intentar registrar auditoría o al consultar salud. | Endpoint `GET /readyz` verifica conexión real a PostgreSQL (responde 200 solo si contesta); fallos al persistir auditoría se manejan con error controlado sin exponer consultas SQL. | Prueba de `GET /readyz` y pruebas E2E contra PostgreSQL real (Testcontainers / `TEST_DATABASE_URL`). |