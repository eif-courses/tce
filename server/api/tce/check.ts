export default defineEventHandler(async (event) => {
  const goBackendUrl = process.env.GO_BACKEND_URL || 'http://localhost:8080'

  // Read the multipart form data
  const formData = await readMultipartFormData(event)

  if (!formData) {
    throw createError({
      statusCode: 400,
      statusMessage: 'No form data provided'
    })
  }

  // Reconstruct FormData for forwarding to Go backend
  const newFormData = new FormData()

  for (const field of formData) {
    if (field.name === 'file') {
      // Handle file field
      const file = new File([field.data], field.filename || 'document.docx', {
        type: field.type || 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
      })
      newFormData.append('file', file)
    } else {
      // Handle text fields
      newFormData.append(field.name, field.data.toString())
    }
  }

  try {
    const response = await fetch(`${goBackendUrl}/api/tce/check`, {
      method: 'POST',
      body: newFormData
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`Backend returned ${response.status}: ${errorText}`)
    }

    const result = await response.json()
    return result
  } catch (error: any) {
    console.error('Check request failed:', error.message)
    throw createError({
      statusCode: 500,
      statusMessage: 'Failed to process document',
      data: {
        error: error.message
      }
    })
  }
})
