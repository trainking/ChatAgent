import { getStore, updateStore, addMessage, onStoreChange } from './store'
import { widgetAuth, sendMessage as apiSendMessage } from './api'
import { connectWS, subscribe, disconnectWS } from './websocket'

let root: HTMLElement | null = null
let shadow: ShadowRoot | null = null
let unsub: (() => void) | null = null
let inputEl: HTMLTextAreaElement | null = null

function h<K extends keyof HTMLElementTagNameMap>(tag: K, attrs: Record<string, string> = {}, ...children: (string | Node)[]): HTMLElementTagNameMap[K] {
  const el = document.createElement(tag)
  for (const [k, v] of Object.entries(attrs)) el.setAttribute(k, v)
  for (const c of children) el.append(typeof c === 'string' ? document.createTextNode(c) : c)
  return el
}

const css = `
:host { all: initial; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
* { box-sizing: border-box; margin: 0; padding: 0; }
.bubble-btn { position: fixed; right: 20px; bottom: 20px; width: 56px; height: 56px; border-radius: 50%; border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 700; color: #fff; box-shadow: 0 4px 12px rgba(0,0,0,0.2); z-index: 9999; transition: transform 0.2s; }
.bubble-btn:hover { transform: scale(1.1); }
.bubble-btn.left { right: auto; left: 20px; }
.window { position: fixed; right: 20px; bottom: 20px; width: 372px; height: 584px; max-height: calc(100vh - 40px); border-radius: 14px; overflow: hidden; display: flex; flex-direction: column; box-shadow: 0 16px 48px rgba(15,23,42,0.22); z-index: 9998; border: 1px solid rgba(15,23,42,0.08); background: #fff; }
.window.left { right: auto; left: 20px; }
.window.embedded { position: relative; right: auto; bottom: auto; width: 100%; height: 100%; max-height: none; }
.header { position: relative; min-height: 112px; padding: 24px 24px 34px; color: #fff; display: flex; align-items: flex-start; gap: 18px; overflow: hidden; flex: 0 0 auto; }
.header:after { content: ""; position: absolute; left: -8%; right: -8%; bottom: -24px; height: 44px; background: #fff; border-radius: 0 0 50% 50%; }
.brand-icon { position: relative; z-index: 1; width: 52px; height: 52px; border-radius: 50%; background: #08265b; color: #fff; display: flex; align-items: center; justify-content: center; flex: 0 0 auto; overflow: hidden; font-size: 11px; font-weight: 700; text-align: center; line-height: 1.1; }
.brand-icon img { width: 100%; height: 100%; object-fit: cover; display: block; }
.header-text { position: relative; z-index: 1; min-width: 0; flex: 1; padding-top: 1px; }
.header .title { font-size: 22px; line-height: 1.15; font-weight: 700; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.header .subtitle { margin-top: 5px; font-size: 14px; line-height: 1.3; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; opacity: 0.96; }
.close-btn { position: relative; z-index: 2; background: rgba(255,255,255,0.16); border: none; border-radius: 50%; color: #fff; cursor: pointer; font-size: 16px; width: 28px; height: 28px; line-height: 28px; flex: 0 0 auto; }
.messages { flex: 1; overflow-y: auto; padding: 18px; display: flex; flex-direction: column; gap: 10px; background: #fff; }
.messages.has-intro { padding-top: 160px; }
.brand-label { color: #6b7280; font-size: 13px; margin: -2px 0 0 8px; }
.intro-card { width: 100%; padding: 18px 20px 20px; border-radius: 14px; background: #f7f7f8; color: #111827; }
.intro-card h3 { font-size: 16px; line-height: 1.35; font-weight: 700; margin-bottom: 20px; }
.intro-card p { font-size: 14px; line-height: 1.45; margin-bottom: 10px; }
.intro-card input { width: 100%; height: 38px; border: 1px solid v-bind; border-radius: 8px; padding: 0 14px; font-size: 14px; outline: none; background: #fff; color: #111827; }
.intro-card .intro-submit { margin-top: 12px; display: flex; justify-content: flex-end; }
.intro-card button { border: none; background: transparent; color: #22c55e; cursor: pointer; font-size: 13px; font-weight: 700; padding: 4px 0; }
.msg { max-width: 85%; padding: 8px 12px; border-radius: 10px; font-size: 14px; line-height: 1.5; word-break: break-word; }
.msg.agent { align-self: flex-start; background: #f0f2f5; border-bottom-left-radius: 4px; }
.msg.visitor { align-self: flex-end; color: #fff; border-bottom-right-radius: 4px; }
.msg.activity { align-self: center; background: #f5f5f5; color: #909399; font-size: 12px; border-radius: 6px; max-width: 90%; text-align: center; }
.msg .time { font-size: 11px; opacity: 0.6; margin-top: 2px; }
.msg-img { max-width: 180px; border-radius: 8px; cursor: pointer; }
.input-area { min-height: 70px; padding: 11px 20px; border-top: 1px solid #e5e7eb; display: flex; gap: 8px; align-items: center; background: #fff; flex: 0 0 auto; }
.input-area textarea { flex: 1; border: none; padding: 4px 0; font-size: 13px; resize: none; outline: none; font-family: inherit; min-height: 34px; max-height: 80px; color: #111827; }
.input-area textarea::placeholder { color: #c4c7cc; }
.send-btn { border: none; border-radius: 6px; padding: 8px 12px; cursor: pointer; color: #fff; font-size: 13px; white-space: nowrap; }
.form { padding: 16px; }
.form input, .form textarea { width: 100%; border: 1px solid #dcdfe6; border-radius: 6px; padding: 8px 10px; font-size: 13px; margin-bottom: 8px; outline: none; font-family: inherit; }
.form input:focus, .form textarea:focus { border-color: v-bind; }
.form textarea { resize: none; height: 60px; }
@media (max-width: 420px) {
  .window { left: 0; right: 0; bottom: 0; width: 100%; height: 100%; max-height: none; border-radius: 0; }
  .window.left { left: 0; }
}
`

