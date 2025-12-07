export default defineNuxtConfig({
  ssr: true,
  devtools: { enabled: true },

  // App configuration
  app: {
    head: {
      title: 'TCE - Document Analysis Engine',
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { hid: 'description', name: 'description', content: 'Academic document analysis engine with AI-powered suggestions' }
      ]
    }
  },

  // Nitro server configuration
  nitro: {
    // Proxy API requests to Go backend
    routeRules: {
      '/api/tce/**': {
        proxy: process.env.GO_BACKEND_URL || 'http://localhost:8080/api/tce/**'
      }
    },
    // Enable CORS for development
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization'
    }
  },

  // Runtimes
  runtimes: {
    httpServer: 'node'
  },

  // Build configuration
  build: {
    transpile: []
  },

  // Environment variables
  runtimeConfig: {
    public: {
      goBackendUrl: process.env.GO_BACKEND_URL || 'http://localhost:8080',
      apiBase: '/api/tce'
    }
  },

  // Module imports
  modules: [],

  // Compatibility
  compatibilityDate: '2025-12-07'
})
