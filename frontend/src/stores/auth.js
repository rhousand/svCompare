import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)

  async function fetchMe() {
    try {
      const res = await fetch('/api/auth/me', { credentials: 'include' })
      if (res.ok) {
        const data = await res.json()
        user.value = data.user
      } else {
        user.value = null
      }
    } catch {
      user.value = null
    }
  }

  async function login(username, password) {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || 'Login failed')
    }
    const data = await res.json()
    user.value = data.user
  }

  async function logout() {
    await fetch('/api/auth/logout', { method: 'POST', credentials: 'include' })
    user.value = null
  }

  return { user, fetchMe, login, logout }
})
