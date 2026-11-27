import { execFileSync } from "node:child_process"
import { resolve } from "node:path"

const root = resolve(import.meta.dirname, "..")
const apiDir = resolve(root, "apps", "api")
const openapiPath = resolve(root, "contracts", "openapi.yaml")

execFileSync("go", ["run", "./cmd/api", "-export-openapi", openapiPath], {
  cwd: apiDir,
  stdio: "inherit",
})

execFileSync(
  "pnpm",
  [
    "--filter",
    "@pingeso-civil/ms-reportes-ui",
    "exec",
    "openapi-typescript",
    openapiPath,
    "-o",
    "src/api/schema.d.ts",
  ],
  {
    cwd: root,
    stdio: "inherit",
    shell: process.platform === "win32",
  },
)
