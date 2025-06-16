<template>
	<div class="editor-container">
		<editor-toolbar></editor-toolbar>
		<div class="flex-1">
			<editor-canvas></editor-canvas>
		</div>
	</div>
</template>

<script lang="ts" setup>
import EditorCanvas from '../components/layout/Canvas.vue'
import EditorToolbar from '../components/toolbar/index.vue'
</script>

<style lang="scss" scoped>
.editor-container {
	display: flex;
	flex-direction: column;
	flex: 1;
}
</style>
