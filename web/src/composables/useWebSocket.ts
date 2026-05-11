import { ref, onUnmounted } from 'vue'

type EventHandler = (event: string, data: unknown) => void

export function useWebSocket() {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const reconnectAttempt = ref(0)
  let maxReconnectDelay = 30000
  let handlers: EventHandler[] = []
  let pendingSubscriptions: string[] = []

  function connect() {
    const token = localStorage.getItem('token')
    if (!token) return

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${location.host}/ws?token=${token}&type=agent`

    const socket = new WebSocket(url)
    ws.value = socket

    socket.onopen = () => {
      connected.value = true
      reconnectAttempt.value = 0
      pendingSubscriptions.forEach((ch) => {
        socket.send(JSON.stringify({ action: 'subscribe', channel: ch }))
      })
    }

    socket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        handlers.forEach((h) => h(msg.event, msg.data))
      } catch {}
    }

    socket.onclose = () => {
      connected.value = false
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempt.value), maxReconnectDelay)
      reconnectAttempt.value++
      setTimeout(connect, delay)
    }

    socket.onerror = () => {
      socket.close()
    }
  }

  function subscribe(channel: string) {
    pendingSubscriptions.push(channel)
    if (connected.value && ws.value) {
      ws.value.send(JSON.stringify({ action: 'subscribe', channel }))
    }
  }

  function unsubscribe(channel: string) {
    pendingSubscriptions = pendingSubscriptions.filter((c) => c !== channel)
    if (connected.value && ws.value) {
      ws.value.send(JSON.stringify({ action: 'unsubscribe', channel }))
    }
  }

  function sendTyping(convId: string) {
    if (connected.value && ws.value) {
      ws.value.send(JSON.stringify({ action: 'typing', channel: `conversation:${convId}` }))
    }
  }

  function onEvent(handler: EventHandler) {
    handlers.push(handler)
  }

  function offEvent(handler: EventHandler) {
    handlers = handlers.filter((h) => h !== handler)
  }

  function disconnect() {
    pendingSubscriptions = []
    handlers = []
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    connected.value = false
  }

  onUnmounted(() => disconnect())

  return { connected, connect, disconnect, subscribe, unsubscribe, sendTyping, onEvent, offEvent }
}
