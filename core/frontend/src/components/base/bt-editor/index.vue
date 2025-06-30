<template>
	<n-spin class="h-full" :show="loading">
		<div ref="containerRef" class="monaco-editor-container"></div>
	</n-spin>
</template>

<script lang="ts" setup>
import monacoLoader from '@monaco-editor/loader'
import type { editor } from 'monaco-editor'

const { language, theme, options } = defineProps({
	language: {
		type: String,
		default: 'html',
	},
	theme: {
		type: String,
		default: 'vs-dark',
	},
	options: {
		type: Object as PropType<editor.IStandaloneEditorConstructionOptions>,
		default: () => ({}),
	},
})

const emit = defineEmits<{
	save: []
}>()

const code = defineModel<string>('value', {
	default: '',
})

const containerRef = ref<HTMLElement | null>(null)

let editorInstance: editor.IStandaloneCodeEditor | null = null

const loading = ref(false)

monacoLoader.config({
	paths: {
		vs: `${location.origin}/static/plugin/monaco`,
	},
})

// 初始化编辑器
const initEditor = async () => {
	if (!containerRef.value || editorInstance) {
		return
	}

	try {
		loading.value = true

		const monaco = await monacoLoader.init()
		editorInstance = monaco.editor.create(containerRef.value, {
			value: code.value,
			language: language,
			theme: theme,
			minimap: { enabled: false },
			automaticLayout: true,
			scrollBeyondLastLine: false,
			fontSize: 14,
			tabSize: 2,
			wordWrap: 'on',
			...options,
		})

		editorInstance.onDidChangeModelContent(() => {
			const value = editorInstance?.getValue() || ''
			code.value = value
		})

		editorInstance.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, function () {
			emit('save')
		})
	} finally {
		loading.value = false
	}
}

// 销毁编辑器
const destroyEditor = () => {
	if (editorInstance) {
		editorInstance.dispose()
		editorInstance = null
	}
}

// 监听属性变化
watch(
	() => code.value,
	newVal => {
		if (editorInstance && newVal !== editorInstance.getValue()) {
			editorInstance.setValue(newVal)
		}
	}
)

onMounted(() => {
	initEditor()
})

onBeforeUnmount(() => {
	destroyEditor()
})

// 暴露编辑器实例
defineExpose({
	getEditor: () => editorInstance,
})
</script>

<style scoped>
.monaco-editor-container {
	width: 100%;
	height: 100%;
	min-height: 400px;
	border-radius: 4px;
	overflow: hidden;
}
</style>
