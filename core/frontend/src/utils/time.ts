import { toNumber } from 'lodash-es'
import { format, startOfDay, endOfDay } from 'date-fns'
import { isString, isDate, isNumber, getNumber } from '@/utils'
import i18n from '@/i18n'

export type Time = Date | number | string

/**
 * @description Format time
 * @param time Timestamp / Date / Time string
 * @param fmt Time format, optional
 * @returns { string } Formatted time
 * @throws { Error } When time is invalid
 */
export const formatTime = (time?: Time, fmt = 'yyyy-MM-dd HH:mm:ss'): string => {
	if (!time) return '--'

	let timestamp: number

	try {
		if (isString(time)) {
			timestamp = isNaN(toNumber(time)) ? new Date(time).getTime() : getNumber(time)
		} else if (isDate(time)) {
			timestamp = time.getTime()
		} else if (isNumber(time)) {
			const str = time.toString()
			timestamp = str.length === 10 ? time * 1000 : time
		} else {
			return '--'
		}

		return format(timestamp, fmt)
	} catch {
		// console.error('Time formatting error:', error)
		return '--'
	}
}

/**
 * @description Get time range for a given date
 * @param { Date } date Date object, defaults to current date
 * @returns { [number, number] } Tuple containing start and end timestamps
 */
export const getDayTimeRange = (date: Date = new Date()): [number, number] => {
	const start = startOfDay(date)
	const end = endOfDay(date)
	return [start.getTime(), end.getTime()]
}

/**
 * @description Convert seconds to the highest time unit only with decimal precision
 * @param seconds Total seconds to convert
 * @param precision Number of decimal places (default: 1)
 * @returns { string } Formatted time string with only the highest unit and decimal (e.g., "2.5天", "5.3小时", "30.8分钟", "45秒")
 */
export const formatDurationHighest = (seconds: number, precision: number = 1): string => {
	if (!seconds || seconds < 0) return `0 ${i18n.global.t('common.unit.seconds')}`

	const days = seconds / (24 * 60 * 60)
	const hours = seconds / (60 * 60)
	const minutes = seconds / 60

	if (days >= 1) return `${days.toFixed(precision)} ${i18n.global.t('common.unit.days')}`
	if (hours >= 1) return `${hours.toFixed(precision)} ${i18n.global.t('common.unit.hours')}`
	if (minutes >= 1) return `${minutes.toFixed(precision)} ${i18n.global.t('common.unit.minutes')}`
	return `${Math.floor(seconds)} ${i18n.global.t('common.unit.seconds')}`
}
