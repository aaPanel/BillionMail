import { createVNode, render } from 'vue'
import { render as renderContent } from 'naive-ui/es/_utils'
import { app } from '@/index'
import type { ConfirmOptions } from './interface'

import Modal from '@/components/ui/bt-modal/index.vue'
import ConfigProvider from '@/components/ui/bt-config-provider/index.vue'

/**
 * @description Command-style call confirm dialog
 * @param options Confirm dialog options
 */
export function confirm(options: ConfirmOptions) {
	const container = document.createElement('div')

	const removeModalElement = () => {
		render(null, container)
		if (container.parentNode) {
			container.parentNode.removeChild(container)
		}
	}

	// Create virtual node
	const modalNode = createVNode(
		Modal,
		{
			show: true,
			width: options.width || 420,
			title: options.title,
			onCancel: options.onCancel,
			onConfirm: options.onConfirm,
			confirmText: options.confirmText,
			confirmType: options.confirmType,
			onAfterLeave: () => {
				removeModalElement()
			},
		},
		{
			default: () => <div class="pt-6px pb-10px text-14px">{renderContent(options.content)}</div>,
		}
	)

	// Create configuration provider node
	const configNode = createVNode(
		ConfigProvider,
		{},
		{
			default: () => [modalNode],
		}
	)

	configNode.appContext = app._context

	// Render to Modal
	render(configNode, container)
}
