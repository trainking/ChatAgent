import request from './request'

export function getContacts(params?: Record<string, unknown>) {
  return request.get('/contacts', { params })
}

export function createContact(data: Record<string, unknown>) {
  return request.post('/contacts', data)
}

export function getContact(id: string) {
  return request.get(`/contacts/${id}`)
}

export function updateContact(id: string, data: Record<string, unknown>) {
  return request.put(`/contacts/${id}`, data)
}

export function mergeContact(id: string, targetId: string) {
  return request.post(`/contacts/${id}/merge`, { target_id: targetId })
}

export function getContactConversations(id: string) {
  return request.get(`/contacts/${id}/conversations`)
}
