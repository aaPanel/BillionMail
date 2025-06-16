<template>
	<div class="block-none">
		<slot></slot>
	</div>
</template>

<script lang="ts" setup></script>

<style lang="scss" scoped>
.block-none {
	display: flex;
	flex-direction: column;
	justify-content: center;
	align-items: center;
	height: 100px;
	padding: 10px;
	background-color: #f9f9f9;
	border: 1px solid #d4d4d4;
	border-radius: 4px;
	color: #a8a8a8;
	font-size: 14px;
}
</style>
