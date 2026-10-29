import { defineConfig, mergeConfig } from 'vite'
import { resolve } from 'node:path'
import storefrontConfig from '../../vite.config'

export default defineConfig((env) => mergeConfig(storefrontConfig(env), {
  build: {
    outDir: 'node_modules/.tmp/mobile-card-dist',
    emptyOutDir: true,
    rollupOptions: { input: resolve(import.meta.dirname, 'index.html') },
  },
}))
