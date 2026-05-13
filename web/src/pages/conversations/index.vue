<template>
  <div class="conversations-page">
    <!-- Left Panel: Conversation List -->
    <div class="conv-list-panel">
      <div class="panel-header">
        <div class="panel-heading">
          <span class="panel-title">{{ $t('conversation.title') }}</span>
          <el-select v-model="filterStatus" size="small" class="status-select" @change="fetchList">
            <el-option :label="$t('conversation.statusOpen')" value="open" />
            <el-option :label="$t('conversation.statusPending')" value="pending" />
            <el-option :label="$t('conversation.statusResolved')" value="resolved" />
            <el-option :label="$t('conversation.statusSnoozed')" value="snoozed" />
          </el-select>
        </div>
        <div class="panel-actions">
          <el-select
            v-model="filterInbox"
            :placeholder="$t('conversation.filterInbox')"
            size="small"
            clearable
            class="inbox-filter"
            @change="fetchList"
          >
            <template #prefix>
              <el-icon><Filter /></el-icon>
            </template>
            <el-option v-for="ib in inboxes" :key="ib.id" :label="ib.name" :value="ib.id" />
          </el-select>
          <el-button circle size="small" class="icon-btn" @click="toggleSort">
            <el-icon><Sort /></el-icon>
          </el-button>
          <el-button circle size="small" class="icon-btn">
            <el-icon><Right /></el-icon>
          </el-button>
        </div>
      </div>

      <!-- Filters -->
      <div class="assignment-tabs">
        <button
          v-for="tab in assignmentTabs"
          :key="tab.value"
          type="button"
          class="assignment-tab"
          :class="{ active: filterAssignee === tab.value }"
          @click="setAssigneeFilter(tab.value)"
        >
          <span>{{ tab.label }}</span>
          <span class="tab-count">{{ tab.count }}</span>
        </button>
      </div>

      <!-- List -->
      <div class="conv-list" v-loading="loading">
        <div
          v-if="convs.length === 0 && !loading"
          class="empty-state"
        >
          {{ $t('common.noData') }}
        </div>
        <div
          v-for="conv in convs"
          :key="conv.id"
          class="conv-item"
          :class="{ active: selectedId === conv.id, unread: conv.unread_count > 0 }"
          @click="selectConversation(conv)"
        >
          <div class="conv-avatar">{{ avatarText(conv) }}</div>
          <div class="conv-item-main">
            <div class="conv-inbox">
              <el-icon><ChatDotRound /></el-icon>
              <span>{{ inboxName(conv.inbox_id) }}</span>
            </div>
            <div class="conv-item-header">
              <span class="conv-contact">{{ conversationTitle(conv) }}</span>
            </div>
            <div class="conv-item-preview">
              <el-icon v-if="isPictureMessage(conv)" class="preview-icon"><Picture /></el-icon>
              <span class="conv-preview-text">{{ getPreview(conv) }}</span>
            </div>
          </div>
          <div class="conv-side-meta">
            <span class="conv-time">{{ formatTime(conv.last_message_at) }}</span>
            <span class="assignee-state" :class="assignmentClass(conv)">
              {{ assignmentLabel(conv) }}
            </span>
            <el-badge v-if="conv.unread_count > 0" :value="conv.unread_count" class="conv-badge" />
          </div>
        </div>
      </div>

      <div class="panel-footer" v-if="total > pageSize">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          small
          layout="prev, next"
          @current-change="fetchList"
        />
      </div>
    </div>

    <!-- Right Panel: Conversation Detail -->
    <div class="conv-detail-panel" v-if="selectedId">
      <!-- Detail Header -->
      <div class="detail-header">
        <div class="detail-header-left">
          <span class="detail-title">{{ selectedConv?.contact_name || selectedConv?.subject || `#${selectedConv?.display_id}` }}</span>
          <span v-if="selectedConv?.contact_email" class="detail-subtitle">{{ selectedConv?.contact_email }}</span>
        </div>
        <div class="detail-header-right">
          <el-select v-model="editStatus" size="small" @change="handleStatusChange">
            <el-option :label="$t('conversation.statusOpen')" value="open" />
            <el-option :label="$t('conversation.statusPending')" value="pending" />
            <el-option :label="$t('conversation.statusResolved')" value="resolved" />
            <el-option :label="$t('conversation.statusSnoozed')" value="snoozed" />
          </el-select>
          <el-select v-model="editPriority" size="small" @change="handlePriorityChange">
            <el-option :label="$t('conversation.priorityLow')" value="low" />
            <el-option :label="$t('conversation.priorityMedium')" value="medium" />
            <el-option :label="$t('conversation.priorityHigh')" value="high" />
            <el-option :label="$t('conversation.priorityUrgent')" value="urgent" />
          </el-select>
          <el-button size="small" @click="showSidebar = !showSidebar">
            <el-icon><InfoFilled /></el-icon>
          </el-button>
        </div>
      </div>

      <!-- Messages -->
      <div class="messages-area" ref="msgContainerRef" @scroll="handleScroll">
        <div v-if="hasMore" class="load-more">
          <el-button size="small" @click="loadMoreMessages" :loading="loadingMore">
            {{ $t('conversation.loadMore') }}
          </el-button>
        </div>
        <div
          v-for="msg in currentMessages"
          :key="msg.id"
          class="msg-row"
          :class="{
            'msg-contact': msg.sender_type === 'contact' && !msg.private,
            'msg-agent': msg.sender_type === 'user' && !msg.private,
            'msg-private': msg.private,
            'msg-activity': msg.message_type === 'activity',
          }"
        >
          <div class="msg-bubble" :class="{ 'msg-private-bg': msg.private }">
            <div v-if="msg.file_url" class="msg-file">
              <a :href="msg.file_url" target="_blank" class="file-link">
                <span v-if="msg.content_type?.startsWith('image/')">
                  <img :src="msg.file_url" class="msg-image" />
                </span>
                <span v-else>
                  {{ msg.file_name || $t('conversation.downloadFile') }}
                </span>
              </a>
            </div>
            <div v-if="msg.content" class="msg-content" v-html="sanitizeMessage(msg.content)" />
            <div class="msg-time">{{ formatTime(msg.created_at) }}</div>
          </div>
          <div v-if="msg.sender_type === 'user'" class="msg-status">
            <span v-if="msg.status === 'sent'">✓</span>
            <span v-else-if="msg.status === 'read'" class="read">✓✓</span>
            <span v-else-if="msg.status === 'failed'" class="failed" @click="retryMessage(msg)">!</span>
          </div>
        </div>
        <div v-if="typingNames.length > 0" class="typing-indicator">
          {{ typingNames.join(', ') }} {{ $t('conversation.typing') }}
        </div>
      </div>

      <!-- Input -->
      <div class="input-area">
        <div class="composer">
          <el-input
            ref="inputRef"
            v-model="inputText"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 6 }"
            :placeholder="$t('conversation.inputPlaceholder')"
            resize="none"
            @keydown.ctrl.enter.prevent="sendText"
            class="msg-input"
          />
          <div class="composer-toolbar">
            <div class="composer-tools">
              <el-upload
                class="file-upload-btn"
                :show-file-list="false"
                :http-request="handleUpload"
                accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.txt,.csv"
              >
                <el-button circle text class="composer-icon-btn">
                  <el-icon><Link /></el-icon>
                </el-button>
              </el-upload>
              <el-popover
                placement="top-start"
                trigger="click"
                width="280"
                popper-class="emoji-popover"
              >
                <div class="emoji-grid">
                  <button
                    v-for="emoji in emojis"
                    :key="emoji"
                    type="button"
                    class="emoji-option"
                    @click="insertEmoji(emoji)"
                  >
                    {{ emoji }}
                  </button>
                </div>
                <template #reference>
                  <el-button circle text class="composer-icon-btn">
                    <el-icon><Sunny /></el-icon>
                  </el-button>
                </template>
              </el-popover>
            </div>
            <el-button
              circle
              class="send-icon-btn"
              :disabled="!canSend"
              :loading="sending"
              @click="sendText"
            >
              <el-icon><Top /></el-icon>
            </el-button>
          </div>
        </div>
      </div>

      <!-- Sidebar -->
      <div class="detail-sidebar" v-if="showSidebar">
        <h4>{{ $t('conversation.info') }}</h4>
        <el-descriptions :column="1" size="small">
          <el-descriptions-item :label="$t('conversation.displayId')">#{{ selectedConv?.display_id }}</el-descriptions-item>
          <el-descriptions-item :label="$t('conversation.createdAt')">{{ selectedConv?.created_at }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <!-- Empty State -->
    <div class="conv-detail-panel empty-detail" v-else>
      <el-empty :description="$t('conversation.selectHint')" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { useConversationsStore } from '@/stores/conversations'
import { useMessagesStore } from '@/stores/messages'
import { useWebSocket } from '@/composables/useWebSocket'
import { getConversations, getConversation, assignConversation, changeStatus, changePriority } from '@/api/conversation'
import { getMessages, sendMessage, retryMessage as apiRetryMessage } from '@/api/message'
import { getInboxes } from '@/api/inbox'
import { uploadFile } from '@/api/upload'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const convStore = useConversationsStore()
const msgStore = useMessagesStore()
const ws = useWebSocket()

// Filter state
const filterInbox = ref('')
const filterStatus = ref('open')
const filterAssignee = ref('')
const inboxes = ref<{ id: string; name: string }[]>([])

// List state
const convs = ref<ConversationItem[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const newestFirst = ref(true)

// Detail state
const selectedId = ref<string | null>(null)
const selectedConv = ref<ConversationItem | null>(null)
const inputText = ref('')
const inputRef = ref<any>(null)
const sending = ref(false)
const showSidebar = ref(false)
const msgContainerRef = ref<HTMLElement | null>(null)
const loadingMore = ref(false)
const hasMore = ref(true)
const editStatus = ref('')
const editPriority = ref('')

import type { ConversationItem } from '@/stores/conversations'

// Computed
const currentMessages = computed(() => selectedId.value ? msgStore.getMessages(selectedId.value) : [])
const typingNames = computed(() => selectedId.value ? msgStore.getTyping(selectedId.value) : [])
const inboxNameMap = computed(() => new Map(inboxes.value.map((ib) => [ib.id, ib.name])))
const canSend = computed(() => inputText.value.trim().length > 0 && !sending.value)
const assignmentTabs = computed(() => [
  { label: (t as any)('conversation.assigneeMe'), value: 'me', count: convs.value.filter((conv) => isMine(conv)).length },
  { label: (t as any)('conversation.assigneeUnassigned'), value: 'unassigned', count: convs.value.filter((conv) => !conv.assignee_id).length },
  { label: (t as any)('conversation.assigneeAll'), value: '', count: total.value || convs.value.length },
])
const emojis = [
  '😀', '😄', '😊', '😍', '😘', '😎', '🥳', '🤔',
  '😅', '😂', '🙂', '🙌', '👍', '👎', '👏', '🙏',
  '💪', '🔥', '✨', '🎉', '❤️', '💙', '✅', '⭐',
  '📌', '📎', '💬', '📷', '🚀', '☕', '🌟', '😢',
]

async function fetchInboxes() {
  try {
    const res = await getInboxes()
    inboxes.value = res.data || []
  } catch {}
}

async function fetchList() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (filterInbox.value) params.inbox_id = filterInbox.value
    if (filterStatus.value) params.status = filterStatus.value
    if (filterAssignee.value) params.assignee_id = filterAssignee.value
    const res = await getConversations(params)
    convs.value = res.data?.list || []
    total.value = res.data?.total || 0
    convStore.setList(convs.value)
  } catch {} finally { loading.value = false }
}

