<template>
	<div class="editor-tools">
		<div>
			<div v-if="false" class="tools-view">
				<div class="view-item checked">
					<i class="i-ri:computer-line text-16px"></i>
				</div>
				<div class="view-item">
					<i class="i-ri:smartphone-line text-16px"></i>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
// import { useModal } from '@/hooks/useModal'
// import EditorPreview from './preview/index.vue'
</script>

<style lang="scss" scoped>
.editor-tools {
	display: flex;
	align-items: center;
	justify-content: space-between;
	height: 40px;
	min-height: 40px;
	padding: 0 16px;
	border-bottom: 1px solid #d8dee8;
}

.tools-view {
	display: flex;
	align-items: center;
	gap: 4px;
	height: 28px;
	padding: 2px;
	border: 1px solid #d9d9d9;
	border-radius: 4px;

	.view-item {
		display: flex;
		align-items: center;
		height: 24px;
		padding: 0 4px;
		border-radius: 2px;
		cursor: pointer;

		&.checked {
			background-color: #f8f8f8;
			color: #3e8bf2;
		}
	}
}
</style>
