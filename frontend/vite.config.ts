import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Under `wails dev` the API is served in-process, so no proxy. For plain browser dev
// against `make run`, start with VITE_API_PROXY_TARGET=http://localhost:8080.
const apiTarget = process.env.VITE_API_PROXY_TARGET;

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: apiTarget
      ? { "/query": { target: apiTarget, changeOrigin: true }, "/backup": { target: apiTarget, changeOrigin: true } }
      : undefined,
  },
});
