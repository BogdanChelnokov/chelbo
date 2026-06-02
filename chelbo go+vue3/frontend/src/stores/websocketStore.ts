import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './authStore'
import { useChatStore } from './chatStore'
import type { Message } from '@/types/chat'

export const useWebSocketStore = defineStore('websocket', () => {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 10
  const reconnectDelay = 1000

  function connect() {
    const authStore = useAuthStore()
    const token = authStore.token
    
    if (!token) {
      console.error('No token for WebSocket connection')
      return
    }

    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/ws?token=${token}`
    console.log('Connecting to WebSocket:', wsUrl)
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      console.log('WebSocket connected')
      connected.value = true
      reconnectAttempts.value = 0
    }

    ws.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        console.log('WebSocket message received:', data)
        handleMessage(data)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    ws.value.onclose = () => {
      console.log('WebSocket disconnected')
      connected.value = false
      reconnect()
    }

    ws.value.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  function reconnect() {
    if (reconnectAttempts.value >= maxReconnectAttempts) {
      console.error('Max reconnect attempts reached')
      return
    }

    setTimeout(() => {
      reconnectAttempts.value++
      console.log(`Reconnecting... Attempt ${reconnectAttempts.value}`)
      connect()
    }, reconnectDelay * Math.pow(2, reconnectAttempts.value - 1))
  }

  function disconnect() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
      connected.value = false
    }
  }

  function send(data: any) {
    if (ws.value && connected.value) {
      ws.value.send(JSON.stringify(data))
    }
  }

  function handleMessage(data: any) {
    const chatStore = useChatStore()
    const authStore = useAuthStore()
    
    switch (data.type) {
      case 'new_message':
        console.log('New message received:', data.message)
        const message = data.message as Message
        
        // Добавляем сообщение в store
        chatStore.addMessage(message.chat_id, message)
        
        // Если мы не в этом чате, увеличиваем счетчик непрочитанных
        if (chatStore.currentChat?.id !== message.chat_id && message.sender_id !== authStore.user?.id) {
          const chat = chatStore.chats.find(c => c.id === message.chat_id)
          if (chat) {
            chat.unread_count = (chat.unread_count || 0) + 1
          }
        }
        break
      
      case 'message_delivered':
        console.log('Message delivered:', data.message_id)
        chatStore.updateMessageStatus(data.message_id, data.chat_id, true, false)
        break
      
      case 'message_read':
        console.log('Message read:', data.message_id)
        chatStore.updateMessageStatus(data.message_id, data.chat_id, true, true)
        break
      
      case 'typing':
        console.log(`User ${data.user_id} is ${data.is_typing ? 'typing' : 'stopped typing'} in chat ${data.chat_id}`)
        break
      
      case 'user_online':
        console.log(`User ${data.user_id} is online`)
        break
      
      default:
        console.log('Unknown message type:', data.type)
    }
  }

  function sendTyping(chatId: number, isTyping: boolean) {
    send({
      type: 'typing',
      chat_id: chatId,
      is_typing: isTyping
    })
  }

  function markAsRead(chatId: number, messageId: number) {
    send({
      type: 'read',
      chat_id: chatId,
      message_id: messageId
    })
  }

  return {
    ws,
    connected,
    connect,
    disconnect,
    send,
    sendTyping,
    markAsRead
  }
})