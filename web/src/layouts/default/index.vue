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
          <span>收件箱</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="layout-header">
        <span class="header-title">客服系统</span>
        <el-dropdown>
          <span class="user-info">
            <el-avatar :size="32" icon="UserFilled" />
            <span class="user-name">管理员</span>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

function handleLogout() {
  localStorage.removeItem('token')
  router.push('/login')
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
