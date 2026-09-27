/** Badging API нет в lib.dom этой сборки. Отказ промиса не показываем в UI. */
type BadgingNavigator = Navigator & {
  setAppBadge?: (contents?: number) => Promise<void>
  clearAppBadge?: () => Promise<void>
}

/** Ставит бейдж иконки PWA. Ноль снимает. Нет API или отказ браузера — молча. */
export function syncAppBadge(count: number): void {
  const nav = navigator as BadgingNavigator
  const pending = count > 0 ? nav.setAppBadge?.(count) : nav.clearAppBadge?.()
  void pending?.catch(() => undefined)
}
