import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'

import { getAccessToken, refreshSession, subscribeAccess } from '../api/client.ts'
import { useAuth } from './useAuth.ts'

function openEvents(token: string): EventSource {
  return new EventSource(`/api/v1/events?access_token=${encodeURIComponent(token)}`)
}

export function useSSE() {
  const { user } = useAuth()
  const qc = useQueryClient()

  useEffect(() => {
    if (!user) {
      return
    }

    let source: EventSource | null = null
    let closed = false
    let reconnecting = false

    const invalidate = () => {
      void qc.invalidateQueries({ queryKey: ['notifications'] })
      void qc.invalidateQueries({ queryKey: ['dashboard'] })
      void qc.invalidateQueries({ queryKey: ['items'] })
      void qc.invalidateQueries({ queryKey: ['calendar'] })
    }

    const connect = (token: string) => {
      if (source) {
        source.onerror = null
        source.close()
      }
      const es = openEvents(token)
      source = es
      es.addEventListener('notification', invalidate)
      es.onerror = () => {
        if (closed || reconnecting) {
          return
        }
        reconnecting = true
        void (async () => {
          const ok = await refreshSession()
          const next = getAccessToken()
          reconnecting = false
          if (ok && next && !closed) {
            connect(next)
          }
        })()
      }
    }

    const first = getAccessToken()
    if (first) {
      connect(first)
    }

    const stop = subscribeAccess((token) => {
      if (closed) {
        return
      }
      if (!token) {
        source?.close()
        source = null
        return
      }
      connect(token)
    })

    return () => {
      closed = true
      stop()
      source?.close()
    }
  }, [user, qc])
}
