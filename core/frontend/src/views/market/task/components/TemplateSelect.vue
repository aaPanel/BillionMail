<template>
	<n-select
		v-model:value="templateId"
		:loading="loading"
		:filterable="true"
		:options="options"
		@update:value="handleUpdateTemplate">
	</n-select>
</template>

<script lang="ts" setup>
import { SelectOption } from 'naive-ui'
import { isArray, isObject } from '@/utils'
import { getTemplateAll } from '@/api/modules/market/template'
import { Template } from '@/views/market/template/interface'

const templateId = defineModel<number | null>('value')

const content = defineModel<string>('content')

const loading = ref(false)

const options = ref<SelectOption[]>([])

const handleUpdateTemplate = (val: number, option: SelectOption) => {
	const { data } = option
	if (isObject<Template>(data)) {
		content.value = data.html_content
	}
}

const getList = async () => {
	try {
		loading.value = true
		const res = await getTemplateAll()
		if (isArray<Template>(res)) {
			options.value = res.map(item => ({
				label: item.temp_name,
				value: item.id,
				data: item,
			}))
			if (templateId.value === null) {
				templateId.value = res[0].id
				content.value = res[0].html_content
			} else {
				const template = res.find(item => item.id === templateId.value)
				if (template) {
					content.value = template.html_content
				}
			}
		}
	} finally {
		loading.value = false
	}
}

getList()
</script>

<style lang="scss" scoped></style>
