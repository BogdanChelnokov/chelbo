<script setup lang="ts">
import { ref, computed, watch, nextTick, onUnmounted, onMounted } from 'vue'
import { useAuthStore } from '@/stores/authStore'
import { useChatStore } from '@/stores/chatStore'
import { useWebSocketStore } from '@/stores/websocketStore'
import { chatsApi } from '@/api/chats'
import MessageBubble from './MessageBubble.vue'
import MessageInput from './MessageInput.vue'
import Loader from '@/components/common/Loader.vue'
import type { Chat, Message } from '@/types/chat'

const props = defineProps<{
  chat: Chat | null
}>()

const emit = defineEmits(['search', 'settings'])

const authStore = useAuthStore()
const chatStore = useChatStore()
const wsStore = useWebSocketStore()

// Keyboard handling
const keyboardHeight = ref(0)
const isKeyboardOpen = ref(false)

const handleResize = () => {
  const viewportHeight = window.visualViewport?.height || window.innerHeight
  const windowHeight = window.innerHeight
  const diff = windowHeight - viewportHeight
  
  if (diff > 150) {
    keyboardHeight.value = diff
    isKeyboardOpen.value = true
  } else {
    keyboardHeight.value = 0
    isKeyboardOpen.value = false
  }
}

const messagesContainer = ref<HTMLElement | null>(null)
const isTyping = ref(false)
const typingTimeout = ref<number | null>(null)
const typingUsers = ref<number[]>([])
const replyTo = ref<Message | null>(null)
const forwardMessageData = ref<Message | null>(null)
const showForwardModal = ref(false)

const reversedMessages = computed(() => {
  return [...chatStore.currentMessages].reverse()
})

function getChatTitle(): string {
  if (!props.chat) return ''
  if (props.chat.type === 'group') {
    return props.chat.title || 'Группа'
  }
  return props.chat.title || 'Пользователь'
}

function getChatTitleForChat(chat: Chat): string {
  if (chat.type === 'group') {
    return chat.title || 'Группа'
  }
  return chat.title || 'Пользователь'
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

async function handleSendMessage(text: string) {
  if (!props.chat) return
  
  try {
    await chatStore.sendMessage(
      props.chat.id, 
      text, 
      undefined, 
      undefined, 
      replyTo.value?.id
    )
    
    clearReplyTo()
    
    // Перезагружаем сообщения, чтобы получить reply_to данные
    await chatStore.loadMessages(props.chat.id, true)
    
    scrollToBottom()
  } catch (error) {
    console.error('Failed to send message:', error)
    alert('Не удалось отправить сообщение')
  }
}

async function handleSendFile(file: File) {
  if (!props.chat) return
  
  const formData = new FormData()
  formData.append('file', file)
  
  try {
    const uploadResponse = await fetch(`http://localhost:8080/api/files/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authStore.token}`
      },
      body: formData
    })
    
    const uploadResult = await uploadResponse.json()
    console.log('Upload result:', uploadResult)
    
    if (uploadResult.success && uploadResult.data) {
      // Используем sendMessage вместо sendFileMessage
      await chatStore.sendMessage(
        props.chat.id, 
        '', 
        uploadResult.data.file_url, 
        uploadResult.data.file_type,
        replyTo.value?.id
      )
      clearReplyTo()
      scrollToBottom()
    } else {
      console.error('Upload failed:', uploadResult)
      alert('Не удалось загрузить файл')
    }
  } catch (error) {
    console.error('Failed to upload file:', error)
    alert('Не удалось отправить файл')
  }
}


async function handleSendVoice(blob: Blob) {
  if (!props.chat) return
  
  const file = new File([blob], `voice_${Date.now()}.webm`, { type: 'audio/webm' })
  await handleSendFile(file)
}

async function handleSendVideo(blob: Blob) {
  if (!props.chat) return
  
  const file = new File([blob], `video_${Date.now()}.webm`, { type: 'video/webm' })
  await handleSendFile(file)
}

function handleTyping(isTypingNow: boolean) {
  if (!props.chat) return
  wsStore.sendTyping(props.chat.id, isTypingNow)
}

function setReplyTo(message: Message) {
  replyTo.value = message
  scrollToBottom()
}

function clearReplyTo() {
  replyTo.value = null
}

function forwardMessage(message: Message) {
  forwardMessageData.value = message
  showForwardModal.value = true
}

async function confirmForward(targetChatId: number) {
  if (!forwardMessageData.value) return
  
  try {
    await chatsApi.forwardMessage(forwardMessageData.value.id, targetChatId)
    showForwardModal.value = false
    forwardMessageData.value = null
  } catch (error) {
    console.error('Failed to forward message:', error)
    alert('Не удалось переслать сообщение')
  }
}

async function deleteMessage(messageId: number) {
  if (!confirm('Удалить сообщение?')) return
  
  try {
    await chatsApi.deleteMessage(messageId)
    if (props.chat) {
      await chatStore.loadMessages(props.chat.id, true)
    }
  } catch (error) {
    console.error('Failed to delete message:', error)
    alert('Не удалось удалить сообщение')
  }
}

// Watch for new messages to scroll
watch(() => chatStore.currentMessages.length, () => {
  scrollToBottom()
})

