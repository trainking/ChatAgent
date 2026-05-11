import request from './request'

export function getConversations(params?: Record<string, unknown>) {
  return request.get('/conversations', { params })
}

export function getConversation(id: string) {
  return request.get(`/conversations/${id}`)
}

export function assignConversation(id: string, assigneeId: string) {
  return request.put(`/conversations/${id}/assign`, { assignee_id: assigneeId })
}

export function changeStatus(id: string, status: string) {
  return request.put(`/conversations/${id}/status`, { status })
}

export function changePriority(id: string, priority: string) {
  return request.put(`/conversations/${id}/priority`, { priority })
}

export function getUnreadCount() {
  return request.get('/conversations/unread-count')
}
