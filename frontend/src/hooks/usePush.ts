import { useCallback, useEffect, useState } from 'react'

import { getVapidPublic, subscribePush, unsubscribePush } from '../api/endpoints.ts'
import { useAuth } from './useAuth.ts'

const swReadyMs = 10_000

function urlBase64ToUint8Array(raw: string): Uint8Array {
  const padding = '='.repeat((4 - (raw.length % 4)) % 4)
  const base64 = (raw + padding).replace(/-/g, '+').replace(/_/g, '/')
  const binary = atob(base64)
  const out = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    out[i] = binary.charCodeAt(i)
  }
  return out
}

function asBytes(raw: BufferSource): Uint8Array {
  if (raw instanceof Uint8Array) {
    return raw
  }
  if (raw instanceof ArrayBuffer) {
    return new Uint8Array(raw)
  }
  return new Uint8Array(raw.buffer, raw.byteOffset, raw.byteLength)
}

function sameApplicationKey(sub: PushSubscription, key: Uint8Array): boolean {
  const current = sub.options.applicationServerKey
  if (!current) {
    return false
  }
  const left = asBytes(current)
  if (left.length !== key.length) {
    return false
  }
  return left.every((b, i) => b === key[i])
}

function pushSupported(): boolean {
  return 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window
}

export function isStandalonePWA(): boolean {
  const nav = navigator as Navigator & { standalone?: boolean }
  return window.matchMedia('(display-mode: standalone)').matches || nav.standalone === true
}

export function isIOSDevice(): boolean {
  return /iPad|iPhone|iPod/.test(navigator.userAgent)
}

async function swRegistration(): Promise<ServiceWorkerRegistration> {
  const existing = await navigator.serviceWorker.getRegistration()
  if (!existing) {
    throw new Error('Нет service worker. Откройте установленное приложение, не вкладку Vite.')
  }
  const ready = navigator.serviceWorker.ready
  const timeout = new Promise<never>((_, reject) => {
    window.setTimeout(() => reject(new Error('Service worker не готов')), swReadyMs)
  })
  return Promise.race([ready, timeout])
}

async function sendSubscription(sub: PushSubscription): Promise<void> {
  const json = sub.toJSON()
  const endpoint = json.endpoint
  const p256dh = json.keys?.p256dh
  const auth = json.keys?.auth
  if (!endpoint || !p256dh || !auth) {
    throw new Error('неполная подписка')
  }
  await subscribePush({ endpoint, keys: { p256dh, auth } })
}

async function currentSubscription(): Promise<PushSubscription | null> {
  const reg = await swRegistration()
  return reg.pushManager.getSubscription()
}

export async function enablePush(): Promise<NotificationPermission> {
  if (!pushSupported()) {
    return 'denied'
  }
  const perm = Notification.permission === 'granted' ? 'granted' : await Notification.requestPermission()
  if (perm !== 'granted') {
    return perm
  }
  const { public_key } = await getVapidPublic()
  if (!public_key) {
    throw new Error('нет VAPID')
  }
  const key = urlBase64ToUint8Array(public_key)
  const reg = await swRegistration()
  let sub = await reg.pushManager.getSubscription()
  if (sub && !sameApplicationKey(sub, key)) {
    await sub.unsubscribe()
    sub = null
  }
  if (!sub) {
    sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: key as BufferSource,
    })
  }
  await sendSubscription(sub)
  return 'granted'
}

export async function disablePush(): Promise<void> {
  if (!pushSupported()) {
    return
  }
  let sub: PushSubscription | null = null
  try {
    sub = await currentSubscription()
  } catch {
    return
  }
  if (!sub) {
    return
  }
  try {
    await unsubscribePush(sub.endpoint)
  } finally {
    await sub.unsubscribe()
  }
}

export function usePush() {
  const { user } = useAuth()
  const [permission, setPermission] = useState<NotificationPermission>(() =>
    pushSupported() ? Notification.permission : 'denied',
  )
  const [busy, setBusy] = useState(false)
  const [subscribed, setSubscribed] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!user || !pushSupported()) {
      return
    }
    if (Notification.permission === 'denied') {
      setPermission('denied')
      return
    }
    void (async () => {
      try {
        const perm = await enablePush()
        setPermission(perm)
        setSubscribed(perm === 'granted')
        setError(null)
      } catch (err) {
        setPermission(Notification.permission)
        setSubscribed(false)
        setError(err instanceof Error ? err.message : 'Не удалось включить пуши')
      }
    })()
  }, [user])

  const request = useCallback(async () => {
    setBusy(true)
    setError(null)
    try {
      const perm = await enablePush()
      setPermission(perm)
      setSubscribed(perm === 'granted')
      if (perm !== 'granted') {
        setError('Разрешение не выдано')
      }
    } catch (err) {
      setSubscribed(false)
      setError(err instanceof Error ? err.message : 'Не удалось включить пуши')
    } finally {
      setBusy(false)
    }
  }, [])

  const disable = useCallback(async () => {
    setBusy(true)
    try {
      await disablePush()
      setSubscribed(false)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Не удалось отписаться')
    } finally {
      setBusy(false)
    }
  }, [])

  return { supported: pushSupported(), permission, subscribed, busy, error, request, disable }
}
