import type { AuthConfig } from './types'

const TOKEN_KEY = 'tas.access_token'
const VERIFIER_KEY = 'tas.pkce_verifier'
const STATE_KEY = 'tas.oauth_state'

export const apiBase = import.meta.env.VITE_API_BASE ?? '/api'

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (!(init.body instanceof FormData) && init.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const response = await fetch(`${apiBase}${path}`, { ...init, headers })
  if (!response.ok) {
    const text = await response.text()
    throw new Error(text || response.statusText)
  }
  if (response.status === 204) {
    return undefined as T
  }
  return response.json() as Promise<T>
}

export function getJSON<T>(path: string) {
  return apiFetch<T>(path)
}

export function postJSON<T>(path: string, body: unknown) {
  return apiFetch<T>(path, { method: 'POST', body: JSON.stringify(body) })
}

export function patchJSON<T>(path: string, body: unknown) {
  return apiFetch<T>(path, { method: 'PATCH', body: JSON.stringify(body) })
}

export function putJSON<T>(path: string, body: unknown) {
  return apiFetch<T>(path, { method: 'PUT', body: JSON.stringify(body) })
}

export function deleteJSON(path: string) {
  return apiFetch<void>(path, { method: 'DELETE' })
}

export async function uploadFile<T>(path: string, file: File) {
  const form = new FormData()
  form.set('file', file)
  return apiFetch<T>(path, { method: 'POST', body: form })
}

export async function fetchAuthConfig() {
  return apiFetch<AuthConfig>('/auth/config')
}

export async function startLogin(config: AuthConfig) {
  const verifier = randomString(96)
  const state = randomString(32)
  localStorage.setItem(VERIFIER_KEY, verifier)
  localStorage.setItem(STATE_KEY, state)
  const challenge = await pkceChallenge(verifier)
  const params = new URLSearchParams({
    client_id: config.client_id,
    redirect_uri: window.location.origin + window.location.pathname,
    response_type: 'code',
    scope: 'openid email profile',
    state,
    code_challenge: challenge,
    code_challenge_method: 'S256'
  })
  window.location.assign(`${trimSlash(config.issuer)}/oauth2/authorize?${params}`)
}

export async function completeLogin(config: AuthConfig): Promise<boolean> {
  const params = new URLSearchParams(window.location.search)
  const code = params.get('code')
  const state = params.get('state')
  if (!code) {
    return false
  }
  const expectedState = localStorage.getItem(STATE_KEY)
  const verifier = localStorage.getItem(VERIFIER_KEY)
  if (!state || state !== expectedState || !verifier) {
    throw new Error('OAuth state validation failed')
  }
  const body = new URLSearchParams({
    client_id: config.client_id,
    grant_type: 'authorization_code',
    code,
    redirect_uri: window.location.origin + window.location.pathname,
    code_verifier: verifier
  })
  const response = await fetch(`${trimSlash(config.issuer)}/oauth2/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body
  })
  if (!response.ok) {
    throw new Error('Token exchange failed')
  }
  const token = (await response.json()) as { access_token?: string }
  if (!token.access_token) {
    throw new Error('FusionAuth did not return an access token')
  }
  setToken(token.access_token)
  localStorage.removeItem(VERIFIER_KEY)
  localStorage.removeItem(STATE_KEY)
  history.replaceState({}, document.title, window.location.pathname)
  return true
}

function randomString(length: number) {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~'
  const bytes = crypto.getRandomValues(new Uint8Array(length))
  return Array.from(bytes, (byte) => alphabet[byte % alphabet.length]).join('')
}

async function pkceChallenge(verifier: string) {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/g, '')
}

function trimSlash(value: string) {
  return value.replace(/\/+$/g, '')
}
