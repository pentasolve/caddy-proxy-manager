let accessToken: string | null = null
let tokenExpiresAt: number = 0

let isRefreshing = false
let refreshPromise: Promise<string | null> | null = null

export async function getValidToken(): Promise<string | null> {
  if (accessToken && tokenExpiresAt > Date.now() + 60000) {
    return accessToken
  }

  return await refreshAccessToken()
}

async function refreshAccessToken(): Promise<string | null> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise
  }

  isRefreshing = true
  refreshPromise = doRefresh()

  try {
    return await refreshPromise
  } finally {
    isRefreshing = false
    refreshPromise = null
  }
}

async function doRefresh(): Promise<string | null> {
  try {
    const res = await fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({})
    })

    if (!res.ok) {
      clearTokens()
      return null
    }

    const data = await res.json()

    accessToken = data.access_token
    tokenExpiresAt = Date.now() + (data.expires_in * 1000)

    return accessToken
  } catch (error) {
    console.error('Token refresh failed:', error)
    clearTokens()
    return null
  }
}

function clearTokens(): void {
  accessToken = null
  tokenExpiresAt = 0
}

export function setTokens(token: string, expiresIn: number): void {
  accessToken = token
  tokenExpiresAt = Date.now() + (expiresIn * 1000)
}

export async function logout(): Promise<void> {
  try {
    await fetch('/api/auth/logout', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({})
    })
  } catch (e) {
  }

  clearTokens()

  window.location.href = '/login'
}

export async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = await getValidToken()

  if (!token) {
    window.location.href = '/login'
    throw new Error('Not authenticated')
  }

  const headers = new Headers(options.headers)
  headers.set('Authorization', `Bearer ${token}`)

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include'
  })

  if (response.status === 401) {
    const newToken = await refreshAccessToken()
    if (newToken) {
      headers.set('Authorization', `Bearer ${newToken}`)
      return fetch(url, {
        ...options,
        headers,
        credentials: 'include'
      })
    }
    window.location.href = '/login'
  }

  return response
}

export function isAuthenticated(): boolean {
  return accessToken !== null && tokenExpiresAt > Date.now()
}

export async function initAuth(): Promise<boolean> {
  const token = await refreshAccessToken()
  return token !== null
}

let refreshTimer: number | null = null

export function startAutoRefresh(): void {
  stopAutoRefresh()

  const checkAndRefresh = async () => {
    if (tokenExpiresAt > 0 && tokenExpiresAt - Date.now() < 2 * 60 * 1000) {
      await refreshAccessToken()
    }
  }

  refreshTimer = window.setInterval(checkAndRefresh, 60 * 1000)
}

export function stopAutoRefresh(): void {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}
