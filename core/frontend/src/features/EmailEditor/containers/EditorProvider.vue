<template>
	<!-- 仅作为上下文提供容器 -->
	<dnd-provider :backend="HTML5Backend">
		<slot></slot>
	</dnd-provider>
</template>

<script lang="ts" setup>
import { DndProvider } from 'vue3-dnd'
import { HTML5Backend } from 'react-dnd-html5-backend'
import { useEmailEditorStore } from '../store'
import { createContext } from '../hooks/useContext'

// 初始化全局状态
const store = useEmailEditorStore()

// 提供上下文
createContext({
	store,
})
</script>
