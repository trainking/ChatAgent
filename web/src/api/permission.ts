import request from './request'

export function getPermissions() {
  return request.get('/permissions')
}

export function getRolePermissions(role: string) {
  return request.get(`/roles/${role}/permissions`)
}

export function setRolePermissions(role: string, permissions: string[]) {
  return request.put(`/roles/${role}/permissions`, { permissions })
}
