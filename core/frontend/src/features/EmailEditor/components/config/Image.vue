<template>
	<Collapse :default-expanded-names="['1', '2', '3']">
		<n-collapse-item title="Image" name="1">
			<div class="pb-10px">
				<ImageForm v-model:value="blockStyle" />
			</div>
		</n-collapse-item>
		<n-collapse-item title="Action" name="2">
			<div class="pb-10px">
				<ActionForm v-model:value="blockStyle" />
			</div>
		</n-collapse-item>
		<n-collapse-item title="Container" name="3">
			<div class="py-8px">
				<ContainerForm v-model:value="containerStyle" />
			</div>
		</n-collapse-item>
	</Collapse>
</template>

<script lang="tsx" setup>
import { useConfig } from '../../hooks/useConfig'
import { useSetData } from '../../hooks/useSetData'
import { useNormalForm } from '../style/useNormalForm'

import Collapse from '../shared/Collapse.vue'
import ImageUrl from '../style/Image.vue'
import ImageLink from '../style/Link.vue'
import ImageTitle from '../style/Title.vue'

const { selectedBlockKey, blockConfigMap } = useConfig()
const { autoSaveFn } = useSetData()

const attr = computed({
	get() {
		return blockConfigMap.value[selectedBlockKey.value].attr
	},
	set(newVal) {
		blockConfigMap.value[selectedBlockKey.value].attr = newVal
	},
})

const blockStyle = computed({
	get() {
		return blockConfigMap.value[selectedBlockKey.value].style
	},
	set(newVal) {
		blockConfigMap.value[selectedBlockKey.value].style = newVal
	},
})

const containerStyle = computed({
	get() {
		return blockConfigMap.value[selectedBlockKey.value].containerStyle
	},
	set(newVal) {
		blockConfigMap.value[selectedBlockKey.value].containerStyle = newVal
	},
})

watch(
	() => [attr.value, blockStyle.value, containerStyle.value],
	() => {
		autoSaveFn()
	},
	{
		deep: true,
	}
)

const [ImageForm] = useNormalForm([
	{
		attrKey: 'custom',
		render: () => <ImageUrl v-model:value={attr.value.src} />,
	},
	{
		attrKey: 'widthAuto',
		key: 'width',
	},
	{
		attrKey: 'custom',
		render: () => <ImageTitle v-model:value={attr.value.alt} />,
	},
])

const [ActionForm] = useNormalForm([
	{
		attrKey: 'custom',
		render: () => <ImageLink v-model:value={attr.value.href} v-model:target={attr.value.target} />,
	},
])

const [ContainerForm] = useNormalForm([
	{
		attrKey: 'textAlign',
	},
	{
		attrKey: 'padding',
	},
])
</script>
