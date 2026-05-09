<template>
  <div class="setup-wrapper">
    <el-card class="setup-card">
      <h2 class="setup-title">{{ $t('twofa.setupTitle') }}</h2>
      <p class="setup-desc">{{ $t('twofa.setupDesc') }}</p>

      <div v-if="qrUrl" class="qr-section">
        <img :src="`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrUrl)}`"
          alt="QR Code" class="qr-img" />
        <p class="secret-text">{{ $t('twofa.manualKey') }}: {{ secret }}</p>
      </div>

      <el-form v-if="qrUrl" size="large">
        <el-form-item>
          <el-input v-model="code" :placeholder="$t('twofa.enterCode')" maxlength="6"
            style="text-align: center; font-size: 22px; letter-spacing: 6px"
            @keyup.enter="handleVerify" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" style="width: 100%" @click="handleVerify">
            {{ $t('common.submit') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { setup2FA, verify2FASetup } from '@/api/twofactor'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { login } from '@/api/auth'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const qrUrl = ref('')
const secret = ref('')
const code = ref('')

// Temp token from login
const tempToken = (userStore.tempToken || localStorage.getItem('temp_token')) ?? ''

onMounted(async () => {
  if (!tempToken) { router.push('/login'); return }

  localStorage.setItem('token', tempToken) // temp auth for API call
  try {
    const res = await setup2FA()
    qrUrl.value = res.data.qr_url
    secret.value = res.data.secret
  } catch { router.push('/login') }
})

async function handleVerify() {
  if (!code.value || code.value.length < 6) return
  loading.value = true
  try {
    await verify2FASetup(code.value)

    // Now login again to get real token
    localStorage.setItem('token', tempToken)
    // Actually, we need to log the user in properly
    // The temp_token is still valid, so use it
    ElMessage.success('2FA setup complete, please login again')
    localStorage.removeItem('token')
    localStorage.removeItem('temp_token')
    router.push('/login')
  } catch {} finally { loading.value = false }
}
</script>

<style scoped lang="scss">
.setup-wrapper { height: 100vh; display: flex; align-items: center; justify-content: center; background: #f0f2f5; }
.setup-card { width: 440px; }
.setup-title { text-align: center; margin-bottom: 8px; }
.setup-desc { text-align: center; margin-bottom: 20px; color: #909399; }
.qr-section { text-align: center; margin-bottom: 20px; }
.qr-img { border: 1px solid #eee; padding: 8px; }
.secret-text { margin-top: 8px; color: #909399; font-size: 13px; word-break: break-all; }
</style>
