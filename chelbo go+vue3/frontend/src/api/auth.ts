import api from './client'
import type { User, LoginRequest, RegisterRequest } from '@/types/user'

export const authApi = {
  register: (data: RegisterRequest) => 
    api.post<{ success: boolean; data: { token: string; user: User } }>('/auth/register', data),
  
  login: (data: LoginRequest) => 
    api.post<{ success: boolean; data: { token: string; user: User } }>('/auth/login', data),
  
  logout: () => 
    api.post<{ success: boolean }>('/auth/logout'),
  
  getMe: () => 
    api.get<{ success: boolean; data: User }>('/auth/me'),
  
  updateProfile: (data: { name?: string; bio?: string }) => 
    api.put<{ success: boolean; data: User }>('/users/me', data)
}