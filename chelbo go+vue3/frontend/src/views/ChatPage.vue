<template>
  <div class="chat-page" :class="{ 'mobile-mode': isMobile }">
    <!-- Список чатов - всегда виден на десктопе, скрывается на мобиле -->
    <div 
      class="chat-list-wrapper" 
      :class="{ 
        'visible': !isMobile || (isMobile && !isChatOpen),
        'hidden': isMobile && isChatOpen
      }"
    >
      <ChatList @select-chat="handleSelectChat" />
    </div>

    <!-- Окно чата - открывается поверх на мобиле -->
    <div 
      class="chat-window-wrapper" 
      :class="{ 
        'visible': !isMobile || (isMobile && isChatOpen),
        'hidden': isMobile && !isChatOpen
      }"
    >
      <div v-if="chatStore.currentChat" class="chat-container">
        <!-- Кнопка назад на мобиле -->
        <button v-if="isMobile" class="mobile-back-btn" @click="closeChat">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="15 18 9 12 15 6" />
          </svg>
          <span>Назад</span>
        </button>
        
        <ChatWindow 
          :chat="chatStore.currentChat" 
          @search="handleSearch"
          @settings="handleChatSettings"
        />
      </div>
      
      <!-- Пустое состояние -->
      <div v-else class="empty-state">
        <div class="empty-content">
          <div class="empty-icon">
            <svg width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
            </svg>
          </div>
          <h2>Chelbo</h2>
          <p>Выберите чат или начните новый диалог</p>
          <button class="btn btn-primary" @click="showNewChatModal = true">
            Начать новый чат
          </button>
        </div>
      </div>
    </div>

    <!-- Модальное окно настроек -->
    <div v-if="showSettings" class="modal-overlay" @click.self="showSettings = false">
      <div class="modal-content settings-modal">
        <div class="modal-header">
          <h3>Настройки профиля</h3>
          <button class="modal-close" @click="showSettings = false">×</button>
        </div>
        
        <div class="avatar-section">
          <div class="avatar-preview">
            <img v-if="authStore.user?.avatar_url" :src="authStore.user.avatar_url" alt="Avatar" />
            <span v-else>{{ authStore.user?.name?.charAt(0).toUpperCase() }}</span>
          </div>
          <button class="btn btn-secondary" @click="triggerAvatarUpload">
            Изменить аватар
          </button>
          <input ref="avatarInput" type="file" accept="image/*" style="display: none" @change="uploadAvatar" />
        </div>
        
        <div class="form-group">
          <label>Имя</label>
          <input v-model="editForm.name" type="text" class="form-input" />
        </div>
        
        <div class="form-group">
          <label>О себе</label>
          <textarea v-model="editForm.bio" rows="3" class="form-input" placeholder="Расскажите о себе..."></textarea>
        </div>
        
        <div class="form-group">
          <label>Email</label>
          <input :value="authStore.user?.email" type="email" class="form-input" disabled />
        </div>
        
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="showSettings = false">Отмена</button>
          <button class="btn btn-primary" @click="saveProfile">Сохранить</button>
        </div>
        
        <button class="logout-btn btn btn-danger" @click="handleLogout">
          Выйти из аккаунта
        </button>
      </div>
    </div>

    <!-- Модальное окно нового чата -->
    <div v-if="showNewChatModal" class="modal-overlay" @click.self="showNewChatModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h3>Новый чат</h3>
          <button class="modal-close" @click="showNewChatModal = false">×</button>
        </div>
        <div class="modal-tabs">
          <button :class="{ active: newChatType === 'private' }" @click="newChatType = 'private'">
            Личный чат
          </button>
          <button :class="{ active: newChatType === 'group' }" @click="newChatType = 'group'">
            Группа
          </button>
        </div>

        <div v-if="newChatType === 'private'" class="private-chat">
          <input
            v-model="userSearchQuery"
            type="text"
            placeholder="Поиск пользователей..."
            class="search-input"
            @input="searchUsers"
          />
          <div class="users-list">
            <div
              v-for="user in searchResults"
              :key="user.id"
              class="user-item"
              @click="createPrivateChat(user.id)"
            >
              <div class="user-avatar">
                <span>{{ getInitials(user.name) }}</span>
              </div>
              <div class="user-info">
                <div class="user-name">{{ user.name }}</div>
                <div class="user-email">{{ user.email }}</div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="group-chat">
          <input
            v-model="groupTitle"
            type="text"
            placeholder="Название группы"
            class="search-input"
          />
          <input
            v-model="groupMembersQuery"
            type="text"
            placeholder="Поиск участников..."
            class="search-input"
            @input="searchGroupMembers"
          />
          <div class="selected-members">
            <div v-for="member in selectedMembers" :key="member.id" class="member-tag">
              {{ member.name }}
              <button @click="removeMember(member.id)">×</button>
            </div>
          </div>
          <div class="users-list">
            <div
              v-for="user in groupSearchResults"
              :key="user.id"
              v-show="!selectedMembers.find(m => m.id === user.id)"
              class="user-item"
              @click="addMember(user)"
            >
              <div class="user-avatar">
                <span>{{ getInitials(user.name) }}</span>
              </div>
              <div class="user-info">
                <div class="user-name">{{ user.name }}</div>
                <div class="user-email">{{ user.email }}</div>
              </div>
            </div>
          </div>
          <button
            class="btn btn-primary create-group-btn"
            :disabled="!groupTitle || selectedMembers.length === 0"
            @click="createGroup"
          >
            Создать группу
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import { useChatStore } from '@/stores/chatStore'
import { useWebSocketStore } from '@/stores/websocketStore'
import { authApi } from '@/api/auth'
import { usersApi } from '@/api/users'
import { chatsApi } from '@/api/chats'
import type { User } from '@/types/user'
import ChatList from '@/components/chat/ChatList.vue'
import ChatWindow from '@/components/chat/ChatWindow.vue'

