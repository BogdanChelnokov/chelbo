import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { chatsApi } from '@/api/chats'
import type { Chat, Message } from '@/types/chat'

export const useChatStore = defineStore('chat', () => {
  const chats = ref<Chat[]>([])
  const currentChat = ref<Chat | null>(null)
  const messages = ref<Map<number, Message[]>>(new Map())
  const loading = ref(false)
  const sendingMessage = ref(false)

  const currentMessages = computed(() => {
    if (!currentChat.value) return []
    const msgs = messages.value.get(currentChat.value.id) || []
    return [...msgs].sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
  })

  async function loadChats() {
    loading.value = true
    try {
      const response = await chatsApi.getChats()
      chats.value = response.data.data || []
    } catch (error) {
      console.error('Failed to load chats:', error)
    } finally {
      loading.value = false
    }
  }

  async function loadMessages(chatId: number, reset = false) {
    if (reset) {
      messages.value.set(chatId, [])
    }
    
    try {
      const response = await chatsApi.getMessages(chatId, 100, 0)
      const newMessages = response.data.data || []
      const sortedMessages = [...newMessages].sort((a, b) => 
        new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      )
      messages.value.set(chatId, sortedMessages)
      
      if (currentChat.value?.id === chatId) {
        await markChatAsRead(chatId)
      }
      
      return sortedMessages
    } catch (error) {
      console.error('Failed to load messages:', error)
      return []
    }
  }

  async function markChatAsRead(chatId: number) {
    try {
      const currentMsgs = messages.value.get(chatId) || []
      const unreadMessages = currentMsgs.filter(m => 
        m.sender_id !== currentChat.value?.id && !m.is_read
      )
      
      for (const msg of unreadMessages) {
        await chatsApi.markAsRead(msg.id)
        msg.is_read = true
      }
      
      updateUnreadCount(chatId, 0)
    } catch (error) {
      console.error('Failed to mark chat as read:', error)
    }
  }

  // Единый метод для отправки сообщений (текст, файл, ответ)
  async function sendMessage(chatId: number, text: string, fileUrl?: string, fileType?: string, replyToId?: number) {
    sendingMessage.value = true
    try {
      const data: any = {}
      if (text && text.trim()) {
        data.text = text
      }
      if (fileUrl) {
        data.file_url = fileUrl
        data.file_type = fileType
      }
      if (replyToId) {
        data.reply_to_id = replyToId
      }
      
      const response = await chatsApi.sendMessage(chatId, data)
      const newMessage = response.data.data!
      
      const currentMsgs = messages.value.get(chatId) || []
      const updatedMsgs = [...currentMsgs, newMessage]
      messages.value.set(chatId, updatedMsgs)
      
      const chat = chats.value.find(c => c.id === chatId)
      if (chat) {
        chat.last_message = newMessage
        chat.updated_at = newMessage.created_at
      }
      
      return newMessage
    } catch (error) {
      console.error('Failed to send message:', error)
      throw error
    } finally {
      sendingMessage.value = false
    }
  }

  function addMessage(chatId: number, message: Message) {
    const currentMsgs = messages.value.get(chatId) || []
    const updatedMsgs = [...currentMsgs, message]
    messages.value.set(chatId, updatedMsgs)
    
    const chat = chats.value.find(c => c.id === chatId)
    if (chat) {
      chat.last_message = message
      chat.updated_at = message.created_at
      
      if (currentChat.value?.id !== chatId && message.sender_id !== currentChat.value?.id) {
        chat.unread_count = (chat.unread_count || 0) + 1
      }
    }
    
    const chatIndex = chats.value.findIndex(c => c.id === chatId)
    if (chatIndex !== -1) {
      const updatedChat = chats.value[chatIndex]
      chats.value.splice(chatIndex, 1)
      chats.value.unshift(updatedChat)
    }
  }

  function updateMessageStatus(messageId: number, chatId: number, isDelivered: boolean, isRead: boolean) {
    const msgs = messages.value.get(chatId) || []
    const msg = msgs.find(m => m.id === messageId)
    if (msg) {
      msg.is_delivered = isDelivered
      msg.is_read = isRead
    }
  }

  function setCurrentChat(chat: Chat | null) {
    currentChat.value = chat
    if (chat) {
      loadMessages(chat.id, true).then(() => {
        markChatAsRead(chat.id)
      })
    }
  }

  function addChat(chat: Chat) {
    if (!chats.value.find(c => c.id === chat.id)) {
      chats.value.unshift(chat)
    }
  }

  function updateUnreadCount(chatId: number, count: number) {
    const chat = chats.value.find(c => c.id === chatId)
    if (chat) {
      chat.unread_count = count
    }
  }

  return {
    chats,
    currentChat,
    currentMessages,
    loading,
    sendingMessage,
    loadChats,
    loadMessages,
    sendMessage,
    addMessage,
    updateMessageStatus,
    setCurrentChat,
    addChat,
    updateUnreadCount,
    markChatAsRead
  }
})