import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    // three.js is inherently large and is isolated on its own lazy-loaded
    // chunk, so raise the advisory threshold above it.
    chunkSizeWarningLimit: 1400,
    rollupOptions: {
      output: {
        manualChunks: {
          // Isolate the heavy three.js stack so it loads on its own chunk
          // (the hero scene is lazy-loaded) and caches independently.
          three: ['three', '@react-three/fiber', '@react-three/drei', '@react-three/postprocessing'],
          highlighter: ['react-syntax-highlighter'],
          vendor: ['react', 'react-dom', 'react-router-dom', 'framer-motion'],
        },
      },
    },
  },
})
