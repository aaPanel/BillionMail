import { createApp } from 'vue'
import i18n from '@/i18n'
import pinia from '@/store'
import router from '@/router'
import App from '@/App.vue'

import '@unocss/reset/normalize.css'
import '@/styles/uno.css'
import '@/styles/index.scss'

const app = createApp(App)

app.use(i18n)
app.use(pinia)
app.use(router)
app.mount('#root')

export { app }
