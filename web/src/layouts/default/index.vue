<template>
  <el-container class="layout-container">
    <el-aside width="220px" class="layout-aside">
      <div class="logo">
        <h2>ChatAgent</h2>
      </div>
      <el-menu
        router
        :default-active="route.path"
        background-color="#001529"
        text-color="#fff"
        active-text-color="#409eff"
      >
        <el-menu-item index="/dashboard/inbox">
          <el-icon><ChatDotRound /></el-icon>
          <span>{{ $t('layout.inbox') }}</span>
        </el-menu-item>
        <el-menu-item v-if="userStore.isAdmin()" index="/dashboard/users">
          <el-icon><UserFilled /></el-icon>
          <span>{{ $t('layout.users') }}</span>
        </el-menu-item>
        <el-menu-item v-if="userStore.isSuperAdmin()" index="/dashboard/roles">
          <el-icon><Setting /></el-icon>
          <span>{{ $t('layout.roles') }}</span>
        </el-menu-item>
        <el-menu-item v-if="userStore.isSuperAdmin()" index="/dashboard/system">
          <el-icon><Tools /></el-icon>
          <span>{{ $t('layout.system') }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="layout-header">
        <span class="header-title">{{ $t('layout.title') }}</span>
        <div class="header-right">
          <el-dropdown @command="handleLangSwitch">
            <el-button size="small" text>
              <el-icon><Switch /></el-icon>
              <span style="margin-left: 4px">{{ locale === 'zh-CN' ? $t('lang.zh') : $t('lang.en') }}</span>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="zh-CN">
                  <span :style="{ fontWeight: locale === 'zh-CN' ? 'bold' : 'normal' }">{{ $t('lang.zh') }}</span>
                </el-dropdown-item>
                <el-dropdown-item command="en-US">
                  <span :style="{ fontWeight: locale === 'en-US' ? 'bold' : 'normal' }">{{ $t('lang.en') }}</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown>
            <span class="user-info">
              <el-avatar :size="32" icon="UserFilled" />
              <span class="user-name">{{ userStore.user?.name || 'User' }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleChangePassword">{{ $t('layout.changePassword') }}</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">{{ $t('layout.logout') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>

    <el-dialog v-model="showPwdDialog" :title="$t('changePassword.title')" width="420px" :close-on-click-modal="false">
      <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="0">
        <el-form-item prop="old_password">
          <el-input v-model="pwdForm.old_password" type="password" :placeholder="$t('changePassword.oldPassword')" show-password />
        </el-form-item>
        <el-form-item prop="new_password">
          <el-input v-model="pwdForm.new_password" type="password" :placeholder="$t('changePassword.newPassword')" show-password />
        </el-form-item>
        <el-form-item prop="confirm_password">
          <el-input v-model="pwdForm.confirm_password" type="password" :placeholder="$t('changePassword.confirmPassword')" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPwdDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="pwdLoading" @click="handlePwdSubmit">{{ $t('changePassword.submit') }}</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { changePassword } from '@/api/auth'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const showPwdDialog = ref(false)
const pwdLoading = ref(false)
const pwdFormRef = ref<FormInstance>()

const pwdForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const validatePwdConfirm = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== pwdForm.new_password) {
    callback(new Error(t('changePassword.mismatch')))
  } else {
    callback()
  }
}

const pwdRules = computed<FormRules>(() => ({
  old_password: [{ required: true, message: t('changePassword.oldPasswordRequired'), trigger: 'blur' }],
  new_password: [
    { required: true, message: t('changePassword.newPasswordRequired'), trigger: 'blur' },
    { min: 6, message: t('changePassword.newPasswordMin'), trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: t('changePassword.confirmRequired'), trigger: 'blur' },
    { validator: validatePwdConfirm, trigger: 'blur' },
  ],
}))

async function handlePwdSubmit() {
  const valid = await pwdFormRef.value?.validate().catch(() => false)
  if (!valid) return

  pwdLoading.value = true
  try {
    await changePassword(pwdForm.old_password, pwdForm.new_password)
    ElMessage.success(t('changePassword.success'))
    userStore.markPasswordChanged()
    showPwdDialog.value = false
    pwdForm.old_password = ''
    pwdForm.new_password = ''
    pwdForm.confirm_password = ''
  } catch {
    // axios interceptor handles error
  } finally {
    pwdLoading.value = false
  }
}

function handleLogout() {
  userStore.clearAuth()
  router.push('/login')
}

function handleChangePassword() {
  showPwdDialog.value = true
}

function handleLangSwitch(lang: string) {
  locale.value = lang
  localStorage.setItem('lang', lang)
}
</script>

<style scoped lang="scss">
.layout-container {
  height: 100%;
}
.layout-aside {
  background: #001529;
  overflow: hidden;
  .logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    h2 {
      color: #fff;
      font-size: 18px;
      white-space: nowrap;
    }
  }
  .el-menu {
    border-right: none;
  }
}
.layout-header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e4e7ed;
  .header-title {
    font-size: 16px;
    font-weight: 500;
  }
  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .user-info {
    display: flex;
    align-items: center;
    cursor: pointer;
    .user-name {
      margin-left: 8px;
    }
  }
}
</style>
