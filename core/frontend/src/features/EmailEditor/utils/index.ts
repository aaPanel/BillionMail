/**
 * @description 获取随机字符串
 * @param length 长度
 * @returns
 */
export function getRandom(length = 12): string {
	const array = new Uint8Array(length)
	crypto.getRandomValues(array)
	return Array.from(array, byte => byte.toString(16).padStart(2, '0'))
		.join('')
		.slice(0, length)
}
