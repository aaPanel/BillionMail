<template>
	<Collapse :default-expanded-names="['1', '2', '3']">
		<n-collapse-item title="Columns" name="1">
			<ColumnsPanel />
		</n-collapse-item>
		<n-collapse-item title="Column Properties" name="2">
			<ColumnPropertiesPanel />
		</n-collapse-item>
		<n-collapse-item title="Row Properties" name="3">
			<RowPropertiesPanel />
		</n-collapse-item>
	</Collapse>
</template>

<script lang="ts" setup>
import Collapse from '../shared/Collapse.vue'
import ColumnsPanel from './columns/ColumnsPanel.vue'
import ColumnPropertiesPanel from './columns/ColumnPropertiesPanel.vue'
import RowPropertiesPanel from './columns/RowPropertiesPanel.vue'
</script>