function getPreview(conv: ConversationItem) {
  if (conv.last_message?.private) return `[${(t as any)('conversation.privateNote')}]`
  if (isPictureMessage(conv) && !conv.last_message?.content) return (t as any)('conversation.pictureMessage')
  const text = stripHtml(conv.last_message?.content || '')
  if (text.length > 50) return text.substring(0, 50) + '...'
  return text || conv.subject || `#${conv.display_id}`
}

function stripHtml(value: string) {
  return value.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
}

function inboxName(inboxId: string) {
  return inboxNameMap.value.get(inboxId) || (t as any)('conversation.unknownInbox')
}

function conversationTitle(conv: ConversationItem) {
  return conv.contact_name || conv.contact_email || conv.subject || `#${conv.display_id}`
}

function avatarText(conv: ConversationItem) {
  const source = conv.contact_name || conv.contact_email || String(conv.display_id)
  const trimmed = stripHtml(source).trim()
  if (!trimmed) return '#'
  if (/^\d+$/.test(trimmed)) return trimmed.slice(-1)
  const parts = trimmed.split(/[\s._-]+/).filter(Boolean)
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase()
  return trimmed.slice(0, 1).toUpperCase()
}

function isMine(conv: ConversationItem) {
  return !!conv.assignee_id && conv.assignee_id === userStore.user?.id
}

