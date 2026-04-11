import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const routePrefix = process.env.VITE_ROUTE_PREFIX || "/";

export default defineConfig({
  base: routePrefix === "/" ? "/" : `${routePrefix}/`,
  plugins: [react()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
    assetsDir: "assets",
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes("node_modules")) return;
          if (id.includes("@mui") || id.includes("@emotion")) return "mui";
          if (id.includes("react")) return "react-vendor";
          return "vendor";
        },
      },
    },
  },
  publicDir: "public"
});
