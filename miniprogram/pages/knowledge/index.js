// pages/knowledge/index.js
const api = require('../../utils/api')

Page({
  data: {
    knowledgeBases: [],
    selectedBase: null,
    knowledgeItems: [],
    loading: false,
    loadingItems: false
  },

  onLoad() {
    this.loadKnowledgeBases()
  },

  async loadKnowledgeBases() {
    this.setData({ loading: true })
    try {
      const response = await api.getKnowledgeBases()
      this.setData({
        knowledgeBases: response.items || [],
        loading: false
      })
    } catch (error) {
      console.error('加载知识库失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
      this.setData({ loading: false })
    }
  },

  async selectBase(e) {
    const baseId = e.currentTarget.dataset.id
    const base = this.data.knowledgeBases.find(b => b.id === baseId)
    
    this.setData({
      selectedBase: base,
      knowledgeItems: [],
      loadingItems: true
    })

    try {
      const response = await api.getKnowledgeItems(baseId)
      this.setData({
        knowledgeItems: response.items || [],
        loadingItems: false
      })
    } catch (error) {
      console.error('加载知识项失败', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
      this.setData({ loadingItems: false })
    }
  },

  backToList() {
    this.setData({
      selectedBase: null,
      knowledgeItems: []
    })
  }
})

