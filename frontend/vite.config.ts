import path from "path"
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import compression from 'vite-plugin-compression'
import { SITE_URL } from './src/lib/site'

// https://vite.dev/config/
export default defineConfig({
  base: '/',
  define: {
    __APP_BASE_URL__: JSON.stringify('/'),
  },
  plugins: [
    react(),
    tailwindcss(),
    {
      name: 'site-url',
      transformIndexHtml: (html) => html.replaceAll('%SITE_URL%', SITE_URL),
    },
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
})
