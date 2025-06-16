<template>
	<EditorProvider>
		<div class="email-editor">
			<EditorMenu></EditorMenu>
			<EditorContainer></EditorContainer>
			<EditorSidebar></EditorSidebar>
		</div>
	</EditorProvider>
</template>

<script lang="ts" setup>
import { useHtml } from './hooks/useHtml'
import { useSetData } from './hooks/useSetData'
import { dragData, dragHtml } from './config/addData'

import EditorProvider from './containers/EditorProvider.vue'
import EditorContainer from './containers/Editor.vue'
import EditorMenu from './containers/EditorMenu.vue'
import EditorSidebar from './containers/EditorSidebar.vue'

const configRef = defineModel<string>('config')

const htmlRef = defineModel<string>('html')

const { setData, getData, resetData, setSaveFn } = useSetData()

const { html, generateHtml } = useHtml()

const save = () => {
	generateHtml()
	htmlRef.value = html.value
	configRef.value = getData()
}

setSaveFn(save)

onMounted(() => {
	if (!configRef.value) {
		configRef.value = dragData
		htmlRef.value = dragHtml
		setData(dragData)
	} else {
		setData(configRef.value)
	}
})

onBeforeUnmount(() => {
	resetData()
})

defineExpose({
	setEmailData: (data: string) => {
		setData(data)
	},
	getEmailData: () => {
		return getData()
	},
	getEmailHTML: () => {
		generateHtml()
		return html.value
	},
})
</script>

<style lang="scss" scoped>
.email-editor {
	display: flex;
	height: 100%;
	border: 1px solid #e4e7eb;
}
</style>
