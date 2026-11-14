import { defineConfig } from "vite";

// While developing, the visual page is served by Vite and the backend runs
// separately on port 8080. This passes the page's live connection through to
// the backend.
export default defineConfig({
  server: {
    proxy: {
      "/ws": { target: "ws://localhost:8080", ws: true },
    },
  },
});
