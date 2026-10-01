<template>
	<n-dropdown size="large" :options="dropdownOptions" @select="handleSelect">
		<n-button class="icon-btn" :bordered="false">
			<i class="icon i-mdi-server-network text-18px"></i>
		</n-button>
	</n-dropdown>

	<n-modal v-model:show="showManageModal" preset="card" :title="t('layout.instances.manage')" style="width: 520px">
		<div v-if="instanceStore.instances.length === 0" class="text-center py-16px color-gray">
			{{ t('layout.instances.noInstances') }}
		</div>
		<div v-else class="instance-list">
			<div v-for="inst in instanceStore.instances" :key="inst.id" class="instance-row">
				<div v-if="editingId !== inst.id" class="instance-info">
					<div class="instance-name">
						{{ inst.name }}
						<n-tag v-if="instanceStore.currentInstance?.id === inst.id" size="small" type="success">
							{{ t('layout.instances.current') }}
						</n-tag>
					</div>
					<div class="instance-url">{{ inst.url }}</div>
				</div>
				<div v-else class="instance-edit-form">
					<n-input v-model:value="editName" size="small" :placeholder="t('layout.instances.namePlaceholder')" />
					<n-input v-model:value="editUrl" size="small" :placeholder="t('layout.instances.urlPlaceholder')" class="mt-4px" />
				</div>
				<div class="instance-actions">
					<template v-if="editingId !== inst.id">
						<n-button size="small" quaternary @click="startEdit(inst)">
							<i class="i-mdi-pencil"></i>
						</n-button>
						<n-popconfirm @positive-click="instanceStore.removeInstance(inst.id)">
							<template #trigger>
								<n-button size="small" quaternary type="error">
									<i class="i-mdi-delete"></i>
								</n-button>
							</template>
							{{ t('layout.instances.deleteConfirm', { name: inst.name }) }}
						</n-popconfirm>
					</template>
					<template v-else>
						<n-button size="small" quaternary type="primary" @click="saveEdit(inst.id)">
							<i class="i-mdi-check"></i>
						</n-button>
						<n-button size="small" quaternary @click="editingId = ''">
							<i class="i-mdi-close"></i>
						</n-button>
					</template>
				</div>
			</div>
		</div>

		<n-divider />

		<div class="add-form">
			<div class="add-form-title">{{ t('layout.instances.add') }}</div>
			<div class="add-form-fields">
				<n-input v-model:value="newName" size="small" :placeholder="t('layout.instances.namePlaceholder')" />
				<n-input
					v-model:value="newUrl"
					size="small"
					:placeholder="t('layout.instances.urlPlaceholder')"
					:status="urlError ? 'error' : undefined"
					class="mt-4px" />
				<div v-if="urlError" class="url-error">{{ urlError }}</div>
				<n-button size="small" type="primary" :disabled="!newName || !newUrl" class="mt-8px" @click="handleAdd">
					{{ t('layout.instances.add') }}
				</n-button>
			</div>
		</div>
	</n-modal>
</template>

<script lang="ts" setup>
import { DropdownOption } from 'naive-ui'
import { useInstanceStore } from '@/store'
import type { BillionMailInstance } from '@/store/modules/instance'

const { t } = useI18n()
const instanceStore = useInstanceStore()

const showManageModal = ref(false)
const newName = ref('')
const newUrl = ref('')
const urlError = ref('')
const editingId = ref('')
const editName = ref('')
const editUrl = ref('')

const dropdownOptions = computed<DropdownOption[]>(() => {
	const items: DropdownOption[] = instanceStore.instances.map(inst => ({
		label: inst.name,
		key: inst.id,
		icon: instanceStore.currentInstance?.id === inst.id
			? () => h('i', { class: 'i-mdi-check text-14px' })
			: undefined,
	}))

	if (items.length > 0) {
		items.push({ type: 'divider', key: 'd1' })
	}

	items.push({
		label: t('layout.instances.manage'),
		key: '__manage__',
		icon: () => h('i', { class: 'i-mdi-cog text-14px' }),
	})

	return items
})

const handleSelect = (key: string) => {
	if (key === '__manage__') {
		showManageModal.value = true
		return
	}
	const inst = instanceStore.instances.find(i => i.id === key)
	if (inst) instanceStore.switchTo(inst)
}

const validateUrl = (raw: string): boolean => {
	try {
		new URL(raw)
		return true
	} catch {
		return false
	}
}

const handleAdd = () => {
	if (!validateUrl(newUrl.value)) {
		urlError.value = t('layout.instances.invalidUrl')
		return
	}
	urlError.value = ''
	instanceStore.addInstance(newName.value, newUrl.value)
	newName.value = ''
	newUrl.value = ''
}

const startEdit = (inst: BillionMailInstance) => {
	editingId.value = inst.id
	editName.value = inst.name
	editUrl.value = inst.url
}

const saveEdit = (id: string) => {
	if (editName.value && editUrl.value && validateUrl(editUrl.value)) {
		instanceStore.updateInstance(id, editName.value, editUrl.value)
	}
	editingId.value = ''
}
</script>

<style lang="scss" scoped>
.instance-list {
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.instance-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 8px;
	border-radius: 4px;
	background: var(--color-bg-2, #f5f5f5);
}

.instance-info {
	flex: 1;
	min-width: 0;
}

.instance-name {
	display: flex;
	align-items: center;
	gap: 6px;
	font-size: 13px;
	font-weight: 500;
}

.instance-url {
	font-size: 11px;
	color: var(--color-text-3, #999);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.instance-edit-form {
	flex: 1;
	min-width: 0;
	margin-right: 8px;
}

.instance-actions {
	display: flex;
	gap: 2px;
	flex-shrink: 0;
}

.add-form-title {
	font-size: 13px;
	font-weight: 600;
	margin-bottom: 8px;
}

.url-error {
	font-size: 11px;
	color: var(--n-error-color, #d03050);
	margin-top: 2px;
}
</style>
