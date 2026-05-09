<template>
  <div class="init-wrapper">
    <el-card class="init-card">
      <h2 class="init-title">
        {{ $t('init.title') }}
      </h2>
      <p class="init-desc">
        {{ $t('init.desc') }}
      </p>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        size="large"
        @keyup.enter="handleInit"
      >
        <el-form-item prop="email">
          <el-input
            v-model="form.email"
            :placeholder="$t('init.email')"
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item prop="name">
          <el-input
            v-model="form.name"
            :placeholder="$t('init.name')"
          >
            <template #prefix>
              <el-icon><UserFilled /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            :placeholder="$t('init.password')"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%"
            @click="handleInit"
          >
            {{ $t('init.submit') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import { initRoot } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({ email: '', name: '', password: '' })

const rules: FormRules = {
  email: [
    { required: true, message: t('login.emailRequired'), trigger: 'blur' },
    { type: 'email', message: t('login.emailInvalid'), trigger: 'blur' },
  ],
  name: [{ required: true, message: t('init.nameRequired'), trigger: 'blur' }],
  password: [
    { required: true, message: t('init.passwordRequired'), trigger: 'blur' },
    { min: 6, message: t('init.passwordMin'), trigger: 'blur' },
  ],
}

async function handleInit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    const res = await initRoot(form.email, form.password, form.name)
    userStore.setAuth(res.data.token, res.data.user)
    router.push('/dashboard')
  } catch { /* axios interceptor */ }
  finally { loading.value = false }
}
</script>

<style scoped lang="scss">
.init-wrapper { height: 100vh; display: flex; align-items: center; justify-content: center; background: #f0f2f5; }
.init-card { width: 440px; }
.init-title { text-align: center; margin-bottom: 8px; color: #303133; }
.init-desc { text-align: center; margin-bottom: 24px; color: #909399; font-size: 14px; }
</style>
