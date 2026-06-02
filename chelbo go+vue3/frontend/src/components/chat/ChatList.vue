<template>
  <div class="chat-list" :class="{ 'mobile': isMobile }">
    <div class="chat-list-header">
      <div class="user-info" @click="$router.push('/settings')">
        <div class="avatar">
          <span>{{ getInitials(authStore.user?.name || '') }}</span>
        </div>
        <div class="user-details" v-if="!isMobile">
          <span class="user-name">{{ authStore.user?.name }}</span>
          <span class="user-status">Online</span>
        </div>
      </div>
      <button class="btn-icon new-chat-btn" @click="showNewChatModal = true">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>
    </div>

    <div class="chat-search">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Поиск чатов..."
        class="search-input"
      />
    </div>

    <div class="chats-list scrollable">
      <div
        v-for="chat in filteredChats"
        :key="chat.id"
        :class="['chat-item', { active: chatStore.currentChat?.id === chat.id }]"
        @click="selectChat(chat)"
      >
        <div class="chat-avatar">
          <span>{{ getChatTitle(chat).charAt(0).toUpperCase() }}</span>
        </div>
        <div class="chat-info">
          <div class="chat-name">{{ getChatTitle(chat) }}</div>
          <div class="chat-last-message">
            <template v-if="chat.last_message">
              {{ chat.last_message.text || 'Вложение' }}
            </template>
            <template v-else>
              Нет сообщений
            </template>
          </div>
        </div>
        <div class="chat-meta">
          <div class="chat-time">{{ formatTime(chat.updated_at) }}</div>
          <div v-if="chat.unread_count > 0" class="unread-badge">
            {{ chat.unread_count > 99 ? '99+' : chat.unread_count }}
          </div>
        </div>
      </div>

      <div v-if="chatStore.chats.length === 0 && !chatStore.loading" class="empty-chats">
        <p>Нет чатов</p>
        <button class="btn btn-primary" @click="showNewChatModal = true">
          Начать новый чат
        </button>
      </div>

      <Loader v-if="chatStore.loading" />
    </div>

    <!-- New Chat Modal -->
    <div v-if="showNewChatModal" class="modal" @click.self="showNewChatModal = false">
      <div class="modal-content" :class="{ 'mobile-modal': isMobile }">
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
            placeholder="Поиск пользователей по email или имени..."
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
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/authStore'
import { useChatStore } from '@/stores/chatStore'
import { usersApi } from '@/api/users'
import { chatsApi } from '@/api/chats'
import type { User } from '@/types/user'
import type { Chat as ChatType } from '@/types/chat'
import Loader from '@/components/common/Loader.vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/ru'

dayjs.extend(relativeTime)
dayjs.locale('ru')

const authStore = useAuthStore()
const chatStore = useChatStore()

// Мобильное определение
const isMobile = computed(() => window.innerWidth <= 768)

const emit = defineEmits(['select-chat'])

const searchQuery = ref('')
const showNewChatModal = ref(false)
const newChatType = ref<'private' | 'group'>('private')
const userSearchQuery = ref('')
const groupMembersQuery = ref('')
const groupTitle = ref('')
const searchResults = ref<User[]>([])
const groupSearchResults = ref<User[]>([])
const selectedMembers = ref<User[]>([])

const filteredChats = computed(() => {
  if (!searchQuery.value) return chatStore.chats
  return chatStore.chats.filter(chat => 
    getChatTitle(chat).toLowerCase().includes(searchQuery.value.toLowerCase())
  )
})

function getChatTitle(chat: ChatType): string {
  if (chat.type === 'group') {
    return chat.title || 'Группа'
  }
  return chat.title || 'Пользователь'
}

function getInitials(name: string): string {
  return name.charAt(0).toUpperCase()
}

function formatTime(date: string): string {
  return dayjs(date).fromNow()
}

function selectChat(chat: ChatType) {
  emit('select-chat', chat)
  if (chat.unread_count > 0) {
    chatStore.updateUnreadCount(chat.id, 0)
  }
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
    searchResults.value = []
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
    groupSearchResults.value = []
  }
}

