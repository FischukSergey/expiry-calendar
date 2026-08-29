import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { registerSW } from 'virtual:pwa-register'

import App from './App.tsx'
import { AuthProvider } from './auth/AuthContext.tsx'
import { ApiError } from './api/client.ts'
import { listenInstallPrompt } from './hooks/useInstallPrompt.ts'
import './index.css'

listenInstallPrompt()
registerSW({ immediate: true })

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 20_000,
      retry: (count, err) => err instanceof ApiError && err.status >= 500 && count < 1,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <App />
      </AuthProvider>
    </QueryClientProvider>
  </StrictMode>,
)
