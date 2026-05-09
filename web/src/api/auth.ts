import request from './request'

export function getStatus() {
  return request.get('/auth/status')
}

export function initRoot(email: string, password: string, name: string) {
  return request.post('/auth/init', { email, password, name })
}

export function login(email: string, password: string) {
  return request.post('/auth/login', { email, password })
}

export function changePassword(old_password: string, new_password: string) {
  return request.post('/auth/change-password', { old_password, new_password })
}

export function getUsers() {
  return request.get('/users')
}

export function createUser(email: string, name: string, password: string, role: string) {
  return request.post('/users', { email, name, password, role })
}

export function updateUser(id: string, data: { name?: string; role?: string; status?: string; must_change_password?: boolean }) {
  return request.put(`/users/${id}`, data)
}

export function resetPassword(id: string) {
  return request.put(`/users/${id}/reset-password`)
}

export function deleteUser(id: string) {
  return request.delete(`/users/${id}`)
}
