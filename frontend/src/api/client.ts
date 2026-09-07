async function parse<T>(res: Response): Promise<T> {
  const text = await res.text()
  let data: unknown
  try {
    data = JSON.parse(text)
  } catch {
    data = { raw: text }
  }
  if (!res.ok) {
    const err = data as { error?: string }
    throw new Error(err.error || res.statusText)
  }
  return data as T
}

export async function api<T>(
  path: string,
  opts: { method?: string; token?: string; body?: unknown } = {},
): Promise<T> {
  const headers: Record<string, string> = {}
  if (opts.body !== undefined) headers['Content-Type'] = 'application/json'
  if (opts.token) headers.Authorization = `Bearer ${opts.token}`
  const res = await fetch(path, {
    method: opts.method || 'GET',
    headers,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
  })
  return parse<T>(res)
}

export async function apiForm<T>(path: string, token: string, form: FormData): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: form,
  })
  return parse<T>(res)
}

export function statusLabel(s?: string) {
  return (
    {
      none: 'не подтверждён',
      pending: 'на проверке',
      verified: 'подтверждён',
      rejected: 'отклонён',
      expired: 'истёк',
    }[s || ''] || s || '—'
  )
}
