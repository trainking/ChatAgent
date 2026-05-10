import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { getStatus } from '@/api/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/init',
    name: 'Init',
    component: () => import('@/pages/init/index.vue'),
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/login/index.vue'),
  },
  {
    path: '/setup-2fa',
    name: 'Setup2FA',
    component: () => import('@/pages/setup-2fa/index.vue'),
  },
  {
    path: '/change-password',
    name: 'ChangePassword',
    component: () => import('@/pages/change-password/index.vue'),
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/layouts/default/index.vue'),
    redirect: '/dashboard/inbox',
    children: [
      {
        path: 'inbox',
        name: 'Inbox',
        component: () => import('@/pages/inbox/index.vue'),
      },
      {
        path: 'inbox/:id',
        name: 'InboxDetail',
        component: () => import('@/pages/inbox/detail.vue'),
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/pages/profile/index.vue'),
        meta: { title: 'profile.title' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/pages/users/index.vue'),
      },
      {
        path: 'system',
        name: 'System',
        redirect: '/dashboard/system/2fa',
        children: [
          {
            path: 'roles',
            name: 'Roles',
            component: () => import('@/pages/roles/index.vue'),
          },
          {
            path: '2fa',
            name: 'System2FA',
            component: () => import('@/pages/system/index.vue'),
          },
        ],
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

let statusChecked = false
let systemInitialized = false

router.beforeEach(async (to, _from, next) => {
  if (to.path === '/init' || to.path === '/setup-2fa') {
    next()
    return
  }

  const token = localStorage.getItem('token')

  if (!token) {
    if (!statusChecked) {
      try {
        const res = await getStatus()
        systemInitialized = res.data.initialized
        statusChecked = true
      } catch {
        systemInitialized = false
      }
    }

    if (!systemInitialized) {
      next('/init')
      return
    }

    if (to.path === '/login') {
      next()
      return
    }

    next('/login')
    return
  }

  if (to.path === '/login' || to.path === '/init') {
    next('/dashboard')
    return
  }

  const user = JSON.parse(localStorage.getItem('user') || '{}')
  if (user.must_change_password && to.path !== '/change-password') {
    next('/change-password')
    return
  }

  if (!user.must_change_password && to.path === '/change-password') {
    next('/dashboard')
    return
  }

  next()
})

export default router
