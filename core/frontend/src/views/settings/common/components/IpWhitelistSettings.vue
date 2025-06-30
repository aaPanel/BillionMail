<template>
    <div>
        <label class="block text-sm font-medium mb-3">{{ t('settings.common.ipWhitelist.title') }}</label>
        <div class="flex justify-between items-center h-34px mb-12px">
            <div class="flex item-center gap-8px">
                <n-switch :value="ipWhitelistEnable" @update:value="handleUpdateEnable">
                    <template #checked>{{ t('settings.common.ipWhitelist.enable') }}</template>
                    <template #unchecked>{{ t('settings.common.ipWhitelist.disable') }}</template>
                </n-switch>
            </div>
            <div class="text-desc">{{ t('settings.common.ipWhitelist.description') }}</div>
        </div>
        <div class="flex gap-8px mb-12px">
            <div class="flex-1">
                <n-input v-model:value="ip" :placeholder="t('settings.common.ipWhitelist.placeholder')"></n-input>
            </div>
            <n-button type="primary" @click="onAddIp">{{ t('settings.common.ipWhitelist.add') }}</n-button>
        </div>
        <n-list v-if="ipWhitelistList.length > 0" hoverable bordered>
            <n-list-item v-for="(item, index) in ipWhitelistList" :key="item.id">
                <span class="text-13px">
                    {{ item.ip }}
                </span>
                <template #suffix>
                    <n-button size="small" quaternary circle @click="onDelIp(item.id, index)">
                        <i class="i-mingcute:close-fill text-16px"></i>
                    </n-button>
                </template>
            </n-list-item>
        </n-list>
        <n-empty v-else></n-empty>
        <div class="mt-12px">
            <n-button @click="onClear">{{ t('settings.common.ipWhitelist.clear') }}</n-button>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { confirm, Message } from '@/utils'
import {
	addIpWhitelist,
	clearIpWhitelist,
	deleteIpWhitelist,
	setSystemConfigKey,
} from '@/api/modules/settings/common'
import { getSettingsStore } from '../store'

const { ipWhitelistEnable, ipWhitelistList, getCommonConfig } = getSettingsStore()

const ip = ref('')

const { t } = useI18n()

const onAddIp = async () => {
    if (ip.value === '') {
        Message.error(t('settings.common.ipWhitelist.validation.ipRequired'))
        return
    }
	await addIpWhitelist({ ip: ip.value })
	getCommonConfig()
	ip.value = ''
}

const onDelIp = (id: number, index: number) => {
	confirm({
		title: t('settings.common.ipWhitelist.confirm.deleteTitle'),
		content: t('settings.common.ipWhitelist.confirm.deleteContent', { ip: ipWhitelistList.value[index].ip }),
		onConfirm: async () => {
			await deleteIpWhitelist({ id })
			ipWhitelistList.value.splice(index, 1)
		},
	})
}

const handleUpdateEnable = (val: boolean) => {
	confirm({
		title: val ? t('settings.common.ipWhitelist.confirm.enableTitle') : t('settings.common.ipWhitelist.confirm.disableTitle'),
		content: val ? t('settings.common.ipWhitelist.confirm.enableContent') : t('settings.common.ipWhitelist.confirm.disableContent'),
		onConfirm: async () => {
			await setSystemConfigKey({ key: 'ip_whitelist_enable', value: `${val}` })
			ipWhitelistEnable.value = val
		},
	})
}

const onClear = async () => {
	confirm({
		title: t('settings.common.ipWhitelist.confirm.clearTitle'),
		content: t('settings.common.ipWhitelist.confirm.clearContent'),
		onConfirm: async () => {
			await clearIpWhitelist()
			getCommonConfig()
		},
	})
}
</script>

<style lang="scss" scoped></style>
