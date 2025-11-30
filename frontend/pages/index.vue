<script setup lang="ts">
import { ref, computed } from 'vue'

type Lang = 'lt' | 'en'
type Program = 'pi' | 'se'

interface Section {
  id: string
  label: string
  requirement: string
  issues: number // -1 = missing, 0 = ok, >0 = count
}

// Rich content types
interface ContentBlock {
  id: string
  sectionId: string
  type: 'paragraph' | 'image' | 'table' | 'heading' | 'list'
  content: TextContent | ImageContent | TableContent | HeadingContent | ListContent
}

interface TextContent {
  text: string
}

interface ImageContent {
  alt: string
  base64?: string
  caption?: string
  width?: number
  height?: number
}

interface TableContent {
  headers: string[]
  rows: string[][]
  caption?: string
}

interface HeadingContent {
  text: string
  level: number // 1-6
}

interface ListContent {
  items: string[]
  ordered: boolean
}

// Legacy paragraph interface (kept for backwards compatibility)
interface Paragraph {
  id: string
  sectionId: string
  text: string
}

interface Comment {
  id: string
  paragraphId: string // Can also reference ContentBlock ID
  category: 'structure' | 'content' | 'language' | 'formatting'
  severity: 'major' | 'minor'
  sectionLabel: string
  title: string
  message: string
}

// === LANGUAGE & PROGRAM STATE ===
const docLanguage = ref<Lang>('lt')
const studyProgram = ref<Program>('pi')

// UI texts based on language
const uiText = computed(() => {
  if (docLanguage.value === 'en') {
    return {
      appTitle: 'Thesis Check Tool',
      appSubtitle:
        'Upload your document and get automatic analysis of structure, content and formatting.',
      summary: 'Summary',
      sections: 'Sections',
      allSections: 'All',
      overallMatch: 'Overall compliance',
      categories: 'Categories',
      regulations: 'Regulations',
      generalReq: 'General requirements',
      piGuidelines: 'PI guidelines',
      seGuidelines: 'SE Final Project',
      issuesCount: 'issues',
      majorIssuesShort: 'major',
      minorIssuesShort: 'minor',
      readingNow: 'You are reading:',
      problemsLabel: 'issues',
      termsHeading: 'List of terms and abbreviations',
      introHeading: 'Introduction',
      taskHeading: 'Practice task and system',
      frHeading: 'Functional requirements',
      comments: 'Comments',
      sortByImportance: 'By severity',
      sortBySection: 'By section',
      severityMajor: 'Major',
      severityMinor: 'Minor',
      uploadButtonIdle: 'Upload document',
      uploadButtonBusy: 'Checking…',
      fileLabel: 'File:',
      summaryTitleTop: 'Summary',
      categoryStructure: 'Structure',
      categoryContent: 'Content / methodology',
      categoryLanguage: 'Language / style',
      categoryFormatting: 'Formatting',
      llmSuggestButton: 'Suggest improved text',
      llmLoading: 'Generating suggestion…',
      llmError: 'Could not generate suggestion. Please try again.'
    }
  }

  // Lithuanian default
  return {
    appTitle: 'Baigiamųjų darbų tikrintuvas',
    appSubtitle:
      'Įkelkite dokumentą ir gaukite automatinę struktūros, turinio ir formatavimo analizę.',
    summary: 'Suvestinė',
    sections: 'Skyriai',
    allSections: 'Visi',
    overallMatch: 'Bendras atitikimas',
    categories: 'Kategorijos',
    regulations: 'Reglamentavimas',
    generalReq: 'Bendrieji reikalavimai',
    piGuidelines: 'PI BD metodiniai',
    seGuidelines: 'SE Final Project',
    issuesCount: 'problemų',
    majorIssuesShort: 'rimtos',
    minorIssuesShort: 'mažesnių',
    readingNow: 'Dabar skaitote:',
    problemsLabel: 'problemų',
    termsHeading: 'Terminų ir santrumpų paaiškinimų sąrašas',
    introHeading: 'Įvadas',
    taskHeading: 'Praktikos užduotis ir sistema',
    frHeading: 'Funkciniai reikalavimai',
    comments: 'Komentarai',
    sortByImportance: 'Pagal svarbumą',
    sortBySection: 'Pagal skyrių',
    severityMajor: 'Didelė',
    severityMinor: 'Maža',
    uploadButtonIdle: 'Įkelti dokumentą',
    uploadButtonBusy: 'Vykdoma patikra…',
    fileLabel: 'Failas:',
    summaryTitleTop: 'Suvestinė',
    categoryStructure: 'Struktūra',
    categoryContent: 'Turinys / metodika',
    categoryLanguage: 'Kalba / stilius',
    categoryFormatting: 'Formatavimas',
    llmSuggestButton: 'Pasiūlyti pataisytą tekstą',
    llmLoading: 'Generuojamas pasiūlymas…',
    llmError: 'Nepavyko sugeneruoti pasiūlymo. Bandykite dar kartą.'
  }
})

