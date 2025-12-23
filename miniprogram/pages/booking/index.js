// pages/booking/index.js
Page({
  data: {
    classes: [],
    selectedDate: '',
    loading: false
  },

  onLoad() {
    // 设置默认日期为今天
    const today = new Date()
    const dateStr = today.toISOString().split('T')[0]
    this.setData({
      selectedDate: dateStr
    })
    this.loadClasses()
  },

  onDateChange(e) {
    this.setData({
      selectedDate: e.detail.value
    })
    this.loadClasses()
  },

  async loadClasses() {
    this.setData({ loading: true })
    try {
      // TODO: 调用API获取课程列表
      // const response = await api.getClasses(this.data.selectedDate)
      // this.setData({ classes: response.items })
      
      // 模拟数据
      this.setData({
        classes: [
          {
            id: '1',
            name: '晨间瑜伽',
            instructor: '张老师',
            startTime: '08:00',
            endTime: '09:00',
            capacity: 20,
            bookedCount: 15,
            available: true
          }
        ],
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

  async bookClass(e) {
    const classId = e.currentTarget.dataset.id
    try {
      // TODO: 调用API预订课程
      // await api.bookClass(classId)
      wx.showToast({
        title: '预订成功',
        icon: 'success'
      })
      this.loadClasses()
    } catch (error) {
      console.error('预订失败', error)
      wx.showToast({
        title: '预订失败',
        icon: 'none'
      })
    }
  }
})

