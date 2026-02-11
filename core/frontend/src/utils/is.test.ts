import { describe, it, expect } from 'vitest'
import { is, isObject, isNumber, isString, isFunction, isUndefined, isArray, isDate } from './is'

describe('is', () => {
  it('matches Object type', () => {
    expect(is({}, 'Object')).toBe(true)
    expect(is([], 'Object')).toBe(false)
    expect(is('str', 'Object')).toBe(false)
  })

  it('matches Number type', () => {
    expect(is(1, 'Number')).toBe(true)
    expect(is(NaN, 'Number')).toBe(true)
    expect(is('1', 'Number')).toBe(false)
  })

  it('matches String type', () => {
    expect(is('hello', 'String')).toBe(true)
    expect(is(123, 'String')).toBe(false)
  })

  it('matches Array type', () => {
    expect(is([], 'Array')).toBe(true)
    expect(is({}, 'Array')).toBe(false)
  })

  it('matches Function type', () => {
    expect(is(() => {}, 'Function')).toBe(true)
    expect(is({}, 'Function')).toBe(false)
  })

  it('matches Date type', () => {
    expect(is(new Date(), 'Date')).toBe(true)
    expect(is('2024-01-01', 'Date')).toBe(false)
  })
})

describe('isObject', () => {
  it('returns true for plain objects', () => {
    expect(isObject({})).toBe(true)
    expect(isObject({ a: 1 })).toBe(true)
  })

  it('returns false for non-objects', () => {
    expect(isObject(null)).toBe(false)
    expect(isObject(undefined)).toBe(false)
    expect(isObject([])).toBe(false)
    expect(isObject('string')).toBe(false)
    expect(isObject(42)).toBe(false)
    expect(isObject(() => {})).toBe(false)
  })
})

describe('isNumber', () => {
  it('returns true for numbers', () => {
    expect(isNumber(0)).toBe(true)
    expect(isNumber(1)).toBe(true)
    expect(isNumber(-1)).toBe(true)
    expect(isNumber(3.14)).toBe(true)
    expect(isNumber(NaN)).toBe(true)
    expect(isNumber(Infinity)).toBe(true)
  })

  it('returns false for non-numbers', () => {
    expect(isNumber('1')).toBe(false)
    expect(isNumber(null)).toBe(false)
    expect(isNumber(undefined)).toBe(false)
    expect(isNumber(true)).toBe(false)
  })
})

describe('isString', () => {
  it('returns true for strings', () => {
    expect(isString('')).toBe(true)
    expect(isString('hello')).toBe(true)
    expect(isString(`template`)).toBe(true)
  })

  it('returns false for non-strings', () => {
    expect(isString(1)).toBe(false)
    expect(isString(null)).toBe(false)
    expect(isString(undefined)).toBe(false)
    expect(isString({})).toBe(false)
  })
})

describe('isFunction', () => {
  it('returns true for functions', () => {
    expect(isFunction(() => {})).toBe(true)
    expect(isFunction(function () {})).toBe(true)
    expect(isFunction(describe)).toBe(true)
  })

  it('returns false for non-functions', () => {
    expect(isFunction({})).toBe(false)
    expect(isFunction('fn')).toBe(false)
    expect(isFunction(null)).toBe(false)
  })
})

describe('isUndefined', () => {
  it('returns true for undefined', () => {
    expect(isUndefined(undefined)).toBe(true)
    expect(isUndefined(void 0)).toBe(true)
  })

  it('returns false for defined values', () => {
    expect(isUndefined(null)).toBe(false)
    expect(isUndefined(0)).toBe(false)
    expect(isUndefined('')).toBe(false)
    expect(isUndefined(false)).toBe(false)
  })
})

describe('isArray', () => {
  it('returns true for arrays', () => {
    expect(isArray([])).toBe(true)
    expect(isArray([1, 2, 3])).toBe(true)
    expect(isArray(new Array(3))).toBe(true)
  })

  it('returns false for non-arrays', () => {
    expect(isArray({})).toBe(false)
    expect(isArray('array')).toBe(false)
    expect(isArray(null)).toBe(false)
    expect(isArray(undefined)).toBe(false)
  })
})

describe('isDate', () => {
  it('returns true for Date objects', () => {
    expect(isDate(new Date())).toBe(true)
    expect(isDate(new Date('2024-01-01'))).toBe(true)
  })

  it('returns false for non-dates', () => {
    expect(isDate('2024-01-01')).toBe(false)
    expect(isDate(1704067200000)).toBe(false)
    expect(isDate(null)).toBe(false)
    expect(isDate({})).toBe(false)
  })
})
