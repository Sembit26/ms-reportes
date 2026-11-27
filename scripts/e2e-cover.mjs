import { execFileSync } from "node:child_process"
import { existsSync, mkdirSync, rmSync } from "node:fs"
import { resolve } from "node:path"

const root = resolve(import.meta.dirname, "..")
const apiDir = resolve(root, "apps", "api")
const covDir = resolve(root, ".gocover")
const binExt = process.platform === "win32" ? ".exe" : ""
const binPath = resolve(covDir, `ms-reportes-e2e${binExt}`)
const covOut = resolve(root, "coverage.out")

if (existsSync(covDir)) {
  rmSync(covDir, { recursive: true, force: true })
}
mkdirSync(covDir, { recursive: true })

execFileSync("go", ["build", "-cover", "-o", binPath, "./cmd/api"], {
  cwd: apiDir,
  stdio: "inherit",
})

execFileSync("go", ["test", "-v", "./e2e/..."], {
  cwd: apiDir,
  stdio: "inherit",
  env: {
    ...process.env,
    API_BINARY: binPath,
    GOCOVERDIR: covDir,
  },
})

execFileSync("go", ["tool", "covdata", "textfmt", `-i=${covDir}`, `-o=${covOut}`], {
  cwd: apiDir,
  stdio: "inherit",
})
