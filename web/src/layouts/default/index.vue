<template>
  <el-container class="pro-layout" :class="{ collapsed }">
    <!-- Sidebar -->
    <el-aside
      :width="collapsed ? '64px' : '202px'"
      class="pro-sider"
    >
      <div
        class="pro-logo"
        :class="{ collapsed }"
      >
        <div class="brand-mark">
          <ChatDotRound />
        </div>
        <span
          v-show="!collapsed"
          class="logo-text"
        >Admin</span>
      </div>

      <el-scrollbar class="side-scroll">
        <el-menu
          :default-active="route.path"
          :collapse="collapsed"
          :collapse-transition="false"
          router
        >
          <el-menu-item index="/dashboard/inbox">
            <el-icon><Message /></el-icon>
            <template #title>
              {{ $t('layout.inbox') }}
            </template>
          </el-menu-item>
          <el-menu-item index="/dashboard/conversations">
            <el-icon><ChatDotRound /></el-icon>
            <template #title>
              {{ $t('layout.conversations') }}
            </template>
          </el-menu-item>
          <el-menu-item index="/dashboard/contacts">
            <el-icon><User /></el-icon>
            <template #title>
              {{ $t('layout.contacts') }}
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

      <div class="side-footer">
        <el-dropdown trigger="click" :hide-on-click="false">
          <span class="user-action" :class="{ collapsed }">
            <span class="avatar-status-wrap">
              <el-avatar
                :size="30"
                :src="userStore.user?.avatar_url || ''"
              >
                {{ userInitial }}
              </el-avatar>
              <span class="avatar-status-dot" :class="statusDotClass" />
            </span>
            <span v-show="!collapsed" class="user-meta">
              <span class="user-name">{{ userStore.user?.name || 'User' }}</span>
              <span class="user-email">{{ userStore.user?.email || '' }}</span>
            </span>
          </span>
          <template #dropdown>
            <el-dropdown-menu class="user-dropdown-menu">
              <div class="dropdown-select-group" @click.stop>
                <label class="dropdown-section">{{ $t('layout.onlineStatus') }}</label>
                <el-select
                  :model-value="statusDotClass"
                  size="small"
                  class="dropdown-select"
                  :teleported="false"
                  @change="handleStatusSwitch"
                >
                  <el-option value="online" :label="$t('onlineStatus.online')">
                    <span class="status-dot online" />{{ $t('onlineStatus.online') }}
                  </el-option>
                  <el-option value="busy" :label="$t('onlineStatus.busy')">
                    <span class="status-dot busy" />{{ $t('onlineStatus.busy') }}
                  </el-option>
                  <el-option value="offline" :label="$t('onlineStatus.offline')">
                    <span class="status-dot offline" />{{ $t('onlineStatus.offline') }}
                  </el-option>
                </el-select>
              </div>
              <div class="dropdown-select-group" @click.stop>
                <label class="dropdown-section">{{ $t('layout.language') }}</label>
                <el-select
                  :model-value="locale"
                  size="small"
                  class="dropdown-select"
                  :teleported="false"
                  @change="handleLangSwitch"
                >
                  <el-option value="zh-CN" :label="$t('lang.zh')" />
                  <el-option value="en-US" :label="$t('lang.en')" />
                </el-select>
              </div>
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
    </el-aside>

    <!-- Main -->
    <el-container class="pro-main">
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

const statusDotClass = computed(() => userStore.user?.online_status || 'offline')
const userInitial = computed(() => (userStore.user?.name || userStore.user?.email || 'U').slice(0, 1).toUpperCase())

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
.pro-layout {
  height: 100vh;
  background: #ffffff;
  color: #111827;
}

.pro-sider {
  position: relative;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-right: 1px solid #e5e7eb;
  transition: width 0.2s;
  overflow: hidden;

  .pro-logo {
    height: 42px;
    padding: 0 8px 0 14px;
    display: flex;
    align-items: center;
    gap: 10px;
    color: #111827;
    overflow: hidden;

    &.collapsed {
      justify-content: center;
      padding: 0;

    }
  }

  .brand-mark {
    flex: 0 0 16px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #1a73e8;
    color: #ffffff;
    display: flex;
    align-items: center;
    justify-content: center;

    :deep(svg) {
      width: 10px;
      height: 10px;
      stroke-width: 4;
    }
  }

  .logo-text {
    flex: 1;
    min-width: 0;
    font-size: 14px;
    font-weight: 700;
    white-space: nowrap;
  }

  .el-menu {
    border-right: none;
    padding: 6px 8px;
    background: transparent;
  }

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    height: 36px;
    margin: 3px 0;
    padding: 0 10px !important;
    border-radius: 8px;
    color: #374151;
    font-size: 14px;
    line-height: 36px;
  }

  :deep(.el-menu-item .el-icon),
  :deep(.el-sub-menu__title .el-icon) {
    color: #6b7280;
  }

  :deep(.el-menu-item:hover),
  :deep(.el-sub-menu__title:hover) {
    background: #f7f8fa;
    color: #111827;
  }

  :deep(.el-menu-item.is-active) {
    background: #f2f4f7;
    color: #006adc;
    font-weight: 700;
  }

  :deep(.el-menu-item.is-active .el-icon) {
    color: #006adc;
  }
}

.side-scroll {
  flex: 1;
  min-height: 0;
}

.side-footer {
  padding: 9px 8px;
  border-top: 1px solid #e5e7eb;
  background: #ffffff;
}

.user-action {
  width: 100%;
  display: flex;
  align-items: center;
  cursor: pointer;
  border-radius: 8px;
}

.user-action {
  gap: 8px;
  min-height: 40px;
  padding: 5px 4px;

  &:hover {
    background: #f7f8fa;
  }

  &.collapsed {
    justify-content: center;
    padding: 5px 0;
  }
}

.avatar-status-wrap {
  position: relative;
  flex: 0 0 auto;
  display: inline-flex;
}

.avatar-status-dot {
  position: absolute;
  right: -1px;
  bottom: -1px;
  width: 9px;
  height: 9px;
  border: 2px solid #ffffff;
  border-radius: 50%;
  background: #9ca3af;

  &.online { background: #10b981; }
  &.offline { background: #9ca3af; }
  &.busy { background: #f59e0b; }
}

.user-meta {
  min-width: 0;
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}

.user-name,
.user-email {
  max-width: 136px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-name {
  color: #111827;
  font-size: 13px;
  font-weight: 700;
}

.user-email {
  margin-top: 3px;
  color: #6b7280;
  font-size: 12px;
}

.pro-main {
  flex-direction: column;
  min-width: 0;
  background: #ffffff;
}

.pro-content {
  height: 100vh;
  padding: 0;
  overflow: auto;
  background: #ffffff;
}

.dropdown-section {
  display: block;
  margin-bottom: 6px;
  color: #9ca3af;
  font-size: 12px;
  font-weight: 700;
}

.dropdown-select-group {
  min-width: 196px;
  padding: 8px 12px;

  & + & {
    border-top: 1px solid #f1f2f4;
  }
}

.dropdown-select {
  width: 100%;
}

.status-dot {
  display: inline-block;
  flex: 0 0 auto;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  background: #9ca3af;
  box-shadow: 0 0 0 2px #ffffff;

  &.online { background: #10b981; }
  &.offline { background: #9ca3af; }
  &.busy { background: #f59e0b; }
}
</style>
