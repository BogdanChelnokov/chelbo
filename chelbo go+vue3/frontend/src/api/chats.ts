import api from './client'
import type { Chat, Message, CreateGroupRequest, SendMessageRequest } from '@/types/chat'

export const chatsApi = {
  getChats: () => 
    api.get<{ success: boolean; data: Chat[] }>('/chats'),
  
  createPrivateChat: (userId: number) => 
    api.post<{ success: boolean; data: { chat_id: number } }>(`/chats/private/${userId}`),
  
  createGroup: (data: CreateGroupRequest) => 
    api.post<{ success: boolean; data: { chat_id: number } }>('/chats/group', data),
  
  getMessages: (chatId: number, limit = 50, offset = 0) => 
    api.get<{ success: boolean; data: Message[] }>(`/chats/${chatId}/messages`, {
      params: { limit, offset }
    }),
  
  sendMessage: (chatId: number, data: SendMessageRequest) => 
    api.post<{ success: boolean; data: Message }>(`/chats/${chatId}/messages`, data),
  
  deleteMessage: (messageId: number) => 
    api.delete<{ success: boolean }>(`/messages/${messageId}`),
  
  markAsRead: (messageId: number) => 
    api.post<{ success: boolean }>(`/messages/${messageId}/read`, {}),
  
  forwardMessage: (messageId: number, targetChatId: number) => 
    api.post<{ success: boolean; data: Message }>(`/messages/${messageId}/forward`, { target_chat_id: targetChatId })
}