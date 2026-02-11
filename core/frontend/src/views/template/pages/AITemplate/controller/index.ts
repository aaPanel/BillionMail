import i18n from '@/i18n'
import { instance } from '@/api'
import { ChatInfo, TemplateStore, UsageInfo } from '../dto'

const { t } = i18n.global

export const instanceOptions = {
	fetchOptions: {
		loading: t('template.ai.loading.message'),
		successMessage: true,
	},
}

/**
 * @description Create new template
 */
export async function createNewTemplate(chatId: string, temp_name: string) {
	try {
		await instance.post('/email_template/create', {
			temp_name,
			add_type: 2,
			chat_id: chatId,
			html_content: '',
		})
	} catch (error) {
		console.warn(error)
		throw error
	}
}

/**
 * @description Initial template info
 */
export async function initialTemplateInfo(store: TemplateStore) {
	await getModelList(store)
	await getChatInfo(store)
}

/**
 * @description Create ai template
 */
export async function createAiTemplate(sourceDomain: string) {
	try {
		const temp_name = `AI_temp_${new Date().getTime()}`
		const res = (await instance.post(
			'/askai/chat/create_chat',
			{
				domain: sourceDomain,
				add_type: 99,
				temp_name,
			},
			instanceOptions
		)) as Record<string, any>
		await createNewTemplate(res.chatId, temp_name)
		return res.chatId
	} catch (error) {
		console.warn(error)
		window.$message?.error('AI template creation failed. Check AI model configuration.')
		return null
	}
}

/**
 * @description Get model list
 */
export async function getModelList(store: TemplateStore) {
	const { modelList, currentModel, currentModelTitle } = store
	try {
		modelList.value = await instance.post('/askai/supplier/models')
		if (modelList.value.length) {
			currentModel.value = modelList.value[0]
			currentModelTitle.value = modelList.value[0].title
		}
	} catch (error) {
		console.warn(error)
	}
}

/**
 * @description Get chat list
 */
export async function getChatList() {
	// const { } = store
	try {
		await instance.post('/askai/chat/get_chat_list')
	} catch (error) {
		console.warn(error)
	}
}

/**
 * @description Get chat info
 */
export async function getChatInfo(store: TemplateStore) {
	const { chatId, chatInfo, currentModel, modelList, currentModelTitle, chatRecord, usageRecord } =
		store
	try {
		chatInfo.value = (await instance.post('/askai/chat/info', { chatId: chatId.value })) as ChatInfo
		// Check chatInfo and assign currentModel and currentModelTitle
		if (chatInfo.value.modelId && modelList.value.length) {
			const findModel = modelList.value.find(item => item.modelId == chatInfo.value.modelId)
			if (findModel) {
				currentModel.value = findModel
				currentModelTitle.value = findModel.title
			}
		}
		// Init chatRecord
		for (let i = 0; i < chatInfo.value.messages.length; i++) {
			const key = `${chatInfo.value.messages[i].content}_+_${chatRecord.value.size}`
			if (chatInfo.value.messages[i].role == 'user') {
				chatRecord.value.set(key, sliceContentToArray(chatInfo.value.messages[i + 1].content))
				usageRecord.value.set(key, chatInfo.value.messages[i + 1].usage)
			}
		}
	} catch (error) {
		console.warn(error)
	}
}

/**
 * @description Stop chat
 */
export async function stopChat(store: TemplateStore) {
	const { chatId, isChat } = store
	try {
		await instance.post('/askai/chat/stop', { chatId: chatId.value }, instanceOptions)
		isChat.value = false
	} catch (error) {
		console.warn(error)
	}
}

/***
 * @description Check shift+enter
 */
function checkShiftEnter(e: KeyboardEvent, content: Ref<string>) {
	if (e.shiftKey && e.key === 'Enter') {
		e.preventDefault()
		const textarea = e.target as HTMLTextAreaElement
		const start = textarea.selectionStart
		const end = textarea.selectionEnd
		content.value = content.value.substring(0, start) + '\n' + content.value.substring(end)
		textarea.selectionStart = textarea.selectionEnd = start + 1
		return true
	}
}

/**
 * Send chat and generate markdown
 */
