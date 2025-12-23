// utils/api.js
const app = getApp()

/**
 * 发起网络请求
 */
function request(url, method, data) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: app.globalData.apiBaseUrl + url,
      method: method,
      data: data,
      header: {
        'content-type': 'application/json'
      },
      success: (res) => {
        if (res.statusCode === 200 || res.statusCode === 201) {
          resolve(res.data)
        } else {
          reject(new Error(`请求失败: ${res.statusCode}`))
        }
      },
      fail: (err) => {
        reject(err)
      }
    })
  })
}

/**
 * AI聊天
 */
function chat(message, history = [], baseId = '') {
  return request('/ai/chat', 'POST', {
    message: message,
    history: history,
    base_id: baseId
  })
}

/**
 * 获取知识库列表
 */
function getKnowledgeBases(limit = 20, offset = 0) {
  return request(`/knowledge-bases?limit=${limit}&offset=${offset}`, 'GET')
}

/**
 * 获取知识项列表
 */
function getKnowledgeItems(baseId, limit = 20, offset = 0) {
  return request(`/knowledge-bases/${baseId}/items?limit=${limit}&offset=${offset}`, 'GET')
}

/**
 * 获取课程列表
 */
function getClasses(startDate = '', endDate = '') {
  let url = '/classes?'
  if (startDate) url += `start_date=${startDate}&`
  if (endDate) url += `end_date=${endDate}`
  return request(url, 'GET')
}

/**
 * 获取课程详情
 */
function getClass(classId) {
  return request(`/classes/${classId}`, 'GET')
}

/**
 * 预订课程
 */
function bookClass(classId, userId, userName) {
  return request(`/classes/${classId}/book`, 'POST', {
    user_id: userId,
    user_name: userName
  })
}

/**
 * 取消预订
 */
function cancelBooking(bookingId) {
  return request(`/classes/bookings/${bookingId}`, 'DELETE')
}

/**
 * 获取用户预订列表
 */
function getUserBookings(userId, limit = 50, offset = 0) {
  return request(`/classes/bookings?user_id=${userId}&limit=${limit}&offset=${offset}`, 'GET')
}

/**
 * 创建评价
 */
function createReview(classId, userId, userName, rating, content, images = []) {
  return request(`/classes/${classId}/reviews`, 'POST', {
    user_id: userId,
    user_name: userName,
    rating: rating,
    content: content,
    images: images
  })
}

/**
 * 获取课程评价列表
 */
function getClassReviews(classId, limit = 20, offset = 0) {
  return request(`/classes/${classId}/reviews?limit=${limit}&offset=${offset}`, 'GET')
}

/**
 * 获取用户评价（从评价列表中筛选）
 */
async function getUserReview(classId, userId) {
  try {
    const response = await getClassReviews(classId, 100, 0)
    const reviews = response.items || []
    const userReview = reviews.find(r => r.user_id === userId)
    return userReview || null
  } catch (error) {
    return null
  }
}

module.exports = {
  request,
  chat,
  getKnowledgeBases,
  getKnowledgeItems,
  getClasses,
  getClass,
  bookClass,
  cancelBooking,
  getUserBookings,
  createReview,
  getClassReviews,
  getUserReview
}
