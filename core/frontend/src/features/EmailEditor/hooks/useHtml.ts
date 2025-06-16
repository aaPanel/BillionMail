import { render, StyleValue } from 'vue'
import { storeToRefs } from 'pinia'
import { useEmailEditorStore } from '../store'
import { BaseConfig, BaseStyle } from '../types/base'
import { useStyle } from './useStyle'
import { copyrightVNode } from '../config/config'

export const useHtml = () => {
	const store = useEmailEditorStore()
	const {
		columnsSource,
		columnsMap,
		columnsConfigMap,
		cellMap,
		cellConfigMap,
		blockMap,
		blockConfigMap,
		pageConfig,
	} = storeToRefs(store)

	const { configToStyle } = useStyle()

	const html = ref('')

	const generateHtml = () => {
		html.value = columnsSourceToTable().outerHTML
	}

	const columnsSourceToTable = () => {
		const columnTrMap = columnToTr()
		const table = document.createElement('table')
		table.style.backgroundColor = pageConfig.value.style.backgroundColor || '#fff'
		table.style.width = '100%'
		columnsSource.value.forEach(columnKey => {
			if (columnTrMap[columnKey]) {
				table.appendChild(columnTrMap[columnKey].dom)
			}
		})
		table.appendChild(cellToCopyright())
		return table
	}

	const columnToTr = () => {
		const tdMap = cellToTd()
		const trMap: Record<string, { dom: HTMLElement }> = {}
		Object.entries(columnsMap.value).forEach(([columnsKey, columns]) => {
			const tr = document.createElement('tr')
			const td = document.createElement('td')
			const div = document.createElement('div')
			setElementStyle(div, columnsConfigMap.value[columnsKey].style)
			div.style.width = pageConfig.value.style.width || 'auto'
			div.style.margin = '0 auto'
			const childTable = document.createElement('table')
			childTable.style.width = '100%'
			const childTableTr = document.createElement('tr')
			columns.children.forEach(cellKey => {
				childTableTr.appendChild(tdMap[cellKey].dom)
			})
			childTable.appendChild(childTableTr)
			div.appendChild(childTable)
			td.appendChild(div)
			tr.appendChild(td)
			trMap[columnsKey] = {
				dom: tr,
			}
		})

		return trMap
	}

	const cellToCopyright = () => {
		const tr = document.createElement('tr')
		const td = document.createElement('td')
		render(copyrightVNode, td)
		tr.appendChild(td)
		return tr
	}

	const cellToTd = () => {
		const blockMap = compToElement()
		const tdMap: Record<string, { dom: HTMLElement }> = {}
		Object.entries(cellMap.value).forEach(([cellKey, cell]) => {
			// 生成cell的单元格td
			const cellTd = document.createElement('td')
			// 设置单元格宽度
			cellTd.style.width = `${cell.width}%`
			// 设置单元格样式
			setElementStyle(cellTd, cellConfigMap.value[cellKey].style)
			// 将组件dom放到单元格内,并设置单元格的样式
			cell.children.forEach(blockKey => {
				if (!blockMap[blockKey]) return
				cellTd.appendChild(blockMap[blockKey].dom)
			})

			tdMap[cellKey] = {
				dom: cellTd,
			}
		})
		return tdMap
	}

	/**
	 * @description 遍历生成所有的组件
	 */
	const compToElement = () => {
		const elementMap: Record<
			string,
			{
				dom: HTMLElement
				config: BaseConfig
			}
		> = {}

		Object.keys(blockMap.value).forEach(blockKey => {
			const block = blockMap.value[blockKey]
			const blockConfig = blockConfigMap.value[blockKey]

			const containerDom = document.createElement('div')
			setElementStyle(containerDom, blockConfig.containerStyle)

			switch (block.type) {
				case 'button':
					containerDom.appendChild(buttonToElement(blockConfig))
					break
				case 'link':
					containerDom.appendChild(linkToElement(blockConfig))
					break
				case 'divider':
					containerDom.appendChild(dividerToElement(blockConfig))
					break
				case 'header':
					containerDom.appendChild(headerToElement(blockConfig))
					break
				case 'text':
					containerDom.appendChild(textToElement(blockConfig))
					break
				case 'image':
					containerDom.appendChild(imageToElement(blockConfig))
					break
				case 'menu':
					containerDom.appendChild(menuToElement(blockConfig))
					break
			}

			elementMap[blockKey] = {
				dom: containerDom,
				config: blockConfig,
			}
		})
		return elementMap
	}

	/**
	 * @description 设置元素样式
	 */
	const setElementStyle = (elDom: HTMLElement, configStyle: BaseStyle) => {
		const style = configToStyle(configStyle)
		Object.entries(style).forEach(([key, value]) => {
			elDom.style[key as keyof StyleValue] = value
		})
	}

	/**
	 * @description 通过内容寻找P标签，并设置样式
	 */
	const setDefaultStyle = (content: string) => {
		return content.replace('<p>', '<p style="margin: 0;">')
	}

	/**
	 * @description 按钮组件
	 */
	const buttonToElement = (config: BaseConfig) => {
		const elDom = document.createElement('a')
		elDom.href = config.attr.href ?? ''
		elDom.target = config.attr.target ?? ''
		elDom.innerHTML = setDefaultStyle(config.attr.content ?? '')
		setElementStyle(elDom, config.style)
		return elDom
	}

	/**
	 * @description 链接组件
	 */
	const linkToElement = (config: BaseConfig) => {
		const elDom = document.createElement('a')
		elDom.href = config.attr.href ?? ''
		elDom.target = config.attr.target ?? ''
		elDom.innerHTML = setDefaultStyle(config.attr.content ?? '')
		setElementStyle(elDom, config.style)
		return elDom
	}

	/**
	 * @description 分割线组件
	 */
	const dividerToElement = (config: BaseConfig) => {
		const elDom = document.createElement('hr')
		setElementStyle(elDom, config.style)
		return elDom
	}

	/**
	 * @description Header组件
	 */
	const headerToElement = (config: BaseConfig) => {
		const elDom = document.createElement('div')
		elDom.innerHTML = setDefaultStyle(config.attr.content ?? '')
		setElementStyle(elDom, config.style)
		return elDom
	}

	/**
	 * @description 文本组件
	 */
	const textToElement = (config: BaseConfig) => {
		const elDom = document.createElement('div')
		elDom.innerHTML = setDefaultStyle(config.attr.content ?? '')
		setElementStyle(elDom, config.style)
		return elDom
	}

	/**
	 * @description 图片组件
	 */
	const imageToElement = (config: BaseConfig) => {
		const elDom = document.createElement('img')
		elDom.src = config.attr.src ?? ''
		elDom.alt = config.attr.alt ?? ''
		if (config.attr.href) {
			const aDom = document.createElement('a')
			aDom.href = config.attr.href ?? ''
			aDom.target = config.attr.target ?? ''
			elDom.style.width = '100%'
			aDom.appendChild(elDom)
			setElementStyle(aDom, config.style)
			return aDom
		}
		setElementStyle(elDom, config.style)
		return elDom
	}

	/**
	 * @description 菜单组件
	 */
	const menuToElement = (config: BaseConfig) => {
		const elDom = document.createElement('div')
		config.attr.links?.forEach(item => {
			const linkDom = document.createElement('a')
			setElementStyle(linkDom, config.style)
			linkDom.href = item.href
			linkDom.target = config.attr.target ?? ''
			linkDom.innerHTML = item.label
			elDom.appendChild(linkDom)
		})
		return elDom
	}

	return {
		html,
		generateHtml,
	}
}
