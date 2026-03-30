import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './styles/index.css'
import 'vue-sonner/style.css' // 全局悬浮提示样式

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')

// 禁止长按右键菜单（移动端 App 体验）
document.addEventListener('contextmenu', (e) => e.preventDefault())
