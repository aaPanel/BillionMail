<template>
	<li :ref="setDrop" class="link-items" :style="{ opacity }">
		<div class="flex-1">
			<n-input v-model:value="label"></n-input>
		</div>
		<div class="flex-1">
			<n-input v-model:value="href"></n-input>
		</div>
		<div class="w-auto cursor-pointer" @click="onDel">
			<i class="i-mdi:delete-outline text-16px"></i>
		</div>
		<div :ref="drag" class="w-auto cursor-grab">
			<i class="i-mdi:format-list-bulleted text-16px"></i>
		</div>
	</li>
</template>

<script lang="ts" setup>
import { useDrag, useDrop } from 'vue3-dnd'
import { MoveLinkFunc, FindLinkFunc } from '../../../types/base'

const { menuId, menuIndex, moveLink, findLink } = defineProps({
	menuId: {
		type: String,
		required: true,
	},
	menuIndex: {
		type: Number,
		required: true,
	},
	moveLink: {
		type: Function as PropType<MoveLinkFunc>,
		required: true,
	},
	findLink: {
		type: Function as PropType<FindLinkFunc>,
		required: true,
	},
})

const emit = defineEmits<{
	del: []
}>()

const label = defineModel<string>('label')

const href = defineModel<string>('href')

const onDel = () => {
	emit('del')
}

const [collect, drag, preview] = useDrag({
	type: 'MenuLink',
	item: () => ({ id: menuId, index: menuIndex }),
	collect: monitor => ({
		isDragging: monitor.isDragging(),
	}),
	end: (item, monitor) => {
		const { id, index } = item
		const didDrop = monitor.didDrop()
		if (!didDrop) {
			moveLink(id, index)
		}
	},
})

const [, drop] = useDrop(() => ({
	accept: 'MenuLink',
	hover({ id: draggedId }: { id: string }) {
		if (draggedId !== menuId) {
			const { index: overIndex } = findLink(menuId)
			moveLink(draggedId, overIndex)
		}
	},
}))

const setDrop = (el: HTMLDivElement) => {
	if (el) {
		preview(drop(el))
	}
	return undefined
}

const opacity = computed(() => (collect.value.isDragging ? 0 : 1))
</script>

<style lang="scss" scoped>
.link-items {
	display: flex;
	align-items: center;
	width: 100%;
	padding: 12px;
	gap: 8px;
	border: 1px solid #e5e7eb;
	border-top: none;

	&:first-child {
		border-top: 1px solid #e5e7eb;
		border-top-left-radius: 4px;
		border-top-right-radius: 4px;
	}

	&:last-child {
		border-bottom-left-radius: 4px;
		border-bottom-right-radius: 4px;
	}
}
</style>
