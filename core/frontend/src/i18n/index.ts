import { createI18n } from 'vue-i18n'
import en from './lang/en.json'
import zh from './lang/zh.json'
import ja from './lang/ja.json'
import tr from './lang/tr.json'

const i18n = createI18n({
	legacy: false,
	globalInjection: true,
	locale: 'en',
	fallbackLocale: 'en',
	messages: {
		en,
		zh,
		ja,
		tr,
	},
})

// 提供切换语言的方法
export const setLanguage = (locale: string) => {
	i18n.global.locale.value = locale as 'en' | 'zh' | 'ja' | 'tr'
}

export default i18n

export { i18n }
