export default defineEventHandler(async (event) => {
  return {
    status: 'ok',
    engine: 'TCE',
    version: '1.0.0',
    runtime: 'Nuxt 4 + Nitro'
  }
})