function assignmentLabel(conv: ConversationItem) {
  if (!conv.assignee_id) return (t as any)('conversation.assigneeUnassigned')
  if (isMine(conv)) return (t as any)('conversation.assigneeMe')
  return conv.assignee_name || conv.assignee_email || (t as any)('conversation.assigneeAssigned')
}

function assignmentClass(conv: ConversationItem) {
  if (!conv.assignee_id) return 'unassigned'
  if (isMine(conv)) return 'mine'
  return 'assigned'
}

function isPictureMessage(conv: ConversationItem) {
  const msg = conv.last_message as any
  return !!msg?.file_url && String(msg.content_type || '').startsWith('image/')
}

function setAssigneeFilter(value: string) {
  filterAssignee.value = value
  page.value = 1
  fetchList()
}

function toggleSort() {
  newestFirst.value = !newestFirst.value
  convs.value = [...convs.value].reverse()
}

function formatTime(time: string) {
  if (!time) return ''
  const d = new Date(time)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60000) return (t as any)('conversation.justNow')
  if (diff < 3600000) return Math.floor(diff / 60000) + (t as any)('conversation.minutesAgo')
  if (diff < 86400000) return Math.floor(diff / 3600000) + (t as any)('conversation.hoursAgo')
  return d.toLocaleDateString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US')
}

