<template>
	<n-upload
		v-model:file-list="fileList"
		:accept="uploadAccept"
		:multiple="false"
		:default-upload="false"
		:show-file-list="false"
		@before-upload="onBeforeUpload"
		@change="onUploadChange">
		<template v-if="mode === 'dragger'">
			<n-upload-dragger>
				<div class="flex-center flex-col">
					<template v-if="fileList.length === 0">
						<upload-prompt :suffix="suffix"></upload-prompt>
					</template>
					<template v-else>
						<upload-progress v-if="uploading" :progress="progress" />
						<upload-error v-if="uploadFailed" />
						<upload-success v-if="uploadComplete" :fileName="fileName" />
					</template>
				</div>
			</n-upload-dragger>
		</template>
		<template v-else>
			<n-button ghost :type="buttonType">
				<slot></slot>
			</n-button>
		</template>
	</n-upload>
</template>

<script lang="ts" setup>
import { PropType } from 'vue'
import { UploadFileInfo } from 'naive-ui'
import { Message } from '@/utils'
import { uploadFile as uploadFileApi } from '@/api/modules/public'

import UploadPrompt from './UploadPrompt.vue'
import UploadProgress from './UploadProgress.vue'
import UploadError from './UploadError.vue'
import UploadSuccess from './UploadSuccess.vue'

const { t } = useI18n()

const { path, accept, maxSize, isUpload } = defineProps({
	path: {
		type: String,
		default: '',
	},
	accept: {
		type: Array as PropType<string[]>,
		default: () => [],
	},
	maxSize: {
		type: Number,
		default: 5, // 5MB
	},
	isUpload: {
		type: Boolean,
		default: true,
	},
	mode: {
		type: String as PropType<'dragger' | 'button'>,
		default: 'dragger',
	},
	buttonType: {
		type: String as PropType<'default' | 'primary' | 'info' | 'success' | 'warning' | 'error'>,
		default: 'default',
	},
})

const emit = defineEmits<{
	change: [file: UploadFileInfo]
}>()

const fileList = ref<UploadFileInfo[]>([])

const fileName = computed(() => {
	return fileList.value[0]?.name || '--'
})

const progress = ref(0)

const uploading = ref(false)

const uploadFailed = ref(false)

const uploadComplete = ref(false)

const uploadAccept = computed(() => {
	return accept.map(item => `.${item}`).join(',')
})

const suffix = computed(() => {
	return accept.map(item => item.toLocaleUpperCase().replace('.', '')).join(' or ')
})

const onBeforeUpload = (data: { file: UploadFileInfo }) => {
	const { file } = data
	if ((file.file?.size ?? 0) > maxSize * 1024 * 1024) {
		Message.error(t('components.upload.error.fileSizeExceeded', { fileName: file.name, maxSize }))
		return false
	}

	return true
}

const onUploadChange = async ({ file }: { file: UploadFileInfo }) => {
	resetUploadStatus()

	if (isUpload) {
		await uploadFile(file)
	} else {
		emit('change', file)
		uploadComplete.value = true
	}
}

const uploadFile = async (file: UploadFileInfo) => {
	try {
		uploading.value = true
		const form = createFormData(file)
		await uploadFileApi(form, updateProgress)
		uploadComplete.value = true
	} catch {
		uploadFailed.value = true
	} finally {
		uploading.value = false
	}
}

const resetUploadStatus = () => {
	progress.value = 0
	uploading.value = false
	uploadFailed.value = false
	uploadComplete.value = false
}

const createFormData = (file: UploadFileInfo) => {
	const form = new FormData()
	form.append('f_path', path)
	form.append('f_name', file.name)
	if (file.file) {
		form.append('f_size', file.file.size.toString())
		form.append('blob', file.file)
	}
	return form
}

const updateProgress = (e: { progress?: number }) => {
	progress.value = (e.progress || 0) * 100
}
</script>

<style lang="scss" scoped>
.n-upload-dragger {
	padding: 32px 0;

	&:hover {
		--n-dragger-color: var(--color-primary-2);
		--n-dragger-border-hover: 1px dashed var(--color-primary-3);

		:deep(.drag-icon) {
			background-color: var(--color-primary-1);
			color: #fff;
		}
	}

	:deep(.drag-icon) {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: 4px;
		background-color: var(--color-bg-4);
		font-size: 16px;
		transition:
			background-color 0.3s ease,
			color 0.3s ease;
	}

	:deep(.drag-text) {
		margin-top: 16px;
		font-size: 14px;
		font-weight: 500;
	}

	:deep(.drag-suffix) {
		margin-top: 4px;
		color: var(--color-text-2);
	}
}
</style>
