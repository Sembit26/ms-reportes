import { afterEach, describe, expect, it, vi } from "vitest"
import { consultarSaludServicio } from "./estado-servicio.api"

describe("consultarSaludServicio", () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it("retorna la respuesta cuando /healthz responde 200", async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        estado: "ok",
        servicio: "ms-reportes",
        version: "dev",
      }),
    })
    vi.stubGlobal("fetch", mockFetch)

    const resultado = await consultarSaludServicio("http://localhost:5111/")
    expect(mockFetch).toHaveBeenCalledWith("http://localhost:5111/healthz", {
      credentials: "include",
    })
    expect(resultado.estado).toBe("ok")
  })

  it("lanza error claro cuando la respuesta no es exitosa", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
      }),
    )

    await expect(consultarSaludServicio("http://localhost:5111")).rejects.toThrow("503")
  })
})
