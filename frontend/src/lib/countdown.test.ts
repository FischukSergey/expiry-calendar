import { describe, expect, it } from 'vitest'

import { countdownText, startedNotAfterExpires } from './countdown.ts'

const today = new Date(Date.UTC(2026, 7, 26))

describe('countdownText', () => {
  it('прячет блок без open-даты', () => {
    expect(countdownText(null, today)).toBeNull()
    expect(countdownText(undefined, today)).toBeNull()
    expect(countdownText('', today)).toBeNull()
  })

  it('считает просрочку, сегодня и будущий день', () => {
    expect(countdownText('2026-08-24', today)).toBe('просрочено на 2 дн.')
    expect(countdownText('2026-08-26', today)).toBe('сегодня')
    expect(countdownText('2026-08-29', today)).toBe('через 3 дн.')
  })
})

describe('startedNotAfterExpires', () => {
  it('разрешает пустое начало и запрещает начало после срока', () => {
    expect(startedNotAfterExpires('', '2026-08-26')).toBe(true)
    expect(startedNotAfterExpires('2026-08-01', '2026-08-26')).toBe(true)
    expect(startedNotAfterExpires('2026-08-26', '2026-08-26')).toBe(true)
    expect(startedNotAfterExpires('2026-09-01', '2026-08-26')).toBe(false)
  })
})
