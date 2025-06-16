<template>
	<div class="editor-menu">
		<EditorMenu></EditorMenu>
	</div>
</template>

<script lang="ts" setup>
import EditorMenu from '../components/menu/index.vue'
</script>

<style lang="scss" scoped>
.editor-menu {
	width: 200px;
	border-right: 1px solid #e4e7eb;
	background-color: #fff;
	overflow: hidden;
}
</style>
