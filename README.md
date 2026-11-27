# MS Reportes

Microservicio del Grupo 1 de PINGESO Civil (`PINGESO-CIVIL/ms-reportes`). Su objetivo es generar reportes finales en PDF o Excel (`xlsx`). El motor recibe los datos del reporte y el identificador o definición de una plantilla, aplica la estructura de esa plantilla a los datos y devuelve el archivo generado.

## Enlaces del repositorio

- **Pipeline CI:** [GitHub Actions](https://github.com/PINGESO-CIVIL/ms-reportes/actions) _(actualizar si el nombre final en la organización cambia)_
- **SonarQube:** pendiente de enlazar URL del proyecto `ms-reportes` tras el alta en Hito 0

## Entrada y salida

- **Entrada:** identificador de la plantilla y datos que se incluirán en el reporte.
- **Salida:** archivo del reporte generado en PDF o Excel (`xlsx`).

## Dependencias directas

- **MS Plantillas de Reporte (Grupo 3):** crea, almacena y versiona las plantillas que definen la estructura de los reportes.
- **MS Gestor de Información (Grupo 2):** carga, valida y organiza la información que se utiliza para generar los reportes.

## Arranque local (Hito 0)

```bash
cp .env.example .env
pnpm install
pnpm exec lefthook install
docker compose up -d --build
```

Para desarrollo con perfil `dev` (`seed` y `mailpit` en `:8025`):

```bash
COMPOSE_PROFILES=dev docker compose up -d --build
```
