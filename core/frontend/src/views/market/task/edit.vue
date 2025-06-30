<template>
	<div class="task-edit-container">
		<div class="flex flex-col flex-1 p-24px overflow-auto">
			<n-breadcrumb class="mb-20px">
				<n-breadcrumb-item>
					<router-link to="/market/task">{{ $t('market.task.title') }}</router-link>
				</n-breadcrumb-item>
				<n-breadcrumb-item>{{ $t('market.task.edit.addTitle') }}</n-breadcrumb-item>
			</n-breadcrumb>
			<div class="task-edit-box" :style="{ minHeight: height + 'px' }">
				<bt-form ref="formRef" class="task-edit-form" :model="form" :rules="rules">
					<div ref="formContentRef">
						<n-card class="mb-24px">
							<n-form-item :label="$t('market.task.edit.from')">
								<from-select v-model:value="form.addresser" v-model:name="form.full_name">
								</from-select>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.displayName')" path="full_name">
								<n-input
									v-model:value="form.full_name"
									:placeholder="$t('market.task.edit.displayNamePlaceholder')">
								</n-input>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.subject')" path="subject">
								<n-input
									v-model:value="form.subject"
									:placeholder="$t('market.task.edit.subjectPlaceholder')">
								</n-input>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.recipients')" type="group_ids">
								<group-select v-model:value="form.group_ids" v-model:label="groupNames">
								</group-select>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.template')" path="template_id">
								<div class="flex-1">
									<template-select
										v-model:value="form.template_id"
										v-model:content="templateContent">
									</template-select>
								</div>
								<n-button text type="primary" class="ml-12px" @click="handleGoTemplate">
									{{ t('common.actions.create') }}
								</n-button>
							</n-form-item>
							<n-form-item v-if="false" :label="$t('market.task.edit.saveOutbox')">
								<n-switch v-model:value="form.is_record" :checked-value="1" :unchecked-value="0">
								</n-switch>
							</n-form-item>
							<n-form-item :label="t('market.task.edit.ipWarmup')">
								<n-switch v-model:value="form.warmup" :checked-value="1" :unchecked-value="0">
								</n-switch>
								<span class="ml-12px text-desc">{{ t('market.task.edit.ipWarmupTip') }}</span>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.unsubscribeLink')">
								<n-switch v-model:value="form.unsubscribe" :checked-value="1" :unchecked-value="0">
								</n-switch>
								<!-- <n-button text type="primary" class="ml-16px" @click="handleViewCase">
							{{ t('market.task.edit.viewCase') }}
						</n-button> -->
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.threads')" :show-feedback="false">
								<n-radio-group v-model:value="threadsType" @update:value="handleUpdateThread">
									<n-radio :value="0">
										{{ $t('market.task.edit.threadsAuto') }}
									</n-radio>
									<n-radio :value="1">
										{{ $t('market.task.edit.threadsCustom') }}
									</n-radio>
								</n-radio-group>
								<div v-show="threadsType === 1" class="w-60px">
									<n-input-number
										v-model:value="form.threads"
										:min="1"
										:max="100"
										:show-button="false">
									</n-input-number>
								</div>
							</n-form-item>
						</n-card>
						<n-card>
							<n-form-item :label="t('market.task.edit.sendTime')" path="start_time">
								<n-radio-group
									v-model:value="sendTimeType"
									class="flex items-center"
									@update:value="handleUpdateSend">
									<n-radio :value="0">{{ t('market.task.edit.sendTimeNow') }}</n-radio>
									<n-radio class="items-center" :value="1"> </n-radio>
									<n-date-picker
										v-model:value="form.start_time"
										class="ml-8px"
										type="datetime"
										:disabled="sendTimeType === 0">
									</n-date-picker>
								</n-radio-group>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.remark')">
								<n-input
									v-model:value="form.remark"
									:placeholder="$t('market.task.edit.remarkPlaceholder')">
								</n-input>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.testEmail')" :show-feedback="false">
								<div class="flex-1 mr-10px">
									<n-input
										v-model:value="testEmail"
										:placeholder="$t('market.task.edit.testEmailPlaceholder')">
									</n-input>
								</div>
								<n-button @click="handleSendTest">
									{{ $t('market.task.edit.sendTest') }}
								</n-button>
							</n-form-item>
						</n-card>
					</div>
				</bt-form>
				<div class="task-preview">
					<n-card
						class="h-full"
						:content-style="{ display: 'flex', flexDirection: 'column', height: '100%' }">
						<div class="preview-header">
							<div class="preview-header-item">
								<div class="preview-label">{{ t('market.task.edit.from') }}:</div>
								<div class="preview-value">{{ form.addresser }}</div>
							</div>
							<div class="preview-header-item">
								<div class="preview-label">{{ t('market.task.edit.to') }}:</div>
								<div class="preview-value">
									{{ groupNames.length > 0 ? groupNames.join(', ') : '--' }}
								</div>
							</div>
							<div class="preview-header-item">
								<div class="preview-label">{{ t('market.task.edit.subject') }}:</div>
								<div class="preview-value">{{ form.subject || '--' }}</div>
							</div>
						</div>
						<div class="preview-content">
							<bt-preview v-if="templateContent" :value="templateContent" />
							<div v-else class="empty-preview">{{ t('market.task.edit.preview.empty') }}</div>
						</div>
					</n-card>
				</div>
			</div>
		</div>
		<!-- Action buttons -->
		<div class="action-buttons">
			<n-button type="primary" @click="handleSubmit">
				{{ $t('common.actions.confirm') }}
			</n-button>
			<n-button class="ml-16px" secondary @click="handleGoBack">
				{{ $t('common.actions.back') }}
			</n-button>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { useElementBounding } from '@vueuse/core'
