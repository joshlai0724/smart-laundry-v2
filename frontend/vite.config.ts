import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig({
  base: "/web/",
  plugins: [react()],
  resolve: {
    alias: [{ find: "@", replacement: path.resolve(__dirname, "src") }],
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '@import "@/style/custom.scss";',
      },
    },
  },
  server: {
    proxy: {
      "/users": {
        target: "http://localhost:80/api/v1/users/",
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/users/, ''),
        // ws: true,
      },
      "/api": {
        target: "http://localhost:80/api/v1/",
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, ''),
        // ws: true,
      },
    },
  },
});
