type QueuedMutation = { id?: number; url: string; method: string; headers: Record<string, string>; body?: string; createdAt: string; blocked?: boolean }

const DB_NAME = 'line-offline'
const STORE = 'mutations'
const syncEvent = 'line:sync-status'

function openDB() {
  return new Promise<IDBDatabase>((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, 1)
    request.onupgradeneeded = () => request.result.createObjectStore(STORE, { keyPath: 'id', autoIncrement: true })
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

async function listMutations() {
  const db = await openDB()
  return new Promise<QueuedMutation[]>((resolve, reject) => {
    const request = db.transaction(STORE, 'readonly').objectStore(STORE).getAll()
    request.onsuccess = () => resolve(request.result as QueuedMutation[])
    request.onerror = () => reject(request.error)
  })
}

async function addMutation(item: QueuedMutation) {
  const db = await openDB()
  await new Promise<void>((resolve, reject) => {
    const request = db.transaction(STORE, 'readwrite').objectStore(STORE).add(item)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

async function removeMutation(id: number) {
  const db = await openDB()
  await new Promise<void>((resolve, reject) => {
    const request = db.transaction(STORE, 'readwrite').objectStore(STORE).delete(id)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error)
  })
}

function emit(detail: { pending: number; blocked?: boolean }) { window.dispatchEvent(new CustomEvent(syncEvent, { detail })) }

export async function pendingMutationCount() { return (await listMutations()).length }

export async function syncOfflineMutations() {
  if (!navigator.onLine) return
  const queued = await listMutations()
  for (const item of queued) {
    if (!item.id || item.blocked) break
    try {
      const response = await window.__lineNativeFetch!(item.url, { method: item.method, headers: item.headers, body: item.body })
      if (response.ok) { await removeMutation(item.id); continue }
      if (response.status === 409) { item.blocked = true; const db = await openDB(); db.transaction(STORE, 'readwrite').objectStore(STORE).put(item); emit({ pending: await pendingMutationCount(), blocked: true }); return }
      if (response.status >= 400 && response.status < 500) { item.blocked = true; const db = await openDB(); db.transaction(STORE, 'readwrite').objectStore(STORE).put(item); emit({ pending: await pendingMutationCount(), blocked: true }); return }
      return
    } catch { return }
  }
  emit({ pending: await pendingMutationCount() })
}

declare global { interface Window { __lineNativeFetch?: typeof fetch } }

export function installOfflineSupport() {
  if (window.__lineNativeFetch) return
  window.__lineNativeFetch = window.fetch.bind(window)
  window.fetch = async (input, init = {}) => {
    try { return await window.__lineNativeFetch!(input, init) } catch (error) {
      const url = typeof input === 'string' ? input : input instanceof Request ? input.url : input.toString()
      const method = (init.method ?? (input instanceof Request ? input.method : 'GET')).toUpperCase()
      const body = typeof init.body === 'string' ? init.body : undefined
      const headers = Object.fromEntries(new Headers(init.headers ?? (input instanceof Request ? input.headers : undefined)).entries())
      const queueable = method !== 'GET' && url.includes('/api/') && !url.includes('/api/auth/') && !url.includes('/api/media/') && body !== undefined && headers['content-type']?.includes('application/json')
      if (!queueable) throw error
      await addMutation({ url, method, headers, body, createdAt: new Date().toISOString() })
      emit({ pending: await pendingMutationCount() })
      return new Response(JSON.stringify({ queued: true }), { status: 202, headers: { 'Content-Type': 'application/json' } })
    }
  }
  window.addEventListener('online', () => { void syncOfflineMutations() })
  window.addEventListener('line:sync-now', () => { void syncOfflineMutations() })
  void navigator.serviceWorker.register('/sw.js')
  void syncOfflineMutations()
}

export { syncEvent }
