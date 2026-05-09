<template>
  <div class="login-wrapper">
    <el-card class="login-card">
      <h2 class="login-title">{{ $t('login.title') }}</h2>

      <template v-if="step === 'password'">
        <el-form ref="formRef" :model="form" :rules="rules" size="large" @keyup.enter="handleLogin">
          <el-form-item prop="email">
            <el-input v-model="form.email" :placeholder="$t('login.email')">
              <template #prefix><el-icon><User /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item prop="password">
            <el-input v-model="form.password" type="password" :placeholder="$t('login.password')" show-password>
              <template #prefix><el-icon><Lock /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="loading" style="width: 100%" @click="handleLogin">
              {{ $t('login.loginBtn') }}
            </el-button>
          </el-form-item>
        </el-form>
      </template>

      <template v-if="step === '2fa'">
        <p class="twofa-desc">{{ $t('twofa.enterCode') }}</p>
        <el-input v-model="totpCode" placeholder="000000" maxlength="6" size="large"
          style="text-align: center; font-size: 22px; letter-spacing: 6px"
          @keyup.enter="handle2FAVerify" />
        <el-button type="primary" :loading="loading" style="width: 100%; margin-top: 16px" @click="handle2FAVerify">
          {{ $t('common.submit') }}
        </el-button>
        <el-button style="width: 100%; margin-top: 8px" @click="step = 'password'">
          {{ $t('common.back') }}
        </el-button>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import { login } from '@/api/auth'
import { verify2FALogin } from '@/api/twofactor'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const step = ref<'password' | '2fa'>('password')
const tempToken = ref('')
const totpCode = ref('')

const form = reactive({ email: '', password: '' })

const rules: FormRules = {
  email: [
    { required: true, message: t('login.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('login.emailInvalid'), trigger: 'blur' },
  ],
  password: [{ required: true, message: t('login.passwordRequired'), trigger: 'blur' }],
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await login(form.email, form.password)
    const d = res.data

    if (d.require_2fa) {
      tempToken.value = d.temp_token
      step.value = '2fa'
      return
    }

    if (d.require_2fa_setup) {
      userStore.setTempAuth(d.temp_token)
      router.push('/setup-2fa')
      return
    }

    userStore.setAuth(d.token, d.user)
    if (d.user.must_change_password) {
      router.push('/change-password')
    } else {
      router.push('/dashboard')
    }
  } catch {} finally { loading.value = false }
}

async function handle2FAVerify() {
  if (!totpCode.value || totpCode.value.length < 6) return

  loading.value = true
  try {
    const res = await verify2FALogin(tempToken.value, totpCode.value)
    const d = res.data
    userStore.setAuth(d.token, d.user)
    if (d.user.must_change_password) {
      router.push('/change-password')
    } else {
      router.push('/dashboard')
    }
  } catch {} finally { loading.value = false }
}
</script>

<style scoped lang="scss">
.login-wrapper { height: 100vh; display: flex; align-items: center; justify-content: center; background: #f0f2f5; }
.login-card { width: 420px; }
.login-title { text-align: center; margin-bottom: 32px; color: #303133; }
.twofa-desc { text-align: center; margin-bottom: 16px; color: #909399; }
</style>
