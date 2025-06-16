<template>
	<div ref="editorBoxRef" class="editor-box">
		<div v-show="showToolbar" ref="toolbarDomRef" class="toolbar"></div>
		<div ref="editorDomRef" class="editor"></div>
	</div>
</template>

<script lang="ts" setup>
import '@wangeditor/editor/dist/css/style.css'
import {
	IDomEditor,
	DomEditor,
	i18nChangeLanguage,
	createEditor,
	createToolbar,
} from '@wangeditor/editor'

const { type } = defineProps({
	type: {
		type: String as PropType<'button' | 'text' | 'header'>,
		default: 'button',
	},
})

// 编辑器实例，必须用 shallowRef
const editorRef = shallowRef<IDomEditor | undefined>(undefined)

const toolbarDomRef = useTemplateRef('toolbarDomRef')

const editorDomRef = useTemplateRef('editorDomRef')

// 内容 HTML
const valueHtml = defineModel<string>('value')

const getToolbarKeys = () => {
	switch (type) {
		case 'button':
			return ['bold', 'underline', 'italic', 'through']
		case 'text':
			return [
				'bold',
				'underline',
				'italic',
				'through',
				'sup',
				'sub',
				'color',
				'bgColor',
				'bulletedList',
				'numberedList',
				'insertLink',
			]
		case 'header':
			return [
				'bold',
				'underline',
				'italic',
				'through',
				'color',
				'bgColor',
				'bulletedList',
				'numberedList',
				'insertLink',
			]
	}
}

i18nChangeLanguage('en')

const showToolbar = ref(false)

let lockShow = false

// 计算工具栏位置
const updateToolbarPosition = () => {
	if (!editorRef.value) return
	if (!toolbarDomRef.value) return

	const editorEl = editorRef.value.getEditableContainer()
	if (!editorEl) return

	const rect = editorEl.getBoundingClientRect()

	// if (!toolbarRef.value.selector) return

	// 获取工具栏元素，计算其宽度
	const toolbarEl = toolbarDomRef.value.querySelector('.w-e-toolbar')
	let toolbarWidth = 300 // 默认宽度估计值

	if (toolbarEl) {
		toolbarWidth = toolbarEl.getBoundingClientRect().width
	}

	// 计算居中位置
	const leftPosition = rect.left + (rect.width - toolbarWidth) / 2

	toolbarDomRef.value.style.top = `${rect.top - 56}px`
	toolbarDomRef.value.style.left = `${leftPosition}px`
}

// 记录editor实例
const handleCreated = (editor: IDomEditor) => {
	editorRef.value = editor

	nextTick(() => {
		// 添加窗口大小变化监听
		window.addEventListener('resize', updateToolbarPosition)

		const toolbar = DomEditor.getToolbar(editor)
		if (toolbar) {
			const linkBtn = toolbar.$box.find('button[data-menu-key="insertLink"]')
			linkBtn.on('click', () => {
				lockShow = true
			})
		}
	})
}

const handleFocus = () => {
	showToolbar.value = true
	nextTick(() => {
		updateToolbarPosition()
	})
}

const handleBlur = () => {
	if (lockShow) {
		lockShow = false
		return
	}
	showToolbar.value = false
}

const initEditor = () => {
	if (!editorDomRef.value) return
	// 创建编辑器
	const editor = createEditor({
		selector: editorDomRef.value,
		html: valueHtml.value,
		mode: 'simple',
		config: {
			autoFocus: false,
			scroll: false,
			placeholder: '',
			onChange(editor: IDomEditor) {
				const html = editor.getHtml()
				// 更新响应式数据，实现编辑器 -> 数据的绑定
				valueHtml.value = html
			},
			onCreated: handleCreated,
			onFocus: handleFocus,
			onBlur: handleBlur,
		},
	})

	if (!toolbarDomRef.value) return
	// 创建工具栏
	createToolbar({
		editor,
		selector: toolbarDomRef.value,
		mode: 'default',
		config: {
			toolbarKeys: getToolbarKeys(),
		},
	})
}

onMounted(() => {
	initEditor()
})

// 组件销毁时，也及时销毁编辑器
onBeforeUnmount(() => {
	const editor = editorRef.value
	if (editor == null) return
	editor.destroy()

	// 移除事件监听
	window.removeEventListener('resize', updateToolbarPosition)
})

defineExpose({
	focus: () => {
		showToolbar.value = true
		editorRef.value?.focus()
	},
	blur: () => {
		showToolbar.value = false
	},
})
</script>

<style lang="scss" scoped>
.editor-box {
	position: relative;
	width: 100%;
}

.toolbar {
	position: fixed;
	top: -100px;
	left: 0;
	z-index: 200;
	display: flex;
	flex-direction: column;

	:deep(.w-e-toolbar) {
		max-width: 500px;
		padding: 0 4px;
		border: 1px solid #ccc;
		margin: 0 auto;

		.w-e-bar {
			overflow: hidden;
		}

		.w-e-bar-item {
			display: inline-flex;
			padding: 4px 2px;

			button .title {
				height: auto;
				border-bottom: none;
			}
		}
	}
}

.editor {
	:deep(.w-e-text-container) {
		background: none;
		color: inherit;

		[data-slate-editor] {
			padding: 0;
			border-top: none;
		}

		p {
			margin: 0;
			line-height: inherit;
		}
	}

	:deep(.w-e-modal) {
		z-index: 100;
	}
}
</style>
