import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts",
    include: ["src/**/*.test.{ts,tsx}"],
  },
  server: {
    host: '0.0.0.0',
    port: 6173,
    watch: {
      usePolling: true,
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          livekit: ["@livekit/components-react", "livekit-client"],
          "react-vendor": ["react", "react-dom", "react-router-dom"],
          forms: ["react-hook-form", "@hookform/resolvers", "zod"],
          "ui-vendor": ["lucide-react", "zustand", "axios"],
        },
      },
    },
    chunkSizeWarningLimit: 600,
    minify: "esbuild",
    sourcemap: false,
    cssCodeSplit: true,
    target: "es2015",
    cssMinify: true,
  },
  optimizeDeps: {
    include: ["react", "react-dom", "react-router-dom"],
    exclude: ["@livekit/components-react", "livekit-client"],
  },
});
