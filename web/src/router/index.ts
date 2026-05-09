import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/login/index.vue'),
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
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
