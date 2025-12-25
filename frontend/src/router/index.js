/**
 * 路由配置
 * 定义应用的页面路由
 */
import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Upload',
    component: () => import('../views/UploadView.vue'),
    meta: { title: '上传图片' }
  },
  {
    path: '/gallery',
    name: 'Gallery',
    component: () => import('../views/GalleryView.vue'),
    meta: { title: '图片管理' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 更新页面标题
router.beforeEach((to, from, next) => {
  document.title = `${to.meta.title || '图床系统'} - 图床`
  next()
})

export default router
