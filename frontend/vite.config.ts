import path from "path"
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import compression from 'vite-plugin-compression'

// https://vite.dev/config/
export default defineConfig({
  base: '/schedule-optimizer/',
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
      "@": path.resolve(__dirname, "./src"),
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
