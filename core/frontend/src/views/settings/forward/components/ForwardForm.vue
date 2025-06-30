<template>
	<modal :title="title" width="480">
		<bt-form ref="formRef" class="pt-8px" :model="form" :rules="rules">
			<n-form-item :label="t('settings.forward.form.status')">
				<n-switch v-model:value="form.active" :checked-value="1" :unchecked-value="0"></n-switch>
			</n-form-item>
			<n-form-item :label="t('settings.forward.form.forwardedUsers')" path="address">
				<n-input
					v-model:value="form.address"
					:disabled="isEdit"
					:placeholder="t('settings.forward.form.forwardedUsersPlaceholder')">
				</n-input>
			</n-form-item>
			<n-form-item v-if="false" :label="t('settings.forward.form.domain')" path="domain">
				<n-input
					v-model:value="form.domain"
					:placeholder="t('settings.forward.form.domainPlaceholder')">
				</n-input>
			</n-form-item>
			<n-form-item :label="t('settings.forward.form.receivingUser')" path="goto">
				<n-input
					v-model:value="form.goto"
					type="textarea"
					:row="4"
					:placeholder="t('settings.forward.form.receivingUserPlaceholder')">
				</n-input>
			</n-form-item>
		</bt-form>
	</modal>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { useModal } from '@/hooks/modal/useModal'
import { addForward, editForward } from '@/api/modules/settings/forward'
import type { Forward } from '../types/base'

const { t } = useI18n()

const isEdit = ref(false)

const title = computed(() =>
	isEdit.value ? t('settings.forward.edit') : t('settings.forward.add')
)

const formRef = useTemplateRef('formRef')

const form = reactive({
	active: 1,
	address: '',
	domain: '',
	goto: '',
})

const rules: FormRules = {}

const [Modal, modalApi] = useModal({
	onChangeState: isOpen => {
		if (isOpen) {
			const state = modalApi.getState<{ isEdit: boolean; row: Forward | null }>()
			const { row } = state
			isEdit.value = state.isEdit
			if (row) {
				form.address = row.address
				form.goto = row.goto
				form.active = row.active
			}
		} else {
			form.address = ''
			form.goto = ''
			form.domain = ''
			form.active = 1
		}
	},
	onConfirm: async () => {
		await formRef.value?.validate()
		if (isEdit.value) {
			await editForward(toRaw(form))
		} else {
			await addForward(toRaw(form))
		}
		const state = modalApi.getState<{ refresh: Function }>()
		state.refresh()
	},
})
</script>

<style lang="scss" scoped></style>
