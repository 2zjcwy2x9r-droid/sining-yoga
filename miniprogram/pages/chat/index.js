// pages/chat/index.js
const api = require('../../utils/api')

Page({
  data: {
    messages: [],
    inputValue: '',
    loading: false
  },

  onLoad() {
    // 初始化消息历史
    this.setData({
      messages: [
        {
          role: 'assistant',
          content: '你好！我是AI助手，可以帮你解答关于瑜伽的问题，也可以帮你预订课程。'
        }
      ]
    })
  },

  onInput(e) {
    this.setData({
      inputValue: e.detail.value
    })
  },

  async sendMessage() {
    const message = this.data.inputValue.trim()
    if (!message) {
      return
    }

    // 添加用户消息
    const userMessage = {
      role: 'user',
      content: message
    }
    this.setData({
      messages: [...this.data.messages, userMessage],
      inputValue: '',
      loading: true
    })

    try {
      // 构建历史消息
      const history = this.data.messages.map(msg => ({
        role: msg.role,
        content: msg.content
      }))

      // 调用AI接口
      const response = await api.chat(message, history)
      
      // 添加AI回复
      const assistantMessage = {
        role: 'assistant',
        content: response.message,
        sources: response.sources || []
      }
      this.setData({
        messages: [...this.data.messages, assistantMessage],
        loading: false
      })
    } catch (error) {
      console.error('发送消息失败', error)
      wx.showToast({
        title: '发送失败',
        icon: 'none'
      })
      this.setData({
        loading: false
      })
    }
  }
})

