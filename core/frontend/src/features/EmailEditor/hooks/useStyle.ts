import { StyleValue } from 'vue'
import { isString } from 'lodash-es'
import { isObject } from '@/utils'
import { BaseStyle, BorderStyle, MoreStyle } from '../types/base'

/**
 * @description border、padding、borderRadius等符合类型的联合类型
 */
export type MoreStyleKey = 'border' | 'padding' | 'borderRadius'

export function useStyle() {
	const configToStyle = (configStyle: BaseStyle) => {
		const style: StyleValue = {}
		Object.entries(configStyle).forEach(([key, value]) => {
			if (isMoreStyleKey(key) && isObject<MoreStyle<string> | MoreStyle<BorderStyle>>(value)) {
				Object.assign(style, translateMoreStyle(key, value))
			}
			if (!isMoreStyleKey(key) && isString(value)) {
				style[key as keyof StyleValue] = value
			}
			if (!isMoreStyleKey(key) && isObject(value)) {
				if (key === 'borderTop' && isObject<BorderStyle>(value)) {
					style.borderTop = `${value.width} ${value.style} ${value.color || 'transparent'}`
				}
			}
		})
		return style
	}

	const isMoreStyleKey = (attrKey: string): attrKey is MoreStyleKey => {
		return attrKey === 'border' || attrKey === 'padding' || attrKey === 'borderRadius'
	}

	const translateMoreStyle = (
		attrKey: MoreStyleKey,
		attrValue: MoreStyle<string> | MoreStyle<BorderStyle>
	) => {
		const styleMap: Record<string, string> = {}
		const posMap = {
			top: 'TopLeft',
			right: 'TopRight',
			bottom: 'BottomRight',
			left: 'BottomLeft',
		}

		if (!attrValue.more) {
			if (attrKey === 'border' && isObject<BorderStyle>(attrValue.all) && attrValue.all.width !== '0') {
				styleMap.border = `${attrValue.all.width} ${attrValue.all.style} ${attrValue.all.color || 'transparent'}`
			} else if ((attrKey === 'padding' || attrKey === 'borderRadius') && isString(attrValue.all)) {
				styleMap[attrKey] = attrValue.all
			}
		} else {
			const positions = ['top', 'right', 'bottom', 'left']
			switch (attrKey) {
				case 'border':
					positions.forEach(pos => {
						const value = attrValue[pos as keyof typeof attrValue]
						if (isObject<BorderStyle>(value) && value.width !== '0') {
							styleMap[`border${capitalizeFirst(pos)}`] =
								`${value.width} ${value.style} ${value.color || 'transparent'}`
						}
					})
					break

				case 'padding':
					positions.forEach(pos => {
						const value = attrValue[pos as keyof typeof attrValue]
						if (isString(value) && value !== '0') {
							styleMap[`padding${capitalizeFirst(pos)}`] = value
						}
					})
					break

				case 'borderRadius':
					positions.forEach(pos => {
						const value = attrValue[pos as keyof typeof attrValue]
						if (isString(value) && value !== '0') {
							styleMap[`border${posMap[pos as keyof typeof posMap]}Radius`] = value
						}
					})
					break
			}
		}

		return styleMap
	}

	const capitalizeFirst = (str: string): string => {
		return str.charAt(0).toUpperCase() + str.slice(1)
	}

	return {
		configToStyle,
	}
}