async function createPrivateChat(userId: number) {
  try {
    const response = await chatsApi.createPrivateChat(userId)
    const chatId = response.data.data!.chat_id
    await chatStore.loadChats()
    const newChat = chatStore.chats.find(c => c.id === chatId)
    if (newChat) {
      selectChat(newChat)
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
    const response = await chatsApi.createGroup({
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
  chatStore.loadChats()
})
</script>

<style scoped>
.chat-list {
  width: 100%;
  height: 100vh;
  background: var(--bg-primary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

/* Мобильная версия - полная ширина */
.chat-list.mobile {
  width: 100vw;
}

.chat-list-header {
  padding: 20px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: var(--primary-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 600;
  cursor: pointer;
  overflow: hidden;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-weight: 600;
  font-size: 14px;
}

.user-status {
  font-size: 12px;
  color: var(--success-color);
}

.new-chat-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-light);
  color: var(--primary-color);
}

.chat-search {
  padding: 12px 16px;
}

.search-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  font-size: 14px;
  background: var(--bg-secondary);
}

.search-input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.chats-list {
  flex: 1;
  overflow-y: auto;
}

.chat-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.2s;
}

.chat-item:hover {
  background: var(--bg-hover);
}

.chat-item.active {
  background: var(--primary-light);
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
  flex-shrink: 0;
}

.chat-info {
  flex: 1;
  min-width: 0;
}

.chat-name {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-last-message {
  font-size: 13px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chat-meta {
  text-align: right;
  flex-shrink: 0;
  min-width: 50px;
}

.chat-time {
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 4px;
  white-space: nowrap;
}

.unread-badge {
  background: var(--primary-color);
  color: white;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
  display: inline-block;
  max-width: 32px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-chats {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.empty-chats p {
  margin-bottom: 16px;
}

/* Modal styles */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 16px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow-y: auto;
  padding: 24px;
}

.mobile-modal {
  width: 95%;
  max-height: 90vh;
  border-radius: 12px;
  padding: 16px;
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
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.modal-tabs button.active {
  background: var(--primary-color);
  color: white;
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
  transition: background 0.2s;
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
  padding: 0;
  line-height: 1;
}

.create-group-btn {
  width: 100%;
  margin-top: 16px;
}

/* Мобильные адаптации для элементов */
@media (max-width: 768px) {
  .chat-list-header {
    padding: 16px;
  }
  
  .chat-search {
    padding: 8px 12px;
  }
  
  .chat-item {
    padding: 10px 12px;
    gap: 10px;
  }
  
  .chat-avatar {
    width: 44px;
    height: 44px;
    font-size: 16px;
  }
  
  .chat-name {
    font-size: 13px;
  }
  
  .chat-last-message {
    font-size: 12px;
  }
  
  .chat-time {
    font-size: 10px;
  }
  
  .unread-badge {
    font-size: 10px;
    padding: 2px 5px;
    min-width: 16px;
  }
}

.chat-list {
  background: var(--bg-primary);
  border-right: 1px solid var(--border-color);
}

.chat-list-header {
  border-bottom: 1px solid var(--border-color);
}

.avatar {
  background: var(--primary-color);
}

.user-name {
  color: var(--text-primary);
}

.user-status {
  color: var(--success-color);
}

.new-chat-btn {
  background: var(--primary-light);
  color: var(--primary-color);
}

.search-input {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
}

.chat-item:hover {
  background: var(--bg-hover);
}

.chat-item.active {
  background: var(--primary-light);
}

.chat-name {
  color: var(--text-primary);
}

.chat-last-message {
  color: var(--text-secondary);
}

.chat-time {
  color: var(--text-muted);
}

.unread-badge {
  background: var(--primary-color);
  color: white;
}

/* Modal */
.modal-content {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
}

.modal-tabs button {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.modal-tabs button.active {
  background: var(--primary-color);
  color: white;
}

.user-item:hover {
  background: var(--bg-hover);
}

.user-avatar {
  background: var(--primary-light);
  color: var(--primary-color);
}

.member-tag {
  background: var(--primary-light);
  color: var(--primary-color);
}
</style>