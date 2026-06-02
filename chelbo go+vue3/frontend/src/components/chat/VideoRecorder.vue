<template>
  <div class="video-recorder">
    <div v-if="!recording && !videoBlob" class="recorder-start">
      <button class="record-btn" @click="startRecording">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10" />
          <polygon points="10 8 16 12 10 16 10 8" />
        </svg>
        <span>Начать запись</span>
      </button>
    </div>

    <div v-if="recording" class="recording-active">
      <video ref="videoPreview" autoplay muted playsinline class="video-preview"></video>
      <div class="recording-overlay">
        <div class="recording-info">
          <div class="recording-dot"></div>
          <span class="recording-time">{{ formatTime(recordingTime) }}</span>
        </div>
        <div class="recording-actions">
          <button class="stop-btn" @click="stopRecording">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="6" y="6" width="12" height="12" />
            </svg>
            <span>Остановить</span>
          </button>
          <button class="cancel-btn" @click="cancelRecording">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
            <span>Отмена</span>
          </button>
        </div>
      </div>
    </div>

    <div v-if="videoBlob && !recording" class="recorder-preview">
      <video :src="videoUrl" controls class="video-preview"></video>
      <div class="preview-actions">
        <button class="send-btn" @click="sendVideo">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
          <span>Отправить</span>
        </button>
        <button class="cancel-btn" @click="clearVideo">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
          <span>Отмена</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onUnmounted } from 'vue'

const emit = defineEmits(['send'])

const recording = ref(false)
const videoBlob = ref<Blob | null>(null)
const videoUrl = ref('')
const mediaRecorder = ref<MediaRecorder | null>(null)
const videoChunks = ref<Blob[]>([])
const videoPreview = ref<HTMLVideoElement | null>(null)
const recordingTime = ref(0)
let timeInterval: ReturnType<typeof setInterval> | null = null
let stream: MediaStream | null = null

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

async function startRecording() {
  try {
    stream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    
    if (videoPreview.value) {
      videoPreview.value.srcObject = stream
    }
    
    mediaRecorder.value = new MediaRecorder(stream)
    videoChunks.value = []
    
    mediaRecorder.value.ondataavailable = (event) => {
      if (event.data.size > 0) {
        videoChunks.value.push(event.data)
      }
    }
    
    mediaRecorder.value.onstop = () => {
      const blob = new Blob(videoChunks.value, { type: 'video/webm' })
      videoBlob.value = blob
      videoUrl.value = URL.createObjectURL(blob)
      recording.value = false
      
      if (timeInterval) {
        clearInterval(timeInterval)
        timeInterval = null
      }
      
      if (stream) {
        stream.getTracks().forEach(track => track.stop())
        stream = null
      }
    }
    
    mediaRecorder.value.start(1000)
    recording.value = true
    recordingTime.value = 0
    
    timeInterval = setInterval(() => {
      recordingTime.value++
    }, 1000)
    
  } catch (error) {
    console.error('Failed to start video recording:', error)
    alert('Не удалось получить доступ к камере и микрофону. Разрешите доступ в браузере.')
  }
}

function stopRecording() {
  if (mediaRecorder.value && mediaRecorder.value.state === 'recording') {
    mediaRecorder.value.stop()
  }
}

function cancelRecording() {
  if (mediaRecorder.value && mediaRecorder.value.state === 'recording') {
    mediaRecorder.value.stop()
  }
  
  if (timeInterval) {
    clearInterval(timeInterval)
    timeInterval = null
  }
  
  recording.value = false
  recordingTime.value = 0
  
  if (stream) {
    stream.getTracks().forEach(track => track.stop())
    stream = null
  }
}

function clearVideo() {
  videoBlob.value = null
  videoUrl.value = ''
}

function sendVideo() {
  if (videoBlob.value) {
    emit('send', videoBlob.value)
    clearVideo()
  }
}

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
  if (stream) {
    stream.getTracks().forEach(track => track.stop())
  }
  if (mediaRecorder.value && mediaRecorder.value.state === 'recording') {
    mediaRecorder.value.stop()
  }
})
</script>

<style scoped>
.video-recorder {
  background: var(--bg-primary);
  border-radius: 20px;
  padding: 24px;
  min-width: 400px;
  text-align: center;
}

.recorder-start {
  display: flex;
  justify-content: center;
}

.record-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px;
  background: var(--primary-light);
  border: 2px solid var(--primary-color);
  border-radius: 50%;
  color: var(--primary-color);
  transition: all 0.2s;
}

.record-btn span {
  font-size: 14px;
  white-space: nowrap;
}

.record-btn:hover {
  background: var(--primary-color);
  color: white;
  transform: scale(1.05);
}

.recording-active {
  position: relative;
}

.video-preview {
  width: 100%;
  max-width: 500px;
  border-radius: 16px;
  background: #000;
}

.recording-overlay {
  position: absolute;
  bottom: 20px;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: linear-gradient(to top, rgba(0,0,0,0.7), transparent);
  border-radius: 16px;
}

.recording-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.recording-dot {
  width: 12px;
  height: 12px;
  background: red;
  border-radius: 50%;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.recording-time {
  font-size: 18px;
  font-weight: 600;
  color: white;
}

.recording-actions {
  display: flex;
  gap: 12px;
}

.stop-btn, .cancel-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 14px;
  background: rgba(0,0,0,0.6);
  color: white;
  backdrop-filter: blur(5px);
}

.stop-btn:hover, .cancel-btn:hover {
  background: rgba(0,0,0,0.8);
}

.recorder-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

.recorder-preview .video-preview {
  max-width: 400px;
  max-height: 300px;
}

.preview-actions {
  display: flex;
  gap: 20px;
}

.send-btn, .preview-actions .cancel-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 12px;
  font-size: 14px;
}

.send-btn {
  background: var(--primary-color);
  color: white;
}

.send-btn:hover {
  background: var(--primary-hover);
}

@media (max-width: 768px) {
  .video-recorder {
    padding: 16px;
    min-width: 300px;
  }
  
  .recording-overlay {
    padding: 8px 16px;
  }
  
  .recording-time {
    font-size: 14px;
  }
  
  .stop-btn, .cancel-btn {
    padding: 6px 12px;
    font-size: 12px;
  }
}
</style>