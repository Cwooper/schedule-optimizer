/// <reference types="vite/client" />

declare global {
  const __APP_BASE_URL__: string

  interface ImportMetaEnv {
    readonly VITE_SITE_URL: string
  }
}

export {}
