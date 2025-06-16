<template>
	<div class="block-panel">
		<div class="block-list">
			<block-item v-for="item of componentsPreset" :key="item.type" :type="item.type" :data="item">
			</block-item>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { MenuBlock } from '../../types/base'
import BlockItem from './block.vue'

/**
 * @description 菜单数据源
 */
const componentsPreset = ref<MenuBlock[]>([
	{
		type: 'columns',
		name: 'Columns',
		key: '0',
	},
	{
		type: 'button',
		name: 'Button',
		key: '0',
	},
	{
		type: 'link',
		name: 'Link',
		key: '0',
	},
	{
		type: 'divider',
		name: 'Divider',
		key: '0',
	},
	{
		type: 'header',
		name: 'Header',
		key: '0',
	},
	{
		type: 'text',
		name: 'Text',
		key: '0',
	},
	{
		type: 'image',
		name: 'Image',
		key: '0',
	},
	{
		type: 'menu',
		name: 'Menu',
		key: '0',
	},
	// {
	// 	type: 'html',
	// 	name: 'HTML',
	// 	key: '0',
	// },
])
</script>

<style lang="scss" scoped>
.block-panel {
	width: 100%;
	height: 100%;
	overflow: auto;
}

.block-title {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 32px;
	margin-top: 8px;
	margin-bottom: 8px;
	font-size: 14px;
	font-weight: bold;
	color: #444;
}

.block-list {
	display: grid;
	justify-content: center;
	grid-template-columns: 74px 74px;
	padding: 16px;
	grid-gap: 16px;
}
</style>