const router = useRouter()
const authStore = useAuthStore()
const chatStore = useChatStore()
const wsStore = useWebSocketStore()

// Мобильное определение
const isMobile = computed(() => window.innerWidth <= 768)
const isChatOpen = ref(false)

const showSettings = ref(false)
const showNewChatModal = ref(false)
const avatarInput = ref<HTMLInputElement | null>(null)
const newChatType = ref<'private' | 'group'>('private')
const userSearchQuery = ref('')
const groupMembersQuery = ref('')
const groupTitle = ref('')
const searchResults = ref<User[]>([])
const groupSearchResults = ref<User[]>([])
const selectedMembers = ref<User[]>([])

const editForm = reactive({
  name: authStore.user?.name || '',
  bio: authStore.user?.bio || ''
})

function getInitials(name: string): string {
  return name.charAt(0).toUpperCase()
}

function handleSelectChat(chat: any) {
  chatStore.setCurrentChat(chat)
  isChatOpen.value = true
}

function closeChat() {
  isChatOpen.value = false
  chatStore.setCurrentChat(null)
}

function handleSearch() {
  console.log('Search messages')
}

function handleChatSettings() {
  showSettings.value = true
}

function triggerAvatarUpload() {
  avatarInput.value?.click()
}

async function uploadAvatar(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return
  alert('Загрузка аватара будет доступна в следующей версии')
}

async function saveProfile() {
  try {
    const response = await authApi.updateProfile({
      name: editForm.name,
      bio: editForm.bio
    })
    if (response.data.data) {
      authStore.updateUser(response.data.data)
      alert('Профиль обновлен')
      showSettings.value = false
    }
  } catch (error) {
    console.error('Failed to update profile:', error)
    alert('Ошибка обновления профиля')
  }
}

async function handleLogout() {
  await authStore.logout()
  wsStore.disconnect()
  router.push('/login')
}

async function searchUsers() {
  if (userSearchQuery.value.length < 2) {
    searchResults.value = []
    return
  }
  try {
    const response = await usersApi.searchUsers(userSearchQuery.value)
    searchResults.value = response.data.data || []
  } catch (error) {
    console.error('Failed to search users:', error)
  }
}

async function searchGroupMembers() {
  if (groupMembersQuery.value.length < 2) {
    groupSearchResults.value = []
    return
  }
  try {
    const response = await usersApi.searchUsers(groupMembersQuery.value)
    groupSearchResults.value = (response.data.data || []).filter(
      u => u.id !== authStore.user?.id
    )
  } catch (error) {
    console.error('Failed to search group members:', error)
  }
}

