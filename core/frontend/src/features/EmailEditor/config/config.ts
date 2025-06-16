import { BlockColumnsType, BlockCellType, BaseConfig, BlockType } from '../types/base'
import { getRandom } from '../utils'

export const columnsData: BlockColumnsType = {
	key: '',
	type: 'columns',
	name: '',
	children: [],
}

export const cellData: BlockCellType = {
	key: '',
	type: 'cell',
	name: '',
	width: 100,
	children: [],
}

export const columnsConfig: BaseConfig = {
	key: '',
	name: 'Columns',
	type: 'columns',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {},
	style: {
		backgroundColor: '',
		padding: {
			more: false,
			all: '0',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
	containerStyle: {
		backgroundColor: '',
	},
}

export const cellConfig: BaseConfig = {
	key: '',
	name: 'Cell',
	type: 'cell',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {},
	style: {
		backgroundColor: '',
		padding: {
			more: false,
			all: '0',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
		border: {
			more: false,
			all: {
				color: '',
				width: '0',
				style: 'solid',
			},
			top: {
				color: '',
				width: '0',
				style: 'solid',
			},
			right: {
				color: '',
				width: '0',
				style: 'solid',
			},
			bottom: {
				color: '',
				width: '0',
				style: 'solid',
			},
			left: {
				color: '',
				width: '0',
				style: 'solid',
			},
		},
	},
	containerStyle: {},
}

export const buttonConfig: BaseConfig = {
	key: '',
	name: 'Button',
	type: 'button',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {
		href: '#',
		target: '_blank',
		content: 'Button Text',
	},
	style: {
		display: 'inline-block',
		backgroundColor: '#20A53A',
		color: '#fff',
		width: 'auto',
		fontWeight: 'normal',
		fontSize: '14px',
		lineHeight: '120%',
		letterSpacing: '0px',
		boxSizing: 'border-box',
		textDecoration: 'none',
		padding: {
			more: true,
			all: '0',
			top: '10px',
			left: '20px',
			right: '20px',
			bottom: '10px',
		},
		border: {
			more: false,
			all: {
				color: '',
				width: '0',
				style: 'solid',
			},
			top: {
				color: '',
				width: '0',
				style: 'solid',
			},
			right: {
				color: '',
				width: '0',
				style: 'solid',
			},
			bottom: {
				color: '',
				width: '0',
				style: 'solid',
			},
			left: {
				color: '',
				width: '0',
				style: 'solid',
			},
		},
		borderRadius: {
			more: false,
			all: '4px',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
	containerStyle: {
		textAlign: 'center',
		padding: {
			more: false,
			all: '10px',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
}

export const linkConfig: BaseConfig = {
	key: '',
	name: 'Link',
	type: 'link',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {
		href: '',
		target: '_blank',
		content: 'Link Text',
	},
	style: {
		display: 'inline-block',
		color: '#333',
		fontWeight: 'normal',
		fontSize: '14px',
		lineHeight: '120%',
		letterSpacing: '0px',
		textDecoration: 'none',
	},
	containerStyle: {
		textAlign: 'center',
		padding: {
			more: false,
			all: '10px',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
}

export const dividerConfig: BaseConfig = {
	key: '',
	name: 'Divider',
	type: 'divider',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {},
	style: {
		display: 'inline-block',
		width: '100%',
		margin: '0px',
		verticalAlign: 'middle',
		boxSizing: 'border-box',
		borderTop: {
			color: '#bbbbbb',
			width: '1px',
			style: 'solid',
		},
	},
	containerStyle: {
		lineHeight: '1px',
		textAlign: 'center',
		padding: {
			more: true,
			all: '0',
			top: '16px',
			left: '10px',
			right: '10px',
			bottom: '16px',
		},
	},
}

export const headerConfig: BaseConfig = {
	key: '',
	name: 'Header',
	type: 'header',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {
		content: 'Header',
	},
	style: {
		display: 'inline-block',
		color: '#333',
		fontWeight: 'normal',
		fontSize: '22px',
		lineHeight: '140%',
		letterSpacing: '0px',
	},
	containerStyle: {
		textAlign: 'left',
		padding: {
			more: false,
			all: '10px',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
}

export const textConfig: BaseConfig = {
	key: '',
	name: 'Text',
	type: 'text',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {
		content: 'Text',
	},
	style: {
		display: 'inline-block',
		color: '#333',
		fontWeight: 'normal',
		fontSize: '14px',
		lineHeight: '120%',
		letterSpacing: '0px',
	},
	containerStyle: {
		textAlign: 'left',
		padding: {
			more: false,
			all: '10px',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
}

export const imageConfig: BaseConfig = {
	key: '',
	name: 'Image',
	type: 'image',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {
		href: '',
		target: '_blank',
		src: '',
		alt: '',
	},
	style: {
		display: 'inline-block',
		width: '100%',
		boxSizing: 'border-box',
		textDecoration: 'none',
	},
	containerStyle: {
		textAlign: 'center',
		padding: {
			more: false,
			all: '10px',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
}

export const menuConfig: BaseConfig = {
	key: '',
	name: 'Menu',
	type: 'menu',
	meta: {
		version: '',
		createdAt: '',
		updatedAt: '',
	},
	attr: {
		links: [
			{ id: getRandom(6), label: 'Menu', href: '' },
			{ id: getRandom(6), label: 'Menu', href: '' },
		],
	},
	style: {
		display: 'inline-block',
		fontWeight: 'normal',
		fontSize: '14px',
		letterSpacing: '0px',
		color: '#333',
		textDecoration: 'none',
		padding: {
			more: true,
			all: '',
			top: '5px',
			left: '15px',
			right: '15px',
			bottom: '5px',
		},
	},
	containerStyle: {
		textAlign: 'left',
		padding: {
			more: false,
			all: '10px',
			top: '0',
			left: '0',
			right: '0',
			bottom: '0',
		},
	},
}

export const baseConfigMap = new Map<BlockType, BaseConfig>([
	['button', buttonConfig],
	['link', linkConfig],
	['divider', dividerConfig],
	['header', headerConfig],
	['text', textConfig],
	['image', imageConfig],
	['menu', menuConfig],
])

export const copyrightVNode = h(
	'div',
	{
		style: {
			display: 'block',
			padding: '10px',
			lineHeight: '140%',
			color: '#999',
			textAlign: 'center',
			fontSize: '12px',
		},
	},
	[
		h('span', 'Powered by '),
		h(
			'a',
			{
				href: 'https://www.billionmail.com',
				target: '_blank',
				style: {
					color: '#999',
					textDecoration: 'underline',
				},
			},
			'BillionMail'
		),
	]
)
