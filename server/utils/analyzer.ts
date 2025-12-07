import { nanoid } from 'nanoid'
import type { ContentBlock, Comment, Lang, Program, RuleSet, Section, Paragraph } from './types'

export interface AnalysisResult {
  comments: Comment[]
  sections: Section[]
}

export class DocumentAnalyzer {
  private rules: RuleSet
  private lang: Lang
  private program: Program

  constructor(rules: RuleSet, lang: Lang, program: Program) {
    this.rules = rules
    this.lang = lang
    this.program = program
  }

  analyze(
    sections: Section[],
    paragraphs: Paragraph[],
    contentBlocks: ContentBlock[]
  ): AnalysisResult {
    const comments: Comment[] = []

    // Apply various rule checks
    this.checkStructure(sections, comments)
    this.checkLanguageStyle(paragraphs, contentBlocks, comments)
    this.checkContentQuality(paragraphs, contentBlocks, comments)
    this.checkFormatting(contentBlocks, comments)

    // Update section issue counts
    const updatedSections = this.updateSectionIssues(sections, comments)

    return {
      comments,
      sections: updatedSections
    }
  }

  private checkStructure(sections: Section[], comments: Comment[]): void {
    // Check if required sections exist
    const requiredSections = this.rules.structure?.required_sections || []

    requiredSections.forEach((requiredSection: string) => {
      const found = sections.some(s => s.label.toLowerCase().includes(requiredSection.toLowerCase()))

      if (!found) {
        comments.push({
          id: `comment-${nanoid(8)}`,
          paragraphId: 'structure',
          category: 'structure',
          severity: 'major',
          sectionLabel: 'Structure',
          title: this.lang === 'en' ? 'Missing required section' : 'Trūksta reikalingo skyriaus',
          message: this.lang === 'en'
            ? `The document is missing the required "${requiredSection}" section.`
            : `Dokumente trūksta reikalingo "${requiredSection}" skyriaus.`
        })
      }
    })
  }

  private checkLanguageStyle(
    paragraphs: Paragraph[],
    contentBlocks: ContentBlock[],
    comments: Comment[]
  ): void {
    const allowedFirstPersonSections = this.rules.language?.allowed_first_person || []
    const informalPatterns = this.rules.language?.informal_patterns || []

    const textContent = [
      ...paragraphs.map(p => ({ id: p.id, text: p.text })),
      ...contentBlocks
        .filter(b => b.type === 'paragraph')
        .map(b => ({ id: b.id, text: (b.content as any).text }))
    ]

    textContent.forEach(({ id, text }) => {
      // Check for informal language
      informalPatterns.forEach((pattern: string) => {
        const regex = new RegExp(pattern, 'gi')
        if (regex.test(text)) {
          comments.push({
            id: `comment-${nanoid(8)}`,
            paragraphId: id,
            category: 'language',
            severity: 'minor',
            sectionLabel: 'Language',
            title: this.lang === 'en' ? 'Informal language detected' : 'Aptikta neformali kalba',
            message: this.lang === 'en'
              ? 'This text uses informal language. Consider using academic language.'
              : 'Šis tekstas naudoja neformialią kalbą. Patartina naudoti akademinę kalbą.'
          })
        }
      })

      // Check for first person usage
      const firstPersonPattern = /\b(I|we|me|us|my|our)\b/gi
      if (firstPersonPattern.test(text) && !allowedFirstPersonSections.includes('introduction')) {
        comments.push({
          id: `comment-${nanoid(8)}`,
          paragraphId: id,
          category: 'language',
          severity: 'minor',
          sectionLabel: 'Language',
          title: this.lang === 'en' ? 'First person usage' : 'Pirmojo asmens vartojimas',
          message: this.lang === 'en'
            ? 'Academic writing should minimize first-person pronouns. Consider using passive voice or third person.'
            : 'Akademinis rašymas turėtų minimaliai naudoti pirmojo asmens žodžius. Apsvarstykite pasyvųjį balsą arba trečiąjį asmenį.'
        })
      }
    })
  }

  private checkContentQuality(
    paragraphs: Paragraph[],
    contentBlocks: ContentBlock[],
    comments: Comment[]
  ): void {
    const maxParagraphLength = this.rules.language?.max_paragraph_length || 500

    const textContent = [
      ...paragraphs,
      ...contentBlocks
        .filter(b => b.type === 'paragraph')
        .map(b => ({
          id: b.id,
          sectionId: b.sectionId,
          text: (b.content as any).text
        }))
    ]

    textContent.forEach((para) => {
      const wordCount = para.text.split(/\s+/).length

      if (wordCount > maxParagraphLength) {
        comments.push({
          id: `comment-${nanoid(8)}`,
          paragraphId: para.id,
          category: 'content',
          severity: 'minor',
          sectionLabel: 'Content',
          title: this.lang === 'en' ? 'Long paragraph' : 'Ilgas paragrafas',
          message: this.lang === 'en'
            ? `This paragraph contains ${wordCount} words. Consider breaking it into smaller paragraphs for better readability.`
            : `Šis paragrafas turi ${wordCount} žodžių. Apsvarstykite jį padalinti į mažesnius paragrafus, kad būtų geriau skaitomas.`
        })
      }
    })
  }

  private checkFormatting(contentBlocks: ContentBlock[], comments: Comment[]): void {
    // Check for tables without captions
    const tablesWithoutCaptions = contentBlocks.filter(
      b => b.type === 'table' && !(b.content as any).caption
    )

    tablesWithoutCaptions.forEach((block) => {
      comments.push({
        id: `comment-${nanoid(8)}`,
        paragraphId: block.id,
        category: 'formatting',
        severity: 'minor',
        sectionLabel: 'Formatting',
        title: this.lang === 'en' ? 'Missing table caption' : 'Trūksta lentelės pavadinimo',
        message: this.lang === 'en'
          ? 'Tables should have descriptive captions.'
          : 'Lentelės turėtų turėti aprašomąsias antraštes.'
      })
    })

    // Check for images without captions
    const imagesWithoutCaptions = contentBlocks.filter(
      b => b.type === 'image' && !(b.content as any).caption
    )

    imagesWithoutCaptions.forEach((block) => {
      comments.push({
        id: `comment-${nanoid(8)}`,
        paragraphId: block.id,
        category: 'formatting',
        severity: 'minor',
        sectionLabel: 'Formatting',
        title: this.lang === 'en' ? 'Missing figure caption' : 'Trūksta paveikslėlio pavadinimo',
        message: this.lang === 'en'
          ? 'Figures should have descriptive captions.'
          : 'Paveikslėliai turėtų turėti aprašomąsias antraštes.'
      })
    })
  }

  private updateSectionIssues(sections: Section[], comments: Comment[]): Section[] {
    return sections.map(section => {
      const sectionComments = comments.filter(c => {
        const sectionLabel = c.sectionLabel.toLowerCase()
        const label = section.label.toLowerCase()
        return sectionLabel.includes(label) || label.includes(sectionLabel)
      })

      return {
        ...section,
        issues: sectionComments.length > 0 ? sectionComments.length : 0
      }
    })
  }
}

export function createAnalyzer(rules: RuleSet, lang: Lang, program: Program): DocumentAnalyzer {
  return new DocumentAnalyzer(rules, lang, program)
}
