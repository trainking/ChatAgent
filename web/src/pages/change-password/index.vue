<template>
  <div class="changepwd-wrapper">
    <el-card class="changepwd-card">
      <h2 class="changepwd-title">
        {{ $t('changePassword.title') }}
      </h2>
      <p class="changepwd-desc">
        {{ $t('changePassword.firstLoginDesc') }}
      </p>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        size="large"
        @keyup.enter="handleChange"
      >
        <el-form-item prop="old_password">
          <el-input
            v-model="form.old_password"
            type="password"
            :placeholder="$t('changePassword.oldPassword')"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item prop="new_password">
          <el-input
            v-model="form.new_password"
            type="password"
            :placeholder="$t('changePassword.newPassword')"
            show-password
          >
            <template #prefix>
              <el-icon><Key /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item prop="confirm_password">
          <el-input
            v-model="form.confirm_password"
            type="password"
            :placeholder="$t('changePassword.confirmPassword')"
            show-password
          >
            <template #prefix>
              <el-icon><Key /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%"
            @click="handleChange"
          >
            {{ $t('changePassword.submit') }}
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
import { changePassword } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({ old_password: '', new_password: '', confirm_password: '' })

const validateConfirm = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== form.new_password) { callback(new Error(t('changePassword.mismatch'))) }
  else { callback() }
}

const rules: FormRules = {
  old_password: [{ required: true, message: t('changePassword.oldPasswordRequired'), trigger: 'blur' }],
  new_password: [
    { required: true, message: t('changePassword.newPasswordRequired'), trigger: 'blur' },
    { min: 6, message: t('changePassword.newPasswordMin'), trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: t('changePassword.confirmRequired'), trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
}

async function handleChange() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await changePassword(form.old_password, form.new_password)
    userStore.markPasswordChanged()
    router.push('/dashboard')
  } catch { /* axios interceptor */ }
  finally { loading.value = false }
}
</script>

<style scoped lang="scss">
.changepwd-wrapper { height: 100vh; display: flex; align-items: center; justify-content: center; background: #f0f2f5; }
.changepwd-card { width: 440px; }
.changepwd-title { text-align: center; margin-bottom: 8px; color: #303133; }
.changepwd-desc { text-align: center; margin-bottom: 24px; color: #e6a23c; font-size: 14px; }
</style>
