import { createApp, watch } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import i18n from './locales'
import './styles/index.scss'

const app = createApp(App)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(i18n)

const elLocaleMap: Record<string, any> = {
  'zh-CN': zhCn,
  'en-US': en,
}
const currentElLocale = elLocaleMap[i18n.global.locale.value] || en
app.use(ElementPlus, { locale: currentElLocale })

watch(
  () => i18n.global.locale.value,
  (val) => {
    const elLocale = elLocaleMap[val] || en
    ;(ElementPlus as any).locale?.(elLocale)
  },
)

app.mount('#app')
