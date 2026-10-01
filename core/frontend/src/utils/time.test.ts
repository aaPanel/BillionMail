import { describe, it, expect, vi } from 'vitest'

// Mock i18n before importing time module
vi.mock('@/i18n', () => ({
  default: {
    global: {
      t: (key: string) => {
        const translations: Record<string, string> = {
          'common.unit.days': 'Days',
          'common.unit.hours': 'Hours',
          'common.unit.minutes': 'Minutes',
          'common.unit.seconds': 'Seconds',
        }
        return translations[key] ?? key
      },
    },
  },
}))

// Mock @/utils to avoid cascading import of base.ts
vi.mock('@/utils', () => ({
  isString: (val: unknown): val is string => typeof val === 'string',
  isDate: (val: unknown): val is Date => val instanceof Date,
  isNumber: (val: unknown): val is number => typeof val === 'number',
  getNumber: (val: unknown) => {
    const n = Number(val)
    return isNaN(n) ? 0 : n
  },
}))

import { formatTime, getDayTimeRange, formatDurationHighest } from './time'

describe('formatTime', () => {
  it('returns "--" for falsy input', () => {
    expect(formatTime()).toBe('--')
    expect(formatTime(0)).toBe('--')
    expect(formatTime('')).toBe('--')
    expect(formatTime(undefined)).toBe('--')
  })

  it('formats Date objects', () => {
    const date = new Date(2024, 0, 15, 10, 30, 45) // Jan 15 2024 10:30:45
    expect(formatTime(date)).toBe('2024-01-15 10:30:45')
  })

  it('formats millisecond timestamps', () => {
    // 2024-01-15 00:00:00 UTC
    const ts = new Date(2024, 0, 15, 0, 0, 0).getTime()
    const result = formatTime(ts)
    expect(result).toMatch(/2024-01-15/)
  })

  it('formats 10-digit (second) timestamps by multiplying by 1000', () => {
    // A 10-digit unix timestamp
    const ts = 1705276800 // 2024-01-15 UTC approx
    const result = formatTime(ts)
    expect(result).toMatch(/2024-01-1/)
  })

  it('formats date strings', () => {
    const result = formatTime('2024-01-15T10:30:00')
    expect(result).toMatch(/2024-01-15/)
  })

  it('formats numeric strings as timestamps', () => {
    const ts = new Date(2024, 0, 15, 12, 0, 0).getTime().toString()
    const result = formatTime(ts)
    expect(result).toMatch(/2024-01-15/)
  })

  it('uses custom format', () => {
    const date = new Date(2024, 0, 15, 10, 30, 45)
    expect(formatTime(date, 'yyyy/MM/dd')).toBe('2024/01/15')
    expect(formatTime(date, 'HH:mm')).toBe('10:30')
  })

  it('returns "--" for invalid date strings', () => {
    expect(formatTime('not-a-date')).toBe('--')
  })
})

describe('getDayTimeRange', () => {
  it('returns start and end of given date', () => {
    const date = new Date(2024, 0, 15, 12, 30, 0) // mid-day
    const [start, end] = getDayTimeRange(date)

    const startDate = new Date(start)
    expect(startDate.getHours()).toBe(0)
    expect(startDate.getMinutes()).toBe(0)
    expect(startDate.getSeconds()).toBe(0)

    const endDate = new Date(end)
    expect(endDate.getHours()).toBe(23)
    expect(endDate.getMinutes()).toBe(59)
    expect(endDate.getSeconds()).toBe(59)
  })

  it('start is before end', () => {
    const [start, end] = getDayTimeRange(new Date())
    expect(start).toBeLessThan(end)
  })

  it('returns tuple of two numbers', () => {
    const result = getDayTimeRange()
    expect(result).toHaveLength(2)
    expect(typeof result[0]).toBe('number')
    expect(typeof result[1]).toBe('number')
  })

  it('both timestamps are on the same date', () => {
    const date = new Date(2024, 5, 20)
    const [start, end] = getDayTimeRange(date)
    const startDate = new Date(start)
    const endDate = new Date(end)
    expect(startDate.getFullYear()).toBe(endDate.getFullYear())
    expect(startDate.getMonth()).toBe(endDate.getMonth())
    expect(startDate.getDate()).toBe(endDate.getDate())
  })
})

describe('formatDurationHighest', () => {
  it('returns "0 Seconds" for 0 or negative', () => {
    expect(formatDurationHighest(0)).toBe('0 Seconds')
    expect(formatDurationHighest(-1)).toBe('0 Seconds')
  })

  it('formats seconds (< 60)', () => {
    expect(formatDurationHighest(30)).toBe('30 Seconds')
    expect(formatDurationHighest(59)).toBe('59 Seconds')
    // Floors seconds
    expect(formatDurationHighest(45.7)).toBe('45 Seconds')
  })

  it('formats minutes (60-3599)', () => {
    expect(formatDurationHighest(60)).toBe('1.0 Minutes')
    expect(formatDurationHighest(90)).toBe('1.5 Minutes')
    expect(formatDurationHighest(150)).toBe('2.5 Minutes')
    expect(formatDurationHighest(3599)).toBe('60.0 Minutes')
  })

  it('formats hours (3600-86399)', () => {
    expect(formatDurationHighest(3600)).toBe('1.0 Hours')
    expect(formatDurationHighest(5400)).toBe('1.5 Hours')
    expect(formatDurationHighest(7200)).toBe('2.0 Hours')
  })

  it('formats days (>= 86400)', () => {
    expect(formatDurationHighest(86400)).toBe('1.0 Days')
    expect(formatDurationHighest(129600)).toBe('1.5 Days')
    expect(formatDurationHighest(172800)).toBe('2.0 Days')
  })

  it('respects precision parameter', () => {
    expect(formatDurationHighest(5400, 0)).toBe('2 Hours')
    expect(formatDurationHighest(5400, 2)).toBe('1.50 Hours')
    expect(formatDurationHighest(5400, 3)).toBe('1.500 Hours')
  })
})
