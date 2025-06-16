<template>
	<div class="mt-2 flex items-center">
		<div class="flex-1 bg-gray-200 rounded-full h-1 mr-2">
			<div
				class="h-1 rounded-full transition-all duration-300"
				:class="strengthClass"
				:style="{ width: strengthWidth }" />
		</div>
		<span class="text-xs text-gray-500">
			{{ t('settings.common.security.passwordStrength') }}: {{ strengthText }}
		</span>
	</div>
</template>

<script lang="ts" setup>
import type { PasswordStrength } from '../types/base'

const { strength } = defineProps({
	strength: {
		type: Object as PropType<PasswordStrength>,
		required: true,
	},
})

const { t } = useI18n()

const strengthClass = computed(() => {
	switch (strength.level) {
		case 'weak':
			return 'bg-error'
		case 'medium':
			return 'bg-warning'
		case 'strong':
			return 'bg-primary'
		default:
			return 'bg-gray-300'
	}
})

const strengthWidth = computed(() => {
	switch (strength.level) {
		case 'weak':
			return '30%'
		case 'medium':
			return '60%'
		case 'strong':
			return '100%'
		default:
			return '0%'
	}
})

const strengthText = computed(() => {
	return t(`settings.common.security.strength.${strength.level}`)
})
</script>
