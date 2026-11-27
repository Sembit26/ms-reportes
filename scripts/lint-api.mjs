import { execFileSync } from "node:child_process"
import { resolve } from "node:path"

const root = resolve(import.meta.dirname, "..")
const apiDir = resolve(root, "apps", "api")

const sinFormato = execFileSync("gofmt", ["-l", "."], {
  cwd: apiDir,
  encoding: "utf8",
}).trim()

if (sinFormato !== "") {
  console.error("Archivos Go sin formato gofmt:\n" + sinFormato)
  process.exit(1)
}

execFileSync("go", ["vet", "./..."], {
  cwd: apiDir,
  stdio: "inherit",
})
