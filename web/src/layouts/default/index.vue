<template>
  <el-container class="pro-layout">
    <!-- Sidebar -->
    <el-aside
      :width="collapsed ? '64px' : '220px'"
      class="pro-sider"
    >
      <div
        class="pro-logo"
        :class="{ collapsed }"
      >
        <el-icon :size="24">
          <ChatDotRound />
        </el-icon>
        <span
          v-show="!collapsed"
          class="logo-text"
        >ChatAgent</span>
      </div>

      <el-scrollbar>
        <el-menu
          :default-active="route.path"
          :collapse="collapsed"
          :collapse-transition="false"
          background-color="#001529"
          text-color="#ffffffa6"
          active-text-color="#fff"
          router
        >
          <el-menu-item index="/dashboard/inbox">
            <el-icon><ChatDotRound /></el-icon>
            <template #title>
              {{ $t('layout.inbox') }}
            </template>
          </el-menu-item>
          <el-menu-item
            v-if="userStore.isAdmin()"
            index="/dashboard/users"
          >
            <el-icon><UserFilled /></el-icon>
            <template #title>
              {{ $t('layout.users') }}
            </template>
          </el-menu-item>
          <el-sub-menu
            v-if="userStore.isSuperAdmin()"
            index="/dashboard/system"
          >
            <template #title>
              <el-icon><Tools /></el-icon>
              <span>{{ $t('layout.system') }}</span>
            </template>
            <el-menu-item index="/dashboard/system/roles">
              <span>{{ $t('layout.roles') }}</span>
            </el-menu-item>
            <el-menu-item index="/dashboard/system/2fa">
              <span>{{ $t('layout.twofa') }}</span>
            </el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <!-- Main -->
    <el-container class="pro-main">
      <!-- Header -->
      <el-header class="pro-header">
        <div class="header-left">
          <el-icon
            class="collapse-btn"
            :size="20"
            @click="collapsed = !collapsed"
          >
            <Fold v-if="!collapsed" /><Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">
              Home
            </el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.matched.length > 1">
              {{ breadcrumbTitle }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown
            trigger="click"
            @command="handleStatusSwitch"
          >
            <span class="action-item">
              <span
                class="status-dot"
                :class="statusDotClass"
              />
              <span>{{ statusLabel }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="online">
                  <span
                    class="status-dot online"
                  />{{ $t('onlineStatus.online') }}
                </el-dropdown-item>
                <el-dropdown-item command="busy">
                  <span
                    class="status-dot busy"
                  />{{ $t('onlineStatus.busy') }}
                </el-dropdown-item>
                <el-dropdown-item command="offline">
                  <span
                    class="status-dot offline"
                  />{{ $t('onlineStatus.offline') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown @command="handleLangSwitch">
            <span class="action-item">
              <el-icon><Switch /></el-icon>
              <span>{{ locale === 'zh-CN' ? $t('lang.zh') : $t('lang.en') }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="zh-CN">
                  <span :class="{ bold: locale === 'zh-CN' }">{{ $t('lang.zh') }}</span>
                </el-dropdown-item>
                <el-dropdown-item command="en-US">
                  <span :class="{ bold: locale === 'en-US' }">{{ $t('lang.en') }}</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <el-dropdown>
            <span class="action-item user-action">
              <el-avatar
                :size="28"
                :src="userStore.user?.avatar_url || ''"
              >
                <el-icon><UserFilled /></el-icon>
              </el-avatar>
              <span class="user-name">{{ userStore.user?.name || 'User' }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="goProfile">
                  <el-icon><User /></el-icon>{{ $t('layout.profile') }}
                </el-dropdown-item>
                <el-dropdown-item @click="showPwdDialog = true">
                  <el-icon><Key /></el-icon>{{ $t('layout.changePassword') }}
                </el-dropdown-item>
                <el-dropdown-item
                  divided
                  @click="handleLogout"
                >
                  <el-icon><SwitchButton /></el-icon>{{ $t('layout.logout') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- Content -->
      <el-main class="pro-content">
        <router-view />
      </el-main>
    </el-container>

    <!-- Password Dialog -->
    <el-dialog
      v-model="showPwdDialog"
      :title="$t('changePassword.title')"
      width="420px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="pwdFormRef"
        :model="pwdForm"
        :rules="pwdRules"
        label-width="0"
      >
        <el-form-item prop="old_password">
          <el-input
            v-model="pwdForm.old_password"
            type="password"
            :placeholder="$t('changePassword.oldPassword')"
            show-password
          />
        </el-form-item>
        <el-form-item prop="new_password">
          <el-input
            v-model="pwdForm.new_password"
            type="password"
            :placeholder="$t('changePassword.newPassword')"
            show-password
          />
        </el-form-item>
        <el-form-item prop="confirm_password">
          <el-input
            v-model="pwdForm.confirm_password"
            type="password"
            :placeholder="$t('changePassword.confirmPassword')"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPwdDialog = false">
          {{ $t('common.cancel') }}
        </el-button>
        <el-button
          type="primary"
          :loading="pwdLoading"
          @click="handlePwdSubmit"
        >
          {{ $t('changePassword.submit') }}
        </el-button>
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
import { logout, updateOnlineStatus } from '@/api/profile'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const collapsed = ref(false)

const breadcrumbTitle = computed(() => {
  const meta = route.matched[route.matched.length - 1]?.meta || {}
  return (meta.title as string) || route.name || ''
})

const statusDotClass = computed(() => userStore.user?.online_status || 'offline')
const statusLabel = computed(() => {
  const s = userStore.user?.online_status || 'offline'
  return (t as any)(`onlineStatus.${s}`)
})

async function handleStatusSwitch(status: string) {
  try {
    await updateOnlineStatus(status)
    if (userStore.user) {
      userStore.user.online_status = status
      localStorage.setItem('user', JSON.stringify(userStore.user))
    }
  } catch {}
}

const showPwdDialog = ref(false)
const pwdLoading = ref(false)
const pwdFormRef = ref<FormInstance>()
const pwdForm = reactive({ old_password: '', new_password: '', confirm_password: '' })

const validatePwdConfirm = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== pwdForm.new_password) callback(new Error(t('changePassword.mismatch')))
  else callback()
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
    pwdForm.old_password = ''; pwdForm.new_password = ''; pwdForm.confirm_password = ''
  } catch {} finally { pwdLoading.value = false }
}

function goProfile() { router.push('/dashboard/profile') }
async function handleLogout() { try { await logout() } catch {} userStore.clearAuth(); router.push('/login') }

function handleLangSwitch(lang: string) { locale.value = lang; localStorage.setItem('lang', lang) }
</script>

<style scoped lang="scss">
.pro-layout { height: 100vh; }

.pro-sider {
  background: #001529;
  transition: width 0.2s;
  overflow: hidden;
  .pro-logo {
    height: 48px;
    margin: 16px;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #fff;
    background: rgba(255, 255, 255, 0.08);
    overflow: hidden;
    transition: all 0.2s;
    &.collapsed { margin: 16px 12px; }
    .logo-text { font-size: 16px; font-weight: 600; white-space: nowrap; }
  }
  .el-menu { border-right: none; }
}

.pro-main { flex-direction: column; min-width: 0; }

.pro-header {
  height: 48px !important;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  z-index: 10;
  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .collapse-btn {
    cursor: pointer;
    color: #606266;
    &:hover { color: #409eff; }
  }
  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .action-item {
    display: flex;
    align-items: center;
    gap: 4px;
    cursor: pointer;
    color: #606266;
    font-size: 14px;
    padding: 4px 8px;
    border-radius: 4px;
    transition: background 0.2s;
    &:hover { background: #f5f5f5; }
  }
  .user-action { padding: 2px 8px 2px 4px; }
  .user-name { margin-left: 6px; max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
}

.pro-content {
  background: #f0f2f5;
  padding: 16px;
  height: calc(100vh - 48px);
  overflow-y: auto;
}

.bold { font-weight: 700; }

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
  background: #909399;
  &.online { background: #67c23a; }
  &.offline { background: #909399; }
  &.busy { background: #e6a23c; }
}
</style>
