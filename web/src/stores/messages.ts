import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface MessageItem {
  id: string
  conversation_id: string
  sender_type: string
  sender_id: string
  content: string
  message_type: string
  content_type: string
  file_url: string
  file_name: string
  file_size: number
  status: string
  private: boolean
  metadata: string
  created_at: string
}

export const useMessagesStore = defineStore('messages', () => {
  const messages = ref<Record<string, MessageItem[]>>({})
  const typingUsers = ref<Record<string, string[]>>({})

  function getMessages(convId: string): MessageItem[] {
    return messages.value[convId] || []
  }

  function setMessages(convId: string, msgs: MessageItem[]) {
    messages.value[convId] = msgs
  }

  function appendMessage(convId: string, msg: MessageItem) {
    if (!messages.value[convId]) {
      messages.value[convId] = []
    }
    const existing = messages.value[convId].find((m) => m.id === msg.id)
    if (!existing) {
      messages.value[convId].push(msg)
    }
  }

  function updateMessage(convId: string, msgId: string, updates: Partial<MessageItem>) {
    const msgs = messages.value[convId]
    if (msgs) {
      const idx = msgs.findIndex((m) => m.id === msgId)
      if (idx >= 0) {
        msgs[idx] = { ...msgs[idx], ...updates }
      }
    }
  }

  function setTyping(convId: string, userName: string, typing: boolean) {
    if (!typingUsers.value[convId]) {
      typingUsers.value[convId] = []
    }
    if (typing) {
      if (!typingUsers.value[convId].includes(userName)) {
        typingUsers.value[convId].push(userName)
      }
    } else {
      typingUsers.value[convId] = typingUsers.value[convId].filter((n) => n !== userName)
    }
  }

  function getTyping(convId: string): string[] {
    return typingUsers.value[convId] || []
  }

  function prependMessages(convId: string, msgs: MessageItem[]) {
    if (!messages.value[convId]) {
      messages.value[convId] = []
    }
    messages.value[convId] = [...msgs, ...messages.value[convId]]
  }

  return {
    messages,
    getMessages,
    setMessages,
    appendMessage,
    updateMessage,
    setTyping,
    getTyping,
    prependMessages,
  }
})
