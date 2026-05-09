import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'

function getDefaultLocale(): string {
  const stored = localStorage.getItem('lang')
  if (stored) return stored
  const browser = navigator.language || ''
  if (browser.startsWith('zh')) return 'zh-CN'
  return 'en-US'
}

const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: 'en-US',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS,
  },
})

export default i18n
