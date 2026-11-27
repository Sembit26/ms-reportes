import { EstadoServicio } from "./componentes/estado-servicio/estado-servicio"

export function App() {
  return (
    <main className="mx-auto max-w-2xl p-6">
      <h1 className="text-2xl font-bold text-slate-900">MS Reportes · App Local</h1>
      <p className="mt-1 text-sm text-slate-600">
        Entorno local de desarrollo para montar microcomponentes (DAS §3).
      </p>
      <div className="mt-6">
        <EstadoServicio baseUrl="" />
      </div>
    </main>
  )
}
