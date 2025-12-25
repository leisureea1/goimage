<script setup>
/**
 * 图片上传页面
 * 支持拖拽上传和文件选择
 * 显示上传进度和结果
 */
import { ref, reactive } from 'vue'
import { uploadImage } from '../api'

// 上传状态
const isDragging = ref(false)
const uploading = ref(false)
const uploadResults = reactive([])
const error = ref('')

// 复制弹窗
const showCopyModal = ref(false)
const copyTarget = ref(null)
const copySuccess = ref('')

// 文件输入引用
const fileInput = ref(null)

// 允许的文件类型
const acceptTypes = ['image/jpeg', 'image/png', 'image/webp']

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
 * 处理文件选择
 */
function handleFileSelect(e) {
  const files = e.target.files
  if (files.length > 0) {
    processFiles(Array.from(files))
  }
  e.target.value = ''
}

/**
 * 处理拖拽进入
 */
function handleDragEnter(e) {
  e.preventDefault()
  isDragging.value = true
}

/**
 * 处理拖拽离开
 */
function handleDragLeave(e) {
  e.preventDefault()
  isDragging.value = false
}

/**
 * 处理拖拽放下
 */
function handleDrop(e) {
  e.preventDefault()
  isDragging.value = false
  
  const files = Array.from(e.dataTransfer.files).filter(
    file => acceptTypes.includes(file.type)
  )
  
  if (files.length > 0) {
    processFiles(files)
  } else {
    error.value = '请上传 JPG、PNG 或 WebP 格式的图片'
  }
}

/**
 * 处理文件上传
 */
async function processFiles(files) {
  error.value = ''
  uploading.value = true

  for (const file of files) {
    // 创建响应式上传任务
    const task = reactive({
      id: Date.now() + Math.random(),
      name: file.name,
      size: file.size,
      progress: 0,
      status: 'uploading',
      result: null,
      error: null
    })
    uploadResults.unshift(task)

    try {
      const result = await uploadImage(file, (progress) => {
        task.progress = progress
      })
      task.status = 'success'
      task.result = result
    } catch (e) {
      task.status = 'error'
      task.error = e.message
    }
  }

  uploading.value = false
}

/**
 * 显示复制选项弹窗
 */
