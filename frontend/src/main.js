/**
 * 应用入口
 * 初始化 Vue 应用和路由
 */
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/main.css'

const app = createApp(App)
app.use(router)
app.mount('#app')
