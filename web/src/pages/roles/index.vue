<template>
  <div class="roles-page">
    <el-card v-loading="loading">
      <template #header>
        <span>{{ $t('roles.title') }}</span>
      </template>

      <el-tabs v-model="activeRole">
        <el-tab-pane
          label="Admin"
          name="admin"
        >
          <template #label>
            <span>{{ $t('users.roleAdmin') }}</span>
          </template>
        </el-tab-pane>
        <el-tab-pane
          label="Agent"
          name="agent"
        >
          <template #label>
            <span>{{ $t('users.roleAgent') }}</span>
          </template>
        </el-tab-pane>
      </el-tabs>

      <el-checkbox-group
        v-model="selected"
        class="perm-list"
      >
        <div
          v-for="p in allPermissions"
          :key="p.code"
          class="perm-item"
        >
          <el-checkbox :label="p.code">
            <span class="perm-name">{{ locale === 'zh-CN' ? p.name : p.code }}</span>
            <span
              v-if="locale === 'zh-CN'"
              class="perm-desc"
            >{{ p.description }}</span>
          </el-checkbox>
        </div>
      </el-checkbox-group>

      <div style="margin-top: 24px">
        <el-button
          type="primary"
          :loading="saving"
          @click="handleSave"
        >
          {{ $t('common.save') }}
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getPermissions, getRolePermissions, setRolePermissions } from '@/api/permission'

const { t, locale } = useI18n()
const loading = ref(false)
const saving = ref(false)
const activeRole = ref('admin')
const selected = ref<string[]>([])
const allPermissions = ref<{ code: string; name: string; description: string }[]>([])

onMounted(async () => {
  loading.value = true
  try {
    const res = await getPermissions()
    allPermissions.value = res.data
  } finally { loading.value = false }
})

watch(activeRole, async (role) => {
  loading.value = true
  try {
    const res = await getRolePermissions(role)
    selected.value = res.data.permissions
  } finally { loading.value = false }
}, { immediate: true })

async function handleSave() {
  saving.value = true
  try {
    await setRolePermissions(activeRole.value, selected.value)
    ElMessage.success(t('common.success'))
  } finally { saving.value = false }
}
</script>

<style scoped lang="scss">
.roles-page {
  .perm-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .perm-item {
    padding: 8px 0;
    border-bottom: 1px solid #f0f0f0;
  }
  .perm-name {
    font-weight: 500;
  }
  .perm-desc {
    margin-left: 8px;
    color: #909399;
    font-size: 12px;
  }
}
</style>
