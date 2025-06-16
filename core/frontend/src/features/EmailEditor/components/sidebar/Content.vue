<template>
	<component :is="getComponentByType" :key="selectedBlockKey"></component>
</template>

<script lang="ts" setup>
import { useConfig } from '../../hooks/useConfig'
import ColumnsConfig from '../config/Columns.vue'
import ButtonConfig from '../config/Button.vue'
import LinkConfig from '../config/Link.vue'
import DividerConfig from '../config/Divider.vue'
import HeaderConfig from '../config/Header.vue'
import TextConfig from '../config/Text.vue'
import ImageConfig from '../config/Image.vue'
import MenuConfig from '../config/Menu.vue'
import NoneConfig from '../config/None.vue'

const { selectedBlockType, selectedBlockKey } = useConfig()

const getComponentByType = computed(() => {
	switch (selectedBlockType.value) {
		case 'columns':
			return ColumnsConfig
		case 'button':
			return ButtonConfig
		case 'link':
			return LinkConfig
		case 'divider':
			return DividerConfig
		case 'header':
			return HeaderConfig
		case 'text':
			return TextConfig
		case 'image':
			return ImageConfig
		case 'menu':
			return MenuConfig
		default:
			return NoneConfig // 默认返回 NoneConfig
	}
})
</script>
