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
    storage: {
      fs: {
        driver: 'fs',
        base: './data'
      }
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
    openaiApiKey: process.env.OPENAI_API_KEY || '',
    openaiModel: process.env.OPENAI_MODEL || 'gpt-4o-mini',
    public: {
      apiBase: '/api/tce'
    }
  },

  // Module imports
  modules: [],

  // Compatibility
  compatibilityDate: '2025-12-07'
})
