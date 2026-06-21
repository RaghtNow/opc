const API_BASE = import.meta.env.VITE_EDU_INSIGHT_API_BASE ?? 'http://127.0.0.1:8088/api'

export type WorkIdentity = {
  id: string
  roleType: string
  roleLabel: string
  primaryLabel: string
  secondaryLabel: string
  scopeType: string
  scopeId: string
  subject: string
}

export type CurrentUser = {
  user: {
    id: string
    name: string
    mobile: string
    status: string
    createdAt: string
  }
  defaultRoleId: string
  workIdentities: WorkIdentity[]
}

export type SendSMSCodeResponse = {
  mobile: string
  scene: string
  expiresIn: number
  devCode?: string
  message: string
}

export type LoginResponse = {
  token: string
  me: CurrentUser
}

export async function sendSMSCode(mobile: string): Promise<SendSMSCodeResponse> {
  const response = await fetch(`${API_BASE}/auth/sms-code`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ mobile, scene: 'login' })
  })
  if (!response.ok) throw await responseError(response, '发送验证码失败')
  return response.json()
}

export async function loginWithSMS(mobile: string, code: string): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE}/auth/login/sms`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ mobile, code, scene: 'login' })
  })
  if (!response.ok) throw await responseError(response, '登录失败')
  return response.json()
}

export async function fetchCurrentUser(token: string): Promise<CurrentUser> {
  const response = await fetch(`${API_BASE}/auth/me`, {
    headers: {
      Authorization: `Bearer ${token}`
    }
  })
  if (!response.ok) throw await responseError(response, '登录状态已失效')
  return response.json()
}

async function responseError(response: Response, fallback: string) {
  const data = await response.json().catch(() => ({}))
  return new Error(data.error ?? data.message ?? fallback)
}
