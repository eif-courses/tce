export type Lang = 'lt' | 'en'
export type Program = 'pi' | 'se'

export interface Section {
  id: string
  label: string
  requirement: string
  issues: number
}

export interface ContentBlock {
  id: string
  sectionId: string
  type: 'paragraph' | 'image' | 'table' | 'heading' | 'list'
  content: TextContent | ImageContent | TableContent | HeadingContent | ListContent
}

export interface TextContent {
  text: string
}

export interface ImageContent {
  alt: string
  base64?: string
  caption?: string
  width?: number
  height?: number
}

export interface TableContent {
  headers: string[]
  rows: string[][]
  caption?: string
}

export interface HeadingContent {
  text: string
  level: number
}

export interface ListContent {
  items: string[]
  ordered: boolean
}

export interface Paragraph {
  id: string
  sectionId: string
  text: string
}

export interface Comment {
  id: string
  paragraphId: string
  category: 'structure' | 'content' | 'language' | 'formatting'
  severity: 'major' | 'minor'
  sectionLabel: string
  title: string
  message: string
}

export interface CheckResponse {
  sections: Section[]
  paragraphs: Paragraph[]
  contentBlocks: ContentBlock[]
  comments: Comment[]
}

export interface RuleSet {
  structure?: any
  sections?: Record<string, any>
  language?: any
  layout?: any
  figures?: any
  tables?: any
  references?: any
  content?: any
  code?: any
  severity?: Record<string, string>
}
