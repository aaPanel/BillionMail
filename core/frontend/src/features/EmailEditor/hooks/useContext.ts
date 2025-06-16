import { useEmailEditorStore } from '../store'

// 上下文暴露内容
export interface EditorContext {
	store: ReturnType<typeof useEmailEditorStore>
}

// 定义注入类型
const contextKey = Symbol('EmailEditorContext')

export function createContext(instance: EditorContext) {
	provide(contextKey, instance)
}

export function useContext(): EditorContext {
	return inject(contextKey) as EditorContext
}
