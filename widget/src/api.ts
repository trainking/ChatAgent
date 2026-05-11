import { getStore, updateStore } from './store'

export async function widgetAuth(fingerprint?: string): Promise<boolean> {
  const s = getStore()
  try {
    const resp = await fetch(s.baseUrl + '/api/v1/widget/auth', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ inbox_id: s.inboxId, fingerprint: fingerprint || '' }),
    })
    if (!resp.ok) return false
    const json = await resp.json()
    if (json.code !== 0) return false

    updateStore({
      pubsubToken: json.data.pubsub_token,
      contactId: json.data.contact_id,
      contactInboxId: json.data.contact_inbox_id,
      welcomeTitle: json.data.welcome_title || '',
      welcomeMessage: json.data.welcome_message || '',
      preChat: !!(json.data.welcome_title || json.data.welcome_message),
    })
    return true
  } catch {
    return false
  }
}

export async function sendMessage(content: string): Promise<{ conversationId?: string } | null> {
  const s = getStore()
  try {
    const resp = await fetch(s.baseUrl + '/api/v1/widget/messages', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pubsub_token: s.pubsubToken, content, content_type: 'text/html' }),
    })
    if (!resp.ok) return null
    const json = await resp.json()
    if (json.code !== 0) return null
    return json.data
  } catch {
    return null
  }
}

export async function sendCSAT(convId: string, score: number, feedback?: string): Promise<boolean> {
  const s = getStore()
  try {
    const resp = await fetch(`${s.baseUrl}/api/v1/conversations/${convId}/csat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ score, feedback }),
    })
    return resp.ok
  } catch {
    return false
  }
}
