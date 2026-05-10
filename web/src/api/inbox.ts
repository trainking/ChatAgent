import request from './request'

export function getInboxes() {
  return request.get('/inboxes')
}

export function getInbox(id: string) {
  return request.get(`/inboxes/${id}`)
}

export function createInbox(data: {
  name: string
  description?: string
  welcome_title?: string
  welcome_message?: string
  inbox_type: string
  status?: string
  collaborators?: string[]
}) {
  return request.post('/inboxes', data)
}

export function updateInbox(id: string, data: {
  description?: string
  welcome_title?: string
  welcome_message?: string
  status?: string
  collaborators?: string[]
}) {
  return request.put(`/inboxes/${id}`, data)
}

export function deleteInbox(id: string) {
  return request.delete(`/inboxes/${id}`)
}