export async function sendChat(store: TemplateStore, e?: KeyboardEvent) {
	const {
		scrollable,
		generateShow,
		chatId,
		currentModel,
		questionContent,
		chatRecord,
		currentChatRecordKey,
		scrollWrapperRef,
		chatScrollRef,
		isChat,
		usageRecord,
		useSpinTax,
		spinTaxLength,
	} = store
	if (isChat.value) return
	if (!questionContent.value) return
	if (e && (e.isComposing || e.keyCode === 229)) return
	if (e && checkShiftEnter(e, questionContent)) return
	const chatRecordKey = `${questionContent.value}_+_${chatRecord.value.size}`
	chatRecord.value.set(chatRecordKey, [])
	currentChatRecordKey.value = chatRecordKey
	let chatContent = questionContent.value
	questionContent.value = ''
	// Split string array
	let resultArray: string[] = []
	// Answer text
	let answerText = ''
	// Character pointer position
	let strPos = 0

	// Check Whether use spinTax
	if (useSpinTax.value) {
		chatContent += `  ${t('template.ai.spintax.instruction', { count: spinTaxLength.value })}`
	}

	/**
	 * @description spliced content
	 */
	function answerTextDeal(sseArr: Array<string>) {
		for (let index = 0; index < sseArr.length; index++) {
			if (sseArr[index]) {
				answerText += JSON.parse(sseArr[index]).content
			}
		}
	}
	try {
		scrollable.value = true
		generateShow.value = true
		isChat.value = true
		nextTick(() =>
			chatScrollRef.value.scrollTo({ left: 0, top: scrollWrapperRef.value.offsetHeight })
		)
		await instance.post(
			'/askai/chat/chat',
			{
				chatId: chatId.value,
				supplierName: currentModel.value.supplierName,
				modelId: currentModel.value.modelId,
				content: chatContent,
				is_text: 'false',
			},
			{
				fetchOptions: {
					cancelResInterceptor: true,
				},
				responseType: 'text',
				onDownloadProgress: progressEvent => {
					generateShow.value = false
					try {
						// Get the native XMLHttpRequest instance and extract the latest response content.
						const xhr = (progressEvent.event as ProgressEvent).target as XMLHttpRequest
						const resText = xhr.responseText
						// Get the incremental characters and record the latest character pointer position.
						const newContent = resText.slice(strPos)
						strPos = resText.length
						resultArray = parseSSEToObjects(newContent) as string[]
						answerTextDeal(resultArray)
						chatRecord.value.set(chatRecordKey, sliceContentToArray(answerText))
						if (scrollable.value) {
							setTimeout(
								() =>
									chatScrollRef.value.scrollTo({
										left: 0,
										top: scrollWrapperRef.value.offsetHeight,
									}),
								200
							)
						}
					} catch (error) {
						console.warn(error)
					}
				},
			}
		)
		scrollable.value = true
		chatScrollRef.value.scrollTo({ left: 0, top: scrollWrapperRef.value.offsetHeight })
		isChat.value = false
		getHtmlTemplateContent(store, (usage: UsageInfo) => {
			usageRecord.value.set(chatRecordKey, usage)
		})
	} catch (error) {
		console.warn(error)
	} finally {
		generateShow.value = false
		resultArray = []
		answerText = ''
		strPos = 0
	}
}

/**
 *
 * @param sseContent @
 */
function parseSSEToObjects(sseContent: string): Array<string> {
	// 1. Split by row
	const events = sseContent.split('\n')
	let jsonList = []
	for (const line of events) {
		// Match lines starting with "data:" (ignore the Spaces before and after)
		const trimmedLine = line.trim()
		if (trimmedLine.startsWith('data:')) {
			// Remove the prefix "data:" to obtain the JSON string
			const jsonStr = trimmedLine.slice('data:'.length).trim()
			if (jsonStr) {
				// Exclude null values (such as the empty data in the last row)
				jsonList.push(jsonStr)
			}
		}
	}
	return jsonList
}

/**
 * @description Segment the character content
 */
function sliceContentToArray(content: string) {
	// Match all code blocks (closed or unclosed)
	// Regular logic: Start with ' ', match to the end of the content or the next ' '(take the longer match)
	const codeBlockRegex = /```[\s\S]*?(?:```|$)/g
	const codeBlocks = content.match(codeBlockRegex) || []

	// Separate the non-code part
	const textParts = content.split(codeBlockRegex)

	// Merge text and code blocks
	const result = []
	const maxLength = Math.max(textParts.length, codeBlocks.length)

	for (let i = 0; i < maxLength; i++) {
		if (i < textParts.length && textParts[i].trim() !== '') {
			result.push(textParts[i])
		}
		if (i < codeBlocks.length) {
			result.push(codeBlocks[i])
		}
	}

	return result
}

/**
 * @description Remove markdown code mark
 */
export function removeHtmlCodeBlockMarkers(content: string) {
	return content.replace(/```html/g, '').replace(/```$/gm, '')
}

/**
 * @description Remove sign code for "<<<<<search" and ">>>>>replace"
 */
export function removeSignCode(content: string) {
	const signCodeRegex = /<<<<<<<\s+SEARCH[\s\S]*?=======([\s\S]*?)>>>>>>>\s+REPLACE/g
	return content.replace(signCodeRegex, '$1')
}

/**
 * @description Get content from title tags
 */
export function getContentFromTitleTags(content: string) {
	const match = content.match(/<title>(.*?)<\/title>/i)
	return match ? match[1] : ''
}

/**
 * @description Get html template code content
 */
export async function getHtmlTemplateContent(
	store: TemplateStore,
	callback?: (usage: UsageInfo) => void
) {
	const { chatId, previewCode, previewTit } = store
	try {
		const codeContent = (await instance.post('/askai/chat/get_html', {
			chatId: chatId.value,
		})) as Record<string, any>
		previewCode.value = codeContent.html_content
		previewTit.value = getContentFromTitleTags(previewCode.value)
		if (callback) {
			callback(codeContent.last_usage)
		}
	} catch (error) {
		console.warn(error)
	}
}

/**
 * @description Save code change
 */
export async function saveCodeChange(store: TemplateStore) {
	const { previewCode, chatId } = store
	try {
		await instance.post(
			'/askai/chat/modify_html',
			{ chatId: chatId.value, content: previewCode.value },
			instanceOptions
		)
		getHtmlTemplateContent(store)
	} catch (error) {
		console.warn(error)
	}
}

/**
 * @description Duplicate template
 */
export async function duplicateTemplate() {}
