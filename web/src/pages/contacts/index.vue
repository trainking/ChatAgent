<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>{{ $t('contact.title') }}</span>
        <el-button type="primary" size="small" @click="showCreate = true">
          {{ $t('contact.createBtn') }}
        </el-button>
      </div>
    </template>

    <div class="filter-bar">
      <el-select v-model="filterInbox" :placeholder="$t('contact.filterInbox')" size="small" clearable @change="fetchList">
        <el-option v-for="ib in inboxes" :key="ib.id" :label="ib.name" :value="ib.id" />
      </el-select>
      <el-select v-model="filterType" :placeholder="$t('contact.filterType')" size="small" clearable @change="fetchList">
        <el-option :label="$t('contact.typeVisitor')" value="visitor" />
        <el-option :label="$t('contact.typeLead')" value="lead" />
        <el-option :label="$t('contact.typeCustomer')" value="customer" />
      </el-select>
      <el-input
        v-model="filterSearch"
        :placeholder="$t('common.search')"
        size="small"
        clearable
        style="width: 200px"
        @clear="fetchList"
        @keyup.enter="fetchList"
      />
      <el-button size="small" @click="fetchList">{{ $t('common.search') }}</el-button>
    </div>

    <el-table :data="contacts" v-loading="loading" stripe style="margin-top: 12px">
      <el-table-column prop="name" :label="$t('contact.name')" min-width="120">
        <template #default="{ row }">
          <router-link :to="`/dashboard/contacts/${row.id}`">{{ row.name }}</router-link>
        </template>
      </el-table-column>
      <el-table-column prop="email" :label="$t('contact.email')" min-width="150" />
      <el-table-column prop="phone" :label="$t('contact.phone')" width="130" />
      <el-table-column prop="contact_type" :label="$t('contact.type')" width="100">
        <template #default="{ row }">
          <el-tag :type="typeTag(row.contact_type)" size="small">{{ typeLabel(row.contact_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="country" :label="$t('contact.country')" width="100" />
      <el-table-column prop="last_activity_at" :label="$t('contact.lastActivity')" width="170">
        <template #default="{ row }">{{ formatTime(row.last_activity_at) }}</template>
      </el-table-column>
    </el-table>

    <div style="margin-top: 12px; text-align: right">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        @current-change="fetchList"
      />
    </div>
  </el-card>

  <!-- Create Dialog -->
  <el-dialog v-model="showCreate" :title="$t('contact.createTitle')" width="480px">
    <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="80px">
      <el-form-item :label="$t('contact.name')" prop="name">
        <el-input v-model="createForm.name" />
      </el-form-item>
      <el-form-item :label="$t('contact.email')">
        <el-input v-model="createForm.email" />
      </el-form-item>
      <el-form-item :label="$t('contact.phone')">
        <el-input v-model="createForm.phone" />
      </el-form-item>
      <el-form-item :label="$t('contact.type')">
        <el-select v-model="createForm.contact_type">
          <el-option :label="$t('contact.typeVisitor')" value="visitor" />
          <el-option :label="$t('contact.typeLead')" value="lead" />
          <el-option :label="$t('contact.typeCustomer')" value="customer" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showCreate = false">{{ $t('common.cancel') }}</el-button>
      <el-button type="primary" @click="handleCreate" :loading="creating">{{ $t('common.create') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getContacts, createContact } from '@/api/contact'
import { getInboxes } from '@/api/inbox'

const { t, locale } = useI18n()

const filterInbox = ref('')
const filterType = ref('')
const filterSearch = ref('')
const contacts = ref<any[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const inboxes = ref<{ id: string; name: string }[]>([])

const showCreate = ref(false)
const creating = ref(false)
const createFormRef = ref<FormInstance>()
const createForm = ref({ name: '', email: '', phone: '', contact_type: 'visitor' })
const createRules: FormRules = { name: [{ required: true, message: 'required', trigger: 'blur' }] }

function typeTag(t: string) { return t === 'visitor' ? 'info' : t === 'lead' ? 'warning' : 'success' }
function typeLabel(type: string) {
  const m: Record<string, string> = { visitor: (t as any)('contact.typeVisitor'), lead: (t as any)('contact.typeLead'), customer: (t as any)('contact.typeCustomer') }
  return m[type] || type
}
function formatTime(t: string) { return t ? new Date(t).toLocaleString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US') : '-' }

async function fetchInboxes() {
  try { const res = await getInboxes(); inboxes.value = res.data || [] } catch {}
}
async function fetchList() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (filterInbox.value) params.inbox_id = filterInbox.value
    if (filterType.value) params.type = filterType.value
    if (filterSearch.value) params.search = filterSearch.value
    const res = await getContacts(params)
    contacts.value = res.data?.list || []
    total.value = res.data?.total || 0
  } catch {} finally { loading.value = false }
}
async function handleCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    await createContact(createForm.value)
    ElMessage.success(t('common.success'))
    showCreate.value = false
    createForm.value = { name: '', email: '', phone: '', contact_type: 'visitor' }
    fetchList()
  } catch {} finally { creating.value = false }
}

onMounted(async () => { await fetchInboxes(); fetchList() })
</script>

<style scoped>
.card-header { display: flex; justify-content: space-between; align-items: center; }
.filter-bar { display: flex; gap: 8px; flex-wrap: wrap; }
</style>
