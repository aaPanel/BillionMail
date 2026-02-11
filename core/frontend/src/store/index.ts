import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import useMenuStore from './modules/menu'
import useUserStore from './modules/user'
import useGlobalStore from './modules/global'
import useThemeStore from './modules/theme'
import useInstanceStore from './modules/instance'

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

export { useMenuStore, useUserStore, useGlobalStore, useThemeStore, useInstanceStore }

export default pinia
