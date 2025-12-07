export default defineEventHandler(async (event) => {
  const goBackendUrl = process.env.GO_BACKEND_URL || 'http://localhost:8080'

  try {
    const body = await readBody(event)

    const response = await fetch(`${goBackendUrl}/api/tce/suggest-rewrite`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`Backend returned ${response.status}: ${errorText}`)
    }

    const result = await response.json()
    return result
  } catch (error: any) {
    console.error('Suggest-rewrite request failed:', error.message)
    throw createError({
      statusCode: 500,
      statusMessage: 'Failed to generate suggestion',
      data: {
        error: error.message
      }
    })
  }
})
