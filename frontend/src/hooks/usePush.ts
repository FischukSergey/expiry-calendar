import { useCallback, useEffect, useState } from 'react'

import { getVapidPublic, subscribePush, unsubscribePush } from '../api/endpoints.ts'
import { useAuth } from './useAuth.ts'

function urlBase64ToUint8Array(raw: string): BufferSource {
  const padding = '='.repeat((4 - (raw.length % 4)) % 4)
  const base64 = (raw + padding).replace(/-/g, '+').replace(/_/g, '/')
  const binary = atob(base64)
  const out = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    out[i] = binary.charCodeAt(i)
  }
  return out
}

function pushSupported(): boolean {
  return 'serviceWorker' in navigator && 'PushManager' in window && 'Notification' in window
}

async function currentSubscription(): Promise<PushSubscription | null> {
  const reg = await navigator.serviceWorker.ready
  return reg.pushManager.getSubscription()
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
  const reg = await navigator.serviceWorker.ready
  let sub = await reg.pushManager.getSubscription()
  if (!sub) {
    sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(public_key),
    })
  }
  await sendSubscription(sub)
  return 'granted'
}

export async function disablePush(): Promise<void> {
  if (!pushSupported()) {
    return
  }
  const sub = await currentSubscription()
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
      } catch {
        setPermission(Notification.permission)
      }
    })()
  }, [user])

  const request = useCallback(async () => {
    setBusy(true)
    try {
      const perm = await enablePush()
      setPermission(perm)
    } finally {
      setBusy(false)
    }
  }, [])

  return { supported: pushSupported(), permission, busy, request }
}
