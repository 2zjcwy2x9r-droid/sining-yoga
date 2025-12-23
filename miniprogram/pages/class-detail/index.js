// pages/class-detail/index.js
const api = require('../../utils/api')

Page({
  data: {
    classId: '',
    classInfo: null,
    reviews: [],
    userBooking: null,
    userReview: null,
    loading: false,
    showReviewModal: false,
    reviewForm: {
      rating: 5,
      content: '',
      images: []
    }
  },

  onLoad(options) {
    if (options.id) {
      this.setData({ classId: options.id })
      this.loadClassDetail()
      this.loadReviews()
      this.loadUserBooking()
    }
  },

  async loadClassDetail() {
    this.setData({ loading: true })
    try {
      const classInfo = await api.getClass(this.data.classId)
      // 格式化时间显示
      if (classInfo.start_time) {
        classInfo.start_time = this.formatTime(classInfo.start_time)
      }
      if (classInfo.end_time) {
        classInfo.end_time = this.formatTime(classInfo.end_time)
      }
      this.setData({
        classInfo: classInfo,
        loading: false
      })
    } catch (error) {
      console.error('加载课程详情失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
      this.setData({ loading: false })
    }
  },

  async loadReviews() {
    try {
      const response = await api.getClassReviews(this.data.classId)
      this.setData({
        reviews: response.items || []
      })
    } catch (error) {
      console.error('加载评价失败', error)
    }
  },

  async loadUserBooking() {
    try {
      const app = getApp()
      const userID = app.globalData.userInfo?.openId || 'test_user'
      const response = await api.getUserBookings(userID)
      
      // 查找当前课程的预订
      const booking = (response.items || []).find(b => b.class_id === this.data.classId)
      if (booking) {
        this.setData({ userBooking: booking })
        // 加载用户评价
        this.loadUserReview()
      }
    } catch (error) {
      console.error('加载用户预订失败', error)
    }
  },

  async loadUserReview() {
    try {
      const app = getApp()
      const userID = app.globalData.userInfo?.openId || 'test_user'
      const review = await api.getUserReview(this.data.classId, userID)
      if (review) {
        this.setData({ userReview: review })
      }
    } catch (error) {
      // 用户未评价，忽略错误
    }
  },

  formatTime(timeStr) {
    if (!timeStr) return ''
    const date = new Date(timeStr)
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${hours}:${minutes}`
  },

  formatDateTime(timeStr) {
    if (!timeStr) return ''
    const date = new Date(timeStr)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day} ${hours}:${minutes}`
  },

  async bookClass() {
    try {
      const app = getApp()
      const userID = app.globalData.userInfo?.openId || 'test_user'
      const userName = app.globalData.userInfo?.nickName || '用户'
      
      wx.showLoading({ title: '预订中...' })
      await api.bookClass(this.data.classId, userID, userName)
      wx.hideLoading()
      
      wx.showToast({
        title: '预订成功',
        icon: 'success'
      })
      
      this.loadClassDetail()
      this.loadUserBooking()
    } catch (error) {
      wx.hideLoading()
      console.error('预订失败', error)
      wx.showToast({
        title: error.message || '预订失败',
        icon: 'none'
      })
    }
  },

  async cancelBooking() {
    wx.showModal({
      title: '确认取消',
      content: '确定要取消预订吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            wx.showLoading({ title: '取消中...' })
            await api.cancelBooking(this.data.userBooking.id)
            wx.hideLoading()
            
            wx.showToast({
              title: '取消成功',
              icon: 'success'
            })
            
            this.setData({ userBooking: null })
            this.loadClassDetail()
          } catch (error) {
            wx.hideLoading()
            console.error('取消预订失败', error)
            wx.showToast({
              title: '取消失败',
              icon: 'none'
            })
          }
        }
      }
    })
  },

  showReview() {
    if (this.data.userReview) {
      // 已评价，显示评价详情
      wx.showModal({
        title: '我的评价',
        content: `评分：${this.data.userReview.rating}星\n${this.data.userReview.content}`,
        showCancel: false
      })
    } else {
      // 未评价，显示评价表单
      this.setData({ showReviewModal: true })
    }
  },

  closeReviewModal() {
    this.setData({ showReviewModal: false })
  },

  onRatingChange(e) {
    this.setData({
      'reviewForm.rating': e.detail.value
    })
  },

  onReviewContentInput(e) {
    this.setData({
      'reviewForm.content': e.detail.value
    })
  },

  chooseImages() {
    wx.chooseImage({
      count: 3,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: async (res) => {
        // 上传图片
        const images = []
        for (let i = 0; i < res.tempFilePaths.length; i++) {
          try {
            const url = await this.uploadImage(res.tempFilePaths[i])
            images.push(url)
          } catch (error) {
            console.error('上传图片失败', error)
          }
        }
        this.setData({
          'reviewForm.images': images
        })
      }
    })
  },

  uploadImage(filePath) {
    return new Promise((resolve, reject) => {
      const app = getApp()
      wx.uploadFile({
        url: app.globalData.apiBaseUrl + '/upload/image',
        filePath: filePath,
        name: 'file',
        success: (res) => {
          const data = JSON.parse(res.data)
          resolve(data.url)
        },
        fail: reject
      })
    })
  },

  async submitReview() {
    if (!this.data.reviewForm.content.trim()) {
      wx.showToast({
        title: '请输入评价内容',
        icon: 'none'
      })
      return
    }

    try {
      const app = getApp()
      const userID = app.globalData.userInfo?.openId || 'test_user'
      const userName = app.globalData.userInfo?.nickName || '用户'
      
      wx.showLoading({ title: '提交中...' })
      await api.createReview(
        this.data.classId,
        userID,
        userName,
        this.data.reviewForm.rating,
        this.data.reviewForm.content,
        this.data.reviewForm.images
      )
      wx.hideLoading()
      
      wx.showToast({
        title: '评价成功',
        icon: 'success'
      })
      
      this.setData({
        showReviewModal: false,
        reviewForm: {
          rating: 5,
          content: '',
          images: []
        }
      })
      
      this.loadReviews()
      this.loadUserReview()
    } catch (error) {
      wx.hideLoading()
      console.error('提交评价失败', error)
      wx.showToast({
        title: '提交失败',
        icon: 'none'
      })
    }
  }
})