function statusTagType(status: string) {
  const m: Record<string, string> = { open: 'danger', pending: 'warning', resolved: 'success', snoozed: 'info' }
  return m[status] || 'info'
}

function statusLabel(status: string) {
  const m: Record<string, string> = {
    open: (t as any)('conversation.statusOpen'),
    pending: (t as any)('conversation.statusPending'),
    resolved: (t as any)('conversation.statusResolved'),
    snoozed: (t as any)('conversation.statusSnoozed'),
  }
  return m[status] || status
}

function priorityTagType(p: string) {
  const m: Record<string, string> = { low: 'info', medium: '', high: 'warning', urgent: 'danger' }
  return m[p] || ''
}

function priorityLabel(p: string) {
  const m: Record<string, string> = {
    low: (t as any)('conversation.priorityLow'),
    medium: (t as any)('conversation.priorityMedium'),
    high: (t as any)('conversation.priorityHigh'),
    urgent: (t as any)('conversation.priorityUrgent'),
  }
  return m[p] || p
}

function sanitizeMessage(html: string) {
  const template = document.createElement('template')
  template.innerHTML = html
  template.content.querySelectorAll('script, iframe, object, embed, link, meta').forEach((node) => node.remove())
  template.content.querySelectorAll('*').forEach((node) => {
    for (const attr of Array.from(node.attributes)) {
      const name = attr.name.toLowerCase()
      const value = attr.value.trim().toLowerCase()
      if (name.startsWith('on') || value.startsWith('javascript:')) {
        node.removeAttribute(attr.name)
      }
    }
  })
  return template.innerHTML
}

async function selectConversation(conv: ConversationItem) {
  selectedId.value = conv.id
  selectedConv.value = conv
  editStatus.value = conv.status
  editPriority.value = conv.priority
  convStore.select(conv.id)
  router.replace(`/dashboard/conversations/${conv.id}`)

  // Reset unread
  if (conv.unread_count > 0) {
    convStore.resetUnread(conv.id)
  }

  // Load messages
  hasMore.value = true
  msgStore.setMessages(conv.id, [])
  await loadMessages()

  // Subscribe to conversation channel
  ws.subscribe(`conversation:${conv.id}`)

  await nextTick()
  scrollToBottom()
}

