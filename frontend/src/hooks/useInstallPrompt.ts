import { useCallback, useEffect, useState } from 'react'

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed' }>
}

let deferred: BeforeInstallPromptEvent | null = null
const listeners = new Set<() => void>()

function notify(): void {
  for (const fn of listeners) {
    fn()
  }
}

export function listenInstallPrompt(): void {
  window.addEventListener('beforeinstallprompt', (event) => {
    event.preventDefault()
    deferred = event as BeforeInstallPromptEvent
    notify()
  })
  window.addEventListener('appinstalled', () => {
    deferred = null
    notify()
  })
}

export function useInstallPrompt() {
  const [, setTick] = useState(0)

  useEffect(() => {
    const onChange = () => setTick((n) => n + 1)
    listeners.add(onChange)
    return () => {
      listeners.delete(onChange)
    }
  }, [])

  const install = useCallback(async () => {
    if (!deferred) {
      return
    }
    await deferred.prompt()
    deferred = null
    notify()
  }, [])

  return { canInstall: Boolean(deferred), install }
}
