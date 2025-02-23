import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  base: "/schedule-optimizer/",
  css: {
    modules: {
      localsConvention: "camelCase",
    },
  },
});