async function loadMessages(beforeId?: string) {
  if (!selectedId.value) return
  try {
    const params: Record<string, unknown> = { limit: 30 }
    if (beforeId) params.before_id = beforeId
    const res = await getMessages(selectedId.value, params)
    const msgs = res.data || []
    if (beforeId) {
      msgStore.prependMessages(selectedId.value, msgs)
    } else {
      msgStore.setMessages(selectedId.value, msgs)
    }
    if (msgs.length < 30) hasMore.value = false
  } catch {}
}

async function loadMoreMessages() {
  if (!selectedId.value) return
  const msgs = msgStore.getMessages(selectedId.value)
  if (msgs.length === 0) return
  loadingMore.value = true
  await loadMessages(msgs[0].id)
  loadingMore.value = false
}

function scrollToBottom() {
  if (msgContainerRef.value) {
    msgContainerRef.value.scrollTop = msgContainerRef.value.scrollHeight
  }
}

function handleScroll() {
  if (!msgContainerRef.value) return
  if (msgContainerRef.value.scrollTop === 0 && hasMore.value) {
    loadMoreMessages()
  }
}

async function sendText() {
  if (!canSend.value || !selectedId.value) return
  sending.value = true
  try {
    await sendMessage(selectedId.value, {
      content: inputText.value,
      content_type: 'text/html',
      private: false,
    })
    inputText.value = ''
  } catch { ElMessage.error(t('common.error')) } finally { sending.value = false }
}

function insertEmoji(emoji: string) {
  const textarea = inputRef.value?.textarea as HTMLTextAreaElement | undefined
  if (!textarea) {
    inputText.value += emoji
    return
  }
  const start = textarea.selectionStart ?? inputText.value.length
  const end = textarea.selectionEnd ?? inputText.value.length
  inputText.value = inputText.value.slice(0, start) + emoji + inputText.value.slice(end)
  nextTick(() => {
    textarea.focus()
    const cursor = start + emoji.length
    textarea.setSelectionRange(cursor, cursor)
  })
}

async function handleUpload(uploadReq: { file: File }) {
  try {
    const res = await uploadFile(uploadReq.file)
    const data = (res as any).data
    if (!selectedId.value) return
    await sendMessage(selectedId.value, {
      content: '',
      content_type: data.type,
      file_url: data.url,
      file_name: data.name,
      file_size: data.size,
      private: false,
    })
  } catch {}
}

async function retryMessage(msg: { id: string }) {
  if (!selectedId.value) return
  try {
    await apiRetryMessage(selectedId.value, msg.id)
  } catch {}
}

async function handleStatusChange() {
  if (!selectedId.value) return
  try {
    await changeStatus(selectedId.value, editStatus.value)
    convStore.updateStatus(selectedId.value, editStatus.value)
    ElMessage.success(t('common.success'))
  } catch {}
}

async function handlePriorityChange() {
  if (!selectedId.value) return
  try {
    await changePriority(selectedId.value, editPriority.value)
    convStore.updatePriority(selectedId.value, editPriority.value)
    ElMessage.success(t('common.success'))
  } catch {}
}

// WebSocket event handling
function handleWSEvent(event: string, data: any) {
  switch (event) {
    case 'message.created': {
      if (data.conversation_id && selectedId.value === data.conversation_id) {
        msgStore.appendMessage(data.conversation_id, data)
        nextTick(() => scrollToBottom())
      }
      // Update conversation in list
      if (data.conversation_id) {
        const conv = convs.value.find((c) => c.id === data.conversation_id)
        if (conv) {
          conv.last_message_at = data.created_at
          conv.last_message = { id: data.id, content: data.content, sender_type: data.sender_type, message_type: data.message_type, private: data.private }
        }
      }
      break
    }
    case 'conversation.status_changed':
      convStore.updateStatus(data.conversation_id, data.new_status)
      break
    case 'conversation.assignee_changed':
      convStore.updateAssignee(data.conversation_id, data.assignee_id)
      break
    case 'conversation.priority_changed':
      convStore.updatePriority(data.conversation_id, data.priority)
      break
  }
}

