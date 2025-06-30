<template>
	<div>
		<label class="block text-sm font-medium text-basic mb-3">
			{{ t('settings.common.system.timezone') }}
		</label>
		<div class="grid items-start grid-cols-1 md:grid-cols-2 gap-4">
			<div>
				<div class="flex gap-12px">
					<div class="flex-1">
						<n-select
							v-model:value="zone"
							:options="zoneOptions"
							filterable
							@update:value="handleUpdateZone">
						</n-select>
					</div>
					<div class="flex-1">
						<n-select
							v-model:value="area"
							:options="areaOptions"
							filterable
							@update:value="handleUpdateArea">
						</n-select>
					</div>
				</div>
				<div class="mt-1 text-xs text-desc">{{ t('settings.common.system.currentTime') }}</div>
			</div>
			<div class="flex items-center">
				<n-alert type="warning" class="w-full">
					<template #icon>
						<n-icon>
							<i class="i-mdi-alert-circle"></i>
						</n-icon>
					</template>
					{{ t('settings.common.system.terminalRequired') }}
				</n-alert>
			</div>
		</div>

		<div class="mt-5">
			<command :value="commandRef"></command>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { SelectOption } from 'naive-ui'
import { isObject } from '@/utils'
import { getTimezoneList } from '@/api/modules/settings/common'
import { getSettingsStore } from '@/views/settings/common/store'

import Command from './Command.vue'

const { t } = useI18n()

const { currentTime, timeCommand } = getSettingsStore()

const zone = ref('')

const area = ref('')

const commandRef = ref('')

const timeZoneData = ref<{ [key: string]: string[] }>({})

const zoneOptions = ref<SelectOption[]>([])

const areaOptions = computed(() => {
	const areas = timeZoneData.value[zone.value]
	if (areas) {
		return areas.map(item => {
			return {
				label: item,
				value: item,
			}
		})
	}
	return []
})

const handleUpdateZone = () => {
	const areas = timeZoneData.value[zone.value]
	area.value = areas ? areas[0] : ''
	handleUpdateArea()
}

const handleUpdateArea = () => {
	if (zone.value && area.value) {
		commandRef.value = timeCommand.value.replace(/TIME ZONE/g, `${zone.value}/${area.value}`)
	}
}

watch(
	() => currentTime.value,
	time => {
		if (time) {
			const times = time.split('/')
			zone.value = times[0]
			area.value = times[1]
			commandRef.value = timeCommand.value.replace(/TIME ZONE/g, `${zone.value}/${area.value}`)
		}
	},
	{
		immediate: true,
	}
)

const getOptions = async () => {
	const res = await getTimezoneList()
	if (isObject<{ [key: string]: string[] }>(res)) {
		timeZoneData.value = res
		zoneOptions.value = Object.keys(res).map(key => {
			return {
				label: key,
				value: key,
			}
		})
	}
}

getOptions()
</script>

<style lang="scss" scoped></style>
