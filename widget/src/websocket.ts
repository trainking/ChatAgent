import { getStore, updateStore, addMessage } from './store'

let socket: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectDelay = 1000
const subscribedChannels: Set<string> = new Set()
const pendingMessages: string[] = []

export function connectWS() {
  const s = getStore()
  if (!s.pubsubToken) return
  if (socket && (socket.readyState === WebSocket.CONNECTING || socket.readyState === WebSocket.OPEN)) return

  const base = new URL(s.baseUrl || location.origin)
  const proto = base.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${proto}//${base.host}/ws?token=${s.pubsubToken}&type=widget`
  socket = new WebSocket(url)

  socket.onopen = () => {
    updateStore({ connected: true })
    reconnectDelay = 1000

    // Resubscribe to all previously subscribed channels
    subscribedChannels.forEach((ch) => {
      socket!.send(JSON.stringify({ action: 'subscribe', channel: ch }))
    })

    // Flush pending messages
    pendingMessages.forEach((msg) => socket!.send(msg))
    pendingMessages.length = 0
  }

  socket.onmessage = (e) => {
    try {
      const msg = JSON.parse(e.data)
      handleEvent(msg.event, msg.data)
    } catch {}
  }

  socket.onclose = () => {
    updateStore({ connected: false })
    scheduleReconnect()
  }

  socket.onerror = () => {
    socket?.close()
  }
}

function scheduleReconnect() {
  if (reconnectTimer) return
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null
    reconnectDelay = Math.min(reconnectDelay * 2, 30000)
    connectWS()
  }, reconnectDelay)
}

function handleEvent(event: string, data: any) {
  switch (event) {
    case 'message.created':
      // Add agent replies and activity messages (not our own echoes)
      if (data.sender_type === 'user' || data.message_type === 'activity') {
        addMessage({
          id: data.id,
          content: data.content || '',
          senderType: data.sender_type || 'user',
          messageType: data.message_type || 'outgoing',
          contentType: data.content_type || 'text/html',
          fileUrl: data.file_url || '',
          fileName: data.file_name || '',
          createdAt: data.created_at || new Date().toISOString(),
          private: data.private || false,
        })
      }
      break
    case 'conversation.status_changed':
      updateStore({ formOpen: data.new_status !== 'resolved' })
      break
  }
}

function send(action: string, channel: string, data?: any) {
  const msg = JSON.stringify({ action, channel, data })
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(msg)
  } else {
    // Queue for when connection opens
    pendingMessages.push(msg)
  }
}

export function subscribe(convId: string) {
  const channel = `conversation:${convId}`
  subscribedChannels.add(channel)
  send('subscribe', channel)
}

export function disconnectWS() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  subscribedChannels.clear()
  pendingMessages.length = 0
  socket?.close()
  socket = null
}