// Watch route param to select conversation
watch(() => route.params.id, async (id) => {
  if (id && typeof id === 'string' && id !== selectedId.value) {
    const conv = convs.value.find((c) => c.id === id)
    if (conv) {
      selectConversation(conv)
    } else {
      try {
        const res = await getConversation(id)
        const detail = res.data
        selectConversation({
          id: detail.id,
          display_id: detail.display_id,
          inbox_id: detail.inbox_id,
          contact_id: detail.contact_id,
          assignee_id: detail.assignee_id,
          status: detail.status,
          priority: detail.priority,
          subject: detail.subject,
          unread_count: detail.unread_count || 0,
          last_message_at: detail.last_message_at,
          waiting_since: detail.waiting_since,
          first_reply_at: detail.first_reply_at,
          resolved_at: detail.resolved_at,
          snoozed_until: detail.snoozed_until,
          contact_name: detail.contact?.name || '',
          contact_email: detail.contact?.email || '',
          assignee_name: detail.assignee?.name || '',
          assignee_email: detail.assignee?.email || '',
          created_at: detail.created_at || '',
          updated_at: detail.updated_at || '',
          last_message: null,
        })
      } catch {}
    }
  }
})

onMounted(async () => {
  await fetchInboxes()
  await fetchList()

  // Connect WebSocket
  ws.connect()
  ws.onEvent(handleWSEvent)

  // Subscribe to inbox channels
  inboxes.value.forEach((ib) => ws.subscribe(`inbox:${ib.id}`))

  // Check route for preselected conversation
  const convId = route.params.id
  if (convId && typeof convId === 'string') {
    const conv = convs.value.find((c) => c.id === convId)
    if (conv) selectConversation(conv)
  }
})
</script>

<style scoped lang="scss">
.conversations-page {
  display: flex;
  height: 100vh;
  background: #ffffff;
}

.conv-list-panel {
  width: 390px;
  min-width: 280px;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  padding: 12px 12px 8px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;

  .panel-title {
    font-size: 16px;
    font-weight: 700;
    color: #111827;
  }
}

.panel-heading,
.panel-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.status-select {
  width: 76px;
}

.inbox-filter {
  width: 34px;

  :deep(.el-input__wrapper) {
    width: 34px;
    height: 28px;
    padding: 0 8px;
    border-radius: 8px;
    background: #f3f4f6;
    box-shadow: none;
  }

  :deep(.el-input__inner),
  :deep(.el-select__caret) {
    display: none;
  }
}

.icon-btn {
  width: 28px;
  height: 28px;
  border: 0;
  background: #f3f4f6;
  color: #374151;
}

.assignment-tabs {
  padding: 6px 12px 8px;
  display: flex;
  align-items: center;
  gap: 18px;
  border-bottom: 1px solid #ebeef5;
}

.assignment-tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 26px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #4b5563;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;

  &.active {
    color: #2563eb;

    &::after {
      content: '';
      position: absolute;
      left: 0;
      right: 0;
      bottom: -9px;
      height: 2px;
      background: #2563eb;
      border-radius: 2px;
    }
  }
}

.tab-count {
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 11px;
  line-height: 18px;
  text-align: center;
}

.conv-list {
  flex: 1;
  overflow-y: auto;
}

