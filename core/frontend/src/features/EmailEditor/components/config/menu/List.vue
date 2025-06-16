<template>
	<div class="w-full">
		<ul class="link-group">
			<menu-link-item
				v-for="(item, index) of links"
				:key="item.id"
				v-model:label="item.label"
				v-model:href="item.href"
				:menu-id="item.id"
				:menu-index="index"
				:find-link="findLink"
				:move-link="moveLink"
				@del="onDelLink(index)">
			</menu-link-item>
		</ul>
		<n-button block class="mt-12px" @click="onAddLink">Add</n-button>
	</div>
</template>

<script lang="ts" setup>
import { getRandom } from '../../../utils'
import { useConfig } from '../../../hooks/useConfig'
import { useSetData } from '../../../hooks/useSetData'

import MenuLinkItem from './Item.vue'

const { selectedBlockKey, blockConfigMap } = useConfig()
const { autoSaveFn } = useSetData()

const links = computed({
	get() {
		return blockConfigMap.value[selectedBlockKey.value].attr.links || []
	},
	set(val) {
		blockConfigMap.value[selectedBlockKey.value].attr.links = val
	},
})

watch(
	() => [links.value],
	() => {
		autoSaveFn()
	},
	{
		deep: true,
	}
)

const onDelLink = (index: number) => {
	links.value.splice(index, 1)
}

const onAddLink = () => {
	links.value.push({ id: getRandom(6), label: '', href: '' })
}

const findLink = (id: string) => {
	const link = links.value.filter((c: { id: string }) => `${c.id}` === id)[0]
	return {
		link,
		index: links.value.indexOf(link),
	}
}

const moveLink = (id: string, atIndex: number) => {
	const { link, index } = findLink(id)
	links.value.splice(index, 1)
	links.value.splice(atIndex, 0, link)
}
</script>

<style lang="scss" scoped>
.link-group {
	display: flex;
	flex-direction: column;
	width: 100%;
	padding-left: 0;
	margin: 0;
}
</style>
