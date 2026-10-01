import { defineStore } from 'pinia'

export interface BillionMailInstance {
	id: string
	name: string
	url: string
}

function normalizeUrl(raw: string): string {
	try {
		return new URL(raw).origin
	} catch {
		return raw.replace(/\/+$/, '')
	}
}

export default defineStore(
	'InstanceStore',
	() => {
		const instances = ref<BillionMailInstance[]>([])

		const currentInstance = computed(() =>
			instances.value.find(i => i.url === window.location.origin)
		)

		const addInstance = (name: string, rawUrl: string) => {
			const url = normalizeUrl(rawUrl)
			const id = Date.now().toString(36) + Math.random().toString(36).slice(2, 6)
			instances.value.push({ id, name, url })
		}

		const updateInstance = (id: string, name: string, rawUrl: string) => {
			const idx = instances.value.findIndex(i => i.id === id)
			if (idx === -1) return
			instances.value[idx] = { id, name, url: normalizeUrl(rawUrl) }
		}

		const removeInstance = (id: string) => {
			instances.value = instances.value.filter(i => i.id !== id)
		}

		const switchTo = (instance: BillionMailInstance) => {
			window.location.href = instance.url
		}

		return {
			instances,
			currentInstance,
			addInstance,
			updateInstance,
			removeInstance,
			switchTo,
		}
	},
	{
		persist: {
			pick: ['instances'],
		},
	}
)

export { normalizeUrl }
