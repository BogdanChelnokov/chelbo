export interface Chat {
  id: number
  type: 'private' | 'group'
  title?: string
  avatar_url?: string
  last_message?: {
    id: number
    text?: string
    created_at: string
    sender_id: number
  } | null
  unread_count: number
  updated_at: string
}

export interface Message {
  id: number
  chat_id: number
  sender_id: number
  text?: string
  file_url?: string
  file_type?: string
  is_delivered: boolean
  is_read: boolean
  created_at: string
  reply_to_id?: number
  reply_to?: Message
  forwarded_from?: number  // Добавляем поле
}

export interface SendMessageRequest {
  text?: string
  reply_to_id?: number
}

export interface CreateGroupRequest {
  title: string
  participant_emails: string[]
}

export interface ApiResponse<T = any> {
  success: boolean
  message?: string
  data?: T
  error?: string
}