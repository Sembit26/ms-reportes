// Package store encapsula el acceso a PostgreSQL sin lógica de negocio.
package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AlmacenSalud verifica la disponibilidad de la base de datos propia.
type AlmacenSalud struct {
	pool *pgxpool.Pool
}

// NuevoAlmacenSalud construye el repositorio de salud sobre el pool PostgreSQL.
func NuevoAlmacenSalud(pool *pgxpool.Pool) *AlmacenSalud {
	return &AlmacenSalud{pool: pool}
}

// VerificarConexion ejecuta un ping contra PostgreSQL.
func (a *AlmacenSalud) VerificarConexion(ctx context.Context) error {
	if a == nil || a.pool == nil {
		return errors.New("conexión a base de datos no inicializada")
	}
	return a.pool.Ping(ctx)
}
