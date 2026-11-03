import path from "path"
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import compression from 'vite-plugin-compression'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // index.html would otherwise ship a literal %VITE_SITE_URL% in its tags.
  if (!loadEnv(mode, import.meta.dirname).VITE_SITE_URL) {
    throw new Error('VITE_SITE_URL is not set; see frontend/.env')
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
