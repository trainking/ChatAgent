<template>
  <div class="inbox-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ $t('inbox.title') }}</span>
          <el-button
            type="primary"
            size="small"
            @click="openCreate"
          >
            {{ $t('inbox.createBtn') }}
          </el-button>
        </div>
      </template>
      <el-table
        v-loading="loading"
        :data="inboxes"
        stripe
        style="width: 100%"
      >
        <el-table-column
          prop="name"
          :label="$t('inbox.name')"
          min-width="120"
        />
        <el-table-column
          prop="inbox_type"
          :label="$t('inbox.type')"
          width="100"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.inbox_type === 'website' ? 'primary' : 'success'"
              size="small"
            >
              {{ row.inbox_type === 'website' ? 'Website' : 'API' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="status"
          :label="$t('inbox.status')"
          width="90"
        >
          <template #default="{ row }">
            <el-tag
              :type="row.status === 'enabled' ? 'success' : 'info'"
              size="small"
            >
              {{ row.status === 'enabled' ? $t('inbox.enabled') : $t('inbox.disabled') }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('inbox.creator')"
          width="120"
        >
          <template #default="{ row }">
            {{ row.creator_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column
          prop="created_at"
          :label="$t('inbox.createdAt')"
          width="170"
        >
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          :label="$t('inbox.actions')"
          width="160"
        >
          <template #default="{ row }">
            <el-button
              size="small"
              @click="openDetail(row.id)"
            >
              {{ $t('inbox.detail') }}
            </el-button>
            <el-button
              size="small"
              type="danger"
              @click="handleDelete(row)"
            >
              {{ $t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create Dialog -->
    <el-dialog
      v-model="showCreate"
      :title="$t('inbox.createTitle')"
      width="620px"
    >
      <el-form
        ref="createFormRef"
        :model="createForm"
        :rules="createRules"
        label-width="90px"
      >
        <el-form-item
          :label="$t('inbox.type')"
          prop="inbox_type"
        >
          <el-radio-group v-model="createForm.inbox_type">
            <el-radio value="website">
              Website
            </el-radio>
            <el-radio value="api">
              API
            </el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          :label="$t('inbox.name')"
          prop="name"
        >
          <el-input
            v-model="createForm.name"
            :placeholder="$t('inbox.namePlaceholder')"
            maxlength="100"
          />
        </el-form-item>
        <el-form-item :label="$t('inbox.description')">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="2"
            :placeholder="$t('inbox.descriptionPlaceholder')"
          />
        </el-form-item>
        <el-form-item :label="$t('inbox.welcomeTitle')">
          <el-input
            v-model="createForm.welcome_title"
            :placeholder="$t('inbox.welcomeTitlePlaceholder')"
            maxlength="32"
            show-word-limit
          />
        </el-form-item>
        <el-form-item :label="$t('inbox.welcomeMessage')">
          <RichTextEditor
            v-model="createForm.welcome_message"
            :maxlength="255"
          />
        </el-form-item>
        <el-form-item :label="$t('inbox.status')">
          <el-switch
            v-model="createFormStatusEnabled"
            :active-text="$t('inbox.enabled')"
            :inactive-text="$t('inbox.disabled')"
          />
        </el-form-item>
        <el-form-item :label="$t('inbox.collaborators')">
          <el-select
            v-model="createForm.collaborators"
            multiple
            filterable
            :placeholder="$t('inbox.collaboratorsPlaceholder')"
            style="width: 100%"
          >
            <el-option
              v-for="u in userOptions"
              :key="u.id"
              :label="u.name + ' (' + u.email + ')'"
              :value="u.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="creating"
          @click="handleCreate"
        >
          {{ $t('common.create') }}
        </el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getInboxes, createInbox, deleteInbox } from '@/api/inbox'
import { getUsers } from '@/api/auth'
import RichTextEditor from '@/components/RichTextEditor.vue'

interface InboxItem {
  id: string
  name: string
  description: string
  welcome_title: string
  welcome_message: string
  inbox_type: string
  status: string
  created_by: string
  creator_name: string
  creator_email: string
  created_at: string
  creator: { id: string; name: string; email: string } | null
  collaborators: { id: string; name: string; email: string }[]
}

const { t, locale } = useI18n()
const router = useRouter()
const loading = ref(false)
const creating = ref(false)
const inboxes = ref<InboxItem[]>([])
const userOptions = ref<{ id: string; name: string; email: string }[]>([])

const showCreate = ref(false)

const createFormRef = ref<FormInstance>()

const createForm = reactive({
  inbox_type: 'website',
  name: '',
  description: '',
  welcome_title: '',
  welcome_message: '',
  collaborators: [] as string[],
})

const createFormStatusEnabled = ref(true)

const createRules: FormRules = {
  inbox_type: [{ required: true, message: 'Type is required', trigger: 'change' }],
  name: [{ required: true, message: 'Name is required', trigger: 'blur' }],
}

onMounted(() => {
  fetchInboxes()
  fetchUsers()
})

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US')
}

async function fetchInboxes() {
  loading.value = true
  try {
    const res = await getInboxes()
    inboxes.value = res.data || []
  } finally {
    loading.value = false
  }
}

async function fetchUsers() {
  try {
    const res = await getUsers()
    userOptions.value = res.data || []
  } catch {}
}

async function openCreate() {
  createForm.inbox_type = 'website'
  createForm.name = ''
  createForm.description = ''
  createForm.welcome_title = ''
  createForm.welcome_message = ''
  createForm.collaborators = []
  createFormStatusEnabled.value = true
  showCreate.value = true
}

async function handleCreate() {
  const valid = await createFormRef.value?.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    await createInbox({
      name: createForm.name,
      description: createForm.description,
      welcome_title: createForm.welcome_title,
      welcome_message: createForm.welcome_message,
      inbox_type: createForm.inbox_type,
      status: createFormStatusEnabled.value ? 'enabled' : 'disabled',
      collaborators: createForm.collaborators,
    })
    ElMessage.success(t('common.success'))
    showCreate.value = false
    fetchInboxes()
  } finally {
    creating.value = false
  }
}

function openDetail(id: string) {
  router.push('/dashboard/inbox/' + id)
}

async function handleDelete(row: InboxItem) {
  try {
    await ElMessageBox.confirm(
      t('inbox.deleteConfirm', { name: row.name }),
      '',
      { type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await deleteInbox(row.id)
    ElMessage.success(t('common.success'))
    fetchInboxes()
  } catch {}
}
</script>

<style scoped lang="scss">
.inbox-page { .card-header { display: flex; justify-content: space-between; align-items: center; } }
</style>
