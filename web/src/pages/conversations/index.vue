<template>
  <div class="conversations-page">
    <!-- Left Panel: Conversation List -->
    <div class="conv-list-panel">
      <div class="panel-header">
        <span class="panel-title">{{ $t('conversation.title') }}</span>
      </div>

      <!-- Filters -->
      <div class="panel-filters">
        <el-select v-model="filterInbox" :placeholder="$t('conversation.filterInbox')" size="small" clearable @change="fetchList">
          <el-option v-for="ib in inboxes" :key="ib.id" :label="ib.name" :value="ib.id" />
        </el-select>
        <el-select v-model="filterStatus" :placeholder="$t('conversation.filterStatus')" size="small" clearable @change="fetchList">
          <el-option :label="$t('conversation.statusOpen')" value="open" />
          <el-option :label="$t('conversation.statusPending')" value="pending" />
          <el-option :label="$t('conversation.statusResolved')" value="resolved" />
          <el-option :label="$t('conversation.statusSnoozed')" value="snoozed" />
        </el-select>
        <el-select v-model="filterAssignee" :placeholder="$t('conversation.filterAssignee')" size="small" clearable @change="fetchList">
          <el-option :label="$t('conversation.assigneeAll')" value="" />
          <el-option :label="$t('conversation.assigneeMe')" value="me" />
          <el-option :label="$t('conversation.assigneeUnassigned')" value="unassigned" />
        </el-select>
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
          <div class="conv-item-main">
            <div class="conv-item-header">
              <span class="conv-contact">{{ conv.contact_name || conv.contact_email || `#${conv.display_id}` }}</span>
              <span class="conv-time">{{ formatTime(conv.last_message_at) }}</span>
            </div>
            <div class="conv-item-preview">
              <span class="conv-preview-text">{{ getPreview(conv) }}</span>
              <el-badge v-if="conv.unread_count > 0" :value="conv.unread_count" class="conv-badge" />
            </div>
          </div>
          <div class="conv-item-meta">
            <el-tag :type="statusTagType(conv.status)" size="small" effect="dark">
              {{ statusLabel(conv.status) }}
            </el-tag>
            <el-tag v-if="conv.priority !== 'medium'" :type="priorityTagType(conv.priority)" size="small">
              {{ priorityLabel(conv.priority) }}
            </el-tag>
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
            <div v-if="msg.content" class="msg-content" v-html="msg.content" />
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
        <el-checkbox v-model="isPrivateNote" size="small" class="private-toggle">
          {{ $t('conversation.privateNote') }}
        </el-checkbox>
        <div class="input-row">
          <el-upload
            class="file-upload-btn"
            :show-file-list="false"
            :http-request="handleUpload"
            accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.txt,.csv"
          >
            <el-button circle size="small"><el-icon><Link /></el-icon></el-button>
          </el-upload>
          <el-input
            v-model="inputText"
            :placeholder="$t('conversation.inputPlaceholder')"
            @keyup.enter.exact="sendText"
            class="msg-input"
          />
          <el-button type="primary" size="small" @click="sendText" :loading="sending">
            {{ $t('conversation.send') }}
          </el-button>
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
const filterStatus = ref('')
const filterAssignee = ref('')
const inboxes = ref<{ id: string; name: string }[]>([])

// List state
const convs = ref<ConversationItem[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

// Detail state
const selectedId = ref<string | null>(null)
const selectedConv = ref<ConversationItem | null>(null)
const inputText = ref('')
const sending = ref(false)
const isPrivateNote = ref(false)
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
  const text = conv.last_message?.content || ''
  if (text.length > 50) return text.substring(0, 50) + '...'
  return text
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
  if (!inputText.value.trim() || !selectedId.value) return
  sending.value = true
  try {
    await sendMessage(selectedId.value, {
      content: inputText.value,
      content_type: 'text/html',
      private: isPrivateNote.value,
    })
    inputText.value = ''
  } catch { ElMessage.error(t('common.error')) } finally { sending.value = false }
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
      private: isPrivateNote.value,
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
  height: calc(100vh - 80px);
  gap: 1px;
  background: #dcdfe6;
}

.conv-list-panel {
  width: 340px;
  min-width: 280px;
  background: #fff;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  padding: 12px 16px;
  border-bottom: 1px solid #ebeef5;
  .panel-title { font-size: 16px; font-weight: 600; }
}

.panel-filters {
  padding: 8px 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  border-bottom: 1px solid #ebeef5;
  .el-select { width: calc(50% - 3px); }
}

.conv-list {
  flex: 1;
  overflow-y: auto;
}

.conv-item {
  padding: 10px 16px;
  border-bottom: 1px solid #f2f3f5;
  cursor: pointer;
  &:hover { background: #f5f7fa; }
  &.active { background: #ecf5ff; border-left: 3px solid #409eff; padding-left: 13px; }
  &.unread { font-weight: 600; }
}

.conv-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
  .conv-contact { font-size: 14px; }
  .conv-time { font-size: 12px; color: #909399; }
}

.conv-item-preview {
  display: flex;
  justify-content: space-between;
  align-items: center;
  .conv-preview-text { font-size: 12px; color: #909399; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 220px; }
}

.conv-item-meta { display: flex; gap: 4px; margin-top: 4px; }

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
  padding: 10px 16px;
  border-top: 1px solid #ebeef5;
}
.input-row { display: flex; gap: 8px; align-items: center; }
.msg-input { flex: 1; }
.private-toggle { margin-bottom: 6px; }

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
