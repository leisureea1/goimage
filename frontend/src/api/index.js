/**
 * API 封装层
 * 统一管理所有后端 API 调用
 * 便于未来迁移到桌面端或移动端
 */

// API 基础配置
const API_BASE = '/api/v1'

// 从 localStorage 获取 Token (如果有)
const getToken = () => localStorage.getItem('api_token') || ''

/**
 * 通用请求方法
 * @param {string} url - 请求路径
 * @param {object} options - fetch 选项
 * @returns {Promise<object>} - API 响应
 */
async function request(url, options = {}) {
  const token = getToken()
  
  const headers = {
    ...options.headers
  }
  
  // 如果有 Token，添加到请求头
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const response = await fetch(`${API_BASE}${url}`, {
    ...options,
    headers
  })

  const data = await response.json()

  // 统一处理错误
  if (data.code !== 0) {
    throw new Error(data.message || '请求失败')
  }

  return data.data
}

/**
 * 上传图片
 * @param {File} file - 图片文件
 * @param {function} onProgress - 进度回调
 * @returns {Promise<object>} - 上传结果
 */
export async function uploadImage(file, onProgress) {
  const formData = new FormData()
  formData.append('file', file)

  // 使用 XMLHttpRequest 以支持进度回调
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    
    xhr.upload.addEventListener('progress', (e) => {
      if (e.lengthComputable && onProgress) {
        const percent = Math.round((e.loaded / e.total) * 100)
        onProgress(percent)
      }
    })

    xhr.addEventListener('load', () => {
      try {
        const data = JSON.parse(xhr.responseText)
        if (data.code === 0) {
          resolve(data.data)
        } else {
          reject(new Error(data.message || '上传失败'))
        }
      } catch (e) {
        reject(new Error('解析响应失败'))
      }
    })

    xhr.addEventListener('error', () => {
      reject(new Error('网络错误'))
    })

    xhr.open('POST', `${API_BASE}/upload`)
    
    const token = getToken()
    if (token) {
      xhr.setRequestHeader('Authorization', `Bearer ${token}`)
    }
    
    xhr.send(formData)
  })
}

/**
 * 获取图片列表
 * @param {number} page - 页码
 * @param {number} pageSize - 每页数量
 * @returns {Promise<object>} - 分页列表
 */
export async function getImages(page = 1, pageSize = 20) {
  return request(`/images?page=${page}&page_size=${pageSize}`)
}

/**
 * 获取单张图片信息
 * @param {string} id - 图片 ID
 * @returns {Promise<object>} - 图片信息
 */
export async function getImage(id) {
  return request(`/image/${id}`)
}

/**
 * 删除图片
 * @param {string} id - 图片 ID
 * @returns {Promise<void>}
 */
export async function deleteImage(id) {
  return request(`/image/${id}`, {
    method: 'DELETE'
  })
}

/**
 * 设置 API Token
 * @param {string} token - API Token
 */
export function setToken(token) {
  if (token) {
    localStorage.setItem('api_token', token)
  } else {
    localStorage.removeItem('api_token')
  }
}
