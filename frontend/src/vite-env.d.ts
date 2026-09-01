/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

interface ImportMetaEnv {
  /** Только локальный стенд. Прод-сборка без флага — без демо-пароля на /login. */
  readonly VITE_DEMO_LOGIN?: string
}
