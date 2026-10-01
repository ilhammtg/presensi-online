<template>
  <header class="app-header">
    <div class="header-inner">
      <!-- Institutional Branding -->
      <div class="brand-container">
        <div class="brand-logo-box">
          <img v-if="theme.logoUrl" :src="theme.logoUrl" alt="Logo" class="custom-logo" />
          <div v-else class="default-logo-crest">
            <i class="fas fa-university"></i>
          </div>
        </div>
        <div class="brand-text">
          <div class="brand-name">{{ theme.institutionName }}</div>
          <div class="brand-sub">{{ theme.institutionTagline }}</div>
        </div>
      </div>

      <!-- Navigation & Profile (Role-Segregated) -->
      <div class="header-right">
        <!-- Live Academic Clock & Calendar Widget (Visible for all roles) -->
        <div class="academic-clock-widget">
          <div class="clock-calendar-item">
            <i class="far fa-calendar-alt text-accent"></i>
            <span class="clock-date-text">{{ currentDateStr }}</span>
          </div>
          <div class="clock-divider"></div>
          <div class="clock-time-item">
            <span class="clock-seconds-pulse"></span>
            <i class="far fa-clock text-primary"></i>
            <span class="clock-time-text font-mono">{{ currentTimeStr }}</span>
            <span class="clock-tz">WIB</span>
          </div>
        </div>

        <!-- Navigation Links: STRICTLY SEGREGATED -->
        <nav class="nav-menu" v-if="auth.isAuthenticated">
          <!-- Dosen Navigation: Only for Dosen -->
          <template v-if="auth.isDosen">
            <router-link to="/dosen" class="nav-item" active-class="active">
              <i class="fas fa-chalkboard-teacher"></i>
              <span>Portal Perkuliahan</span>
            </router-link>
          </template>

          <!-- Admin Prodi Navigation: Only for Admin Prodi -->
          <template v-if="auth.isAdminProdi">
            <router-link to="/admin/prodi" class="nav-item" active-class="active">
              <i class="fas fa-chart-pie"></i>
              <span>Monitoring Prodi</span>
            </router-link>
          </template>

          <!-- Superadmin Navigation: Full system & switcher -->
          <template v-if="auth.isSuperAdmin">
            <router-link to="/admin/prodi" class="nav-item" active-class="active">
              <i class="fas fa-chart-pie"></i>
              <span>Monitoring Prodi</span>
            </router-link>
            <router-link to="/admin/system" class="nav-item" active-class="active">
              <i class="fas fa-tools"></i>
              <span>Sistem & Tema</span>
            </router-link>
          </template>
        </nav>

        <!-- User Profile Pill -->
        <div class="user-profile" v-if="auth.isAuthenticated">
          <div class="user-avatar-box" @click="showProfileModal = true" role="button" title="Buka Profil & Pengaturan Akun">
            <img v-if="auth.user?.avatar_url" :src="auth.user.avatar_url" alt="Avatar" class="avatar-img" />
            <div v-else class="user-avatar">
              <i class="fas" :class="auth.isDosen ? 'fa-user-tie' : auth.isAdminProdi ? 'fa-university' : 'fa-user-shield'"></i>
            </div>
          </div>
          <div class="user-meta" @click="showProfileModal = true" role="button" title="Buka Profil & Pengaturan Akun">
            <span class="user-name">{{ auth.user?.name }}</span>
            <span class="user-role-badge">
              {{ auth.isDosen ? 'Dosen Pengampu' : auth.isAdminProdi ? 'Admin Program Studi' : 'Super Administrator' }}
            </span>
          </div>
          <button class="btn-profile" @click="showProfileModal = true" title="Profil & Ganti Kata Sandi">
            <i class="fas fa-user-cog"></i>
          </button>
          <button class="btn-logout" @click="handleLogout" title="Keluar dari Sistem">
            <i class="fas fa-sign-out-alt"></i>
          </button>
        </div>
      </div>
    </div>
    <div class="header-gold-line"></div>

    <!-- Profile & Password Modal -->
    <ProfileModal v-if="showProfileModal" @close="showProfileModal = false" />
  </header>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'
