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
.bubble-btn { position: fixed; right: 20px; bottom: 20px; width: 56px; height: 56px; border-radius: 50%; border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; font-size: 24px; color: #fff; box-shadow: 0 4px 12px rgba(0,0,0,0.2); z-index: 9999; transition: transform 0.2s; }
.bubble-btn:hover { transform: scale(1.1); }
.bubble-btn.left { right: auto; left: 20px; }
.badge { position: absolute; top: -4px; right: -4px; min-width: 20px; height: 20px; border-radius: 10px; background: #f56c6c; color: #fff; font-size: 12px; line-height: 20px; text-align: center; }
.window { position: fixed; right: 20px; bottom: 90px; width: 360px; max-height: 520px; border-radius: 12px; overflow: hidden; display: flex; flex-direction: column; box-shadow: 0 8px 32px rgba(0,0,0,0.15); z-index: 9998; }
.window.left { right: auto; left: 20px; }
.header { padding: 14px 16px; color: #fff; display: flex; align-items: center; gap: 8px; }
.header .status { width: 8px; height: 8px; border-radius: 50%; background: #67c23a; flex-shrink: 0; }
.header .title { font-size: 15px; font-weight: 600; flex: 1; }
.close-btn { background: none; border: none; color: #fff; cursor: pointer; font-size: 20px; padding: 4px; }
.welcome { padding: 16px; font-size: 14px; color: #606266; overflow-y: auto; }
.messages { flex: 1; overflow-y: auto; padding: 12px; display: flex; flex-direction: column; gap: 8px; }
.msg { max-width: 85%; padding: 8px 12px; border-radius: 10px; font-size: 14px; line-height: 1.5; word-break: break-word; }
.msg.agent { align-self: flex-start; background: #f0f2f5; border-bottom-left-radius: 4px; }
.msg.visitor { align-self: flex-end; color: #fff; border-bottom-right-radius: 4px; }
.msg.activity { align-self: center; background: #f5f5f5; color: #909399; font-size: 12px; border-radius: 6px; max-width: 90%; text-align: center; }
.msg .time { font-size: 11px; opacity: 0.6; margin-top: 2px; }
.msg-img { max-width: 180px; border-radius: 8px; cursor: pointer; }
.input-area { padding: 10px 12px; border-top: 1px solid #ebeef5; display: flex; gap: 6px; align-items: flex-end; }
.input-area textarea { flex: 1; border: 1px solid #dcdfe6; border-radius: 6px; padding: 8px 10px; font-size: 13px; resize: none; outline: none; font-family: inherit; min-height: 36px; max-height: 80px; }
.input-area textarea:focus { border-color: v-bind; }
.send-btn { border: none; border-radius: 6px; padding: 8px 16px; cursor: pointer; color: #fff; font-size: 13px; white-space: nowrap; }
.form { padding: 16px; }
.form input, .form textarea { width: 100%; border: 1px solid #dcdfe6; border-radius: 6px; padding: 8px 10px; font-size: 13px; margin-bottom: 8px; outline: none; font-family: inherit; }
.form input:focus, .form textarea:focus { border-color: v-bind; }
.form textarea { resize: none; height: 60px; }
.csat { padding: 12px; text-align: center; }
.csat .emojis { display: flex; justify-content: center; gap: 12px; margin-bottom: 8px; }
.csat .emoji-btn { font-size: 28px; background: none; border: none; cursor: pointer; transition: transform 0.2s; padding: 4px; }
.csat .emoji-btn:hover { transform: scale(1.3); }
.csat .thanks { font-size: 24px; }
`
// We use CSS custom property for brand color — set via JS

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

  const style = document.createElement('style')
  style.textContent = css.replace(/v-bind/g, s.brandColor)
  shadow.append(style)

  if (s.mode === 'bubble' || s.mode === 'bubble-left') {
    renderBubble()
  } else {
    renderWindow()
  }

  root.appendChild(host)

  unsub = onStoreChange(() => {
    if (!shadow) return
    const existing = shadow.querySelector('.window')
    if (s.mode === 'bubble' || s.mode === 'bubble-left') {
      if (!existing && !s.open) return
      shadow.innerHTML = ''
      shadow.append(style.cloneNode(true))
      if (s.open) renderWindow()
      else renderBubble()
    } else {
      shadow.innerHTML = ''
      shadow.append(style.cloneNode(true))
      renderWindow()
    }
  })
}

function renderBubble() {
  if (!shadow) return
  const s = getStore()
  const btn = h('button', { class: `bubble-btn ${s.mode === 'bubble-left' ? 'left' : ''}`, style: `background:${s.brandColor}` },
    '💬'
  )
  btn.onclick = () => {
    updateStore({ open: true, formOpen: !s.pubsubToken || s.preChat })
    shadow!.innerHTML = ''
    renderWindow()
  }
  shadow.append(btn)
}

function renderWindow() {
  if (!shadow) return
  const s = getStore()
  const pos = s.mode === 'bubble-left' ? ' left' : ''

  const win = h('div', { class: `window${pos}`, style: 'background:#fff' })

  // Header
  const header = h('div', { class: 'header', style: `background:${s.brandColor}` },
    h('span', { class: 'status' }),
    h('span', { class: 'title' }, s.welcomeTitle || 'Chat'),
  )
  if (s.mode === 'bubble' || s.mode === 'bubble-left') {
    const closeBtn = h('button', { class: 'close-btn' }, '×')
    closeBtn.onclick = () => {
      updateStore({ open: false })
    }
    header.append(closeBtn)
  }
  win.append(header)

  // Messages area
  const msgContainer = h('div', { class: 'messages' })
  // Always show welcome message as first bubble if configured
  if (s.welcomeMessage) {
    msgContainer.append(
      h('div', { class: 'msg agent', style: 'background:#f0f2f5' },
        h('div', {}, s.welcomeMessage),
        h('div', { class: 'time' }, '')
      )
    )
  }
  // Show real messages
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

  // Pre-chat form or input
  if (s.formOpen && !s.pubsubToken) {
    const form = h('div', { class: 'form' })
    const nameInput = h('input', { type: 'text', placeholder: 'Name (optional)' }) as HTMLInputElement
    const emailInput = h('input', { type: 'email', placeholder: 'Email (optional)' }) as HTMLInputElement
    const msgInput = h('textarea', { placeholder: 'Message...' }) as HTMLTextAreaElement
    const sendBtn = h('button', { class: 'send-btn', style: `background:${s.brandColor}` }, 'Send')

    sendBtn.onclick = async () => {
      const msg = msgInput.value.trim()
      if (!msg) return
      let fp = localStorage.getItem('chatagent_fp')
      if (!fp) { fp = crypto.randomUUID(); localStorage.setItem('chatagent_fp', fp) }
      const ok = await widgetAuth(fp)
      if (ok) {
        connectWS()
        updateStore({ formOpen: false })
        form.remove()
        renderInput(win)
        sendTextMessage(msg, { name: nameInput.value.trim(), email: emailInput.value.trim() })
      }
    }
    form.append(nameInput, emailInput, msgInput, sendBtn)
    win.append(form)
  } else if (s.pubsubToken) {
    renderInput(win)
  }

  shadow.append(win)

  // Auto-scroll
  const msgDiv = win.querySelector('.messages')
  if (msgDiv) msgDiv.scrollTop = msgDiv.scrollHeight
}

function renderMessages(parent: HTMLElement) {
  const s = getStore()
  const msgDiv = h('div', { class: 'messages' })

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
    msgDiv.append(bubble)
  })

  parent.append(msgDiv)
}

function renderInput(parent: HTMLElement) {
  const s = getStore()
  const area = h('div', { class: 'input-area' })
  inputEl = h('textarea', { placeholder: 'Type a message...', rows: '1' }) as HTMLTextAreaElement
  inputEl.onkeydown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      const text = inputEl!.value.trim()
      if (text) sendTextMessage(text)
    }
  }
  const btn = h('button', { class: 'send-btn', style: `background:${s.brandColor}` }, 'Send')
  btn.onclick = () => {
    const text = inputEl!.value.trim()
    if (text) sendTextMessage(text)
  }
  area.append(inputEl, btn)
  parent.append(area)
}

async function sendTextMessage(text: string, visitor?: { name?: string; email?: string }) {
  // Optimistic local echo
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

  // Refresh UI
  if (shadow) {
    const msgs = shadow.querySelector('.messages')
    if (msgs) msgs.scrollTop = msgs.scrollHeight
  }

  // Send to server
  const result = await apiSendMessage(text, visitor)
  if (result && result.conversationId) {
    updateStore({ conversationId: result.conversationId })
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
