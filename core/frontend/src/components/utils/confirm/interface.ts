import { VNodeChild } from 'vue'
import { Type } from 'naive-ui/es/button/src/interface'

export interface ConfirmOptions {
	title?: string
	width?: number | string
	content?: string | (() => VNodeChild)
	confirmText?: string
	confirmType?: Type
	onCancel?: () => unknown
	onConfirm?: () => void | boolean | Promise<unknown>
}