import { confirm, Message } from '@/utils'
import { addTask, sendTestEmail } from '@/api/modules/market/task'

import FromSelect from './components/FromSelect.vue'
import GroupSelect from './components/GroupSelect.vue'
import TemplateSelect from './components/TemplateSelect.vue'

const { t } = useI18n()

const router = useRouter()

const formRef = useTemplateRef('formRef')

const formContentRef = useTemplateRef('formContentRef')

const { height } = useElementBounding(formContentRef)

// 表单数据
const form = reactive({
	addresser: null as null | string,
	full_name: '',
	subject: '',
	group_ids: [] as number[],
	template_id: null as null | number,
	is_record: 1,
	unsubscribe: 1,
	warmup: 0,
	threads: 0,
	start_time: null as number | null,
	remark: '',
})

const rules: FormRules = {
	full_name: {
		trigger: ['blur', 'input'],
		validator: () => {
			if (form.full_name === '') {
				return new Error(t('market.task.edit.validation.displayNameRequired'))
			}
			return true
		},
	},
	subject: {
		trigger: ['blur', 'input'],
		validator: () => {
			if (form.subject === '') {
				return new Error(t('market.task.edit.validation.subjectRequired'))
			}
			return true
		},
	},
	group_ids: {
		trigger: 'change',
		validator: () => {
			if (form.group_ids.length === 0) {
				return new Error(t('market.task.edit.validation.recipientsRequired'))
			}
			return true
		},
	},
	template_id: {
		trigger: 'change',
		validator: () => {
			if (form.template_id === null) {
				return new Error(t('market.task.edit.validation.templateRequired'))
			}
			return true
		},
	},
	start_time: {
		validator: () => {
			if (sendTimeType.value === 1 && form.start_time === null) {
				return new Error(t('market.task.edit.validation.sendTimeRequired'))
			}
			return true
		},
	},
}

// 分组名称
const groupNames = ref<string[]>([])

// 线程类型
const threadsType = ref(0)

// 发送时间类型
const sendTimeType = ref(0)

// 测试邮件
const testEmail = ref('')

// 模板内容
const templateContent = ref('')

// 跳转模板页面
const handleGoTemplate = () => {
	router.push('/market/template')
}

// 查看案例
// const handleViewCase = () => {}

// 更新线程类型
const handleUpdateThread = (val: number) => {
	if (val === 0) {
		form.threads = 0
	} else {
		form.threads = 5
	}
}

const handleUpdateSend = () => {
	if (sendTimeType.value === 0) {
		form.start_time = null
	}
}

// 发送测试邮件
const handleSendTest = async () => {
	if (!testEmail.value) {
		Message.error(t('market.task.edit.validation.testEmailRequired'))
		return
	}
	if (!form.subject) {
		Message.error(t('market.task.edit.validation.subjectRequired'))
		return
	}
	if (!form.template_id) {
		Message.error(t('market.task.edit.validation.templateRequired'))
		return
	}

	await sendTestEmail({
		addresser: form.addresser || '',
		subject: form.subject,
		recipient: testEmail.value,
		template_id: form.template_id || 0,
	})
}

const handleGoBack = () => {
	if (form.subject || form.group_ids.length > 0) {
		confirm({
			title: t('market.task.edit.discard.title'),
			content: t('market.task.edit.discard.content'),
			onConfirm: () => {
				router.go(-1)
			},
		})
	} else {
		router.go(-1)
	}
}

const getParams = () => {
	let startTime = form.start_time
	if (sendTimeType.value === 0 || startTime === null) {
		startTime = Date.now()
	}
	return {
		track_open: 1,
		track_click: 1,
		addresser: form.addresser || '',
		full_name: form.full_name,
		subject: form.subject,
		group_ids: form.group_ids,
		template_id: form.template_id || 0,
		is_record: form.is_record,
		unsubscribe: form.unsubscribe,
		warmup: form.warmup,
		threads: form.threads,
		start_time: startTime / 1000,
		remark: form.remark,
	}
}

const handleSubmit = async () => {
	await formRef.value?.validate()
	await addTask(getParams())
	router.push('/market/task')
}
</script>

<style lang="scss" scoped>
.task-edit-container {
	position: relative;
	display: flex;
	flex-direction: column;
	height: 100%;
	overflow: hidden;

	.task-edit-box {
		flex: 1;
		display: flex;
		gap: 24px;
		min-height: 0;
	}

	.task-edit-form {
		flex: 1;
	}

	.task-preview {
		flex: 1;
		overflow: hidden;

		.preview-header {
			width: 420px;
			margin: 0 auto;
			margin-bottom: 16px;

			.preview-header-item {
				display: flex;
				margin-bottom: 4px;

				.preview-label {
					min-width: 150px;
					text-align: right;
					font-weight: 500;
					margin-right: 8px;
				}

				.preview-value {
					flex: 1;
					color: #666;
				}
			}
		}

		.preview-content {
			flex: 1;
			min-height: 400px;
			overflow: hidden;

			.empty-preview {
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100%;
				color: #999;
			}
		}
	}
}

.action-buttons {
	width: 100%;
	padding: 23px 24px;
	background-color: var(--color-bg-1);
	border-top: 1px solid var(--color-border-1);
}
</style>
