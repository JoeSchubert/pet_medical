import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import en from './locales/en.json'
import es from './locales/es.json'
import fr from './locales/fr.json'
import de from './locales/de.json'

type Messages = Record<string, string>

const locales: Record<string, Messages> = {
  en: en as Messages,
  es: es as Messages,
  fr: fr as Messages,
  de: de as Messages,
}

function getMessages(lang: string): Messages {
  const normalized = lang.split('-')[0].toLowerCase()
  return locales[normalized] ?? locales.en
}

function replaceParams(text: string, params?: Record<string, string>): string {
  if (!params) return text
  return text.replace(/\{\{(\w+)\}\}/g, (_, key) => params[key] ?? `{{${key}}}`)
}

interface I18nContextValue {
  language: string
  setLanguage: (lang: string) => void
  t: (key: string, params?: Record<string, string>) => string
}

const I18nContext = createContext<I18nContextValue | null>(null)

export function I18nProvider({
  children,
  defaultLanguage = 'en',
}: {
  children: React.ReactNode
  defaultLanguage?: string
}) {
  const [language, setLanguage] = useState(defaultLanguage)
  useEffect(() => {
    setLanguage(defaultLanguage)
  }, [defaultLanguage])
  const messages = useMemo(() => getMessages(language), [language])
  const t = useCallback(
    (key: string, params?: Record<string, string>) => {
      const text = messages[key] ?? key
      return replaceParams(text, params)
    },
    [messages]
  )
  const value = useMemo(
    () => ({ language, setLanguage, t }),
    [language, t]
  )
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useTranslation() {
  const ctx = useContext(I18nContext)
  if (!ctx) {
    const fallback = (key: string, params?: Record<string, string>) => {
      const text = (en as Record<string, string>)[key] ?? key
      return replaceParams(text, params)
    }
    return { language: 'en', setLanguage: () => {}, t: fallback }
  }
  return ctx
}
