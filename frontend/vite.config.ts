import path from "path"
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import compression from 'vite-plugin-compression'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // A bad value ships broken meta tags and makes ics.ts throw at page load.
  const siteUrl = loadEnv(mode, import.meta.dirname).VITE_SITE_URL
  if (!URL.canParse(siteUrl) || new URL(siteUrl).origin !== siteUrl) {
    throw new Error(`VITE_SITE_URL must be an origin like https://example.org, got "${siteUrl ?? ""}"`)
  }
  return {
    base: '/',
    define: {
      __APP_BASE_URL__: JSON.stringify('/'),
    },
    plugins: [
      react(),
      tailwindcss(),
      compression({
        algorithm: 'gzip',
        ext: '.gz',
        threshold: 1024, // Only compress files > 1kB
      }),
    ],
    resolve: {
      alias: {
        "@": path.resolve(import.meta.dirname, "./src"),
      },
    },
    server: {
      proxy: {
        "/api": {
          target: "http://localhost:48920",
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: '../backend/internal/static/dist',
      emptyOutDir: true,
    },
  }
})