import ProfileModal from './ProfileModal.vue'

const auth = useAuthStore()
const theme = useThemeStore()
const router = useRouter()
const showProfileModal = ref(false)

const currentDateStr = ref('')
const currentTimeStr = ref('')
let clockTimer = null

function updateDateTime() {
  const now = new Date()
  
  // Format Tanggal: "Jumat, 25 Sep 2026"
  currentDateStr.value = now.toLocaleDateString('id-ID', {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
  
  // Format Jam Digital: "02:05:12"
  const h = String(now.getHours()).padStart(2, '0')
  const m = String(now.getMinutes()).padStart(2, '0')
  const s = String(now.getSeconds()).padStart(2, '0')
  currentTimeStr.value = `${h}:${m}:${s}`
}

onMounted(() => {
  updateDateTime()
  clockTimer = setInterval(updateDateTime, 1000)
})

onUnmounted(() => {
  if (clockTimer) clearInterval(clockTimer)
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.app-header {
  background-color: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle);
  position: relative;
  margin-bottom: 24px;
}

.header-gold-line {
  height: 3px;
  background: linear-gradient(90deg, var(--brand-primary) 0%, var(--brand-accent) 50%, var(--brand-primary) 100%);
  width: 100%;
}

.header-inner {
  max-width: 1320px;
  margin: 0 auto;
  padding: 12px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand-container {
  display: flex;
  align-items: center;
  gap: 14px;
}

.brand-logo-box {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  background-color: var(--brand-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 1.25rem;
  border: 1px solid var(--brand-accent);
}

.custom-logo {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.brand-name {
  font-size: 1.1rem;
  font-weight: 800;
  color: var(--text-main);
  line-height: 1.2;
}

.brand-sub {
  font-size: 0.775rem;
  color: var(--brand-primary);
  font-weight: 600;
  letter-spacing: 0.02em;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

/* Academic Clock & Calendar Widget */
.academic-clock-widget {
  display: flex;
  align-items: center;
  gap: 12px;
  background-color: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: 30px;
  padding: 6px 14px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.clock-calendar-item,
.clock-time-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-main);
}

.clock-date-text {
  color: var(--text-muted);
}

.clock-divider {
  width: 1px;
  height: 14px;
  background-color: var(--border-subtle);
}

.clock-seconds-pulse {
  width: 6px;
  height: 6px;
  background-color: #22c55e;
  border-radius: 50%;
  animation: pulse-dot 1.2s infinite;
}

@keyframes pulse-dot {
  0% {
    transform: scale(0.9);
    opacity: 0.7;
  }
  50% {
    transform: scale(1.2);
    opacity: 1;
  }
  100% {
    transform: scale(0.9);
    opacity: 0.7;
  }
}

.clock-time-text {
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: 0.05em;
}

.clock-tz {
  font-size: 0.65rem;
  font-weight: 800;
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
  padding: 1px 4px;
  border-radius: 3px;
}

@media (max-width: 900px) {
  .clock-calendar-item {
    display: none;
  }
  .clock-divider {
    display: none;
  }
}

.nav-menu {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-muted);
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.nav-item:hover {
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
}

.nav-item.active {
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
  border-color: var(--brand-primary-border);
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-left: 16px;
  border-left: 1px solid var(--border-subtle);
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: var(--brand-primary-light);
  color: var(--brand-primary);
  border: 1px solid var(--brand-primary-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.95rem;
}

.user-meta {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.2;
}

.user-role-badge {
  font-size: 0.725rem;
  color: var(--text-muted);
  font-weight: 500;
}

.user-avatar-box {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.avatar-img {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--brand-accent);
}

.btn-profile,
.btn-logout {
  background: transparent;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
  margin-left: 4px;
}

.btn-profile:hover {
  background-color: var(--brand-primary-light);
  border-color: var(--brand-primary);
  color: var(--brand-primary);
}

.btn-logout:hover {
  background-color: #fee2e2;
  border-color: #fca5a5;
  color: #dc2626;
}
</style>
