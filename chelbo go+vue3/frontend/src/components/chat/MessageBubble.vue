<template>
  <div 
    class="message-wrapper"
    :class="{ 
      'message-own': isOwn, 
      'message-other': !isOwn 
    }"
    :data-message-id="message.id"
  >

    <!-- Индикатор пересланного сообщения -->
    <div v-if="message.forwarded_from" class="forward-indicator">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polyline points="13 3 22 12 13 21" />
        <polyline points="2 3 11 12 2 21" />
      </svg>
      <span>Переслано от {{ getForwardedName(message.forwarded_from) }}</span>
    </div>

    <!-- Блок с ответом -->
    <div v-if="message.reply_to" class="reply-quote" @click="scrollToReply(message.reply_to.id)">
      <div class="quote-line"></div>
      <div class="quote-content">
        <div class="quote-header">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 15 9 9 15 15" />
            <line x1="21" y1="9" x2="21" y2="21" />
          </svg>
          <span class="quote-name">{{ getReplySenderName(message.reply_to.sender_id) }}</span>
        </div>
        <div class="quote-text">
          {{ getMessagePreview(message.reply_to) }}
        </div>
      </div>
    </div>

    <!-- Основное сообщение -->
    <div class="message-bubble" @contextmenu.prevent="openContextMenu" @touchstart="handleTouchStart" @touchend="handleTouchEnd">
      <div class="message-content">
        <!-- Аудио сообщение -->
        <div v-if="message.file_type === 'audio' || message.file_type === 'voice'" class="audio-message">
          <div class="audio-player">
            <button class="play-btn" @click="togglePlay">
              <svg v-if="!isPlaying" width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5 3 19 12 5 21 5 3" />
              </svg>
              <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                <rect x="6" y="4" width="4" height="16" />
                <rect x="14" y="4" width="4" height="16" />
              </svg>
            </button>
            
            <div class="progress-container" @click="seekAudio">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
              </div>
              <div class="time-info">
                <span class="current-time">{{ formatDuration(currentTime) }}</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Видео сообщение -->
        <div v-else-if="message.file_type === 'video'" class="video-message" @click="openFile(message.file_url)">
          <video v-if="message.file_url" :src="message.file_url" class="video-thumbnail"></video>
          <div class="video-overlay">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="white">
              <polygon points="5 3 19 12 5 21 5 3" />
            </svg>
          </div>
        </div>
        
        <!-- Изображение -->
        <div v-else-if="message.file_type === 'image'" class="image-message" @click="openFile(message.file_url)">
          <img v-if="message.file_url" :src="message.file_url" alt="Image" />
        </div>
        
        <!-- Документ -->
        <div v-else-if="message.file_url" class="document-message" @click="openFile(message.file_url)">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" />
            <polyline points="13 2 13 9 20 9" />
          </svg>
          <span>Документ</span>
        </div>
        
        <!-- Текст -->
        <div v-if="message.text" class="message-text">{{ message.text }}</div>
        
        <div class="message-meta">
          <span class="message-time">{{ formatTime(message.created_at) }}</span>
          <span v-if="isOwn" class="message-status">
            <span v-if="message.is_read">✓✓</span>
            <span v-else-if="message.is_delivered">✓✓</span>
            <span v-else>✓</span>
          </span>
        </div>
      </div>
    </div>

    <!-- Context Menu -->
    <div v-if="contextMenuVisible" class="context-menu" :style="contextMenuStyle" @click.stop>
      <button class="menu-item" @click="handleReply">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 15 9 9 15 15" />
          <line x1="21" y1="9" x2="21" y2="21" />
        </svg>
        Ответить
      </button>
      <button class="menu-item" @click="handleForward">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="13 3 22 12 13 21" />
          <polyline points="2 3 11 12 2 21" />
        </svg>
        Переслать
      </button>
      <button v-if="isOwn" class="menu-item danger" @click="handleDelete">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6" />
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
          <line x1="10" y1="11" x2="10" y2="17" />
          <line x1="14" y1="11" x2="14" y2="17" />
        </svg>
        Удалить
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Message } from '@/types/chat'
import { useAuthStore } from '@/stores/authStore'
import { useChatStore } from '@/stores/chatStore'
import dayjs from 'dayjs'

const props = defineProps<{
  message: Message
  isOwn: boolean
}>()

const emit = defineEmits(['reply', 'forward', 'delete'])

const authStore = useAuthStore()
const chatStore = useChatStore()
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
let longPressTimer: ReturnType<typeof setTimeout> | null = null

// Аудио плеер
const audio = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const progressPercent = computed(() => (currentTime.value / duration.value) * 100 || 0)

const contextMenuStyle = computed(() => ({
  left: `${contextMenuPosition.value.x}px`,
  top: `${contextMenuPosition.value.y}px`
}))

