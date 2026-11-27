import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { resolve } from "node:path"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": resolve(import.meta.dirname, "src"),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: process.env.VITE_DEV_API_URL || "http://localhost:5111",
        changeOrigin: true,
      },
      "/healthz": {
        target: process.env.VITE_DEV_API_URL || "http://localhost:5111",
        changeOrigin: true,
      },
    },
  },
})
