<template>
	<div class="editor-sidebar">
		<config-tabs> </config-tabs>
	</div>
</template>

<script lang="ts" setup>
import ConfigTabs from '../components/sidebar/ConfigTabs.vue'
</script>

<style lang="scss" scoped>
.editor-sidebar {
	height: 100%;
	width: 320px;
	border-left: 1px solid #e4e7eb;
	background-color: #fff;
	overflow: hidden;
}
</style>