function formatTime(date: string): string {
  return dayjs(date).locale('ru').format('HH:mm')
}

function getForwardedName(senderId: number): string {
  if (senderId === authStore.user?.id) return 'Вас'
  const chat = chatStore.currentChat
  if (chat && chat.type === 'group') {
    return 'участника группы'
  }
  return 'собеседника'
}

function formatDuration(seconds: number): string {
  if (!seconds || isNaN(seconds) || !isFinite(seconds) || seconds < 0) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

function getMessagePreview(msg: Message): string {
  if (msg.text) {
    return msg.text.length > 50 ? msg.text.substring(0, 50) + '...' : msg.text
  }
  if (msg.file_url) {
    if (msg.file_type === 'image') return '📷 Фото'
    if (msg.file_type === 'video') return '🎥 Видео'
    if (msg.file_type === 'audio' || msg.file_type === 'voice') return '🎙️ Голосовое сообщение'
    return '📎 Вложение'
  }
  return 'Сообщение'
}

function getReplySenderName(senderId: number): string {
  if (senderId === authStore.user?.id) return 'Вы'
  const chat = chatStore.currentChat
  if (chat && chat.type === 'group') {
    return 'Участник'
  }
  return 'Собеседник'
}

function scrollToReply(messageId: number) {
  const element = document.querySelector(`[data-message-id="${messageId}"]`)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'center' })
    element.classList.add('highlight-message')
    setTimeout(() => {
      element.classList.remove('highlight-message')
    }, 1000)
  }
}

function openFile(url: string | undefined) {
  if (url) {
    window.open(url, '_blank')
  }
}

function initAudio() {
  if (props.message.file_url && (props.message.file_type === 'audio' || props.message.file_type === 'voice')) {
    audio.value = new Audio(props.message.file_url)
    
    audio.value.addEventListener('loadedmetadata', () => {
      duration.value = audio.value?.duration || 0
      console.log('Audio duration:', duration.value)
    })
    
    audio.value.addEventListener('timeupdate', () => {
      currentTime.value = audio.value?.currentTime || 0
    })
    
    audio.value.addEventListener('ended', () => {
      isPlaying.value = false
      currentTime.value = 0
    })
    
    audio.value.addEventListener('error', (e) => {
      console.error('Audio loading error:', e)
      duration.value = 0
    })
    
    // Принудительно загружаем метаданные
    audio.value.load()
  }
}

function togglePlay() {
  if (!audio.value) return
  
  if (isPlaying.value) {
    audio.value.pause()
    isPlaying.value = false
  } else {
    // Останавливаем все другие аудио
    document.querySelectorAll('audio').forEach(a => {
      if (a !== audio.value) {
        a.pause()
      }
    })
    audio.value.play().catch(err => {
      console.error('Playback failed:', err)
    })
    isPlaying.value = true
  }
}

function seekAudio(event: MouseEvent) {
  if (!audio.value || !duration.value || duration.value <= 0) return
  
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  if (!rect.width) return
  
  let x = event.clientX - rect.left
  x = Math.max(0, Math.min(rect.width, x))
  const percent = x / rect.width
  const newTime = percent * duration.value
  
  if (isFinite(newTime) && newTime >= 0 && newTime <= duration.value) {
    audio.value.currentTime = newTime
    currentTime.value = newTime
  }
}

function handleTouchStart(event: TouchEvent) {
  longPressTimer = setTimeout(() => {
    const touch = event.touches[0]
    contextMenuPosition.value = { x: touch.clientX, y: touch.clientY }
    contextMenuVisible.value = true
    setTimeout(() => {
      contextMenuVisible.value = false
    }, 3000)
  }, 500)
}

function handleTouchEnd() {
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
}

function openContextMenu(event: MouseEvent) {
  event.preventDefault()
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
  contextMenuVisible.value = true
  setTimeout(() => {
    document.addEventListener('click', closeContextMenu)
  }, 0)
}

function closeContextMenu() {
  contextMenuVisible.value = false
  document.removeEventListener('click', closeContextMenu)
}

function handleReply() {
  emit('reply', props.message)
  closeContextMenu()
}

function handleForward() {
  emit('forward', props.message)
  closeContextMenu()
}

function handleDelete() {
  if (confirm('Удалить сообщение?')) {
    emit('delete', props.message.id)
  }
  closeContextMenu()
}

onMounted(() => {
  document.addEventListener('click', closeContextMenu)
  initAudio()
})

onUnmounted(() => {
  if (longPressTimer) {
    clearTimeout(longPressTimer)
    longPressTimer = null
  }
  document.removeEventListener('click', closeContextMenu)
  if (audio.value) {
    audio.value.pause()
    audio.value = null
  }
})
</script>