// === DATA ===
const sections = ref<Section[]>([])
const contentBlocks = ref<ContentBlock[]>([])
const paragraphs = ref<Paragraph[]>([]) // Legacy support
const comments = ref<Comment[]>([])

// === UI / UPLOAD STATE ===
const selectedCommentId = ref<string | null>(null)
const sortMode = ref<'severity' | 'section'>('severity')
const activeSectionId = ref<string | 'all'>('all')

const fileInput = ref<HTMLInputElement | null>(null)
const isChecking = ref(false)
const lastFileName = ref<string | null>(null)
const errorMessage = ref<string | null>(null)

// === LLM SUGGESTION STATE (per comment) ===
const suggestionByComment = ref<Record<string, string>>({})
const suggestionLoading = ref<Record<string, boolean>>({})
const suggestionError = ref<Record<string, string>>({})

// === COMPUTED ===
const selectedBlockId = computed(() => {
  return comments.value.find(x => x.id === selectedCommentId.value)?.paragraphId ?? null
})

const sortedComments = computed(() => {
  const arr = [...comments.value]
  if (sortMode.value === 'severity') {
    const severityOrder = { major: 0, minor: 1 } as const
    return arr.sort((a, b) => severityOrder[a.severity] - severityOrder[b.severity])
  }
  return arr.sort((a, b) => a.sectionLabel.localeCompare(b.sectionLabel))
})

const filteredSortedComments = computed(() => {
  if (activeSectionId.value === 'all') return sortedComments.value
  return sortedComments.value.filter((comment) => {
    const block = contentBlocks.value.find(b => b.id === comment.paragraphId)
    return block?.sectionId === activeSectionId.value
  })
})

const statsOverall = computed(() => ({
  compliance: 82,
  totalIssues: comments.value.length,
  major: comments.value.filter(c => c.severity === 'major').length,
  minor: comments.value.filter(c => c.severity === 'minor').length
}))

const categoryStats = computed(() => [
  { name: uiText.value.categoryStructure, key: 'structure', color: 'amber', count: comments.value.filter(c => c.category === 'structure').length },
  { name: uiText.value.categoryContent, key: 'content', color: 'sky', count: comments.value.filter(c => c.category === 'content').length },
  { name: uiText.value.categoryLanguage, key: 'language', color: 'emerald', count: comments.value.filter(c => c.category === 'language').length },
  { name: uiText.value.categoryFormatting, key: 'formatting', color: 'rose', count: comments.value.filter(c => c.category === 'formatting').length }
])

