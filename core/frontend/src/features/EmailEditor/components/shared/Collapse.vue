<template>
	<n-collapse arrow-placement="right" display-directive="show">
		<slot></slot>
	</n-collapse>
</template>

<style lang="scss" scoped>
.n-collapse {
	--n-title-padding: 0;
	--n-item-margin: 0;
	--n-text-color: #444;
	--n-arrow-color: #d4d4d4;
	--n-title-text-color: #333;
	--n-divider-color: #d4d4d4;

	:deep(.n-collapse-item) {
		background-color: #fff;

		> .n-collapse-item__header {
			&:hover {
				background-color: #f9f9f9;
			}

			.n-collapse-item__header-main {
				justify-content: space-between;
				padding: 12px 16px;
				font-weight: bold;
			}
		}

		.n-collapse-item__content-inner {
			padding: 16px;
		}
	}
}
</style>
