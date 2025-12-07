import { suggestRewrite } from '~/server/utils/llm'
import type { Lang } from '~/server/utils/types'

export default defineEventHandler(async (event) => {
  try {
    const body = await readBody(event)

    const {
      paragraphText,
      commentTitle,
      commentMessage,
      lang = 'lt'
    } = body

    if (!paragraphText) {
      throw createError({
        statusCode: 400,
        statusMessage: 'Paragraph text is required'
      })
    }

    const suggestion = await suggestRewrite(
      paragraphText,
      commentTitle,
      commentMessage,
      lang as Lang
    )

    return {
      suggestion
    }
  } catch (error: any) {
    console.error('Suggest-rewrite request failed:', error)
    throw createError({
      statusCode: 500,
      statusMessage: 'Failed to generate suggestion',
      data: {
        error: error.message
      }
    })
  }
})
