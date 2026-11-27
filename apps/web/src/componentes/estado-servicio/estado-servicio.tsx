import { useEstadoServicio } from "./use-estado-servicio"

export interface PropiedadesEstadoServicio {
  /** URL base del backend (por ejemplo "" para usar proxy /api en mismo origen). */
  baseUrl: string
}

/** Microcomponente autocontenido para verificar conectividad en desarrollo. */
export function EstadoServicio({ baseUrl }: PropiedadesEstadoServicio) {
  const { datos, cargando, error } = useEstadoServicio(baseUrl)

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
      <h2 className="text-sm font-semibold text-slate-700">Estado de MS Reportes</h2>
      {cargando && <p className="mt-2 text-xs text-slate-500">Consultando salud...</p>}
      {error && <p className="mt-2 text-xs text-red-600">{error}</p>}
      {datos && (
        <p className="mt-2 text-xs text-emerald-700">
          Servicio <strong>{datos.servicio}</strong> ({datos.version}): {datos.estado}
        </p>
      )}
    </section>
  )
}
