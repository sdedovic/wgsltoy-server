import react from "@vitejs/plugin-react";
import path from "node:path";
import type { Plugin } from "vite";
import { patchCssModules } from "vite-css-modules";
import { defineConfig } from "vitest/config";

const devServer = "http://localhost:5173";

// Vite reloads open pages when a template changes, since they're .html files under its root. It can't
// see the Go server restart though, so `just dev` touches tmp/reload once the server is back.
function reloadOnServerRestart(): Plugin {
  const sentinel = path.resolve(import.meta.dirname, "../tmp/reload");
  return {
    name: "reload-on-server-restart",
    apply: "serve",
    configureServer(server) {
      server.watcher.add(sentinel);
      const reload = (file: string) => {
        if (file === sentinel) {
          server.ws.send({ type: "full-reload", path: "*" });
        }
      };
      server.watcher.on("change", reload);
      server.watcher.on("add", reload);
    },
  };
}

export default defineConfig({
  base: "/static/",
  input: ["src/site.ts", "src/editor.tsx"],
  plugins: [react(), patchCssModules({ generateSourceTypes: true }), reloadOnServerRestart()],
  server: {
    port: 5173,
    strictPort: true,
    origin: devServer,
  },
  build: {
    manifest: true,
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: true,
    // every browser with WebGPU supports modulepreload
    modulePreload: { polyfill: false },
  },
  test: {
    include: ["src/**/*.test.ts"],
  },
});
