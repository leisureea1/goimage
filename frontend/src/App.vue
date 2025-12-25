<script setup>
/**
 * 应用根组件
 * 包含导航栏和路由视图
 */
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
</script>

<template>
  <div class="app">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <h1 class="logo">📷 图床</h1>
        <nav class="nav">
          <router-link 
            to="/" 
            :class="{ active: route.path === '/' }"
          >
            上传
          </router-link>
          <router-link 
            to="/gallery" 
            :class="{ active: route.path === '/gallery' }"
          >
            管理
          </router-link>
        </nav>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: white;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.logo {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
}

.nav {
  display: flex;
  gap: 24px;
}

.nav a {
  color: var(--text-secondary);
  font-size: 15px;
  padding: 8px 0;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.nav a:hover,
.nav a.active {
  color: var(--primary-color);
}

.nav a.active {
  border-bottom-color: var(--primary-color);
}

.main {
  flex: 1;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 24px 20px;
}
</style>