<style scoped>
.message-wrapper {
  display: flex;
  flex-direction: column;
  margin-bottom: 12px;
  position: relative;
}

.message-own {
  align-items: flex-end;
}

.message-other {
  align-items: flex-start;
}

/* Цитата ответа */
.reply-quote {
  display: flex;
  align-items: center;
  gap: 10px;
  max-width: 85%;
  margin-bottom: 6px;
  padding: 8px 12px;
  background: rgba(157, 19, 6, 0.15);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.reply-quote:hover {
  background: rgba(157, 19, 6, 0.25);
  transform: scale(1.01);
}

.quote-line {
  width: 3px;
  height: 36px;
  background: var(--primary-color);
  border-radius: 2px;
  flex-shrink: 0;
}

.quote-content {
  flex: 1;
  overflow: hidden;
}

.quote-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.quote-header svg {
  width: 12px;
  height: 12px;
  stroke: var(--primary-color);
  flex-shrink: 0;
}

.quote-name {
  font-size: 12px;
  font-weight: 700;
  color: var(--primary-color);
}

.quote-text {
  font-size: 11px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Основное сообщение */
.message-bubble {
  max-width: 70%;
  position: relative;
}

.message-content {
  padding: 8px 12px;
  border-radius: 18px;
  position: relative;
}

.message-own .message-content {
  background: var(--primary-color);
  color: white;
  border-bottom-right-radius: 4px;
}

.message-other .message-content {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border-bottom-left-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.message-text {
  font-size: 14px;
  line-height: 1.4;
  word-wrap: break-word;
}

.message-meta {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  margin-top: 4px;
  font-size: 10px;
  opacity: 0.7;
}

.message-status {
  font-size: 12px;
}

/* Аудио плеер - темный, с большой полосой */
.audio-message {
  min-width: 260px;
}

.audio-player {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #1a1a1a;
  border-radius: 28px;
  padding: 8px 16px;
}

.message-own .audio-player {
  background: #2a1a18;
}

.message-other .audio-player {
  background: #1a1a1a;
}

.play-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--primary-color);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: transform 0.2s;
}

.play-btn:hover {
  transform: scale(1.05);
}

.progress-container {
  flex: 1;
  cursor: pointer;
}

.progress-bar {
  height: 6px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 6px;
}

.progress-fill {
  height: 100%;
  background: var(--primary-color);
  border-radius: 3px;
  transition: width 0.1s linear;
}

.time-info {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: rgba(255, 255, 255, 0.6);
}

.current-time {
  font-family: monospace;
}

.duration {
  display: none;
  font-family: monospace;
}

/* Видео сообщение */
.video-message {
  position: relative;
  cursor: pointer;
  max-width: 250px;
  border-radius: 12px;
  overflow: hidden;
}

.video-thumbnail {
  width: 100%;
  aspect-ratio: 16/9;
  object-fit: cover;
  background: #000;
}

.video-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 48px;
  height: 48px;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Изображение */
.image-message img {
  max-width: 200px;
  max-height: 200px;
  border-radius: 12px;
  object-fit: cover;
  cursor: pointer;
}

/* Документ */
.document-message {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 12px;
  cursor: pointer;
  min-width: 180px;
}

.message-own .document-message {
  background: rgba(255, 255, 255, 0.1);
}

/* Подсветка при переходе */
.highlight-message {
  animation: highlight 0.5s ease;
}

@keyframes highlight {
  0% { background: transparent; }
  50% { background: rgba(157, 19, 6, 0.25); border-radius: 12px; }
  100% { background: transparent; }
}

/* Контекстное меню */
.context-menu {
  position: fixed;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  overflow: hidden;
  z-index: 1000;
  min-width: 160px;
  animation: fadeIn 0.15s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-primary);
  transition: background 0.2s;
  text-align: left;
}

.menu-item:hover {
  background: var(--bg-hover);
}

.menu-item.danger {
  color: var(--danger-color);
}

.menu-item.danger:hover {
  background: rgba(157, 19, 6, 0.15);
}

.forward-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
  font-size: 11px;
  color: var(--text-muted);
}

.forward-indicator svg {
  width: 12px;
  height: 12px;
}

@media (max-width: 768px) {
  .message-bubble {
    max-width: 85%;
  }
  
  .message-content {
    padding: 6px 10px;
  }
  
  .message-text {
    font-size: 13px;
  }
  
  .audio-message {
    min-width: 220px;
  }
  
  .audio-player {
    padding: 6px 12px;
    gap: 10px;
  }
  
  .play-btn {
    width: 36px;
    height: 36px;
  }
  
  .progress-bar {
    height: 5px;
  }
  
  .image-message img {
    max-width: 150px;
    max-height: 150px;
  }
  
  .video-message {
    max-width: 200px;
  }
}
</style>