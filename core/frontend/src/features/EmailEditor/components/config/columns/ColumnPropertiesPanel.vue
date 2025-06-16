<template>
	<div>
		<div class="mb-8px">
			<n-button @click="addCell"> Add </n-button>
			<n-button class="ml-10px" :disabled="cells.length <= 1" @click="delCell">Delete</n-button>
		</div>
		<n-tabs ref="tabsRef" v-model:value="activeCellIndex" class="mb-8px" type="line">
			<n-tab-pane v-for="(cell, index) in cells" :key="cell.key" :name="index" :tab="`Column ${index + 1}`">
			</n-tab-pane>
		</n-tabs>
		<div class="py-16px">
			<style-form v-if="currentCellKey" v-model:value="currentCellStyle"> </style-form>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useConfig } from '../../../hooks/useConfig'
import { useSetData } from '../../../hooks/useSetData'
import { useNormalForm } from '../../style/useNormalForm'

const { selectedBlockKey, columnsMap, cellMap, cellConfigMap, applyNewWidthColumnsInCell, delColumnsInCell } =
	useConfig()
const { autoSaveFn } = useSetData()

const tabsRef = useTemplateRef('tabsRef')

const columnsConfig = computed(() => {
	return columnsMap.value[selectedBlockKey.value]
})

const cells = computed(() => {
	return columnsConfig.value.children.map(key => {
		return cellMap.value[key]
	})
})

const activeCellIndex = ref(0)

const addCell = () => {
	// 收集所有的cell宽度
	const cellWidths: number[] = []
	cells.value.forEach(cell => {
		cellWidths.push(cell.width)
	})

	const targetTotal = 100 // 目标总和
	const newValue = Math.floor(targetTotal / (cellWidths.length + 1)) // 新项占总和的三分之一
	// 计算当前总和
	const currentTotal = cellWidths.reduce((sum, num) => sum + num, 0)
	// 计算需要调整的总和
	const totalAfterAddition = currentTotal + newValue
	const adjustment = targetTotal - totalAfterAddition
	// 计算比例因子
	const factor = adjustment / currentTotal
	// 添加新项
	cellWidths.push(newValue)
	// 按比例调整元素
	for (let i = 0; i < cellWidths.length - 1; i++) {
		cellWidths[i] = Math.floor(cellWidths[i] + cellWidths[i] * factor) // 调整每个元素
	}

	// 应用新的宽度到当前cell
	applyNewWidthColumnsInCell(selectedBlockKey.value, cellWidths)

	// 更新当前激活的cell
	activeCellIndex.value = cellWidths.length - 1
	nextTick(() => {
		tabsRef.value?.syncBarPosition()
	})
}

const delCell = () => {
	if (cells.value.length <= 1) return

	// 先收集所有cell的宽度集合（此处需要形成map结构，这样才能精准找到应该删除哪一个cell）
	const cellWidthMap: Record<string, number> = cells.value.reduce((p, v) => {
		const cellWidth = v.width
		return { ...p, [v.key]: cellWidth }
	}, {})

	const delCell = cells.value[activeCellIndex.value]
	// 根据cellKey删除map中对应的宽度
	delete cellWidthMap[delCell.key]
	// 得到新的宽度集合
	let widths = Object.values(cellWidthMap)
	// 目标总和
	const targetTotal = 100
	// 计算当前总和
	const currentTotal = widths.reduce((sum, num) => sum + num, 0)
	// 计算需要增加的值
	const adjustment = targetTotal - currentTotal
	// 计算比例因子
	const factor = adjustment / currentTotal
	// 按比例调整剩余项
	widths = widths.map(num => Math.floor(num * (1 + factor)))
	// 确保总和为 100
	const finalTotal = widths.reduce((sum, num) => sum + num, 0)
	const totalAdjustment = targetTotal - finalTotal
	// 如果有误差，均匀调整最后一项
	if (totalAdjustment !== 0) {
		widths[widths.length - 1] += totalAdjustment
	}

	delColumnsInCell(selectedBlockKey.value, delCell.key)
	applyNewWidthColumnsInCell(selectedBlockKey.value, widths)

	// 更新当前激活的cell
	if (activeCellIndex.value >= widths.length) {
		activeCellIndex.value = widths.length - 1
	}
	nextTick(() => {
		tabsRef.value?.syncBarPosition()
	})
}

const currentCellKey = computed(() => {
	return cells.value[activeCellIndex.value]?.key || ''
})

const currentCellStyle = computed({
	get() {
		return cellConfigMap.value[currentCellKey.value].style
	},
	set(newVal) {
		cellConfigMap.value[currentCellKey.value].style = newVal
		autoSaveFn()
	},
})

const [StyleForm] = useNormalForm([
	{
		attrKey: 'backgroundColor',
		label: 'Background Color',
	},
	{
		attrKey: 'padding',
		label: 'Padding',
	},
	{
		attrKey: 'border',
		label: 'Border',
	},
])
</script>

<style lang="scss" scoped>
.n-tabs {
	--n-tab-padding: 13px 16px 11px;
	--n-tab-gap: 0;
	--n-pane-padding-top: 0;
	--n-tab-font-size: 12px;
}

.n-tab-pane {
	flex: 1;
	overflow: auto;
}
</style>