export function mount(container?: HTMLElement) {
  const s = getStore()
  if (s.mode === 'embedded' && container) {
    root = container
  } else {
    if (root) return
    root = document.body
  }

  const host = document.createElement('div')
  host.id = 'chatagent-widget-root'
  shadow = host.attachShadow({ mode: 'open' })

  const render = () => {
    if (!shadow) return
    const current = getStore()
    const style = document.createElement('style')
    style.textContent = css.replace(/v-bind/g, current.brandColor)
    shadow.innerHTML = ''
    shadow.append(style)
    if ((current.mode === 'bubble' || current.mode === 'bubble-left') && !current.open) renderBubble()
    else renderWindow()
  }

  render()
  root.appendChild(host)
  unsub = onStoreChange(render)
}

function renderBubble() {
  if (!shadow) return
  const s = getStore()
  const btn = h('button', { class: `bubble-btn ${s.mode === 'bubble-left' ? 'left' : ''}`, style: `background:${s.brandColor}` }, 'Chat')
  btn.onclick = () => updateStore({ open: true, formOpen: !s.pubsubToken || s.preChat })
  shadow.append(btn)
}

function renderWindow() {
  if (!shadow) return
  const s = getStore()
  const pos = s.mode === 'bubble-left' ? ' left' : ''
  const modeClass = s.mode === 'embedded' ? ' embedded' : ''
  const title = s.welcomeTitle || 'Chat'
  const subtitle = stripHtml(s.welcomeMessage) || 'How can we help?'
  const win = h('div', { class: `window${pos}${modeClass}` })

  const header = h('div', { class: 'header', style: `background:${s.brandColor}` },
    renderIcon(title, s.icon),
    h('div', { class: 'header-text' },
      h('div', { class: 'title' }, title),
      h('div', { class: 'subtitle' }, subtitle),
    ),
  )
  if (s.mode === 'bubble' || s.mode === 'bubble-left') {
    const closeBtn = h('button', { class: 'close-btn', title: 'Close' }, 'x')
    closeBtn.onclick = () => updateStore({ open: false })
    header.append(closeBtn)
  }
  win.append(header)

  const showIntro = s.preChat && !s.introSubmitted && s.conversations.length === 0
  const msgContainer = h('div', { class: `messages${showIntro ? ' has-intro' : ''}` })
  if (showIntro) {
    msgContainer.append(h('div', { class: 'brand-label' }, title))
    msgContainer.append(renderIntroCard(title, s.welcomeMessage))
  }

  s.conversations.forEach(m => {
    const cls = m.messageType === 'activity' ? 'activity' : m.senderType === 'user' ? 'agent' : 'visitor'
    const bubble = h('div', { class: `msg ${cls}`, style: cls === 'visitor' ? `background:${s.brandColor}` : '' })
    if (m.fileUrl && m.contentType?.startsWith('image/')) {
      bubble.append(h('img', { src: m.fileUrl, class: 'msg-img' }))
    } else if (m.fileUrl) {
      bubble.append(h('a', { href: m.fileUrl, target: '_blank' }, m.fileName || 'Download'))
    }
    if (m.content) bubble.append(h('div', {}, m.content))
    bubble.append(h('div', { class: 'time' }, formatTime(m.createdAt)))
    msgContainer.append(bubble)
  })
  win.append(msgContainer)

  if (s.formOpen && !s.pubsubToken) {
    renderPreAuthForm(win)
  } else if (s.pubsubToken) {
    renderInput(win)
  }

  shadow.append(win)
  const msgDiv = win.querySelector('.messages')
  if (msgDiv) msgDiv.scrollTop = msgDiv.scrollHeight
}

