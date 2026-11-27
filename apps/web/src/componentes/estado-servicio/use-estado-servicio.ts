import { useEffect, useState } from "react"
import { consultarSaludServicio, type RespuestaSalud } from "./estado-servicio.api"

export interface EstadoHookSalud {
  datos: RespuestaSalud | null
  cargando: boolean
  error: string | null
}

/** Gestiona el ciclo de consulta de salud del microcomponente. */
export function useEstadoServicio(baseUrl: string): EstadoHookSalud {
  const [datos, setDatos] = useState<RespuestaSalud | null>(null)
  const [cargando, setCargando] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let activo = true
    setCargando(true)
    consultarSaludServicio(baseUrl)
      .then((res) => {
        if (activo) {
          setDatos(res)
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (activo) {
          setError(err instanceof Error ? err.message : "Error desconocido")
        }
      })
      .finally(() => {
        if (activo) {
          setCargando(false)
        }
      })
    return () => {
      activo = false
    }
  }, [baseUrl])

  return { datos, cargando, error }
}
