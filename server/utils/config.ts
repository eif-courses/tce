import { readFileSync } from 'fs'
import { resolve } from 'path'
import yaml from 'js-yaml'
import type { RuleSet, Lang, Program } from './types'

const configCache: Map<string, RuleSet> = new Map()

export function loadRules(program: Program, lang: Lang): RuleSet {
  const key = `${program}_${lang}`

  // Check cache first
  if (configCache.has(key)) {
    return configCache.get(key)!
  }

  try {
    const configPath = resolve(process.cwd(), 'config', `rules_${program}_${lang}.yaml`)
    const fileContent = readFileSync(configPath, 'utf-8')
    const rules = yaml.load(fileContent) as RuleSet

    configCache.set(key, rules)
    return rules
  } catch (error) {
    console.error(`Failed to load rules for ${program}_${lang}:`, error)
    // Return empty rules object on failure
    return {}
  }
}

export function getRulesDescription(rules: RuleSet): Record<string, string> {
  const descriptions: Record<string, string> = {}

  // Extract rule descriptions from severity levels
  if (rules.severity) {
    Object.assign(descriptions, rules.severity)
  }

  return descriptions
}
