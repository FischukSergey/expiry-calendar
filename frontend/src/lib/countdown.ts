/** Текст отсчёта до ближайшего неоплаченного вхождения. Нет даты — null. */
export function countdownText(nextOpenAt: string | null | undefined, today = new Date()): string | null {
  if (!nextOpenAt) {
    return null
  }
  const due = utcDay(nextOpenAt)
  if (due === null) {
    return null
  }
  const now = Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), today.getUTCDate())
  const days = Math.round((due - now) / 86_400_000)
  if (days < 0) {
    return `просрочено на ${-days} дн.`
  }
  if (days === 0) {
    return 'сегодня'
  }
  return `через ${days} дн.`
}

/** Начало периода не позже срока оплаты. Пустая дата начала допустима. */
export function startedNotAfterExpires(startedAt: string, expiresAt: string): boolean {
  if (!startedAt || !expiresAt) {
    return true
  }
  return startedAt <= expiresAt
}

function utcDay(iso: string): number | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(iso)
  if (!match) {
    return null
  }
  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  return Date.UTC(year, month - 1, day)
}
