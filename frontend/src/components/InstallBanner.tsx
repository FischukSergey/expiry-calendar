import { Button } from './ui.tsx'
import { useInstallPrompt } from '../hooks/useInstallPrompt.ts'

export function InstallBanner() {
  const { canInstall, install } = useInstallPrompt()
  if (!canInstall) {
    return null
  }
  return (
    <div className="flex flex-col items-stretch justify-between gap-2 border-b border-teal-900 bg-teal-950/40 px-4 py-2 text-sm sm:flex-row sm:items-center">
      <span>Установить Duekeep на устройство</span>
      <Button type="button" className="shrink-0" onClick={() => void install()}>
        Установить
      </Button>
    </div>
  )
}
