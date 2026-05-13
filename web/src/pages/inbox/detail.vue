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
            <el-form-item :label="$t('inbox.icon')">
              <el-input
                v-model="editForm.icon"
                :placeholder="$t('inbox.iconPlaceholder')"
                maxlength="500"
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
          <div v-if="detailInbox?.inbox_type === 'website'" class="access-ref-content">
            <el-alert
              type="info"
              :closable="false"
              style="margin-bottom: 16px"
            >
              <template #title>
                {{ $t('inbox.accessRefDesc') }}
              </template>
            </el-alert>

            <h4>{{ $t('inbox.accessRefSnippet') }}</h4>
            <div class="code-block">
              <div class="code-header">
                <span>HTML</span>
                <el-button size="small" text @click="copySnippet">
                  <el-icon><DocumentCopy /></el-icon>{{ copyBtnText }}
                </el-button>
              </div>
              <pre class="code-pre"><code>{{ snippetCode }}</code></pre>
            </div>

            <h4 style="margin-top: 24px">{{ $t('inbox.accessRefParams') }}</h4>
            <el-table :data="paramTable" size="small" border style="margin-top: 12px">
              <el-table-column prop="attr" label="Attribute" width="180" />
              <el-table-column prop="type" label="Type" width="100" />
              <el-table-column prop="desc" :label="$t('inbox.accessRefParamDesc')" />
            </el-table>

            <h4 style="margin-top: 24px">{{ $t('inbox.accessRefPreview') }}</h4>
            <el-alert
              type="warning"
              :closable="false"
              :title="$t('inbox.accessRefPreviewHint')"
              style="margin-bottom: 12px"
            />
          </div>
          <el-empty v-else :description="$t('inbox.accessRefNotWebsite')" />
        </el-tab-pane>
        <el-tab-pane :label="$t('inbox.forwardRules')" name="rules">
          <el-empty :description="$t('inbox.forwardRulesPlaceholder')" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
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
  icon: string
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
  icon: '',
  welcome_title: '',
  welcome_message: '',
  collaborators: [] as string[],
})

const editFormStatusEnabled = ref(true)

const snippetCode = computed(() => {
  const host = location.host
  return `<script\n  src="//${host}/widget.js"\n  data-inbox-id="${detailInbox.value?.id || 'YOUR_INBOX_ID'}"\n  data-mode="bubble"\n  data-color="#409EFF"\n  data-position="right"\n  async\n><\\/script>`
})

const paramTable = computed(() => [
  { attr: 'data-inbox-id', type: 'string', desc: t('inbox.accessRefParamInboxId') },
  { attr: 'data-mode', type: 'enum', desc: t('inbox.accessRefParamMode') },
  { attr: 'data-color', type: 'color', desc: t('inbox.accessRefParamColor') },
  { attr: 'data-position', type: 'enum', desc: t('inbox.accessRefParamPosition') },
])

const copyBtnText = ref(t('common.copy'))
async function copySnippet() {
  try {
    await navigator.clipboard.writeText(snippetCode.value)
    copyBtnText.value = t('common.copied')
    setTimeout(() => { copyBtnText.value = t('common.copy') }, 2000)
  } catch {}
}

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
    editForm.icon = res.data.icon || ''
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
      icon: editForm.icon,
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

.access-ref-content {
  h4 {
    font-size: 15px;
    font-weight: 600;
    margin-bottom: 8px;
  }
}

.code-block {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  overflow: hidden;
  .code-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 14px;
    background: #f8fafc;
    border-bottom: 1px solid #e2e8f0;
    font-size: 13px;
    color: #64748b;
  }
  .code-pre {
    padding: 14px;
    margin: 0;
    background: #1e293b;
    color: #e2e8f0;
    font-size: 13px;
    line-height: 1.7;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
    font-family: 'SF Mono', 'Fira Code', monospace;
  }
}
</style>
