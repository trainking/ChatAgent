import { defineStore } from 'pinia'
import { ref } from 'vue'

interface User {
  id: string
  email: string
  name: string
  role: string
  avatar_url: string
  must_change_password: boolean
}

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const tempToken = ref(localStorage.getItem('temp_token') || '')
  const user = ref<User | null>(JSON.parse(localStorage.getItem('user') || 'null'))

  function setAuth(t: string, u: User) {
    token.value = t
    user.value = u
    localStorage.setItem('token', t)
    localStorage.setItem('user', JSON.stringify(u))
    localStorage.removeItem('temp_token')
  }

  function setTempAuth(t: string) {
    tempToken.value = t
    localStorage.setItem('temp_token', t)
  }

  function clearAuth() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function markPasswordChanged() {
    if (user.value) {
      user.value.must_change_password = false
      localStorage.setItem('user', JSON.stringify(user.value))
    }
  }

  function isLoggedIn() {
    return !!token.value
  }

  function isAdmin() {
    return user.value?.role === 'admin' || user.value?.role === 'super_admin'
  }

  function isSuperAdmin() {
    return user.value?.role === 'super_admin'
  }

  return { token, tempToken, user, setAuth, setTempAuth, clearAuth, markPasswordChanged, isLoggedIn, isAdmin, isSuperAdmin }
})
