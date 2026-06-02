import api from './client'
import type { User } from '@/types/user'

export const usersApi = {
  searchUsers: (query: string, limit = 20) => 
    api.get<{ success: boolean; data: User[] }>('/users/search', {
      params: { q: query, limit }
    }),
  
  getUserById: (userId: number) => 
    api.get<{ success: boolean; data: User }>(`/users/${userId}`)
}