async function createPrivateChat(userId: number) {
  try {
    const response = await chatsApi.createPrivateChat(userId)
    await chatStore.loadChats()
    const newChat = chatStore.chats.find(c => c.id === response.data.data!.chat_id)
    if (newChat) {
      handleSelectChat(newChat)
    }
    showNewChatModal.value = false
    userSearchQuery.value = ''
    searchResults.value = []
  } catch (error) {
    console.error('Failed to create chat:', error)
    alert('Не удалось создать чат')
  }
}

function addMember(user: User) {
  if (!selectedMembers.value.find(m => m.id === user.id)) {
    selectedMembers.value.push(user)
  }
  groupMembersQuery.value = ''
  groupSearchResults.value = []
}

function removeMember(userId: number) {
  selectedMembers.value = selectedMembers.value.filter(m => m.id !== userId)
}

async function createGroup() {
  if (!groupTitle.value || selectedMembers.value.length === 0) return
  try {
    await chatsApi.createGroup({
      title: groupTitle.value,
      participant_emails: selectedMembers.value.map(m => m.email)
    })
    await chatStore.loadChats()
    showNewChatModal.value = false
    groupTitle.value = ''
    selectedMembers.value = []
    groupMembersQuery.value = ''
  } catch (error) {
    console.error('Failed to create group:', error)
    alert('Не удалось создать группу')
  }
}

onMounted(() => {
  wsStore.connect()
  chatStore.loadChats()
})

onUnmounted(() => {
  wsStore.disconnect()
})
</script>

<style scoped>
.chat-page {
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: var(--bg-secondary);
}

/* Десктоп версия */
.chat-list-wrapper {
  width: 320px;
  flex-shrink: 0;
  background: var(--bg-primary);
  border-right: 1px solid var(--border-color);
  transition: transform 0.3s ease;
}

.chat-window-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  transition: transform 0.3s ease;
}

.chat-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.mobile-back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border-color);
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  color: var(--primary-color);
  width: 100%;
  text-align: left;
}

/* Мобильная версия */
@media (max-width: 768px) {
  .chat-list-wrapper {
    position: fixed;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    z-index: 10;
    transform: translateX(0);
  }
  
  .chat-list-wrapper.hidden {
    transform: translateX(-100%);
  }
  
  .chat-list-wrapper.visible {
    transform: translateX(0);
  }
  
  .chat-window-wrapper {
    position: fixed;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    z-index: 20;
    background: var(--bg-secondary);
    transform: translateX(100%);
  }
  
  .chat-window-wrapper.visible {
    transform: translateX(0);
  }
  
  .chat-window-wrapper.hidden {
    transform: translateX(100%);
  }
}

/* Пустое состояние */
.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary);
}

.empty-content {
  text-align: center;
  padding: 40px;
}

.empty-icon {
  width: 120px;
  height: 120px;
  margin: 0 auto 24px;
  background: var(--bg-primary);
  border-radius: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary-color);
}

.empty-content h2 {
  font-size: 24px;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.empty-content p {
  color: var(--text-secondary);
  margin-bottom: 24px;
}

/* Модальные окна */
.modal-overlay {
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
  max-height: 90vh;
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

.modal-tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.modal-tabs button {
  flex: 1;
  padding: 10px;
  border-radius: 8px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-weight: 500;
  cursor: pointer;
}

.modal-tabs button.active {
  background: var(--primary-color);
  color: white;
}

.avatar-section {
  text-align: center;
  margin-bottom: 24px;
}

.avatar-preview {
  width: 100px;
  height: 100px;
  margin: 0 auto 12px;
  border-radius: 50%;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 36px;
  font-weight: 600;
  overflow: hidden;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: var(--text-secondary);
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.form-input:focus {
  border-color: var(--primary-color);
  outline: none;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.logout-btn {
  width: 100%;
  margin-top: 20px;
}

.search-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.users-list {
  max-height: 300px;
  overflow-y: auto;
  margin-top: 12px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  cursor: pointer;
  border-radius: 8px;
}

.user-item:hover {
  background: var(--bg-hover);
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: var(--primary-color);
}

.selected-members {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 12px 0;
}

.member-tag {
  background: var(--primary-light);
  color: var(--primary-color);
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.member-tag button {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: var(--primary-color);
}

.create-group-btn {
  width: 100%;
  margin-top: 16px;
}
</style>