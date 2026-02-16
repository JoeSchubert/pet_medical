import { logApi } from '../lib/log'

const API_BASE = '/api'

let accessToken: string | null = null

/** Store access token so we can send it in Authorization header (e.g. for uploads where cookies may not be sent on some mobile browsers). */
export function setAccessToken(token: string) {
  accessToken = token
  logApi('setAccessToken stored')
}
export function clearAccessToken() {
  accessToken = null
  logApi('clearAccessToken')
}
export function getAccessToken(): string | null {
  return accessToken
}

function authHeaders(): Record<string, string> {
  const t = getAccessToken()
  return t ? { Authorization: `Bearer ${t}` } : {}
}

export async function fetchApi(
  path: string,
  options: RequestInit = {}
): Promise<Response> {
  const url = path.startsWith('http') ? path : `${API_BASE}${path}`
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...authHeaders(),
    ...(options.headers as Record<string, string>),
  }
  logApi('fetch', options.method ?? 'GET', path)
  const res = await fetch(url, { ...options, credentials: 'include', headers })
  logApi('fetch response', path, res.status, res.statusText)
  if (res.status === 401) {
    logApi('got 401, attempting refresh')
    const refreshed = await refreshToken()
    if (refreshed) return fetch(url, { ...options, credentials: 'include', headers: { ...headers, ...authHeaders(), ...(options.headers as Record<string, string>) } })
  }
  return res
}

/** Call refresh endpoint; cookies are set by server. Returns true if new cookies were set. Updates in-memory token from response. */
async function refreshToken(): Promise<boolean> {
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  })
  if (!res.ok) return false
  try {
    const data = (await res.json()) as { access_token?: string }
    if (data.access_token) setAccessToken(data.access_token)
  } catch {
    /* ignore */
  }
  return true
}

/** Refresh and return user (for initial load when access cookie may be expired). Stores new access token. */
export async function refresh(): Promise<{ access_token: string; expires_in: number; user: User }> {
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  })
  if (!res.ok) throw new Error('Refresh failed')
  const data = await res.json() as { access_token: string; expires_in: number; user: User }
  if (data.access_token) setAccessToken(data.access_token)
  return data
}

export interface User {
  id: string
  display_name: string
  email: string
  role: string
  weight_unit?: 'lbs' | 'kg'
  currency?: string
  language?: string
}

export interface UserListEntry {
  id: string
  display_name: string
  email: string
  role: string
  created_at: string
}

export interface LoginResponse {
  access_token: string
  expires_in: number
  user: User
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  logApi('login POST', email)
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  logApi('login response', res.status, res.statusText)
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    const msg = (err as { error?: string }).error || 'Login failed'
    logApi('login failed', msg)
    throw new Error(msg)
  }
  const data: LoginResponse = await res.json()
  logApi('login success', 'user=', data.user?.display_name, 'hasAccessToken=', !!data.access_token)
  setAccessToken(data.access_token)
  return data
}

