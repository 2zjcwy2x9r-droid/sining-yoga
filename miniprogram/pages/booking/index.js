// pages/booking/index.js
const api = require('../../utils/api')

Page({
  data: {
    weekSchedule: [],
    selectedDayIndex: 0,
    selectedDate: '',
    classes: [],
    loading: false,
    weekDays: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  },

  onLoad() {
    this.initWeekSchedule()
    this.loadWeekClasses()
  },

  onShow() {
    // 每次显示页面时刷新课程列表
    this.loadWeekClasses()
  },

  // 初始化周一到周日的日期
  initWeekSchedule() {
    const now = new Date()
    const monday = new Date(now)
    monday.setDate(now.getDate() - now.getDay() + 1) // 本周一
    monday.setHours(0, 0, 0, 0)
    
    const weekSchedule = []
    for (let i = 0; i < 7; i++) {
      const currentDate = new Date(monday)
      currentDate.setDate(monday.getDate() + i)
      const dateStr = this.formatDate(currentDate)
      
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      const targetDate = new Date(currentDate)
      targetDate.setHours(0, 0, 0, 0)
      const diff = Math.floor((targetDate - today) / (1000 * 60 * 60 * 24))
      
      const month = currentDate.getMonth() + 1
      const day = currentDate.getDate()
      
      let dateLabel = ''
      if (diff === 0) dateLabel = '今日'
      else if (diff === 1) dateLabel = '明日'
      else {
        dateLabel = this.data.weekDays[i]
      }
      
      weekSchedule.push({
        dayIndex: i,
        dayName: this.data.weekDays[i],
        date: dateStr,
        dateLabel: dateLabel,
        fullDate: `${month}月${day}日`
      })
    }
    
    // 设置默认选中今天
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const todayIndex = weekSchedule.findIndex(day => {
      const dayDate = new Date(day.date)
      dayDate.setHours(0, 0, 0, 0)
      return dayDate.getTime() === today.getTime()
    })
    
    this.setData({
      weekSchedule: weekSchedule,
      selectedDayIndex: todayIndex >= 0 ? todayIndex : 0,
      selectedDate: weekSchedule[todayIndex >= 0 ? todayIndex : 0].date
    })
  },

  // 选择某一天
  selectDay(e) {
    const dayIndex = e.currentTarget.dataset.index
    const selectedDay = this.data.weekSchedule[dayIndex]
    
    this.setData({
      selectedDayIndex: dayIndex,
      selectedDate: selectedDay.date
    })
    
    this.loadClasses()
  },

  // 加载本周所有课程
  async loadWeekClasses() {
    this.setData({ loading: true })
    try {
      const now = new Date()
      const monday = new Date(now)
      monday.setDate(now.getDate() - now.getDay() + 1)
      monday.setHours(0, 0, 0, 0)
      
      const sunday = new Date(monday)
      sunday.setDate(monday.getDate() + 6)
      sunday.setHours(23, 59, 59, 999)

      const startDate = this.formatDate(monday)
      const endDate = this.formatDate(sunday)

      const response = await api.getClasses(startDate, endDate)
      
      // 将课程按日期分组
      const classesByDate = this.groupClassesByDate(response.items || [])
      
      // 更新周计划中的课程数量
      const weekSchedule = this.data.weekSchedule.map(day => {
        const dayClasses = classesByDate[day.date] || []
        return {
          ...day,
          classCount: dayClasses.length
        }
      })
      
      this.setData({
        weekSchedule: weekSchedule,
        loading: false
      })
      
      // 加载选中日期的课程
      this.loadClasses()
    } catch (error) {
      console.error('加载课程失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
      this.setData({ loading: false })
    }
  },

  // 加载选中日期的课程
  async loadClasses() {
    if (!this.data.selectedDate) return
    
    this.setData({ loading: true })
    try {
      const response = await api.getClasses(this.data.selectedDate, this.data.selectedDate)
      
      const classes = (response.items || []).map(cls => {
        const bookedCount = cls.booked_count || 0
        const capacity = cls.capacity || 0
        return {
          ...cls,
          startTime: this.formatTime(cls.start_time),
          endTime: this.formatTime(cls.end_time),
          duration: this.calculateDuration(cls.start_time, cls.end_time),
          available: bookedCount < capacity,
          bookedCount: bookedCount,
          capacity: capacity
        }
      })
      
      // 按开始时间排序
      classes.sort((a, b) => {
        const timeA = new Date(a.start_time).getTime()
        const timeB = new Date(b.start_time).getTime()
        return timeA - timeB
      })
      
      this.setData({
        classes: classes,
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

  // 按日期分组课程
  groupClassesByDate(classes) {
    const grouped = {}
    classes.forEach(cls => {
      const date = new Date(cls.start_time)
      const dateStr = this.formatDate(date)
      if (!grouped[dateStr]) {
        grouped[dateStr] = []
      }
      grouped[dateStr].push(cls)
    })
    return grouped
  },

  formatDate(date) {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  },

  formatTime(timeStr) {
    if (!timeStr) return ''
    const date = new Date(timeStr)
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${hours}:${minutes}`
  },

  calculateDuration(startTime, endTime) {
    if (!startTime || !endTime) return 0
    const start = new Date(startTime).getTime()
    const end = new Date(endTime).getTime()
    return Math.round((end - start) / (1000 * 60))
  },

  // 预订课程
  async bookClass(e) {
    const classId = e.currentTarget.dataset.id
    const app = getApp()
    const userInfo = app.globalData.userInfo
    
    if (!userInfo) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      return
    }
    
    try {
      await api.bookClass(classId, userInfo.id || 'user1', userInfo.nickName || '用户')
      wx.showToast({
        title: '预订成功',
        icon: 'success'
      })
      this.loadClasses()
      this.loadWeekClasses() // 刷新周计划
    } catch (error) {
      console.error('预订失败', error)
      wx.showToast({
        title: error.message || '预订失败',
        icon: 'none'
      })
    }
  },

  // 阻止事件冒泡
  stopPropagation() {
    // 空函数，阻止事件冒泡
  },

  // 跳转到课程详情
  navigateToClassDetail(e) {
    const classId = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/class-detail/index?id=${classId}`
    })
  }
})

