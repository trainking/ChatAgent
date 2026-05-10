import request from './request'

export function getProfile() {
  return request.get('/profile')
}

export function updateProfile(name: string) {
  return request.put('/profile', { name })
}

export function uploadAvatar(file: File) {
  const formData = new FormData()
  formData.append('avatar', file)
  return request.post('/profile/avatar', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function getActivities(page: number = 1, pageSize: number = 20) {
  return request.get('/profile/activities', { params: { page, page_size: pageSize } })
}

export function updateOnlineStatus(onlineStatus: string) {
  return request.put('/profile/status', { online_status: onlineStatus })
}

export function logout() {
  return request.post('/auth/logout')
}
