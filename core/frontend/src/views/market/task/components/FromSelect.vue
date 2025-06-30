<template>
	<div class="flex-1">
		<n-select
			v-model:value="domain"
			:loading="loading"
			:options="domainOptions"
			@update:value="handleUpdateDomain">
		</n-select>
	</div>
	<div class="flex-1 ml-10px">
		<n-select
			v-model:value="sender"
			:loading="loading"
			:options="senderOptions"
			@update:value="handleUpdateSender">
		</n-select>
	</div>
</template>

<script lang="ts" setup>
import { SelectOption } from 'naive-ui'
import { isArray, isObject } from '@/utils'
import { getDomainAll } from '@/api/modules/domain'
import { getMailboxList } from '@/api/modules/mailbox'
import type { MailDomain } from '@/views/domain/interface'
import type { MailBox } from '@/views/mailbox/interface'

const sender = defineModel<string | null>('value')

const name = defineModel<string>('name')

const loading = ref<boolean>(false)

const domain = defineModel<string | null>('domain')

const domainOptions = ref<SelectOption[]>([])

const mailboxList = ref<MailBox[]>([])

const senderOptions = computed(() => {
	return mailboxList.value
		.filter(item => item.domain === domain.value)
		.map(item => ({
			label: item.username,
			value: item.username,
			data: item,
		}))
})

const handleUpdateDomain = async () => {
	await nextTick()
	if (senderOptions.value.length > 0) {
		sender.value = senderOptions.value[0].value
		handleUpdateSender(sender.value, senderOptions.value[0])
	} else {
		sender.value = null
		name.value = ''
	}
}

const handleUpdateSender = (val: string, option: SelectOption) => {
	const { data } = option
	if (isObject<MailBox>(data)) {
		name.value = data.full_name
	}
}

// 获取域名列表
const getDomainOptions = async () => {
	const domainRes = await getDomainAll()
	if (isArray<MailDomain>(domainRes)) {
		domainOptions.value = domainRes.map(item => ({
			label: item.domain,
			value: item.domain,
		}))
		if (domain.value === null && domainRes.length > 0) {
			domain.value = domainRes[0].domain
		}
	}
}

// 获取邮箱列表
const getSenderOptions = async () => {
	const res = await getMailboxList({ page: 1, page_size: 1000, domain: '' })
	if (isObject<{ list: MailBox[] }>(res)) {
		mailboxList.value = res.list
	}
}

const initData = async () => {
	try {
		loading.value = true
		await getSenderOptions()
		await getDomainOptions()
		await nextTick()
		if (sender.value === null) {
			handleUpdateDomain()
		}
	} finally {
		loading.value = false
	}
}

initData()
</script>

<style lang="scss" scoped></style>
