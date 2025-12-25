<script setup>
/**
 * 图片管理页面
 * 展示所有已上传的图片
 * 支持复制链接和删除操作
 */
import { ref, onMounted } from 'vue'
import { getImages, deleteImage } from '../api'

// 状态
const images = ref([])
const loading = ref(true)
const error = ref('')
const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0,
  totalPages: 0
})

// 删除确认
const deleteConfirm = ref(null)

// 复制弹窗
const showCopyModal = ref(false)
const copyTarget = ref(null)
const copySuccess = ref('')

/**
 * 获取完整 URL
 */
function getFullUrl(url) {
  return window.location.origin + url
}

/**
 * 获取各种格式的链接
 */
function getLinkFormats(url) {
  const fullUrl = getFullUrl(url)
  return {
    url: fullUrl,
    markdown: `![](${fullUrl})`,
    html: `<img src="${fullUrl}" alt="image" />`,
    bbcode: `[img]${fullUrl}[/img]`
  }
}

/**
 * 加载图片列表
 */
async function loadImages(page = 1) {
  loading.value = true
  error.value = ''

  try {
    const data = await getImages(page, pagination.value.pageSize)
    images.value = data.items || []
    pagination.value = {
      page: data.page,
      pageSize: data.page_size,
      total: data.total,
      totalPages: data.total_pages
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

/**
 * 显示复制选项弹窗
 */
function showCopyOptions(image) {
  copyTarget.value = {
    image,
    formats: getLinkFormats(image.url)
  }
  showCopyModal.value = true
  copySuccess.value = ''
}

/**
 * 复制到剪贴板
 */
async function copyToClipboard(text, label) {
  try {
    await navigator.clipboard.writeText(text)
    copySuccess.value = label
    setTimeout(() => {
      copySuccess.value = ''
    }, 2000)
  } catch (e) {
    error.value = '复制失败'
  }
}

/**
 * 关闭复制弹窗
 */
function closeCopyModal() {
  showCopyModal.value = false
  copyTarget.value = null
}

/**
 * 显示删除确认
 */
function showDeleteConfirm(image) {
  deleteConfirm.value = image
}

/**
 * 取消删除
 */
function cancelDelete() {
  deleteConfirm.value = null
}

/**
 * 确认删除
 */
async function confirmDelete() {
  if (!deleteConfirm.value) return

  const id = deleteConfirm.value.id
  try {
    await deleteImage(id)
    // 从列表中移除
    images.value = images.value.filter(img => img.id !== id)
    pagination.value.total--
    deleteConfirm.value = null
  } catch (e) {
    error.value = e.message
  }
}

/**
 * 格式化文件大小
 */
function formatSize(bytes) {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
}

/**
 * 格式化日期
 */
function formatDate(dateStr) {
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 页面加载时获取数据
onMounted(() => {
  loadImages()
})
</script>

<template>
  <div class="gallery-page">
    <div class="page-header">
      <h2>图片管理</h2>
      <span class="total-count">共 {{ pagination.total }} 张图片</span>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="message message-error">
      {{ error }}
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading">
      加载中...
    </div>

    <!-- 空状态 -->
    <div v-else-if="images.length === 0" class="empty">
      <div class="empty-icon">🖼️</div>
      <p>还没有上传任何图片</p>
      <router-link to="/" class="btn-primary" style="display: inline-block; margin-top: 16px;">
        去上传
      </router-link>
    </div>

    <!-- 图片网格 -->
    <div v-else class="image-grid">
      <div 
        v-for="image in images" 
        :key="image.id"
        class="image-card card"
        @contextmenu.prevent="showCopyOptions(image)"
      >
        <div class="image-preview">
          <img :src="image.url" :alt="image.id" />
        </div>
        <div class="image-info">
          <div class="image-meta">
            <span>{{ image.original_format.toUpperCase() }}</span>
            <span>{{ formatSize(image.processed_size) }}</span>
            <span>{{ image.width }}×{{ image.height }}</span>
          </div>
          <div class="image-date">
            {{ formatDate(image.created_at) }}
          </div>
        </div>
        <div class="image-actions">
          <button class="btn-primary" @click="showCopyOptions(image)">
            复制链接
          </button>
          <button class="btn-danger" @click="showDeleteConfirm(image)">
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="pagination.totalPages > 1" class="pagination">
      <button 
        class="btn-outline"
        :disabled="pagination.page <= 1"
        @click="loadImages(pagination.page - 1)"
      >
        上一页
      </button>
      <span class="page-info">
        {{ pagination.page }} / {{ pagination.totalPages }}
      </span>
      <button 
        class="btn-outline"
        :disabled="pagination.page >= pagination.totalPages"
        @click="loadImages(pagination.page + 1)"
      >
        下一页
      </button>
    </div>

    <!-- 复制选项弹窗 -->
    <div v-if="showCopyModal" class="modal-overlay" @click="closeCopyModal">
      <div class="modal copy-modal" @click.stop>
        <h3>复制链接</h3>
        <div class="copy-preview">
          <img :src="copyTarget?.image.url" :alt="copyTarget?.image.id" />
        </div>
        <div class="copy-options">
          <div class="copy-option">
            <div class="copy-label">URL</div>
            <div class="copy-content">{{ copyTarget?.formats.url }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'URL' }"
              @click="copyToClipboard(copyTarget?.formats.url, 'URL')"
            >
              {{ copySuccess === 'URL' ? '已复制' : '复制' }}
            </button>
          </div>
          <div class="copy-option">
            <div class="copy-label">Markdown</div>
            <div class="copy-content">{{ copyTarget?.formats.markdown }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'Markdown' }"
              @click="copyToClipboard(copyTarget?.formats.markdown, 'Markdown')"
            >
              {{ copySuccess === 'Markdown' ? '已复制' : '复制' }}
            </button>
          </div>
          <div class="copy-option">
            <div class="copy-label">HTML</div>
            <div class="copy-content">{{ copyTarget?.formats.html }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'HTML' }"
              @click="copyToClipboard(copyTarget?.formats.html, 'HTML')"
            >
              {{ copySuccess === 'HTML' ? '已复制' : '复制' }}
            </button>
          </div>
          <div class="copy-option">
            <div class="copy-label">BBCode</div>
            <div class="copy-content">{{ copyTarget?.formats.bbcode }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'BBCode' }"
              @click="copyToClipboard(copyTarget?.formats.bbcode, 'BBCode')"
            >
              {{ copySuccess === 'BBCode' ? '已复制' : '复制' }}
            </button>
          </div>
        </div>
        <button class="btn-outline close-btn" @click="closeCopyModal">关闭</button>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="deleteConfirm" class="modal-overlay" @click="cancelDelete">
      <div class="modal" @click.stop>
        <h3>确认删除</h3>
        <p>确定要删除这张图片吗？此操作不可恢复。</p>
        <div class="modal-preview">
          <img :src="deleteConfirm.url" :alt="deleteConfirm.id" />
        </div>
        <div class="modal-actions">
          <button class="btn-outline" @click="cancelDelete">取消</button>
          <button class="btn-danger" @click="confirmDelete">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.gallery-page {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  font-size: 20px;
  font-weight: 600;
}

.total-count {
  color: var(--text-secondary);
  font-size: 14px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

.image-card {
  overflow: hidden;
  padding: 0;
}

.image-preview {
  aspect-ratio: 4/3;
  overflow: hidden;
  background-color: var(--bg-color);
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s;
}

.image-card:hover .image-preview img {
  transform: scale(1.05);
}

.image-info {
  padding: 12px 16px;
}

.image-meta {
  display: flex;
  gap: 12px;
  font-size: 13px;
  color: var(--text-secondary);
}

.image-date {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.image-actions {
  display: flex;
  gap: 8px;
  padding: 0 16px 16px;
}

.image-actions button {
  flex: 1;
  font-size: 13px;
  padding: 6px 12px;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 32px;
}

.page-info {
  color: var(--text-secondary);
  font-size: 14px;
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 8px;
  padding: 24px;
  max-width: 400px;
  width: 90%;
}

.modal h3 {
  font-size: 18px;
  margin-bottom: 12px;
}

.modal p {
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.modal-preview {
  margin-bottom: 20px;
  border-radius: 4px;
  overflow: hidden;
}

.modal-preview img {
  width: 100%;
  max-height: 200px;
  object-fit: cover;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

/* 复制弹窗样式 */
.copy-modal {
  max-width: 500px;
}

.copy-preview {
  margin-bottom: 16px;
  border-radius: 8px;
  overflow: hidden;
  max-height: 150px;
}

.copy-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.copy-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.copy-option {
  display: flex;
  align-items: center;
  gap: 12px;
}

.copy-label {
  width: 80px;
  font-weight: 500;
  flex-shrink: 0;
}

.copy-content {
  flex: 1;
  padding: 8px 12px;
  background-color: var(--bg-color);
  border-radius: 4px;
  font-size: 13px;
  font-family: monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 13px;
  flex-shrink: 0;
}

.btn-success {
  background-color: var(--success-color);
}

.close-btn {
  width: 100%;
}
</style>
