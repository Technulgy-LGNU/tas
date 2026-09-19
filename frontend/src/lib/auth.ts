import { ref } from 'vue'

export interface User {
  id: string
  email: string
  username: string
  roles: string[] | null
  localDevelopment?: boolean
}

export const currentUser = ref<User | null>(null)

export function safeReturnTo(value: unknown): string {
  return typeof value === 'string' && value.startsWith('/') && !value.startsWith('//') &&
    !/[\\\r\n]/.test(value) && !value.startsWith('/login') ? value : '/'
}

// Use this helper for private API requests so expired sessions also leave the current page.
export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  if (!path.startsWith('/api/')) throw new Error('API requests must use a local /api/ path.')
  const headers = new Headers(init.headers)
  if (!['GET', 'HEAD', 'OPTIONS'].includes((init.method ?? 'GET').toUpperCase())) {
    headers.set('X-TAS-CSRF', '1')
  }
  const response = await fetch(path, { ...init, headers, credentials: 'same-origin', cache: 'no-store' })
  if (response.status === 401) {
    currentUser.value = null
    const returnTo = safeReturnTo(window.location.hash.slice(1))
    if (!window.location.hash.startsWith('#/login')) {
      window.location.hash = `/login?returnTo=${encodeURIComponent(returnTo)}`
    }
  }
  return response
}

export async function checkSession(): Promise<boolean> {
  const response = await fetch('/api/v1/auth/me', { credentials: 'same-origin', cache: 'no-store' })
  if (response.status === 401) {
    currentUser.value = null
    return false
  }
  if (!response.ok) throw new Error('The sign-in service is unavailable. Please try again.')
  const data: { user: User } = await response.json()
  currentUser.value = data.user
  return true
}

export function login(returnTo: unknown): void {
  window.location.assign(`/auth/login?returnTo=${encodeURIComponent(safeReturnTo(returnTo))}`)
}

export async function logout(): Promise<void> {
  const response = await apiFetch('/api/v1/auth/logout', { method: 'POST' })
  if (!response.ok) throw new Error('Could not sign out. Please try again.')
  const data: { logout_url: string } = await response.json()
  currentUser.value = null
  window.location.assign(data.logout_url)
}
