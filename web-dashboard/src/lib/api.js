export const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/+$/, '')

export function apiUrl(path = '') {
  if (!path) return API_BASE_URL

  if (/^https?:\/\//i.test(path) || /^ws?:\/\//i.test(path)) {
    return path
  }

  const isApiRoute = /^\/?v1\b|^\/?ws\b/.test(path)
  if (!isApiRoute) {
    return path
  }

  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return API_BASE_URL ? `${API_BASE_URL}${normalizedPath}` : normalizedPath
}

export function wsUrl(path = '') {
  if (!path) return API_BASE_URL ? API_BASE_URL.replace(/^http/, 'ws') : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}`

  if (/^wss?:\/\//i.test(path)) {
    return path
  }

  const isWsRoute = /^\/?v1\/ws\b|^\/?ws\b/.test(path)
  if (!isWsRoute) {
    return path
  }

  const baseUrl = API_BASE_URL ? API_BASE_URL.replace(/^http/, 'ws') : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}`
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  return `${baseUrl}${normalizedPath}`
}

if (typeof window !== 'undefined') {
  const originalFetch = window.fetch.bind(window)

  window.fetch = (input, init) => {
    if (typeof input === 'string') {
      return originalFetch(apiUrl(input), init)
    }

    if (input instanceof URL) {
      const nextUrl = apiUrl(`${input.pathname}${input.search}${input.hash}`)
      return originalFetch(new URL(nextUrl, window.location.origin), init)
    }

    if (input && typeof input === 'object' && 'url' in input) {
      const request = input
      request.url = apiUrl(request.url)
      return originalFetch(request, init)
    }

    return originalFetch(input, init)
  }
}