// === METHODS ===
function selectComment(comment: Comment) {
  selectedCommentId.value = comment.id
  const el = document.getElementById(comment.paragraphId)
  el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

function scrollToSection(sectionId: string) {
  activeSectionId.value = sectionId
  const el = document.getElementById(`sec-${sectionId}`)
  el?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function getIssueIndicator(issues: number) {
  if (issues > 0) return { text: issues.toString(), color: 'amber' }
  if (issues === 0) return { text: '✓', color: 'emerald' }
  return { text: 'nėra', color: 'rose' }
}

function triggerUpload() {
  fileInput.value?.click()
}

const apiBase = 'http://localhost:8080' // you can move this to runtime config later

async function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  errorMessage.value = null
  lastFileName.value = file.name
  isChecking.value = true
  selectedCommentId.value = null
  activeSectionId.value = 'all'
  suggestionByComment.value = {}
  suggestionLoading.value = {}
  suggestionError.value = {}

  const formData = new FormData()
  formData.append('file', file)
  formData.append('lang', docLanguage.value)
  formData.append('program', studyProgram.value)

  try {
    const result = await $fetch<{
      sections: Section[]
      paragraphs: Paragraph[]
      contentBlocks: ContentBlock[]
      comments: Comment[]
    }>(`${apiBase}/api/tce/check`, {
      method: 'POST',
      body: formData
    })

    sections.value = result.sections
    paragraphs.value = result.paragraphs || []
    contentBlocks.value = result.contentBlocks || []
    comments.value = result.comments
  } catch (err: any) {
    console.error('Check failed', err)
    errorMessage.value
      = err?.data?.error
        || err?.message
        || (docLanguage.value === 'en'
          ? 'An error occurred while checking the document. Please try again.'
          : 'Įvyko klaida tikrinant dokumentą. Bandykite dar kartą.')
  } finally {
    isChecking.value = false
    if (target) target.value = ''
  }
}

// === LLM: request rewrite for a specific comment ===
async function requestSuggestion(comment: Comment) {
  const block = contentBlocks.value.find(b => b.id === comment.paragraphId)
  if (!block) return

  let paragraphText = ''
  if (block.type === 'paragraph' && 'text' in block.content) {
    paragraphText = (block.content as TextContent).text
  }

  const cid = comment.id
  suggestionError.value[cid] = ''
  suggestionLoading.value[cid] = true

  try {
    const resp = await $fetch<{ suggestion: string }>(`${apiBase}/api/tce/suggest-rewrite`, {
      method: 'POST',
      body: {
        paragraphId: comment.paragraphId,
        paragraphText: paragraphText,
        lang: docLanguage.value,
        program: studyProgram.value,
        commentTitle: comment.title,
        commentMessage: comment.message
      }
    })

    suggestionByComment.value[cid] = resp.suggestion
  } catch (err: any) {
    console.error('LLM suggestion failed', err)
    suggestionError.value[cid]
      = err?.data?.error
        || err?.message
        || uiText.value.llmError
  } finally {
    suggestionLoading.value[cid] = false
  }
}

// Helper to get category color
function getCategoryDotClass(category: string) {
  const categoryColorMap: Record<string, string> = {
    structure: 'bg-amber-500',
    content: 'bg-sky-500',
    language: 'bg-emerald-500',
    formatting: 'bg-rose-500'
  }
  return categoryColorMap[category] || 'bg-gray-500'
}

// Helper to determine MIME type from base64 data
function getImageMimeType(base64: string): string {
  if (base64.startsWith('/9j/')) return 'image/jpeg'
  if (base64.startsWith('iVBORw0KGgo')) return 'image/png'
  if (base64.startsWith('R0lGOD')) return 'image/gif'
  if (base64.startsWith('UklGR')) return 'image/webp'
  return 'image/png' // default
}
</script>

<template>
  <!-- full-screen app -->
  <div class="h-screen bg-slate-100 flex flex-col">
    <!-- HEADER -->
    <header class="h-14 bg-white border-b border-slate-200 flex items-center justify-between px-6 text-xs">
      <div class="flex items-center gap-3">
        <span
          class="inline-flex items-center justify-center h-8 w-8 bg-slate-900 text-white text-xs font-bold rounded"
        >
          TC
        </span>
        <div>
          <p class="text-[13px] font-semibold text-slate-900">
            {{ uiText.appTitle }}
          </p>
          <p class="text-[11px] text-slate-500">
            {{ uiText.appSubtitle }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <!-- Program selector -->
        <USelectMenu
          v-model="studyProgram"
          :options="[
            { label: 'PI – Programų inžinerija', value: 'pi' },
            { label: 'SE – Software Engineering', value: 'se' }
          ]"
          size="xs"
          class="w-52"
          value-attribute="value"
          option-attribute="label"
        />

        <!-- Language selector -->
        <USelectMenu
          v-model="docLanguage"
          :options="[
            { label: 'Lietuvių', value: 'lt' },
            { label: 'English', value: 'en' }
          ]"
          size="xs"
          class="w-32"
          value-attribute="value"
          option-attribute="label"
        />

        <div
          v-if="lastFileName"
          class="text-[11px] text-slate-500 max-w-[220px] truncate"
        >
          <span class="font-medium text-slate-700">{{ uiText.fileLabel }}</span>
          {{ lastFileName }}
        </div>

        <input
          ref="fileInput"
          type="file"
          accept=".doc,.docx"
          class="hidden"
          @change="handleFileChange"
        >

        <UButton
          color="primary"
          size="sm"
          :loading="isChecking"
          @click="triggerUpload"
        >
          {{ isChecking ? uiText.uploadButtonBusy : uiText.uploadButtonIdle }}
        </UButton>
      </div>
    </header>

    <!-- ERROR BANNER -->
    <div
      v-if="errorMessage"
      class="bg-rose-50 border-b border-rose-200 text-rose-800 text-xs px-6 py-2"
    >
      {{ errorMessage }}
    </div>

    <!-- MAIN: 3 columns -->
    <main class="grid grid-cols-12 h-[calc(100vh-56px)]">
      <!-- LEFT: SUMMARY -->
      <aside class="col-span-3 bg-slate-50 border-r border-slate-200 h-full overflow-y-auto">
        <div class="px-4 py-3 border-b border-slate-200 sticky top-0 bg-slate-50 z-10">
          <h2 class="text-xs font-bold uppercase tracking-wider text-slate-600">
            {{ uiText.summaryTitleTop }}
          </h2>
        </div>

        <div class="p-4 space-y-4">
          <!-- sections -->
          <UCard>
            <template #header>
              <div class="flex items-center justify-between">
                <div class="text-xs font-bold text-slate-600 uppercase tracking-wider">
                  {{ uiText.sections }}
                </div>
                <button
                  type="button"
                  class="text-[11px] text-slate-500 hover:text-slate-900"
                  :class="activeSectionId === 'all' && 'font-semibold'"
                  @click="activeSectionId = 'all'"
                >
                  {{ uiText.allSections }}
                </button>
              </div>
            </template>
            <div class="space-y-1">
              <button
                v-for="section in sections"
                :key="section.id"
                type="button"
                class="w-full flex justify-between items-center px-2 py-1.5 text-sm hover:bg-slate-100 rounded"
                :class="[
                  section.issues === -1 && 'opacity-50 cursor-default',
                  activeSectionId === section.id ? 'bg-slate-100 font-semibold' : 'text-slate-700'
                ]"
                @click="section.issues !== -1 && scrollToSection(section.id)"
              >
                <span>{{ section.label }}</span>
                <span class="flex items-center gap-2">
                  <span class="text-xs text-slate-400">{{ section.requirement }}</span>
                  <UBadge
                    :color="section.issues > 0 ? 'amber' : section.issues === 0 ? 'green' : 'red'"
                    size="xs"
                  >
                    {{ getIssueIndicator(section.issues).text }}
                  </UBadge>
                </span>
              </button>
            </div>
          </UCard>

          <!-- overall -->
          <UCard>
            <template #header>
              <div class="text-xs text-slate-600">
                {{ uiText.overallMatch }}
              </div>
            </template>
            <div class="flex items-baseline justify-between">
              <span class="text-2xl font-bold text-slate-900">{{ statsOverall.compliance }}%</span>
              <span class="text-xs text-slate-500">
                {{ statsOverall.totalIssues }} {{ uiText.issuesCount }}
              </span>
            </div>
            <p class="text-xs text-slate-500 mt-2">
              {{ statsOverall.major }} {{ uiText.majorIssuesShort }} •
              {{ statsOverall.minor }} {{ uiText.minorIssuesShort }}
            </p>
          </UCard>

          <!-- categories -->
          <UCard>
            <template #header>
              <div class="text-xs font-bold text-slate-600 uppercase tracking-wider">
                {{ uiText.categories }}
              </div>
            </template>
            <div class="space-y-3">
              <div
                v-for="cat in categoryStats"
                :key="cat.key"
                class="flex items-center justify-between"
              >
                <span class="flex items-center gap-2">
                  <span
                    class="h-2 w-2 rounded-full"
                    :class="getCategoryDotClass(cat.key)"
                  />
                  <span class="text-sm text-slate-700">{{ cat.name }}</span>
                </span>
                <span class="text-xs text-slate-500">{{ cat.count }} {{ uiText.issuesCount }}</span>
              </div>
            </div>
          </UCard>

          <!-- regs -->
          <UCard>
            <template #header>
              <div class="text-xs font-bold text-slate-600 uppercase tracking-wider">
                {{ uiText.regulations }}
              </div>
            </template>
            <div class="space-y-2">
              <div class="flex justify-between text-sm">
                <span class="text-slate-700">{{ uiText.generalReq }}</span>
                <span class="text-xs text-slate-500">12 / 15</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-slate-700">{{ uiText.piGuidelines }}</span>
                <span class="text-xs text-amber-600 font-medium">3 trūkumai</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-slate-700">{{ uiText.seGuidelines }}</span>
                <span class="text-xs text-rose-600 font-medium">2 skyriai trūksta</span>
              </div>
            </div>
          </UCard>
        </div>
      </aside>

      <!-- CENTER: DOCUMENT -->
      <section class="col-span-6 bg-white h-full overflow-y-auto">
        <div
          class="sticky top-0 bg-white/95 backdrop-blur-sm border-b border-slate-200 px-8 py-2 flex justify-between items-center z-10 text-xs"
        >
          <div class="text-slate-600">
            {{ uiText.readingNow }}
            <span class="font-semibold text-slate-900">{{ uiText.introHeading }}</span>
          </div>
          <UBadge
            color="neutral"
            variant="subtle"
          >
            {{ statsOverall.totalIssues }} {{ uiText.problemsLabel }}
          </UBadge>
        </div>

        <div class="max-w-3xl mx-auto px-10 py-8 space-y-8 text-sm">
          <!-- Render content blocks if available -->
          <template v-if="contentBlocks.length > 0">
            <template v-for="(section, index) in sections" :key="section.id">
              <UDivider v-if="index > 0" />
              <section :id="`sec-${section.id}`">
                <h2 class="text-2xl font-bold text-slate-900 mb-4">
                  {{ section.label }}
                </h2>
                <div class="space-y-4">
                  <!-- Render each content block -->
                  <template v-for="block in contentBlocks.filter(b => b.sectionId === section.id)" :key="block.id">
                    <!-- Paragraph -->
                    <p
                      v-if="block.type === 'paragraph'"
                      :id="block.id"
                      class="leading-relaxed text-slate-700 transition-all"
                      :class="
                        selectedBlockId === block.id
                          && 'bg-blue-50 border-l-4 border-blue-600 pl-3 -ml-3 py-1'
                      "
                    >
                      {{ (block.content as TextContent).text }}
                    </p>

                    <!-- Heading -->
                    <component
                      :is="`h${Math.min((block.content as HeadingContent).level || 3, 6)}`"
                      v-else-if="block.type === 'heading'"
                      :id="block.id"
                      class="font-bold text-slate-900 transition-all"
                      :class="[
                        (block.content as HeadingContent).level === 1 && 'text-2xl',
                        (block.content as HeadingContent).level === 2 && 'text-xl',
                        (block.content as HeadingContent).level === 3 && 'text-lg',
                        (block.content as HeadingContent).level >= 4 && 'text-base',
                        selectedBlockId === block.id && 'bg-blue-50 border-l-4 border-blue-600 pl-3 -ml-3 py-1'
                      ]"
                    >
                      {{ (block.content as HeadingContent).text }}
                    </component>

                    <!-- Image -->
                    <figure
                      v-else-if="block.type === 'image'"
                      :id="block.id"
                      class="my-6 transition-all"
                      :class="
                        selectedBlockId === block.id
                          && 'bg-blue-50 border-l-4 border-blue-600 pl-3 -ml-3 py-1'
                      "
                    >
                      <img
                        v-if="(block.content as ImageContent).base64"
                        :src="`data:${getImageMimeType((block.content as ImageContent).base64!)};base64,${(block.content as ImageContent).base64}`"
                        :alt="(block.content as ImageContent).alt"
                        class="max-w-full h-auto rounded border border-slate-200"
                        :style="{
                          maxWidth: (block.content as ImageContent).width ? `${(block.content as ImageContent).width}px` : '100%'
                        }"
                      >
                      <div v-else class="bg-slate-100 border border-slate-300 rounded p-8 text-center text-slate-500">
                        <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        <p class="mt-2 text-xs">{{ (block.content as ImageContent).alt || 'Paveiksliukas' }}</p>
                      </div>
                      <figcaption
                        v-if="(block.content as ImageContent).caption"
                        class="mt-2 text-xs text-slate-500 text-center"
                      >
                        {{ (block.content as ImageContent).caption }}
                      </figcaption>
                    </figure>

                    <!-- Table -->
                    <div
                      v-else-if="block.type === 'table'"
                      :id="block.id"
                      class="my-6 overflow-x-auto transition-all"
                      :class="
                        selectedBlockId === block.id
                          && 'bg-blue-50 border-l-4 border-blue-600 pl-3 -ml-3 py-1'
                      "
                    >
                      <table class="min-w-full divide-y divide-slate-200 border border-slate-200 text-sm">
                        <thead class="bg-slate-50">
                          <tr>
                            <th
                              v-for="(header, idx) in (block.content as TableContent).headers"
                              :key="idx"
                              class="px-4 py-2 text-left text-xs font-medium text-slate-700 uppercase tracking-wider"
                            >
                              {{ header }}
                            </th>
                          </tr>
                        </thead>
                        <tbody class="bg-white divide-y divide-slate-200">
                          <tr
                            v-for="(row, rowIdx) in (block.content as TableContent).rows"
                            :key="rowIdx"
                          >
                            <td
                              v-for="(cell, cellIdx) in row"
                              :key="cellIdx"
                              class="px-4 py-2 text-slate-700"
                            >
                              {{ cell }}
                            </td>
                          </tr>
                        </tbody>
                      </table>
                      <p
                        v-if="(block.content as TableContent).caption"
                        class="mt-2 text-xs text-slate-500 text-center"
                      >
                        {{ (block.content as TableContent).caption }}
                      </p>
                    </div>

                    <!-- List -->
                    <component
                      :is="(block.content as ListContent).ordered ? 'ol' : 'ul'"
                      v-else-if="block.type === 'list'"
                      :id="block.id"
                      class="space-y-2 transition-all"
                      :class="[
                        (block.content as ListContent).ordered ? 'list-decimal list-inside' : 'list-disc list-inside',
                        selectedBlockId === block.id && 'bg-blue-50 border-l-4 border-blue-600 pl-3 -ml-3 py-1'
                      ]"
                    >
                      <li
                        v-for="(item, itemIdx) in (block.content as ListContent).items"
                        :key="itemIdx"
                        class="text-slate-700 leading-relaxed"
                      >
                        {{ item }}
                      </li>
                    </component>
                  </template>
                </div>
              </section>
            </template>
          </template>

          <!-- Fallback to legacy paragraph rendering if no content blocks -->
          <template v-else-if="paragraphs.length > 0">
            <template v-for="(section, index) in sections" :key="section.id">
              <UDivider v-if="index > 0" />
              <section :id="`sec-${section.id}`">
                <h2 class="text-2xl font-bold text-slate-900 mb-4">
                  {{ section.label }}
                </h2>
                <div class="space-y-4">
                  <p
                    v-for="p in paragraphs.filter((x) => x.sectionId === section.id)"
                    :id="p.id"
                    :key="p.id"
                    class="leading-relaxed text-slate-700 transition-all"
                    :class="
                      selectedBlockId === p.id
                        && 'bg-blue-50 border-l-4 border-blue-600 pl-3 -ml-3 py-1'
                    "
                  >
                    {{ p.text }}
                  </p>
                </div>
              </section>
            </template>
          </template>

          <!-- Empty state -->
          <div v-else class="text-center py-12 text-slate-500">
            <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p class="mt-4">{{ docLanguage === 'en' ? 'Upload a document to begin analysis' : 'Įkelkite dokumentą analizės pradžiai' }}</p>
          </div>
        </div>
      </section>

      <!-- RIGHT: COMMENTS -->
      <aside class="col-span-3 bg-slate-50 border-l border-slate-200 h-full flex flex-col">
        <div
          class="bg-white border-b border-slate-200 px-4 py-3 flex items-center justify-between sticky top-0 z-10"
        >
          <h3 class="text-xs font-bold uppercase tracking-wider text-slate-600">
            {{ uiText.comments }}
          </h3>
          <USelectMenu
            v-model="sortMode"
            :options="[
              { label: uiText.sortByImportance, value: 'severity' },
              { label: uiText.sortBySection, value: 'section' }
            ]"
            size="xs"
            class="w-36"
            value-attribute="value"
            option-attribute="label"
          />
        </div>

        <div class="flex-1 overflow-y-auto px-3 py-3 space-y-3">
          <div
            v-for="comment in filteredSortedComments"
            :key="comment.id"
          >
            <UCard
              :ui="{
                base:
                  selectedCommentId === comment.id
                    ? 'ring-2 ring-blue-500 bg-blue-50 cursor-pointer'
                    : 'cursor-pointer',
                body: { padding: 'p-3' }
              }"
              class="hover:shadow-md transition-shadow"
              @click="selectComment(comment)"
            >
              <div class="flex items-start justify-between gap-2 mb-2">
                <span class="flex items-center gap-2">
                  <span
                    class="h-2 w-2 rounded-full flex-shrink-0 mt-0.5"
                    :class="getCategoryDotClass(comment.category)"
                  />
                  <UBadge
                    :color="comment.severity === 'major' ? 'error' : 'warning'"
                    size="xs"
                    variant="subtle"
                  >
                    {{
                      comment.severity === 'major'
                        ? uiText.severityMajor
                        : uiText.severityMinor
                    }}
                  </UBadge>
                </span>
                <span class="text-xs text-slate-500 flex-shrink-0">
                  {{ comment.sectionLabel }}
                </span>
              </div>
              <p class="text-sm font-medium text-slate-900 line-clamp-2">
                {{ comment.title }}
              </p>
              <p class="text-xs text-slate-600 mt-1 leading-relaxed whitespace-pre-line">
                {{ comment.message }}
              </p>

              <!-- LLM SUGGEST BUTTON + RESULT -->
              <div class="mt-3 space-y-2">
                <UButton
                  size="xs"
                  variant="ghost"
                  class="text-emerald-700 hover:text-emerald-800 px-0"
                  :loading="suggestionLoading[comment.id]"
                  @click.stop="requestSuggestion(comment)"
                >
                  {{ suggestionLoading[comment.id] ? uiText.llmLoading : uiText.llmSuggestButton }}
                </UButton>

                <p
                  v-if="suggestionError[comment.id]"
                  class="text-[11px] text-rose-600"
                >
                  {{ suggestionError[comment.id] }}
                </p>

                <div
                  v-if="suggestionByComment[comment.id]"
                  class="text-[11px] leading-relaxed bg-emerald-50 border border-emerald-200 rounded px-2 py-2 text-emerald-900 whitespace-pre-line"
                >
                  {{ suggestionByComment[comment.id] }}
                </div>
              </div>
            </UCard>
          </div>

          <!-- Empty state for comments -->
          <div v-if="filteredSortedComments.length === 0" class="text-center py-12 text-slate-500">
            <svg class="mx-auto h-12 w-12 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
            </svg>
            <p class="mt-4 text-xs">{{ docLanguage === 'en' ? 'No comments yet' : 'Kol kas komentarų nėra' }}</p>
          </div>
        </div>
      </aside>
    </main>
  </div>
</template>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
