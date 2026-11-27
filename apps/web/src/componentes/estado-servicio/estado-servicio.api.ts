export interface RespuestaSalud {
  estado: string
  servicio: string
  version: string
}

/** Consulta el estado de salud del backend usando la URL base recibida por props. */
export async function consultarSaludServicio(baseUrl: string): Promise<RespuestaSalud> {
  const ruta = `${baseUrl.replace(/\/+$/, "")}/healthz`
  const respuesta = await fetch(ruta, {
    credentials: "include",
  })
  if (!respuesta.ok) {
    throw new Error(`Error al consultar salud: ${respuesta.status}`)
  }
  return (await respuesta.json()) as RespuestaSalud
}
