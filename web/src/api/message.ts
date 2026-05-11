import request from './request'

export function getMessages(convId: string, params?: Record<string, unknown>) {
  return request.get(`/conversations/${convId}/messages`, { params })
}

export function sendMessage(convId: string, data: Record<string, unknown>) {
  return request.post(`/conversations/${convId}/messages`, data)
}

export function retryMessage(convId: string, msgId: string) {
  return request.post(`/conversations/${convId}/messages/${msgId}/retry`)
}
