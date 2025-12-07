export default defineEventHandler(async (event) => {
  const goBackendUrl = process.env.GO_BACKEND_URL || 'http://localhost:8080'

  try {
    const response = await $fetch(`${goBackendUrl}/api/tce/health`)
    return response
  } catch (error: any) {
    console.error('Health check failed:', error.message)
    return {
      status: 'error',
      message: 'Failed to connect to backend'
    }
  }
})
