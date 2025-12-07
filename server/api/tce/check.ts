import { parseDOCX } from '~/server/utils/docparser'
import { loadRules } from '~/server/utils/config'
import { createAnalyzer } from '~/server/utils/analyzer'
import type { Lang, Program } from '~/server/utils/types'

export default defineEventHandler(async (event) => {
  try {
    // Read the multipart form data
    const formData = await readMultipartFormData(event)

    if (!formData) {
      throw createError({
        statusCode: 400,
        statusMessage: 'No form data provided'
      })
    }

    // Extract file and parameters
    let fileData: Buffer | null = null
    let lang: Lang = 'lt'
    let program: Program = 'pi'

    for (const field of formData) {
      if (field.name === 'file') {
        fileData = field.data
      } else if (field.name === 'lang') {
        lang = field.data.toString() as Lang
      } else if (field.name === 'program') {
        program = field.data.toString() as Program
      }
    }

    if (!fileData) {
      throw createError({
        statusCode: 400,
        statusMessage: 'No file provided'
      })
    }

    // Parse DOCX
    const parsed = await parseDOCX(fileData)

    // Load rules and analyze
    const rules = loadRules(program, lang)
    const analyzer = createAnalyzer(rules, lang, program)
    const analysis = analyzer.analyze(parsed.sections, parsed.paragraphs, parsed.contentBlocks)

    return {
      sections: analysis.sections,
      paragraphs: parsed.paragraphs,
      contentBlocks: parsed.contentBlocks,
      comments: analysis.comments
    }
  } catch (error: any) {
    console.error('Check request failed:', error)
    throw createError({
      statusCode: 500,
      statusMessage: 'Failed to process document',
      data: {
        error: error.message
      }
    })
  }
})