export async function logout(): Promise<void> {
  await fetch(`${API_BASE}/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  })
  clearAccessToken()
}

export async function changePassword(currentPassword: string, newPassword: string): Promise<void> {
  const res = await fetchApi('/auth/change-password', {
    method: 'PUT',
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    const msg = (err as { error?: string }).error || 'Failed to change password'
    throw new Error(msg)
  }
}

export async function getMe(): Promise<User> {
  logApi('getMe() called')
  const res = await fetchApi('/auth/me')
  logApi('getMe() response', res.status)
  if (!res.ok) throw new Error('Not authenticated')
  const user = await res.json()
  logApi('getMe() success', user?.display_name)
  return user
}

export interface Pet {
  id: string
  user_id: string
  name: string
  species?: string
  breed?: string
  date_of_birth?: string
  gender?: string
  fixed?: boolean
  color?: string
  microchip_id?: string
  microchip_company?: string
  notes?: string
  photo_url?: string
  created_at: string
  updated_at: string
}

export interface Vaccination {
  id: string
  pet_id: string
  name: string
  administered_at: string
  next_due?: string
  cost_usd?: number
  veterinarian?: string
  batch_number?: string
  notes?: string
  created_at: string
  updated_at: string
}

export interface WeightEntry {
  id: string
  pet_id: string
  weight_lbs: number
  entry_unit?: 'lbs' | 'kg'
  measured_at: string
  approximate?: boolean
  notes?: string
  created_at: string
}

export interface Document {
  id: string
  pet_id: string
  name: string
  doc_type?: string
  file_path: string
  file_size?: number
  mime_type?: string
  notes?: string
  created_at: string
}

export const petsApi = {
  list: () =>
    fetchApi('/pets')
      .then((r) => r.json())
      .then((data: Pet[] | null) => (Array.isArray(data) ? data : [])) as Promise<Pet[]>,
  get: (id: string) => fetchApi(`/pets/${id}`).then((r) => r.json()) as Promise<Pet>,
  create: (body: Partial<Pet>) =>
    fetchApi('/pets', { method: 'POST', body: JSON.stringify(body) }).then((r) => r.json()) as Promise<Pet>,
  update: (id: string, body: Partial<Pet>) =>
    fetchApi(`/pets/${id}`, { method: 'PUT', body: JSON.stringify(body) }).then((r) => r.json()) as Promise<Pet>,
  delete: (id: string) => fetchApi(`/pets/${id}`, { method: 'DELETE' }),
}

export const vaccinationsApi = {
  list: (petId: string) =>
    fetchApi(`/pets/${petId}/vaccinations`)
      .then((r) => r.json())
      .then((data: Vaccination[] | null) => (Array.isArray(data) ? data : [])) as Promise<Vaccination[]>,
  create: (petId: string, body: Partial<Vaccination>) =>
    fetchApi(`/pets/${petId}/vaccinations`, { method: 'POST', body: JSON.stringify(body) }).then((r) =>
      r.json()
    ) as Promise<Vaccination>,
  update: (petId: string, id: string, body: Partial<Vaccination>) =>
    fetchApi(`/pets/${petId}/vaccinations/${id}`, { method: 'PUT', body: JSON.stringify(body) }).then((r) =>
      r.json()
    ) as Promise<Vaccination>,
  delete: (petId: string, id: string) =>
    fetchApi(`/pets/${petId}/vaccinations/${id}`, { method: 'DELETE' }),
}

export const weightsApi = {
  list: (petId: string) =>
    fetchApi(`/pets/${petId}/weights`)
      .then((r) => r.json())
      .then((data: WeightEntry[] | null) => (Array.isArray(data) ? data : [])) as Promise<WeightEntry[]>,
  create: (petId: string, body: { weight_lbs?: number; weight_kg?: number; entry_unit?: WeightUnit; measured_at: string; approximate?: boolean; notes?: string }) =>
    fetchApi(`/pets/${petId}/weights`, { method: 'POST', body: JSON.stringify(body) }).then((r) =>
      r.json()
    ) as Promise<WeightEntry>,
  delete: (petId: string, id: string) =>
    fetchApi(`/pets/${petId}/weights/${id}`, { method: 'DELETE' }),
}

export interface PetPhoto {
  id: string
  pet_id: string
  file_path: string
  display_order: number
  created_at: string
}

export const usersApi = {
  list: () =>
    fetchApi('/users')
      .then((r) => r.json())
      .then((data: UserListEntry[] | null) => (Array.isArray(data) ? data : [])) as Promise<UserListEntry[]>,
  create: (displayName: string, email: string, password: string, role: string) =>
    fetchApi('/users', {
      method: 'POST',
      body: JSON.stringify({ display_name: displayName, email, password, role: role || 'user' }),
    }).then((r) => {
      if (!r.ok) return r.json().then((err: { error?: string }) => { throw new Error(err.error || 'Failed to create user') })
      return r.json()
    }) as Promise<{ id: string; display_name: string; email: string; role: string }>,
  updateRole: (userId: string, role: string) =>
    fetchApi(`/users/${userId}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    }),
}

export const photosApi = {
  list: (petId: string) =>
    fetchApi(`/pets/${petId}/photos`)
      .then((r) => r.json())
      .then((data: PetPhoto[] | null) => (Array.isArray(data) ? data : [])) as Promise<PetPhoto[]>,
  upload: async (petId: string, file: File): Promise<PetPhoto> => {
    const form = new FormData()
    form.append('file', file)
    const headers: HeadersInit = { ...authHeaders() }
    const res = await fetch(`${API_BASE}/pets/${petId}/photos`, {
      method: 'POST',
      credentials: 'include',
      headers,
      body: form,
    })
    if (res.status === 401) {
      const refreshed = await refreshToken()
      if (refreshed) {
        const retry = await fetch(`${API_BASE}/pets/${petId}/photos`, {
          method: 'POST',
          credentials: 'include',
          headers: { ...authHeaders() },
          body: form,
        })
        if (!retry.ok) {
          const body = await retry.json().catch(() => ({}))
          throw new Error((body as { error?: string }).error || 'Upload failed')
        }
        return retry.json()
      }
    }
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      const msg = (body as { error?: string }).error || 'Upload failed'
      throw new Error(msg)
    }
    return res.json()
  },
  setAvatar: (petId: string, photoId: string) =>
    fetchApi(`/pets/${petId}/photos/${photoId}/avatar`, { method: 'PUT' }).then((r) => r.json()),
  delete: (petId: string, id: string) => fetchApi(`/pets/${petId}/photos/${id}`, { method: 'DELETE' }),
}

