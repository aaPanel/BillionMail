<template>
	<n-config-provider
		abstract
		inline-theme-disabled
		:theme="themeRef"
		:locale="locale"
		:date-locale="dateLocale"
		:theme-overrides="themeOverrides">
		<slot></slot>
	</n-config-provider>
</template>

<script lang="ts" setup>
import { storeToRefs } from 'pinia'
import { enUS, zhCN, jaJP, dateEnUS, dateZhCN, dateJaJP, darkTheme } from 'naive-ui'
import { useGlobalStore, useThemeStore } from '@/store'

const globalStore = useGlobalStore()
const { lang } = storeToRefs(globalStore)

const themeStore = useThemeStore()
const { theme, themeOverrides } = storeToRefs(themeStore)

const langMap = {
	zh: {
		locale: zhCN,
		dateLocale: dateZhCN,
	},
	en: {
		locale: enUS,
		dateLocale: dateEnUS,
	},
	ja: {
		locale: jaJP,
		dateLocale: dateJaJP,
	},
}

const themeRef = computed(() => {
	if (theme.value === 'dark') {
		return darkTheme
	}
	return null
})

const locale = computed(() => {
	const langObj = langMap[lang.value as keyof typeof langMap]
	return langObj.locale
})

const dateLocale = computed(() => {
	const langObj = langMap[lang.value as keyof typeof langMap]
	return langObj.dateLocale
})

// const themeOverrides: GlobalThemeOverrides = {
// 	common: {
// 		lineHeight: 'normal',
// 		fontSize: '12px',
// 		fontSizeSmall: '12px',
// 		fontSizeMedium: '12px',
// 		fontSizeLarge: '14px',
// 		borderRadius: '4px',
// 		primaryColor: getCssVar('--color-primary-1'),
// 		primaryColorHover: '#1D9534',
// 		successColor: '#20A53A',
// 		successColorHover: '#1D9534',
// 		warningColor: '#F0AD4E',
// 		warningColorHover: '#C6892E',
// 		errorColor: '#ef0808',
// 		errorColorHover: '#C9302C',
// 	},
// 	Layout: {
// 		color: '#f4f7fa',
// 		textColor: '#333',
// 		headerColor: 'rgba(244, 247, 250, 0.7)',
// 		headerTextColor: '#5b6b79',
// 	},
// 	Menu: {
// 		fontSize: '14px',
// 	},
// 	Form: {
// 		feedbackHeightMedium: '20px',
// 		feedbackHeightLarge: '22px',
// 		feedbackFontSizeMedium: '12px',
// 		feedbackFontSizeLarge: '12px',
// 		feedbackPadding: '2px 0 0',
// 		labelFontSizeLeftMedium: '12px',
// 		labelPaddingHorizontal: '0 20px 0 0',
// 		labelFontSizeTopMedium: '12px',
// 	},
// 	Dialog: {
// 		contentMargin: '16px 0',
// 	},
// 	Radio: {
// 		labelPadding: '0 16px 0 8px',
// 		buttonColorActive: '#20A53A',
// 		buttonTextColorActive: '#fff',
// 	},
// 	DataTable: {
// 		thPaddingMedium: '10px',
// 		tdPaddingMedium: '10px',
// 	},
// 	Breadcrumb: {
// 		fontSize: '14px',
// 	},
// }
</script>
