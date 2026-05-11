<template>
  <div v-loading="loading">
    <el-page-header @back="$router.push('/dashboard/contacts')">
      <template #content>
        <span>{{ contact?.name || $t('contact.detail') }}</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 16px" v-if="contact">
      <template #header>{{ $t('contact.info') }}</template>
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="$t('contact.name')">{{ contact.name }}</el-descriptions-item>
        <el-descriptions-item :label="$t('contact.email')">{{ contact.email || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('contact.phone')">{{ contact.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('contact.type')">
          <el-tag :type="contact.contact_type === 'visitor' ? 'info' : contact.contact_type === 'lead' ? 'warning' : 'success'" size="small">
            {{ typeLabel(contact.contact_type) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('contact.country')">{{ contact.country || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('contact.city')">{{ contact.city || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('contact.browser')">{{ contact.browser || '-' }}</el-descriptions-item>
        <el-descriptions-item :label="$t('contact.os')">{{ contact.os || '-' }}</el-descriptions-item>
      </el-descriptions>

      <div style="margin-top: 16px; display: flex; gap: 8px">
        <el-button type="primary" size="small" @click="showEdit = true">{{ $t('common.edit') }}</el-button>
        <el-button v-if="!contact.blocked" type="warning" size="small" @click="handleBlock">{{ $t('contact.block') }}</el-button>
        <el-button v-else type="success" size="small" @click="handleUnblock">{{ $t('contact.unblock') }}</el-button>
      </div>
    </el-card>

    <!-- Conversations -->
    <el-card style="margin-top: 16px" v-if="contact">
      <template #header>{{ $t('contact.conversations') }}</template>
      <el-table :data="conversations" stripe size="small">
        <el-table-column :label="$t('conversation.displayId')" width="80">
          <template #default="{ row }">#{{ row.display_id }}</template>
        </el-table-column>
        <el-table-column prop="subject" :label="$t('conversation.subject')" min-width="200" />
        <el-table-column prop="status" :label="$t('conversation.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_message_at" :label="$t('conversation.lastMessage')" width="170">
          <template #default="{ row }">{{ formatTime(row.last_message_at) }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Edit Dialog -->
    <el-dialog v-model="showEdit" :title="$t('common.edit')" width="480px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item :label="$t('contact.name')">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item :label="$t('contact.email')">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item :label="$t('contact.phone')">
          <el-input v-model="editForm.phone" />
        </el-form-item>
        <el-form-item :label="$t('contact.type')">
          <el-select v-model="editForm.contact_type">
            <el-option :label="$t('contact.typeVisitor')" value="visitor" />
            <el-option :label="$t('contact.typeLead')" value="lead" />
            <el-option :label="$t('contact.typeCustomer')" value="customer" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleUpdate" :loading="updating">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getContact, updateContact, getContactConversations } from '@/api/contact'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()

const contact = ref<any>(null)
const conversations = ref<any[]>([])
const loading = ref(true)
const showEdit = ref(false)
const updating = ref(false)
const editForm = ref({ name: '', email: '', phone: '', contact_type: '' })

function typeLabel(type: string) {
  const m: Record<string, string> = { visitor: (t as any)('contact.typeVisitor'), lead: (t as any)('contact.typeLead'), customer: (t as any)('contact.typeCustomer') }
  return m[type] || type
}
function statusTag(s: string) {
  const m: Record<string, any> = { open: 'danger', pending: 'warning', resolved: 'success', snoozed: 'info' }
  return m[s] || 'info'
}
function formatTime(t: string) { return t ? new Date(t).toLocaleString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US') : '-' }

async function fetchDetail() {
  loading.value = true
  try {
    const id = route.params.id as string
    const res = await getContact(id)
    contact.value = res.data
    editForm.value = {
      name: res.data.name || '',
      email: res.data.email || '',
      phone: res.data.phone || '',
      contact_type: res.data.contact_type || 'visitor',
    }
    const cr = await getContactConversations(id)
    conversations.value = cr.data || []
  } catch {} finally { loading.value = false }
}

async function handleUpdate() {
  updating.value = true
  try {
    await updateContact(route.params.id as string, editForm.value)
    ElMessage.success(t('common.success'))
    showEdit.value = false
    fetchDetail()
  } catch {} finally { updating.value = false }
}

async function handleBlock() {
  try {
    await updateContact(route.params.id as string, { blocked: true } as any)
    ElMessage.success(t('common.success'))
    fetchDetail()
  } catch {}
}

async function handleUnblock() {
  try {
    await updateContact(route.params.id as string, { blocked: false } as any)
    ElMessage.success(t('common.success'))
    fetchDetail()
  } catch {}
}

onMounted(fetchDetail)
</script>