export const documentsApi = {
  list: (petId: string, opts?: { sort?: 'date' | 'name'; search?: string }) => {
    const params = new URLSearchParams()
    if (opts?.sort) params.set('sort', opts.sort)
    if (opts?.search) params.set('search', opts.search)
    const q = params.toString()
    return fetchApi(`/pets/${petId}/documents${q ? `?${q}` : ''}`)
      .then((r) => r.json())
      .then((data: Document[] | null) => (Array.isArray(data) ? data : [])) as Promise<Document[]>
  },
  create: async (petId: string, name: string, file: File): Promise<Document> => {
    const form = new FormData()
    form.append('name', name)
    form.append('file', file)
    const headers: HeadersInit = { ...authHeaders() }
    const res = await fetch(`${API_BASE}/pets/${petId}/documents`, {
      method: 'POST',
      credentials: 'include',
      headers,
      body: form,
    })
    if (res.status === 401) {
      const refreshed = await refreshToken()
      if (refreshed) {
        const retry = await fetch(`${API_BASE}/pets/${petId}/documents`, {
          method: 'POST',
          credentials: 'include',
          headers: { ...authHeaders() },
          body: form,
        })
        if (!retry.ok) {
          const body = await retry.json().catch(() => ({}))
          throw new Error((body as { error?: string }).error || 'Upload failed')
        }
        return retry.json()
      }
    }
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      const msg = (body as { error?: string }).error || 'Upload failed'
      throw new Error(msg)
    }
    return res.json()
  },
  update: (petId: string, id: string, name: string) =>
    fetchApi(`/pets/${petId}/documents/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) }).then((r) =>
      r.json()
    ) as Promise<Document>,
  delete: (petId: string, id: string) =>
    fetchApi(`/pets/${petId}/documents/${id}`, { method: 'DELETE' }),
}

export type WeightUnit = 'lbs' | 'kg'

export interface Settings {
  weight_unit: WeightUnit
  currency: string
  language: string
  email?: string
  display_name?: string
  role?: string
  is_only_admin?: boolean
}

export const settingsApi = {
  get: () => fetchApi('/settings').then((r) => r.json()) as Promise<Settings>,
  update: (body: Settings) =>
    fetchApi('/settings', { method: 'PUT', body: JSON.stringify(body) }).then((r) => r.json()) as Promise<Settings>,
  getForUser: (userId: string) =>
    fetchApi(`/users/${userId}/settings`).then((r) => r.json()) as Promise<Settings>,
  updateForUser: (userId: string, body: Settings) =>
    fetchApi(`/users/${userId}/settings`, { method: 'PUT', body: JSON.stringify(body) }).then((r) =>
      r.json()
    ) as Promise<Settings>,
}

export interface CustomOptionsResponse {
  species: string[]
  breeds: Record<string, string[]>
  vaccinations: Record<string, string[]>
  /** species -> vaccine name -> duration in months (for expiry hint) */
  vaccination_durations?: Record<string, Record<string, number>>
}

export const customOptionsApi = {
  get: () => fetchApi('/custom-options').then((r) => r.json()) as Promise<CustomOptionsResponse>,
  add: (optionType: 'species' | 'breed' | 'vaccination', value: string, context?: string) =>
    fetchApi('/custom-options', {
      method: 'POST',
      body: JSON.stringify({ option_type: optionType, value, context: context ?? '' }),
    }),
}

export interface DefaultOptionItem {
  id: string
  option_type: string
  value: string
  context: string
  sort_order: number
  duration_months?: number
}

export const defaultOptionsApi = {
  list: () =>
    fetchApi('/admin/default-options').then((r) => r.json()) as Promise<DefaultOptionItem[]>,
  create: (body: Omit<DefaultOptionItem, 'id'>) =>
    fetchApi('/admin/default-options', {
      method: 'POST',
      body: JSON.stringify(body),
    }).then((r) => r.json()) as Promise<DefaultOptionItem>,
  update: (id: string, body: Partial<DefaultOptionItem>) =>
    fetchApi(`/admin/default-options/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }).then((r) => r.json()) as Promise<DefaultOptionItem>,
  delete: (id: string) =>
    fetchApi(`/admin/default-options/${id}`, { method: 'DELETE' }),
}
