import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface ConversationItem {
  id: string
  display_id: number
  inbox_id: string
  contact_id: string
  assignee_id: string | null
  status: string
  priority: string
  subject: string
  unread_count: number
  last_message_at: string
  waiting_since: string | null
  first_reply_at: string | null
  resolved_at: string | null
  snoozed_until: string | null
  created_at: string
  updated_at: string
  contact_name: string
  contact_email: string
  last_message: {
    id: string
    content: string
    sender_type: string
    message_type: string
    private: boolean
  } | null
}

export const useConversationsStore = defineStore('conversations', () => {
  const list = ref<ConversationItem[]>([])
  const selectedId = ref<string | null>(null)
  const unreadCount = ref(0)

  function setList(newList: ConversationItem[]) {
    list.value = newList
  }

  function addOrUpdate(conv: ConversationItem) {
    const idx = list.value.findIndex((c) => c.id === conv.id)
    if (idx >= 0) {
      list.value[idx] = { ...list.value[idx], ...conv }
    } else {
      list.value.unshift(conv)
    }
  }

  function updateStatus(id: string, status: string) {
    const conv = list.value.find((c) => c.id === id)
    if (conv) conv.status = status
  }

  function updateAssignee(id: string, assigneeId: string | null) {
    const conv = list.value.find((c) => c.id === id)
    if (conv) conv.assignee_id = assigneeId
  }

  function updatePriority(id: string, priority: string) {
    const conv = list.value.find((c) => c.id === id)
    if (conv) conv.priority = priority
  }

  function incrementUnread(id: string) {
    const conv = list.value.find((c) => c.id === id)
    if (conv) conv.unread_count++
  }

  function resetUnread(id: string) {
    const conv = list.value.find((c) => c.id === id)
    if (conv) conv.unread_count = 0
  }

  function remove(id: string) {
    list.value = list.value.filter((c) => c.id !== id)
  }

  function select(id: string | null) {
    selectedId.value = id
  }

  return {
    list,
    selectedId,
    unreadCount,
    setList,
    addOrUpdate,
    updateStatus,
    updateAssignee,
    updatePriority,
    incrementUnread,
    resetUnread,
    remove,
    select,
  }
})
