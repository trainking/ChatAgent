import { updateStore, getStore } from './store'
import { widgetAuth } from './api'
import { connectWS, subscribe, disconnectWS } from './websocket'
import { mount, unmount } from './ui'

interface WidgetConfig {
  inboxId: string
  mode: 'bubble' | 'embedded'
  brandColor: string
  position: 'left' | 'right'
}

function init(config: WidgetConfig) {
  const baseUrl = getScriptBaseUrl()

  updateStore({
    inboxId: config.inboxId,
    mode: config.mode === 'embedded' ? 'embedded' : config.position === 'left' ? 'bubble-left' : 'bubble',
    baseUrl: baseUrl,
    brandColor: config.brandColor || '#409EFF',
  })

  // Check if returning visitor
  const token = localStorage.getItem('chatagent_token_' + config.inboxId)
  const fp = getFingerprint()

  if (token) {
    updateStore({ pubsubToken: token })
    connectWS()
  }

  // Mount first, then refresh inbox presentation from auth.
  mount()

  // Auto-auth with fingerprint
  widgetAuth(fp).then(ok => {
    if (ok) {
      const s = getStore()
      if (s.pubsubToken) {
        localStorage.setItem('chatagent_token_' + config.inboxId, s.pubsubToken)
        connectWS()
      }
    }
  })
}

function getFingerprint(): string {
  let fp = localStorage.getItem('chatagent_fp')
  if (!fp) {
    fp = crypto.randomUUID()
    localStorage.setItem('chatagent_fp', fp)
  }
  return fp
}

function getScriptBaseUrl(): string {
  const script = document.currentScript as HTMLScriptElement | null
  if (script && script.src) {
    try {
      const url = new URL(script.src)
      return url.origin
    } catch {}
  }
  // Fallback: same origin
  return location.origin
}

// Auto-initialize from script tag data attributes
if (typeof document !== 'undefined') {
  const script = document.currentScript as HTMLScriptElement | null
  if (script) {
    const inboxId = script.getAttribute('data-inbox-id') || ''
    const mode = (script.getAttribute('data-mode') as 'bubble' | 'embedded') || 'bubble'
    const brandColor = script.getAttribute('data-color') || '#409EFF'
    const position = (script.getAttribute('data-position') as 'left' | 'right') || 'right'

    if (inboxId) {
      if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => init({ inboxId, mode, brandColor, position }))
      } else {
        init({ inboxId, mode, brandColor, position })
      }
    }
  }
}

// Expose programmatic API
export { init, unmount, getStore, updateStore }
