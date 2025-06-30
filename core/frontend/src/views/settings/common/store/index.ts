import { defineStore, storeToRefs } from 'pinia'
import { differenceInDays } from 'date-fns'
import { formatTime, isObject } from '@/utils'
import { getSystemConfig } from '@/api/modules/settings/common'
import type { SystemInfo, PasswordStrength } from '../types/base'

export const useSettingsStore = defineStore('SettingsCommonStore', () => {
	const currentUsername = ref('')

	const passwordMismatch = ref(false)
	const passwordStrength = ref<PasswordStrength>({ level: 'weak', score: 0 })

	const currentPath = ref('')

	const securityForm = reactive({
		username: '',
		newPassword: '',
		confirmPassword: '',
		securityPath: '',
	})

	const networkForm = reactive({
		domain: '',
		sslEnabled: true,
		sslCert: '',
		sslKey: '',
	})

	const sslInfo = reactive({
		status: false,
		expireTime: '',
		expireDays: 0,
	})

	const currentDomain = ref('')

	const currentPort = ref('')

	const portCommand = ref('')

	const currentTime = ref('')

	const timeCommand = ref('')

	const serverIp = ref('')

	const ipWhitelistEnable = ref(false)
	const ipWhitelistList = ref<{ id: number; ip: string }[]>([])

	const checkPasswordStrength = () => {
		const password = securityForm.newPassword
		let score = 0

		if (password.length >= 8) score++
		if (password.match(/[a-z]/)) score++
		if (password.match(/[A-Z]/)) score++
		if (password.match(/[0-9]/)) score++
		if (password.match(/[^a-zA-Z0-9]/)) score++

		if (score <= 2) {
			passwordStrength.value = { level: 'weak', score }
		} else if (score <= 4) {
			passwordStrength.value = { level: 'medium', score }
		} else {
			passwordStrength.value = { level: 'strong', score }
		}
	}

	// 获取设置信息
	const getCommonConfig = async () => {
		const res = await getSystemConfig()
		if (isObject<SystemInfo>(res)) {
			currentUsername.value = res.admin_username
			currentPath.value = res.safe_path
			currentDomain.value = res.billionmail_hostname
			serverIp.value = res.server_ip
			ipWhitelistEnable.value = res.ip_whitelist_enable
			ipWhitelistList.value = res.ip_whitelist

			if (res.manage_ports) {
				currentPort.value = `${res.manage_ports.https}`
				portCommand.value = res.manage_ports.command_https
			}
			if (res.manage_timezone) {
				currentTime.value = res.manage_timezone.timezone
				timeCommand.value = res.manage_timezone.command
			}
			if (res.ssl) {
				networkForm.sslCert = res.ssl.certPem
				networkForm.sslKey = res.ssl.privateKey
				sslInfo.status = res.ssl.status

				const expireDate = new Date(res.ssl.notAfter)
				sslInfo.expireTime = formatTime(expireDate, 'yyyy-MM-dd')
				sslInfo.expireDays = differenceInDays(expireDate, new Date())
			}
		}
	}

	const reset = () => {
		securityForm.username = ''
		securityForm.newPassword = ''
		securityForm.confirmPassword = ''
		securityForm.securityPath = ''
		networkForm.domain = ''
		networkForm.sslEnabled = true
		networkForm.sslCert = ''
		networkForm.sslKey = ''
		sslInfo.status = false
		sslInfo.expireTime = ''
		sslInfo.expireDays = 0
		currentPort.value = ''
		portCommand.value = ''
		currentTime.value = ''
		timeCommand.value = ''
		serverIp.value = ''
		passwordMismatch.value = false
		passwordStrength.value = { level: 'weak', score: 0 }
	}

	return {
		// 状态
		currentUsername,
		currentPath,
		passwordMismatch,
		passwordStrength,
		securityForm,
		networkForm,
		sslInfo,
		currentDomain,
		currentPort,
		portCommand,
		currentTime,
		timeCommand,
		serverIp,
		ipWhitelistEnable,
		ipWhitelistList,

		// 方法
		checkPasswordStrength,
		getCommonConfig,
		reset,
	}
})

export const getSettingsStore = () => {
	const store = useSettingsStore()
	return {
		...store,
		...storeToRefs(store),
	}
}