function renderIcon(title: string, icon: string): HTMLElement {
  const wrap = h('div', { class: 'brand-icon' })
  if (isImageIcon(icon)) {
    wrap.append(h('img', { src: icon, alt: title }))
  } else {
    wrap.append(document.createTextNode(icon || initials(title)))
  }
  return wrap
}

function renderIntroCard(title: string, message: string): HTMLElement {
  const s = getStore()
  const card = h('div', { class: 'intro-card' },
    h('h3', {}, `Hi there and Welcome to ${title}!`),
    h('p', {}, stripHtml(message) || "Let's get acquainted, what's your name?"),
  )
  const nameInput = h('input', { type: 'text', placeholder: 'Enter your name', value: s.visitorName }) as HTMLInputElement
  const submit = h('button', { type: 'button' }, 'Submit >')
  submit.onclick = () => updateStore({ visitorName: nameInput.value.trim(), introSubmitted: true })
  nameInput.onkeydown = (e) => {
    if (e.key === 'Enter') submit.click()
  }
  card.append(nameInput, h('div', { class: 'intro-submit' }, submit))
  return card
}

function renderPreAuthForm(parent: HTMLElement) {
  const s = getStore()
  const form = h('div', { class: 'form' })
  const nameInput = h('input', { type: 'text', placeholder: 'Enter your name' }) as HTMLInputElement
  const emailInput = h('input', { type: 'email', placeholder: 'Email (optional)' }) as HTMLInputElement
  const msgInput = h('textarea', { placeholder: 'Enter details in the input field' }) as HTMLTextAreaElement
  const sendBtn = h('button', { class: 'send-btn', style: `background:${s.brandColor}` }, 'Send')

  sendBtn.onclick = async () => {
    const msg = msgInput.value.trim()
    if (!msg) return
    let fp = localStorage.getItem('chatagent_fp')
    if (!fp) { fp = crypto.randomUUID(); localStorage.setItem('chatagent_fp', fp) }
    const ok = await widgetAuth(fp)
    if (ok) {
      connectWS()
      updateStore({ formOpen: false, visitorName: nameInput.value.trim(), introSubmitted: true })
      sendTextMessage(msg, { name: nameInput.value.trim(), email: emailInput.value.trim() })
    }
  }
  form.append(nameInput, emailInput, msgInput, sendBtn)
  parent.append(form)
}

function renderInput(parent: HTMLElement) {
  const s = getStore()
  const area = h('div', { class: 'input-area' })
  inputEl = h('textarea', { placeholder: 'Enter details in the input field', rows: '1' }) as HTMLTextAreaElement
  inputEl.onkeydown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      const text = inputEl!.value.trim()
      if (text) sendTextMessage(text, { name: s.visitorName })
    }
  }
  const btn = h('button', { class: 'send-btn', style: `background:${s.brandColor}` }, 'Send')
  btn.onclick = () => {
    const text = inputEl!.value.trim()
    if (text) sendTextMessage(text, { name: s.visitorName })
  }
  area.append(inputEl, btn)
  parent.append(area)
}

async function sendTextMessage(text: string, visitor?: { name?: string; email?: string }) {
  addMessage({
    id: 'local-' + Date.now(),
    content: text,
    senderType: 'contact',
    messageType: 'incoming',
    contentType: 'text/html',
    fileUrl: '',
    fileName: '',
    createdAt: new Date().toISOString(),
    private: false,
  })
  if (inputEl) inputEl.value = ''

  const result = await apiSendMessage(text, visitor)
  if (result && result.conversationId) {
    updateStore({ conversationId: result.conversationId, introSubmitted: true })
    subscribe(result.conversationId)
  }
}

function formatTime(t: string): string {
  if (!t) return ''
  const d = new Date(t)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return 'just now'
  if (diff < 3600000) return Math.floor(diff / 60000) + 'm ago'
  return d.toLocaleTimeString()
}

function stripHtml(value: string): string {
  return (value || '').replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
}

function isImageIcon(value: string): boolean {
  return /^(https?:\/\/|data:image\/|blob:|\/)/.test(value || '')
}

function initials(value: string): string {
  const text = stripHtml(value).trim()
  if (!text) return 'CA'
  const parts = text.split(/\s+/).filter(Boolean)
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase()
  return text.slice(0, 2).toUpperCase()
}

export function unmount() {
  unsub?.()
  disconnectWS()
  if (root) {
    const host = root.querySelector('#chatagent-widget-root')
    if (host) root.removeChild(host)
  }
  root = null
  shadow = null
}
