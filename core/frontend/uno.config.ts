import { defineConfig, presetUno, presetAttributify, transformerDirectives } from 'unocss'

export default defineConfig({
	presets: [presetUno({ preflight: false }), presetAttributify()],
	transformers: [transformerDirectives({ enforce: 'pre' })],
	theme: {
		colors: {
			primary: '#20a53a',
			primaryHover: '#1D9534',
			error: '#ef0808',
			warning: '#f0ad4e',
		},
	},
	shortcuts: {
		'flex-center': 'flex items-center justify-center',
	},
})
