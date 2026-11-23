# MS Reportes

Microservicio del Grupo 1 de PINGESO Civil. Su objetivo es generar reportes
finales en PDF o Excel. El motor recibe los datos del reporte y el identificador
de una plantilla, aplica la estructura de esa plantilla a los datos y devuelve
el archivo generado.

## Entrada y salida

- **Entrada:** identificador de la plantilla y datos que se incluirán en el reporte.
- **Salida:** archivo del reporte generado en PDF o Excel.

## Dependencias directas

- **MS Plantillas de Reporte (Grupo 3):** crea, almacena y versiona las
  plantillas que definen la estructura de los reportes.
- **MS Gestor de Información (Grupo 2):** carga, valida y organiza la
  información que se utiliza para generar los reportes.

## Desarrollo inicial

Para probar el motor de forma independiente, utilizaremos un archivo JSON
estático que represente una plantilla y un conjunto de datos ficticios.
A medida que los equipos proveedores publiquen sus contratos OpenAPI,
los incorporaremos para probar las integraciones.