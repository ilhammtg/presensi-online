import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const loading = ref(false)
  const error = ref('')

  const isAuthenticated = computed(() => !!token.value)
  const isDosen = computed(() => user.value?.role === 'dosen')
  const isAdminProdi = computed(() => user.value?.role === 'admin_prodi')
  const isSuperAdmin = computed(() => user.value?.role === 'superadmin')
  const isAdmin = computed(() => isAdminProdi.value || isSuperAdmin.value)

  async function login(usernameOrEmail, password) {
    loading.value = true
    error.value = ''
    try {
      const res = await fetch('/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          username: usernameOrEmail,
          email: usernameOrEmail, 
          password 
        })
      })

      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.error || 'Login gagal')
      }

      token.value = data.access_token
      user.value = data.user
      localStorage.setItem('token', data.access_token)
      localStorage.setItem('user', JSON.stringify(data.user))
      return true
    } catch (err) {
      error.value = err.message
      return false
    } finally {
      loading.value = false
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function fetchProfile() {
    if (!token.value) return null
    try {
      const res = await fetch('/v1/auth/me', {
        headers: { Authorization: `Bearer ${token.value}` }
      })
      if (res.ok) {
        const data = await res.json()
        user.value = { ...user.value, ...data }
        localStorage.setItem('user', JSON.stringify(user.value))
        return data
      }
    } catch (err) {
      console.error('Failed to fetch profile:', err)
    }
    return null
  }

  async function changePassword(oldPassword, newPassword, confirmPassword) {
    const res = await fetch('/v1/auth/change-password', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token.value}`
      },
      body: JSON.stringify({
        old_password: oldPassword,
        new_password: newPassword,
        confirm_password: confirmPassword
      })
    })
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Gagal mengubah kata sandi')
    }
    return data.message
  }

  async function updateProfilePhoto(avatarUrl) {
    const res = await fetch('/v1/auth/profile-photo', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token.value}`
      },
      body: JSON.stringify({ avatar_url: avatarUrl })
    })
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Gagal menyimpan foto profil')
    }
    user.value = { ...user.value, avatar_url: avatarUrl }
    localStorage.setItem('user', JSON.stringify(user.value))
    return data.message
  }

  return {
    token,
    user,
    loading,
    error,
    isAuthenticated,
    isDosen,
    isAdmin,
    isAdminProdi,
    isSuperAdmin,
    login,
    logout,
    fetchProfile,
    changePassword,
    updateProfilePhoto
  }
})