.conv-item {
  position: relative;
  display: flex;
  gap: 10px;
  min-height: 86px;
  padding: 12px 12px;
  border-bottom: 1px solid #eef0f3;
  cursor: pointer;

  &:hover { background: #f7f8fa; }
  &.active { background: #f3f4f6; }
  &.unread .conv-contact { font-weight: 700; }
}

.conv-avatar {
  flex: 0 0 32px;
  width: 32px;
  height: 32px;
  margin-top: 22px;
  border-radius: 50%;
  background: #dfe7ff;
  color: #3b64d8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
}

.conv-item-main {
  min-width: 0;
  flex: 1;
}

.conv-inbox {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-bottom: 4px;
  color: #6b7280;
  font-size: 13px;

  .el-icon {
    font-size: 13px;
    color: #4b5563;
  }

  span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.conv-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;

  .conv-contact {
    min-width: 0;
    color: #111827;
    font-size: 14px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.conv-item-preview {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;

  .conv-preview-text {
    min-width: 0;
    color: #6b7280;
    font-size: 14px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.preview-icon {
  flex: 0 0 auto;
  color: #6b7280;
  font-size: 15px;
}

.conv-side-meta {
  flex: 0 0 72px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 7px;
  padding-top: 24px;
}

.conv-time {
  color: #6b7280;
  font-size: 11px;
  white-space: nowrap;
}

.assignee-state {
  max-width: 72px;
  color: #64748b;
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  &.mine { color: #2563eb; }
  &.unassigned { color: #b45309; }
}

.conv-badge {
  line-height: 1;
}

.panel-footer { padding: 8px; text-align: center; border-top: 1px solid #ebeef5; }

.empty-state { padding: 40px; text-align: center; color: #909399; }

// Right Panel
.conv-detail-panel {
  flex: 1;
  background: #fff;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.detail-header {
  padding: 10px 16px;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.detail-header-left {
  .detail-title { font-size: 15px; font-weight: 600; }
  .detail-subtitle { display: block; font-size: 12px; color: #909399; }
}

.detail-header-right { display: flex; gap: 8px; }

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.load-more { text-align: center; padding: 8px; }

.msg-row {
  margin-bottom: 12px;
  display: flex;
  align-items: flex-end;
  &.msg-contact { justify-content: flex-start; }
  &.msg-agent { justify-content: flex-end; }
  &.msg-private { justify-content: flex-end; }
  &.msg-activity { justify-content: center; }
}

.msg-bubble {
  max-width: 70%;
  padding: 8px 14px;
  border-radius: 12px;
  background: #f0f2f5;
  .msg-contact & { background: #f0f2f5; border-bottom-left-radius: 4px; }
  .msg-agent & { background: #409eff; color: #fff; border-bottom-right-radius: 4px; }
  .msg-private-bg { background: #fdf6ec; border: 1px dashed #e6a23c; }
  .msg-activity & { background: #f5f5f5; color: #909399; font-size: 12px; }
}

.msg-content { font-size: 14px; line-height: 1.5; word-break: break-word; }
.msg-image { max-width: 200px; border-radius: 8px; cursor: pointer; }
.msg-time { font-size: 11px; margin-top: 2px; opacity: 0.6; }
.msg-status { font-size: 11px; margin-left: 4px; color: #909399; .read { color: #409eff; } .failed { color: #f56c6c; cursor: pointer; } }

.file-link { color: inherit; text-decoration: underline; }

.typing-indicator { padding: 4px 16px; font-size: 12px; color: #909399; font-style: italic; }

.input-area {
  padding: 12px 16px;
  border-top: 1px solid #ebeef5;
  background: #fff;
}

.composer {
  border: 1px solid #dcdfe6;
  border-radius: 14px;
  background: #fff;
  padding: 8px;
  transition: border-color 0.2s, box-shadow 0.2s;

  &:focus-within {
    border-color: #409eff;
    box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.12);
  }
}

.msg-input {
  :deep(.el-textarea__inner) {
    min-height: 46px !important;
    padding: 2px 4px 8px;
    border: 0;
    box-shadow: none;
    color: #111827;
    font-size: 14px;
    line-height: 1.5;
  }
}

.composer-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-height: 32px;
}

.composer-tools {
  display: flex;
  align-items: center;
  gap: 2px;
}

.composer-icon-btn {
  width: 28px;
  height: 28px;
  color: #6b7280;

  &:hover {
    background: #f3f4f6;
    color: #1f2937;
  }
}

.send-icon-btn {
  width: 30px;
  height: 30px;
  border: 0;
  background: #0d6efd;
  color: #fff;

  &:hover,
  &:focus {
    background: #0b5ed7;
    color: #fff;
  }

  &.is-disabled,
  &.is-disabled:hover,
  &.is-disabled:focus {
    background: #e5e7eb;
    color: #9ca3af;
    cursor: not-allowed;
  }
}

.emoji-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 4px;
}

.emoji-option {
  width: 28px;
  height: 28px;
  border: 0;
  border-radius: 6px;
  background: transparent;
  font-size: 18px;
  line-height: 28px;
  cursor: pointer;

  &:hover {
    background: #f3f4f6;
  }
}

.detail-sidebar {
  position: absolute;
  right: 0;
  top: 0;
  width: 260px;
  height: 100%;
  background: #fff;
  border-left: 1px solid #ebeef5;
  padding: 16px;
  overflow-y: auto;
  z-index: 5;
}

.empty-detail {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