// Watch for chat change to load messages
watch(() => props.chat?.id, async (newId, oldId) => {
  if (newId && newId !== oldId) {
    await chatStore.loadMessages(newId, true)
    await chatStore.markChatAsRead(newId)
    scrollToBottom()
    clearReplyTo()
  }
})

watch(isKeyboardOpen, (open) => {
  if (open) {
    setTimeout(() => {
      scrollToBottom()
    }, 100)
  }
})

onMounted(() => {
  if ('visualViewport' in window) {
    window.visualViewport?.addEventListener('resize', handleResize)
  }
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (typingTimeout.value) {
    clearTimeout(typingTimeout.value)
  }
  if ('visualViewport' in window) {
    window.visualViewport?.removeEventListener('resize', handleResize)
  }
  window.removeEventListener('resize', handleResize)
})
</script>

<template>
  <div class="chat-window" :class="{ 'keyboard-open': isKeyboardOpen }" :style="{ paddingBottom: keyboardHeight + 'px' }">
    <!-- Chat Header -->
    <div class="chat-header">
      <div class="chat-info">
        <div class="chat-avatar">
          <span>{{ getChatTitle().charAt(0).toUpperCase() }}</span>
        </div>
        <div class="chat-details">
          <h3>{{ getChatTitle() }}</h3>
          <span class="chat-status">{{ typingUsers.length > 0 ? 'Печатает...' : 'Онлайн' }}</span>
        </div>
      </div>
      <div class="chat-actions">
        <button class="btn-icon" @click="$emit('search')">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </button>
        <button class="btn-icon" @click="$emit('settings')">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Messages Area -->
    <div ref="messagesContainer" class="messages-area scrollable">
      <div v-if="chatStore.loading" class="loading-messages">
        <Loader />
      </div>
      
      <div v-else-if="chatStore.currentMessages.length === 0" class="empty-messages">
        <p>Нет сообщений</p>
        <p class="hint">Напишите первое сообщение!</p>
      </div>
      
      <div v-else class="messages-list">
        <MessageBubble
          v-for="message in chatStore.currentMessages"
          :key="message.id"
          :data-message-id="message.id"
          :message="message"
          :isOwn="message.sender_id === authStore.user?.id"
          @reply="setReplyTo"
          @forward="forwardMessage"
          @delete="deleteMessage"
        />
      </div>
    </div>

    <!-- Typing Indicator -->
    <div v-if="isTyping" class="typing-indicator">
      <span>Кто-то печатает...</span>
    </div>

    <!-- Message Input -->
    <MessageInput 
      @send="handleSendMessage" 
      @send-file="handleSendFile"
      @send-voice="handleSendVoice"
      @send-video="handleSendVideo"
      @typing="handleTyping"
      @cancel-reply="clearReplyTo"
      :replyTo="replyTo"
    />
    
    <!-- Forward Modal -->
    <div v-if="showForwardModal" class="modal" @click.self="showForwardModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Переслать сообщение</h3>
          <button class="modal-close" @click="showForwardModal = false">×</button>
        </div>
        <div class="chats-list-forward">
          <div
            v-for="chat in chatStore.chats"
            :key="chat.id"
            class="chat-item-forward"
            @click="confirmForward(chat.id)"
          >
            <div class="chat-avatar">
              <span>{{ getChatTitleForChat(chat).charAt(0).toUpperCase() }}</span>
            </div>
            <div class="chat-name">{{ getChatTitleForChat(chat) }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-window {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--bg-secondary);
  transition: padding-bottom 0.3s ease;
}

@media (max-width: 768px) {
  .chat-window {
    height: 100dvh;
  }
  
  .chat-window.keyboard-open {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: auto;
    max-height: 100dvh;
  }
}

.chat-header {
  padding: 16px 20px;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.chat-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.chat-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-dark) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  font-size: 18px;
}

.chat-details h3 {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--text-primary);
}

.chat-status {
  font-size: 12px;
  color: var(--text-muted);
}

.chat-actions {
  display: flex;
  gap: 8px;
}

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.empty-messages {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.empty-messages .hint {
  font-size: 12px;
  margin-top: 8px;
  color: var(--text-muted);
}

.loading-messages {
  display: flex;
  justify-content: center;
  padding: 40px;
}

.typing-indicator {
  padding: 8px 20px;
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
  flex-shrink: 0;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
  padding: 24px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  font-size: 20px;
  margin: 0;
  color: var(--text-primary);
}

.modal-close {
  font-size: 24px;
  cursor: pointer;
  color: var(--text-muted);
  background: none;
  border: none;
}

.chats-list-forward {
  max-height: 400px;
  overflow-y: auto;
}

.chat-item-forward {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  cursor: pointer;
  border-radius: 8px;
  transition: background 0.2s;
}

.chat-item-forward:hover {
  background: var(--bg-hover);
}

.chat-item-forward .chat-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-dark) 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
}

.chat-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

@media (max-width: 768px) {
  .chat-header {
    padding: 12px 16px;
  }
  
  .chat-avatar {
    width: 40px;
    height: 40px;
    font-size: 16px;
  }
  
  .chat-details h3 {
    font-size: 15px;
  }
  
  .messages-area {
    padding: 12px;
  }
}
</style>