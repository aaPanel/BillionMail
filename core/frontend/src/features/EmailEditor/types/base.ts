export type ColumnType = 'columns'

export type CellType = 'cell'

export type ButtonType = 'button'

export type LinkType = 'link'

export type DividerType = 'divider'

export type HeaderType = 'header'

export type TextType = 'text'

export type ImageType = 'image'

export type MenuType = 'menu'

export type HtmlType = 'html'

/**
 * @description 所有块类型的联合类型
 */
export type BlockType =
	| ColumnType // 列
	| 'placeholder' // 占位符
	| CellType // 单元格
	| ButtonType // 按钮
	| LinkType // 链接
	| DividerType // 分隔线
	| HeaderType // 标题
	| TextType // 文本
	| ImageType // 图片
	| MenuType // 菜单
	| HtmlType // HTML

/**
 * @description 基础块的数据结构
 */
export type BlockBaseType = {
	key: string
	name: string
	type: BlockType
}

/**
 * @description 列的数据结构
 */
export type BlockColumnsType = BlockBaseType & {
	children: string[]
}

/**
 * @description 单元格的数据结构
 */
export type BlockCellType = BlockBaseType & {
	width: number
	children: string[]
}

/**
 * @description 基础元数据配置
 */
export type BaseMeta = {
	createdAt: string // 创建时间
	updatedAt: string // 最后修改时间
	version: string // 版本号（用于撤销/重做）
	locked?: boolean // 是否锁定编辑
}

/**
 * @description 更多配置
 */
export type MoreStyle<K> = {
	more: boolean
	all: K
	top: K
	right: K
	bottom: K
	left: K
}

/**
 * @description 边框样式配置
 */
export type BorderStyle = {
	color: string
	width: string
	style: string
}

/**
 * @description 基础样式配置
 */
export type BaseStyle = {
	display?: 'inline-block' | 'block' | 'inline' // 显示方式
	width?: string // 宽度
	height?: string // 高度
	backgroundColor?: string // 背景色
	textAlign?: 'left' | 'center' | 'right' // 文字对齐
	padding?: MoreStyle<string> // 内边距
	border?: MoreStyle<BorderStyle> // 边框
	borderRadius?: MoreStyle<string> // 圆角
	fontWeight?: string // 字体粗细
	fontSize?: string // 字体大小
	lineHeight?: string // 行高
	letterSpacing?: string // 字间距
	color?: string // 字体颜色
	verticalAlign?: 'top' | 'middle' | 'bottom' // 垂直对齐
	borderTop?: BorderStyle // 上边框
	layout?: 'horizontal' | 'vertical' // 布局
	boxSizing?: 'border-box' | 'content-box' // 盒模型
	textDecoration?: string
	margin?: string
}

/**
 * @description 基础属性
 */
export type BaseAttr = {
	href?: string
	target?: string
	content?: string
	src?: string
	alt?: string
	links?: { id: string; label: string; href: string }[]
}

/**
 * @description 基础配置
 */
export type BaseConfig = {
	// 唯一key
	key: string
	// 名称
	name: string
	// 类型
	type: BlockType
	// 元数据
	meta: BaseMeta
	// 样式配置
	style: BaseStyle
	// 容器样式
	containerStyle: BaseStyle
	// 属性
	attr: BaseAttr
}

/**
 * @description PageConfig页面的数据结构
 */
export type PageConfig = {
	// 元数据
	meta: BaseMeta
	// 样式配置
	style: BaseStyle
}

/**
 * @description ColumnMap列的数据表
 */
export type ColumnMap = {
	[key: string]: BlockColumnsType
}

/**
 * @description ColumnMap列的数据表
 */
export type ColumnConfigMap = {
	[key: string]: BaseConfig
}

/**
 * @description CellMap单元格的数据表
 */
export type CellMap = {
	[key: string]: BlockCellType
}

/**
 * @description CellConfigMap单元格的数据表
 */
export type CellConfigMap = {
	[key: string]: BaseConfig
}

/**
 * @description BlockMap组件的数据表
 */
export type BlockMap = {
	[key: string]: BlockBaseType
}

/**
 * @description BlockConfigMap组件的数据表
 */
export type BlockConfigMap = {
	[key: string]: BaseConfig
}

/**
 * @description 菜单组件信息
 */
export type MenuBlock = {
	// 菜单类型，必选
	type: BlockType
	// 菜单名称，必选
	name: string
	// 菜单组件唯一key，必选
	key: string
}

/**
 * @description 拖拽添加列的数据结构
 */
export type DropAddColumnBlock = {
	type: 'ADD_COLUMNS_BLOCK'
	blockData: MenuBlock // 块数据
}

/**
 * @description 拖拽添加块的数据结构
 */
export type DropAddBaseBlock = {
	type: 'ADD_BASE_BLOCK'
	blockData: MenuBlock // 块数据
}

/**
 * @description 拖拽排序列的数据结构
 */
export type DropSortColumnBlock = {
	type: 'SORT_COLUMNS_BLOCK'
	sourceIndex: number // 源索引
	blockData: BlockColumnsType // 块数据
}

/**
 * @description 拖拽排序块的数据结构
 */
export type DropSortBaseBlock = {
	type: 'SORT_BASE_BLOCK'
	sourceIndex: number // 源索引
	sourceCellKey: string // 源单元格
	blockData: BlockBaseType // 块数据
}

/**
 * @description 拖拽块的数据结构
 */
export type DropBaseBlock =
	| DropAddColumnBlock
	| DropAddBaseBlock
	| DropSortBaseBlock
	| DropSortColumnBlock

export type MoveLinkFunc = (id: string, to: number) => void

export type FindLinkFunc = (id: string) => { index: number }
