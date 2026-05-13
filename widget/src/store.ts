export interface Store {
  inboxId: string
  mode: string
  baseUrl: string
  pubsubToken: string | null
  contactId: string | null
  contactInboxId: string | null
  icon: string
  welcomeTitle: string
  welcomeMessage: string
  brandColor: string
  preChat: boolean
  conversationId: string | null
  conversations: Message[]
  connected: boolean
  open: boolean
  formOpen: boolean
  visitorName: string
  introSubmitted: boolean
}

export interface Message {
  id: string
  content: string
  senderType: string
  messageType: string
  contentType: string
  fileUrl: string
  fileName: string
  createdAt: string
  private: boolean
}

let store: Store = {
  inboxId: '',
  mode: 'bubble',
  baseUrl: '',
  pubsubToken: null,
  contactId: null,
  contactInboxId: null,
  icon: '',
  welcomeTitle: '',
  welcomeMessage: '',
  brandColor: '#409EFF',
  preChat: false,
  conversationId: null,
  conversations: [],
  connected: false,
  open: false,
  formOpen: true,
  visitorName: '',
  introSubmitted: false,
}

const listeners: Set<() => void> = new Set()

export function getStore(): Store { return store }

export function updateStore(partial: Partial<Store>) {
  Object.assign(store, partial)
  listeners.forEach(fn => fn())
}

export function addMessage(msg: Message) {
  store.conversations.push(msg)
  listeners.forEach(fn => fn())
}

export function onStoreChange(fn: () => void) {
  listeners.add(fn)
  return () => listeners.delete(fn)
}
