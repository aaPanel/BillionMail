import { defineStore } from 'pinia'
import { GlobalThemeOverrides } from 'naive-ui'

export default defineStore(
	'ThemeStore',
	() => {
		const theme = ref<'light' | 'dark'>('light')

		const getCssVar = (name: string) => {
			return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
		}

		const getThemeOverrides = (): GlobalThemeOverrides => {
			return {
				common: {
					lineHeight: 'normal',
					fontSize: '12px',
					fontSizeSmall: '12px',
					fontSizeMedium: '12px',
					fontSizeLarge: '14px',
					borderRadius: '4px',
					baseColor: '#fff',
					textColor1: getCssVar('--color-text-1'),
					textColor2: getCssVar('--color-text-2'),
					primaryColor: getCssVar('--color-primary-1'),
					primaryColorHover: getCssVar('--color-primary-hover-1'),
					successColor: getCssVar('--color-primary-1'),
					successColorHover: getCssVar('--color-primary-hover-1'),
					warningColor: getCssVar('--color-warning-1'),
					errorColor: getCssVar('--color-error-1'),
				},
				Layout: {
					color: getCssVar('--color-bg-2'),
					textColor: getCssVar('--color-text-1'),
					headerColor: getCssVar('--color-bg-1'),
					siderColor: getCssVar('--color-bg-1'),
					siderTextColor: getCssVar('--color-sider-text-1'),
				},
				Menu: {
					fontSize: '14px',
					itemTextColor: getCssVar('--color-menu-1'),
					itemIconColor: getCssVar('--color-menu-1'),
					itemColorActive: getCssVar('--color-menu-active-1'),
					itemTextColorActive: getCssVar('--color-menu-active-2'),
					itemIconColorActive: getCssVar('--color-menu-active-2'),
					itemColorActiveHover: getCssVar('--color-menu-active-1'),
					itemTextColorActiveHover: getCssVar('--color-menu-active-2'),
					itemIconColorActiveHover: getCssVar('--color-menu-active-2'),
					itemColorActiveCollapsed: getCssVar('--color-menu-active-1'),
				},
				Card: {
					color: getCssVar('--color-bg-1'),
					borderColor: 'transparent',
					borderRadius: '6px',
				},
				Form: {
					feedbackHeightMedium: '20px',
					feedbackHeightLarge: '22px',
					feedbackFontSizeMedium: '12px',
					feedbackFontSizeLarge: '12px',
					feedbackPadding: '2px 0 0',
					labelFontSizeTopMedium: '13px',
					labelFontSizeLeftMedium: '13px',
					labelPaddingHorizontal: '0 20px 0 0',
					labelFontWeight: '700',
				},
				Radio: {
					buttonColor: getCssVar('--color-radio-1'),
					buttonColorActive: getCssVar('--color-radio-2'),
					buttonTextColor: getCssVar('--color-text-2'),
					buttonTextColorHover: getCssVar('--color-text-1'),
					buttonTextColorActive: getCssVar('--color-text-1'),
					buttonBorderColor: 'transparent',
					buttonBorderColorHover: 'transparent',
					buttonBorderColorActive: 'transparent',
					buttonBoxShadowHover: 'none',
					buttonBoxShadowFocus: 'none',
					buttonBorderRadius: '4px',
					labelPadding: '0 0 0 8px',
				},
				Dialog: {
					contentMargin: '16px 0',
					textColor: getCssVar('--color-text-1'),
				},
				DataTable: {
					thPaddingMedium: '10px',
					tdPaddingMedium: '10px',
					thColor: getCssVar('--color-table-th-1'),
					tdColor: getCssVar('--color-table-td-1'),
					tdColorHover: getCssVar('--color-table-th-1'),
					borderRadius: '4px',
				},
				Breadcrumb: {
					fontSize: '14px',
				},
				Switch: {
					railColorActive: getCssVar('--color-primary-1'),
				},
				Tabs: {
					tabBorderRadius: '4px',
				},
				Progress: {
					textColorLineInner: '#fff',
				},
			}
		}

		const themeOverrides = ref<GlobalThemeOverrides>(getThemeOverrides())

		const setTheme = (val: 'light' | 'dark') => {
			theme.value = val
		}

		const changeTheme = () => {
			const isDarkMode = theme.value === 'dark'
			document.documentElement.setAttribute('theme-mode', isDarkMode ? 'dark' : '')
			nextTick(() => {
				themeOverrides.value = getThemeOverrides()
			})
		}

		return {
			theme,
			themeOverrides,
			setTheme,
			changeTheme,
		}
	},
	{
		persist: {
			pick: ['theme'],
		},
	}
)