function showCopyOptions(url) {
  copyTarget.value = getLinkFormats(url)
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
 * 关闭弹窗
 */
function closeCopyModal() {
  showCopyModal.value = false
  copyTarget.value = null
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
 * 清除上传记录
 */
function clearResults() {
  uploadResults.splice(0, uploadResults.length)
}
</script>

<template>
  <div class="upload-page">
    <!-- 上传区域 -->
    <div 
      class="upload-zone card"
      :class="{ dragging: isDragging }"
      @dragenter="handleDragEnter"
      @dragover.prevent
      @dragleave="handleDragLeave"
      @drop="handleDrop"
      @click="fileInput?.click()"
    >
      <input 
        ref="fileInput"
        type="file" 
        accept="image/jpeg,image/png,image/webp"
        multiple
        @change="handleFileSelect"
        style="display: none"
      />
      
      <div class="upload-icon">📤</div>
      <p class="upload-text">
        拖拽图片到此处，或点击选择文件
      </p>
      <p class="upload-hint">
        支持 JPG、PNG、WebP 格式，最大 10MB
      </p>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="message message-error">
      {{ error }}
    </div>

    <!-- 上传结果列表 -->
    <div v-if="uploadResults.length > 0" class="results">
      <div class="results-header">
        <h3>上传记录</h3>
        <button class="btn-outline" @click="clearResults">清除</button>
      </div>

      <div class="result-list">
        <div 
          v-for="item in uploadResults" 
          :key="item.id"
          class="result-item card"
        >
          <!-- 上传中 -->
          <template v-if="item.status === 'uploading'">
            <div class="result-info">
              <span class="result-name">{{ item.name }}</span>
              <span class="result-size">{{ formatSize(item.size) }}</span>
            </div>
            <div class="progress-bar">
              <div 
                class="progress-fill" 
                :style="{ width: item.progress + '%' }"
              ></div>
            </div>
            <span class="progress-text">{{ item.progress }}%</span>
          </template>

          <!-- 上传成功 -->
          <template v-else-if="item.status === 'success'">
            <img 
              :src="item.result.url" 
              :alt="item.name"
              class="result-thumb"
            />
            <div class="result-info">
              <span class="result-name">{{ item.name }}</span>
              <div class="result-meta">
                <span>{{ item.result.original_format.toUpperCase() }} → WebP</span>
                <span>{{ formatSize(item.result.original_size) }} → {{ formatSize(item.result.processed_size) }}</span>
                <span>{{ item.result.width }} × {{ item.result.height }}</span>
              </div>
            </div>
            <div class="result-actions">
              <button 
                class="btn-primary"
                @click.stop="showCopyOptions(item.result.url)"
              >
                复制链接
              </button>
            </div>
          </template>

          <!-- 上传失败 -->
          <template v-else-if="item.status === 'error'">
            <div class="result-info">
              <span class="result-name error">{{ item.name }}</span>
              <span class="result-error">{{ item.error }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- 复制选项弹窗 -->
    <div v-if="showCopyModal" class="modal-overlay" @click="closeCopyModal">
      <div class="modal copy-modal" @click.stop>
        <h3>复制链接</h3>
        <div class="copy-options">
          <div class="copy-option">
            <div class="copy-label">URL</div>
            <div class="copy-content">{{ copyTarget?.url }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'URL' }"
              @click="copyToClipboard(copyTarget?.url, 'URL')"
            >
              {{ copySuccess === 'URL' ? '已复制' : '复制' }}
            </button>
          </div>
          <div class="copy-option">
            <div class="copy-label">Markdown</div>
            <div class="copy-content">{{ copyTarget?.markdown }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'Markdown' }"
              @click="copyToClipboard(copyTarget?.markdown, 'Markdown')"
            >
              {{ copySuccess === 'Markdown' ? '已复制' : '复制' }}
            </button>
          </div>
          <div class="copy-option">
            <div class="copy-label">HTML</div>
            <div class="copy-content">{{ copyTarget?.html }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'HTML' }"
              @click="copyToClipboard(copyTarget?.html, 'HTML')"
            >
              {{ copySuccess === 'HTML' ? '已复制' : '复制' }}
            </button>
          </div>
          <div class="copy-option">
            <div class="copy-label">BBCode</div>
            <div class="copy-content">{{ copyTarget?.bbcode }}</div>
            <button 
              class="btn-primary btn-sm"
              :class="{ 'btn-success': copySuccess === 'BBCode' }"
              @click="copyToClipboard(copyTarget?.bbcode, 'BBCode')"
            >
              {{ copySuccess === 'BBCode' ? '已复制' : '复制' }}
            </button>
          </div>
        </div>
        <button class="btn-outline close-btn" @click="closeCopyModal">关闭</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.upload-page {
  max-width: 800px;
  margin: 0 auto;
}

.upload-zone {
  border: 2px dashed var(--border-color);
  border-radius: 12px;
  padding: 60px 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-zone:hover,
.upload-zone.dragging {
  border-color: var(--primary-color);
  background-color: rgba(74, 144, 217, 0.04);
}

.upload-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.upload-text {
  font-size: 16px;
  color: var(--text-color);
  margin-bottom: 8px;
}

.upload-hint {
  font-size: 14px;
  color: var(--text-secondary);
}

.results {
  margin-top: 24px;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.results-header h3 {
  font-size: 16px;
  font-weight: 500;
}

.result-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
}

.result-thumb {
  width: 60px;
  height: 60px;
  object-fit: cover;
  border-radius: 4px;
  flex-shrink: 0;
}

.result-info {
  flex: 1;
  min-width: 0;
}

.result-name {
  display: block;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.result-name.error {
  color: var(--danger-color);
}

.result-size {
  font-size: 13px;
  color: var(--text-secondary);
}

.result-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.result-error {
  font-size: 13px;
  color: var(--danger-color);
}

.result-actions {
  flex-shrink: 0;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background-color: var(--border-color);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: var(--primary-color);
  transition: width 0.2s;
}

.progress-text {
  font-size: 13px;
  color: var(--text-secondary);
  width: 40px;
  text-align: right;
}

/* 复制弹窗样式 */
.copy-modal {
  max-width: 500px;
  width: 90%;
}

.copy-modal h3 {
  margin-bottom: 20px;
}

.copy-options {
  display: flex;
  flex-direction: column;
  gap: 16px;
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
