import { Slots, VNodeChild } from 'vue'

import FormHead from './Head.vue'
import FormColor from './Color.vue'
import FormWidth from './Width.vue'
import FormWidthAuto from './WidthAuto.vue'
import FormWidthNumber from './WidthNumber.vue'
import FormFontWeight from './FontWeight.vue'
import FormFontSize from './FontSize.vue'
import FormLineHeight from './LineHeight.vue'
import FormLetterSpacing from './LetterSpacing.vue'
import FormTextAlign from './TextAlign.vue'
import FormLine from './Line.vue'
import FormBorder from './Border.vue'
import FormPadding from './Padding.vue'
import FormBorderRadius from './Rounded.vue'
import FormLayout from './Layout.vue'

type FormOption = {
	attrKey: string
	key?: string
	label?: string
	render?: () => VNodeChild
}

export function useNormalForm(options: FormOption[]) {
	const StyleForm = defineComponent({
		setup(props, { slots }) {
			return () => <CustomForm options={options}>{slots}</CustomForm>
		},
	})

	return [StyleForm]
}

const CustomForm = defineComponent({
	props: {
		value: {
			type: Object as PropType<Record<string, unknown>>,
			default: () => ({}),
		},
		options: {
			type: Array as PropType<FormOption[]>,
			default: () => [],
		},
	},
	setup(props, { slots }) {
		return () => (
			<>
				{props.options.map(option => {
					return getFormItem(option, props.value, slots)
				})}
			</>
		)
	},
})

function getFormItem(option: FormOption, formData: Record<string, unknown>, slots: Slots) {
	const { key, attrKey, label } = option

	const formKey = key || attrKey

	switch (attrKey) {
		case 'color':
		case 'backgroundColor':
			return <FormColor v-model:value={formData[formKey]} label={label} />
		case 'head':
			return <FormHead v-model:value={formData[formKey]} />
		case 'padding':
			return <FormPadding v-model:value={formData[formKey]} label={label} />
		case 'border':
			return <FormBorder v-model:value={formData[formKey]} label={label} />
		case 'borderTop':
			return <FormLine v-model:value={formData[formKey]} />
		case 'width':
			return <FormWidth v-model:value={formData[formKey]} />
		case 'widthAuto':
			return <FormWidthAuto v-model:value={formData[formKey]} />
		case 'widthNumber':
			return <FormWidthNumber v-model:value={formData[formKey]} />
		case 'fontWeight':
			return <FormFontWeight v-model:value={formData[formKey]} />
		case 'fontSize':
			return <FormFontSize v-model:value={formData[formKey]} />
		case 'lineHeight':
			return <FormLineHeight v-model:value={formData[formKey]} />
		case 'letterSpacing':
			return <FormLetterSpacing v-model:value={formData[formKey]} />
		case 'textAlign':
			return <FormTextAlign v-model:value={formData[formKey]} label={label} />
		case 'borderRadius':
			return <FormBorderRadius v-model:value={formData[formKey]} />
		case 'layout':
			return <FormLayout v-model:value={formData[formKey]} />
		case 'custom':
			return customSlotAndRender(option, slots)
	}
}

function customSlotAndRender(option: FormOption, slots: Slots) {
	if (option.render) {
		return <>{option.render()}</>
	} else if (option.key) {
		return slots[option.key]?.()
	}
}
