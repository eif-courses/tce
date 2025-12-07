import mammoth from 'mammoth'
import { nanoid } from 'nanoid'
import type { ContentBlock, Paragraph, Section } from './types'

export interface ParsedDocument {
  sections: Section[]
  paragraphs: Paragraph[]
  contentBlocks: ContentBlock[]
  text: string
}

export async function parseDOCX(fileBuffer: Buffer): Promise<ParsedDocument> {
  try {
    const result = await mammoth.extractRawText({ buffer: fileBuffer })
    const text = result.value

    // Parse into basic sections and paragraphs
    const sections = detectSections(text)
    const paragraphs = createParagraphs(text, sections)
    const contentBlocks = createContentBlocks(text, sections)

    return {
      sections,
      paragraphs,
      contentBlocks,
      text
    }
  } catch (error) {
    console.error('DOCX parsing error:', error)
    throw new Error('Failed to parse DOCX file')
  }
}

function detectSections(text: string): Section[] {
  // Simple section detection based on common academic thesis structure
  const commonSections = [
    { pattern: /^(?:introduction|įvadas)/im, label: 'Introduction', lt: 'Įvadas' },
    { pattern: /^(?:methodology|metodika)/im, label: 'Methodology', lt: 'Metodika' },
    { pattern: /^(?:results|rezultatai)/im, label: 'Results', lt: 'Rezultatai' },
    { pattern: /^(?:discussion|diskusija)/im, label: 'Discussion', lt: 'Diskusija' },
    { pattern: /^(?:conclusion|conclusions|išvados)/im, label: 'Conclusion', lt: 'Išvados' },
    { pattern: /^(?:references|literatūra)/im, label: 'References', lt: 'Literatūra' }
  ]

  const sections: Section[] = []
  const lines = text.split('\n')

  lines.forEach((line, index) => {
    const trimmed = line.trim()
    if (trimmed.length === 0) return

    commonSections.forEach(section => {
      if (section.pattern.test(trimmed)) {
        sections.push({
          id: `section-${sections.length}`,
          label: section.label,
          requirement: '',
          issues: 0
        })
      }
    })
  })

  // If no sections found, create a default one
  if (sections.length === 0) {
    sections.push({
      id: 'section-default',
      label: 'Content',
      requirement: '',
      issues: 0
    })
  }

  return sections
}

function createParagraphs(text: string, sections: Section[]): Paragraph[] {
  const paragraphs: Paragraph[] = []
  const lines = text.split('\n').filter(line => line.trim().length > 0)

  let currentSectionId = sections[0]?.id || 'section-default'

  lines.forEach((line, index) => {
    // Check if this line is a section header
    for (const section of sections) {
      if (section.label.toLowerCase().includes(line.toLowerCase()) ||
          line.toLowerCase().includes(section.label.toLowerCase())) {
        currentSectionId = section.id
        return
      }
    }

    // Skip short lines (likely titles/headers)
    if (line.trim().length < 10) return

    paragraphs.push({
      id: `para-${nanoid(8)}`,
      sectionId: currentSectionId,
      text: line.trim()
    })
  })

  return paragraphs
}

function createContentBlocks(text: string, sections: Section[]): ContentBlock[] {
  const blocks: ContentBlock[] = []
  const lines = text.split('\n')

  let currentSectionId = sections[0]?.id || 'section-default'
  let blockIndex = 0

  lines.forEach((line) => {
    const trimmed = line.trim()

    // Skip empty lines
    if (trimmed.length === 0) return

    // Check if this is a section header
    for (const section of sections) {
      if (trimmed.toLowerCase().includes(section.label.toLowerCase())) {
        currentSectionId = section.id
        return
      }
    }

    // Skip very short lines (headers)
    if (trimmed.length < 10) return

    // Create paragraph block
    blocks.push({
      id: `block-${nanoid(8)}`,
      sectionId: currentSectionId,
      type: 'paragraph',
      content: {
        text: trimmed
      }
    })

    blockIndex++
  })

  return blocks
}
