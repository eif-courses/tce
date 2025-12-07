import { OpenAI } from 'openai'
import type { Lang } from './types'

let clientInstance: OpenAI | null = null

function getOpenAIClient(): OpenAI {
  if (!clientInstance) {
    const apiKey = process.env.OPENAI_API_KEY
    if (!apiKey) {
      throw new Error('OPENAI_API_KEY environment variable is not set')
    }
    clientInstance = new OpenAI({ apiKey })
  }
  return clientInstance
}

const systemPrompts: Record<Lang, string> = {
  lt: `Tu esi akademinio rašymo asistentai. Jūsų darbas yra pasiūlyti gerus akademinius parašymus studentams.
Kai jums duodamas tekstas ir komentaras, turite pasiūlyti geresnę paraišą, kuri:
- Atitinka akademinę kalbą
- Išsaugo pradinio teksto reikšmę
- Yra aiškus ir tiksli
- Naudoja profesionalę terminologiją
- Neatrankus ir formalu

Atsakykite tik su pasiūlymu, be papildomo aiškinimo.`,
  en: `You are an academic writing assistant. Your job is to suggest improved academic paragraphs to students.
When given a text and a comment, you should suggest a better version that:
- Follows academic language standards
- Preserves the original meaning
- Is clear and concise
- Uses professional terminology
- Is objective and formal

Respond only with the suggestion, without additional explanation.`
}

export async function suggestRewrite(
  paragraphText: string,
  commentTitle: string,
  commentMessage: string,
  lang: Lang
): Promise<string> {
  try {
    const client = getOpenAIClient()

    const userMessage = `
Text: "${paragraphText}"

Issue: ${commentTitle}
Description: ${commentMessage}

Please suggest an improved version of this text that addresses the issue above.`

    const response = await client.messages.create({
      model: process.env.OPENAI_MODEL || 'gpt-4o-mini',
      max_tokens: 500,
      system: systemPrompts[lang],
      messages: [
        {
          role: 'user',
          content: userMessage
        }
      ]
    })

    // Extract text from response
    const firstContent = response.content[0]
    if (firstContent && firstContent.type === 'text') {
      return firstContent.text
    }

    return 'Unable to generate suggestion'
  } catch (error) {
    console.error('OpenAI API error:', error)
    throw new Error('Failed to generate suggestion')
  }
}
