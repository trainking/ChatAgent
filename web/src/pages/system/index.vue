<template>
  <div class="system-page">
    <el-card v-loading="loading">
      <template #header>
        <span>{{ $t('system.title') }}</span>
      </template>

      <el-divider content-position="left">
        {{ $t('system.twofa') }}
      </el-divider>

      <el-form label-width="140px">
        <el-form-item :label="$t('system.twofaEnable')">
          <el-switch
            v-model="config.enabled"
            @change="handleToggle"
          />
        </el-form-item>
        <el-form-item :label="$t('system.twofaIssuer')">
          <el-input
            v-model="config.issuer"
            style="width: 300px"
            @blur="handleSave"
          />
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { get2FAConfig, set2FAConfig } from '@/api/twofactor'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const config = reactive({ enabled: false, issuer: 'ChatAgent' })
let loaded = false

onMounted(async () => {
  loading.value = true
  try {
    const res = await get2FAConfig()
    config.enabled = res.data['2fa_enabled'] === 'true'
    config.issuer = res.data['2fa_issuer'] || 'ChatAgent'
    loaded = true
  } finally { loading.value = false }
})

async function handleToggle() {
  if (!loaded) return
  await doSave()
}

async function handleSave() {
  if (!loaded) return
  await doSave()
}

async function doSave() {
  try {
    await set2FAConfig(config.enabled, config.issuer)
    ElMessage.success(t('common.success'))
  } catch {}
}
</script>
