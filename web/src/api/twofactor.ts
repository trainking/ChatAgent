import request from './request'

export function get2FAConfig() {
  return request.get('/system/2fa/config')
}

export function set2FAConfig(enabled: boolean, issuer: string) {
  return request.put('/system/2fa/config', { enabled, issuer })
}

export function get2FAStatus() {
  return request.get('/auth/2fa/status')
}

export function setup2FA() {
  return request.post('/auth/2fa/setup')
}

export function verify2FASetup(code: string) {
  return request.post('/auth/2fa/verify-setup', { code })
}

export function verify2FALogin(temp_token: string, code: string) {
  return request.post('/auth/2fa/verify', { temp_token, code })
}
