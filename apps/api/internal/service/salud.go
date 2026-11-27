// Package service contiene las reglas de negocio del microservicio sin HTTP ni SQL.
package service

import (
	"context"
	"errors"
	"time"
)

const tiempoEsperaSalud = 2 * time.Second

// VerificadorConexion abstrae la comprobación de disponibilidad del store.
type VerificadorConexion interface {
	VerificarConexion(ctx context.Context) error
}

// ServicioSalud evalúa el estado operativo de las dependencias propias.
type ServicioSalud struct {
	verificador VerificadorConexion
}

// NuevoServicioSalud crea el servicio de disponibilidad.
func NuevoServicioSalud(verificador VerificadorConexion) *ServicioSalud {
	return &ServicioSalud{verificador: verificador}
}

// VerificarDisponibilidad confirma que la capa de persistencia responda dentro del plazo.
func (s *ServicioSalud) VerificarDisponibilidad(ctx context.Context) error {
	if s == nil || s.verificador == nil {
		return errors.New("verificador de disponibilidad no configurado")
	}
	ctxConLimite := ctx
	if _, tieneLimite := ctx.Deadline(); !tieneLimite {
		var cancelar context.CancelFunc
		ctxConLimite, cancelar = context.WithTimeout(ctx, tiempoEsperaSalud)
		defer cancelar()
	}
	return s.verificador.VerificarConexion(ctxConLimite)
}
