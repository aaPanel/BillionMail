<template>
	<div class="style-container">
		<div class="style-rows">
			<div class="style-label">{{ label }}</div>
			<div class="style-value">
				<slot name="value"></slot>
			</div>
		</div>
		<slot name="options"></slot>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	label: {
		type: String,
		default: '',
	},
})
</script>

<style lang="scss" scoped>
.style-container {
	padding: 16px 0;
	border-top: 1px solid #eeeeee;

	&:first-of-type {
		padding-top: 0;
		border-top: none;
	}

	&:last-of-type {
		padding-bottom: 0;
	}

	.style-rows {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;

		.style-label {
			display: flex;
			align-items: center;
			min-height: 32px;
			line-height: 24px;
			font-weight: 500;
		}

		.style-value {
			display: flex;
			align-items: center;
			min-height: 32px;
		}
	}
}
</style>
