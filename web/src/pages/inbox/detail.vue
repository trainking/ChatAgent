<template>
  <div class="inbox-detail-page">
    <div class="page-header">
      <el-button @click="goBack" :icon="'ArrowLeft'">
        {{ $t('common.back') }}
      </el-button>
      <h2>{{ detailInbox?.name }}</h2>
    </div>

    <el-card v-if="detailInbox">
      <el-tabs v-model="detailTab">
        <el-tab-pane :label="$t('inbox.basicInfo')" name="info">
          <el-form
            ref="editFormRef"
            :model="editForm"
            label-width="120px"
            style="max-width: 560px"
          >
            <el-form-item :label="$t('inbox.type')">
              <el-tag
                :type="detailInbox.inbox_type === 'website' ? 'primary' : 'success'"
                size="small"
              >
                {{ detailInbox.inbox_type === 'website' ? 'Website' : 'API' }}
              </el-tag>
            </el-form-item>
            <el-form-item :label="$t('inbox.name')">
              <el-input :model-value="detailInbox.name" disabled />
            </el-form-item>
            <el-form-item :label="$t('inbox.description')">
              <el-input
                v-model="editForm.description"
                type="textarea"
                :rows="2"
                :placeholder="$t('inbox.descriptionPlaceholder')"
              />
            </el-form-item>
            <el-form-item :label="$t('inbox.welcomeTitle')">
              <el-input
                v-model="editForm.welcome_title"
                :placeholder="$t('inbox.welcomeTitlePlaceholder')"
                maxlength="32"
                show-word-limit
              />
            </el-form-item>
            <el-form-item :label="$t('inbox.welcomeMessage')">
              <RichTextEditor
                v-model="editForm.welcome_message"
                :maxlength="255"
              />
            </el-form-item>
            <el-form-item :label="$t('inbox.status')">
              <el-switch
                v-model="editFormStatusEnabled"
                :active-text="$t('inbox.enabled')"
                :inactive-text="$t('inbox.disabled')"
              />
            </el-form-item>
            <el-form-item :label="$t('inbox.collaborators')">
              <el-select
                v-model="editForm.collaborators"
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
            <el-form-item>
              <el-button type="primary" :loading="saving" @click="handleUpdate">
                {{ $t('common.save') }}
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        <el-tab-pane :label="$t('inbox.accessRef')" name="access">
          <el-empty :description="$t('inbox.accessRefPlaceholder')" />
        </el-tab-pane>
        <el-tab-pane :label="$t('inbox.forwardRules')" name="rules">
          <el-empty :description="$t('inbox.forwardRulesPlaceholder')" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { getInbox, updateInbox } from '@/api/inbox'
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

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const saving = ref(false)
const detailTab = ref('info')
const detailInbox = ref<InboxItem | null>(null)
const userOptions = ref<{ id: string; name: string; email: string }[]>([])

const editFormRef = ref<FormInstance>()

const editForm = reactive({
  description: '',
  welcome_title: '',
  welcome_message: '',
  collaborators: [] as string[],
})

const editFormStatusEnabled = ref(true)

onMounted(() => {
  fetchDetail()
  fetchUsers()
})

function goBack() {
  router.push('/dashboard/inbox')
}

async function fetchDetail() {
  try {
    const res = await getInbox(route.params.id as string)
    detailInbox.value = res.data
    editForm.description = res.data.description || ''
    editForm.welcome_title = res.data.welcome_title || ''
    editForm.welcome_message = res.data.welcome_message || ''
    editForm.collaborators = (res.data.collaborators || []).map((c: any) => c.id)
    editFormStatusEnabled.value = res.data.status === 'enabled'
  } catch {}
}

async function fetchUsers() {
  try {
    const res = await getUsers()
    userOptions.value = res.data || []
  } catch {}
}

async function handleUpdate() {
  if (!detailInbox.value) return
  saving.value = true
  try {
    await updateInbox(detailInbox.value.id, {
      description: editForm.description,
      welcome_title: editForm.welcome_title,
      welcome_message: editForm.welcome_message,
      status: editFormStatusEnabled.value ? 'enabled' : 'disabled',
      collaborators: editForm.collaborators,
    })
    ElMessage.success(t('common.success'))
    router.push('/dashboard/inbox')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped lang="scss">
.inbox-detail-page {
  .page-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 16px;
    h2 { margin: 0; }
  }
}
</style>
