// pages/index/index.js
const api = require('../../utils/api')

Page({
  data: {
    userInfo: null,
    weekSchedule: [],
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
      
      // 按周一到周日分组课程
      const weekSchedule = this.groupClassesByWeekDay(response.items || [], monday)
      
      this.setData({
        weekSchedule: weekSchedule,
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

  // 按周一到周日分组，即使某天没有课程也要显示
  groupClassesByWeekDay(classes, mondayDate) {
    const weekSchedule = []
    
    // 初始化周一到周日的结构
    for (let i = 0; i < 7; i++) {
      const currentDate = new Date(mondayDate)
      currentDate.setDate(mondayDate.getDate() + i)
      const dateStr = this.formatDate(currentDate)
      
      const dayData = {
        dayIndex: i,
        dayName: this.data.weekDays[i],
        date: dateStr,
        dateLabel: this.getDateLabel(currentDate),
        classes: []
      }
      
      // 找到该日期的所有课程
      classes.forEach(cls => {
        const classDate = new Date(cls.start_time)
        const classDateStr = this.formatDate(classDate)
        
        if (classDateStr === dateStr) {
          // 格式化课程信息
          const bookedCount = cls.booked_count || 0
          const capacity = cls.capacity || 0
          const formattedClass = {
            ...cls,
            startTime: this.formatTime(cls.start_time),
            endTime: this.formatTime(cls.end_time),
            duration: this.calculateDuration(cls.start_time, cls.end_time),
            timePeriod: this.getTimePeriod(cls.start_time),
            available: bookedCount < capacity
          }
          dayData.classes.push(formattedClass)
        }
      })
      
      // 按开始时间排序（上午到晚上）
      dayData.classes.sort((a, b) => {
        const timeA = new Date(a.start_time).getTime()
        const timeB = new Date(b.start_time).getTime()
        return timeA - timeB
      })
      
      weekSchedule.push(dayData)
    }
    
    return weekSchedule
  },

  // 计算课程时长（分钟）
  calculateDuration(startTime, endTime) {
    if (!startTime || !endTime) return 0
    const start = new Date(startTime).getTime()
    const end = new Date(endTime).getTime()
    return Math.round((end - start) / (1000 * 60))
  },

  // 获取时间段（上午/下午/晚上）
  getTimePeriod(timeStr) {
    if (!timeStr) return ''
    const date = new Date(timeStr)
    const hours = date.getHours()
    if (hours < 12) return '上午'
    if (hours < 18) return '下午'
    return '晚上'
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
    
    const month = date.getMonth() + 1
    const day = date.getDate()
    return `${month}/${day}`
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
  }
})
