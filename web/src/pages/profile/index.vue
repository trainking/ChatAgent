<template>
  <div class="profile-page">
    <!-- Editable Section -->
    <el-card>
      <template #header>
        <span>{{ $t('profile.editTitle') }}</span>
      </template>
      <div class="edit-section">
        <div class="avatar-area">
          <el-avatar
            :size="80"
            :src="avatarPreview || avatarUrl"
            icon="UserFilled"
          />
          <el-upload
            :show-file-list="false"
            :before-upload="handleBeforeUpload"
            :http-request="handleAvatarUpload"
            accept=".jpg,.jpeg,.png,.gif"
          >
            <el-button
              size="small"
              type="primary"
              :loading="avatarUploading"
              style="margin-top: 12px"
            >
              {{ $t('profile.uploadAvatar') }}
            </el-button>
          </el-upload>
        </div>
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-width="80px"
          class="edit-form"
        >
          <el-form-item
            :label="$t('profile.nickname')"
            prop="name"
          >
            <el-input
              v-model="form.name"
              :placeholder="$t('profile.nicknamePlaceholder')"
              maxlength="50"
              style="max-width: 320px"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :loading="saving"
              @click="handleSave"
            >
              {{ $t('common.save') }}
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-card>

    <!-- Display Section -->
    <el-card style="margin-top: 16px">
      <template #header>
        <span>{{ $t('profile.infoTitle') }}</span>
      </template>
      <el-descriptions
        :column="2"
        border
      >
        <el-descriptions-item :label="$t('profile.email')">
          {{ user?.email }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('profile.role')">
          <el-tag
            v-if="user?.role === 'super_admin'"
            type="danger"
            size="small"
          >
            {{ $t('users.roleSuperAdmin') }}
          </el-tag>
          <el-tag
            v-else-if="user?.role === 'admin'"
            type="warning"
            size="small"
          >
            {{ $t('users.roleAdmin') }}
          </el-tag>
          <el-tag
            v-else
            type="primary"
            size="small"
          >
            {{ $t('users.roleAgent') }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('profile.createdAt')">
          {{ formatTime(user?.created_at) }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('profile.lastLogin')">
          {{ user?.last_login_at ? formatTime(user.last_login_at) : '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- Activity Log -->
    <el-card style="margin-top: 16px">
      <template #header>
        <span>{{ $t('profile.activityLog') }}</span>
      </template>
      <el-table
        v-loading="loadingActivities"
        :data="activities"
        stripe
        style="width: 100%"
      >
        <el-table-column
          prop="created_at"
          :label="$t('profile.activityTime')"
          width="180"
        >
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column
          prop="event"
          :label="$t('profile.activityEvent')"
          width="160"
        >
          <template #default="{ row }">
            <el-tag
              :type="eventTagType(row.event)"
              size="small"
            >
              {{ row.event }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="module"
          :label="$t('profile.activityModule')"
          width="120"
        />
        <el-table-column
          prop="content"
          :label="$t('profile.activityContent')"
          min-width="300"
        />
      </el-table>
      <div
        v-if="activityTotal > pageSize"
        style="margin-top: 12px; text-align: center"
      >
        <el-pagination
          v-model:current-page="currentPage"
          :page-size="pageSize"
          :total="activityTotal"
          layout="total, prev, pager, next"
          @current-change="fetchActivities"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { getProfile, updateProfile, uploadAvatar, getActivities } from '@/api/profile'
import { useUserStore } from '@/stores/user'

interface ActivityLog {
  id: string
  user_id: string
  event: string
  module: string
  content: string
  created_at: string
}

const { t, locale } = useI18n()
const userStore = useUserStore()

const user = computed(() => userStore.user)

const formRef = ref<FormInstance>()
const form = reactive({ name: '' })
const rules: FormRules = {
  name: [{ required: true, message: 'Name is required', trigger: 'blur' }],
}

const saving = ref(false)
const avatarUrl = ref('')
const avatarPreview = ref('')
const avatarUploading = ref(false)

const activities = ref<ActivityLog[]>([])
const loadingActivities = ref(false)
const currentPage = ref(1)
const pageSize = 20
const activityTotal = ref(0)

onMounted(() => {
  form.name = userStore.user?.name || ''
  avatarUrl.value = userStore.user?.avatar_url || ''
  fetchActivities()
})

function formatTime(t: string | undefined) {
  if (!t) return '-'
  return new Date(t).toLocaleString(locale.value === 'zh-CN' ? 'zh-CN' : 'en-US')
}

async function handleSave() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  saving.value = true
  try {
    await updateProfile(form.name)
    if (userStore.user) {
      userStore.user.name = form.name
      localStorage.setItem('user', JSON.stringify(userStore.user))
    }
    ElMessage.success(t('common.success'))
    fetchActivities()
  } finally {
    saving.value = false
  }
}

function handleBeforeUpload(file: File) {
  const isImage = ['image/jpeg', 'image/png', 'image/gif'].includes(file.type)
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error(t('profile.avatarFormatError'))
    return false
  }
  if (!isLt2M) {
    ElMessage.error(t('profile.avatarSizeError'))
    return false
  }
  return true
}

async function handleAvatarUpload(options: { file: File }) {
  avatarUploading.value = true
  try {
    const res = await uploadAvatar(options.file)
    const url = res.data.avatar_url
    avatarUrl.value = url
    avatarPreview.value = url
    if (userStore.user) {
      userStore.user.avatar_url = url
      localStorage.setItem('user', JSON.stringify(userStore.user))
    }
    ElMessage.success(t('common.success'))
  } finally {
    avatarUploading.value = false
  }
}

async function fetchActivities(page?: number) {
  if (page) currentPage.value = page
  loadingActivities.value = true
  try {
    const res = await getActivities(currentPage.value, pageSize)
    activities.value = res.data?.list || []
    activityTotal.value = res.data?.total || 0
  } finally {
    loadingActivities.value = false
  }
}

function eventTagType(event: string) {
  const map: Record<string, string> = {
    login: 'success',
    logout: 'info',
    change_password: 'warning',
    update_profile: 'primary',
    upload_avatar: 'primary',
  }
  return map[event] || ''
}
</script>

<style scoped lang="scss">
.profile-page {
  max-width: 900px;
  margin: 0 auto;
}
.edit-section {
  display: flex;
  gap: 40px;
  align-items: flex-start;
}
.avatar-area {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.edit-form {
  flex: 1;
}
</style>
