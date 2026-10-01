<template>
  <div class="login-page">
    <div class="login-container">
      <!-- Institutional Crest & Header -->
      <div class="brand-header">
        <div class="logo-emblem">
          <i class="fas fa-university"></i>
        </div>
        <h1 class="portal-title">{{ theme.institutionName }}</h1>
        <p class="portal-subtitle">{{ theme.institutionTagline }}</p>
      </div>

      <!-- Main Login Card (Flat & High-Contrast) -->
      <div class="flat-card login-card">
        <div class="card-gold-bar"></div>
        
        <div class="card-body">
          <h2 class="card-title">Masuk ke Portal</h2>
          <p class="card-instruction">Silakan masukkan kredensial akun akademik Anda.</p>

          <!-- Error Alert Banner -->
          <div v-if="auth.error" class="alert-banner">
            <i class="fas fa-exclamation-circle alert-icon"></i>
            <span>{{ auth.error }}</span>
          </div>

          <!-- Form -->
          <form @submit.prevent="handleLogin" class="login-form">
            <div class="form-group">
              <label for="username">NPM (Mahasiswa) / NIDN (Dosen) / Username</label>
              <div class="input-group">
                <i class="fas fa-id-card input-icon"></i>
                <input 
                  id="username"
                  type="text" 
                  v-model="username" 
                  class="input-field input-with-icon font-mono" 
                  placeholder="Contoh: 25552010096 atau 0002018502" 
                  required
                />
              </div>
              <span class="text-xs text-muted" style="margin-top: 4px; display: block;">
                Login default kampus menggunakan Nomor Pokok Mahasiswa (NPM) atau NIDN Dosen.
              </span>
            </div>

            <div class="form-group">
              <label for="password">Kata Sandi</label>
              <div class="input-group">
                <i class="fas fa-lock input-icon"></i>
                <input 
                  id="password"
                  type="password" 
                  v-model="password" 
                  class="input-field input-with-icon" 
                  placeholder="Masukkan kata sandi" 
                  required
                />
              </div>
            </div>

            <button type="submit" class="btn btn-primary btn-block" :disabled="auth.loading">
              <i v-if="auth.loading" class="fas fa-circle-notch fa-spin"></i>
              <span>{{ auth.loading ? 'Memverifikasi...' : 'Masuk ke Sistem' }}</span>
            </button>
          </form>
        </div>
      </div>

      <!-- Institutional Footer -->
      <footer class="login-footer">
        <p>&copy; 2026 {{ theme.institutionName }}. Hak Cipta Dilindungi.</p>
      </footer>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'

const auth = useAuthStore()
const theme = useThemeStore()
const router = useRouter()

const username = ref('')
const password = ref('')

async function handleLogin() {
  const success = await auth.login(username.value, password.value)
  if (success) {
    if (auth.isDosen) {
      router.push('/dosen')
    } else {
      router.push('/admin')
    }
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background-color: var(--bg-page);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
}

.login-container {
  width: 100%;
  max-width: 440px;
}

.brand-header {
  text-align: center;
  margin-bottom: 24px;
}

.logo-emblem {
  width: 60px;
  height: 60px;
  margin: 0 auto 14px;
  background-color: var(--brand-primary);
  border: 2px solid var(--brand-accent);
  border-radius: 14px;
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.75rem;
}

.portal-title {
  font-size: 1.4rem;
  font-weight: 800;
  color: var(--text-main);
  line-height: 1.2;
}

.portal-subtitle {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-top: 4px;
}

.login-card {
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  background-color: var(--bg-surface);
  box-shadow: 0 4px 12px -2px rgba(15, 23, 42, 0.08);
}

.card-gold-bar {
  height: 4px;
  background: linear-gradient(90deg, var(--brand-primary) 0%, var(--brand-accent) 100%);
}

.card-body {
  padding: 32px 28px;
}

.card-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
}

.card-instruction {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 2px;
  margin-bottom: 20px;
}

.alert-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  background-color: var(--status-danger-bg);
  border: 1px solid var(--status-danger-border);
  color: var(--status-danger-text);
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  margin-bottom: 18px;
}

.alert-icon {
  font-size: 1rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--text-body);
}

.btn-block {
  width: 100%;
  padding: 11px;
  font-size: 0.95rem;
  margin-top: 8px;
}


.login-footer {
  text-align: center;
  margin-top: 20px;
  font-size: 0.75rem;
  color: var(--text-dim);
}
</style>
