// pages/index/index.js
const api = require('../../utils/api')

Page({
  data: {
    userInfo: null,
    classes: [],
    loading: false,
    weekDays: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  },

  onLoad() {
    // 获取用户信息
    const app = getApp()
    this.setData({
      userInfo: app.globalData.userInfo
    })
    this.loadWeekClasses()
  },

  onShow() {
    // 每次显示页面时刷新课程列表
    this.loadWeekClasses()
  },

  async loadWeekClasses() {
    this.setData({ loading: true })
    try {
      // 获取本周课程
      const now = new Date()
      const monday = new Date(now)
      monday.setDate(now.getDate() - now.getDay() + 1) // 本周一
      monday.setHours(0, 0, 0, 0)
      
      const sunday = new Date(monday)
      sunday.setDate(monday.getDate() + 6) // 本周日
      sunday.setHours(23, 59, 59, 999)

      const startDate = this.formatDate(monday)
      const endDate = this.formatDate(sunday)

      const response = await api.getClasses(startDate, endDate)
      
      // 按日期分组课程
      const groupedClasses = this.groupClassesByDate(response.items || [])
      
      this.setData({
        classes: groupedClasses,
        loading: false
      })
    } catch (error) {
      console.error('加载课程失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
      this.setData({ loading: false })
    }
  },

  formatDate(date) {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  },

  groupClassesByDate(classes) {
    const grouped = {}
    classes.forEach(cls => {
      const date = new Date(cls.start_time)
      const dateStr = this.formatDate(date)
      if (!grouped[dateStr]) {
        grouped[dateStr] = {
          date: dateStr,
          dateLabel: this.getDateLabel(date),
          classes: []
        }
      }
      // 格式化时间显示
      if (cls.start_time) {
        cls.start_time = this.formatTime(cls.start_time)
      }
      if (cls.end_time) {
        cls.end_time = this.formatTime(cls.end_time)
      }
      grouped[dateStr].classes.push(cls)
    })
    
    // 转换为数组并按日期排序
    return Object.values(grouped).sort((a, b) => a.date.localeCompare(b.date))
  },

  getDateLabel(date) {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const targetDate = new Date(date)
    targetDate.setHours(0, 0, 0, 0)
    
    const diff = Math.floor((targetDate - today) / (1000 * 60 * 60 * 24))
    
    if (diff === 0) return '今天'
    if (diff === 1) return '明天'
    if (diff === 2) return '后天'
    
    const weekDay = this.data.weekDays[date.getDay() === 0 ? 6 : date.getDay() - 1]
    const month = date.getMonth() + 1
    const day = date.getDate()
    return `${month}月${day}日 ${weekDay}`
  },

  formatTime(timeStr) {
    if (!timeStr) return ''
    const date = new Date(timeStr)
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${hours}:${minutes}`
  },

  navigateToClassDetail(e) {
    const classId = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/class-detail/index?id=${classId}`
    })
  },

  navigateToChat() {
    wx.switchTab({
      url: '/pages/chat/index'
    })
  }
})
