import { defineConfig, mergeConfig } from 'vite'
import { resolve } from 'node:path'
import storefrontConfig from '../../vite.config'

export default defineConfig((env) => mergeConfig(storefrontConfig(env), {
  build: {
    outDir: 'node_modules/.tmp/mobile-card-dist',
    emptyOutDir: true,
    rollupOptions: {
      input: ['index.html', 'checkout.html', 'product-description.html'].map(file => resolve(import.meta.dirname, file)),
    },
  },
}))
