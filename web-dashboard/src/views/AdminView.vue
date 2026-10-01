<template>
  <div class="admin-portal">
    <Navbar />

    <main class="portal-main">
      <!-- Admin Navigation Tabs -->
      <div class="tab-bar">
        <button 
          class="tab-btn" 
          :class="{ 'active': activeTab === 'monitoring' }"
          @click="activeTab = 'monitoring'"
        >
          <i class="fas fa-chart-pie"></i>
          <span>Monitoring Jadwal & Presensi</span>
        </button>

        <button 
          class="tab-btn" 
          :class="{ 'active': activeTab === 'theme' }"
          @click="activeTab = 'theme'"
        >
          <i class="fas fa-palette"></i>
          <span>Pengaturan Identitas & Tema Kampus</span>
        </button>
      </div>

      <!-- TAB 1: MONITORING PRODI -->
      <div v-if="activeTab === 'monitoring'" class="tab-content">
        <!-- Stat Cards Grid -->
        <section class="stats-grid">
          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-primary">
              <i class="fas fa-calendar-alt"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Jadwal Kelas Aktif</span>
              <span class="stat-number">1 Kelas</span>
              <span class="stat-sub">Semester Ganjil 2026/2027</span>
            </div>
          </div>

          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-accent">
              <i class="fas fa-map-marked-alt"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Ruangan Geofence</span>
              <span class="stat-number">LAB-KOM-1</span>
              <span class="stat-sub">Gedung B • Radius 35 Meter</span>
            </div>
          </div>

          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-info">
              <i class="fas fa-user-graduate"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Peserta Terdaftar</span>
              <span class="stat-number">1 Mahasiswa</span>
              <span class="stat-sub">KRS Terverifikasi Aktif</span>
            </div>
          </div>

          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-success">
              <i class="fas fa-shield-alt"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Sinkronisasi Kampus</span>
              <span class="stat-number">Aktif & Sinkron</span>
              <span class="stat-sub">AES-256-GCM Terverifikasi</span>
            </div>
          </div>
        </section>

        <!-- Master Schedules Table -->
        <section class="flat-card table-section">
          <div class="section-heading">
            <div>
              <h2 class="section-title">Master Jadwal Kuliah & Geofence</h2>
              <p class="section-desc">Data sinkronisasi dari sistem akademik kampus</p>
            </div>
            <button class="btn btn-secondary btn-sm" @click="refreshData">
              <i class="fas fa-sync-alt"></i>
              <span>Segarkan Data</span>
            </button>
          </div>

          <div class="table-responsive">
            <table class="academic-table">
              <thead>
                <tr>
                  <th>Mata Kuliah</th>
                  <th>Dosen Pengampu</th>
                  <th>Ruangan & Koordinat</th>
                  <th>Jadwal</th>
                  <th>Peserta KRS</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td>
                    <div class="font-bold">Pemrograman Sistem Terdistribusi</div>
                    <div class="text-sm text-muted">INF302 • S1 Informatika</div>
                  </td>
                  <td>
                    <div>Dr. Irwan Setiawan, M.Kom.</div>
                    <div class="text-sm font-mono text-muted">NIDN: 198801102015041001</div>
                  </td>
                  <td>
                    <div class="font-semibold">LAB-KOM-1 (Gedung B)</div>
                    <div class="text-sm font-mono text-muted">Lat: 5.201452, Lon: 96.702145 (35m)</div>
                  </td>
                  <td>
                    <span class="badge badge-info">Senin</span>
                    <div class="text-sm text-muted" style="margin-top: 3px;">08:00 - 10:30 WIB</div>
                  </td>
                  <td>
                    <span class="badge badge-success">1 Mahasiswa</span>
                  </td>
                  <td>
                    <span class="badge badge-success">Siap Perkuliahan</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <!-- TAB 2: PENGATURAN IDENTITAS & TEMA KAMPUS -->
      <div v-else class="tab-content">
        <div class="theme-settings-grid">
          <!-- Settings Form -->
          <section class="flat-card settings-card">
            <div class="settings-header">
              <h2 class="section-title">Pengaturan Identitas & Palet Warna</h2>
              <p class="section-desc">
                Kustomisasi warna primer, warna aksen, nama instansi, dan logo kampus untuk menyesuaikan dengan panduan visual institusi Anda.
              </p>
            </div>

            <!-- Success Alert -->
            <div v-if="savedSuccess" class="alert-banner alert-success">
              <i class="fas fa-check-circle"></i>
              <span>Pengaturan identitas kampus berhasil disimpan dan diterapkan ke seluruh sistem.</span>
            </div>

            <form @submit.prevent="handleSaveTheme" class="settings-form">
              <div class="form-group">
                <label>Nama Perguruan Tinggi / Institusi</label>
                <input 
                  type="text" 
                  v-model="formTheme.institutionName" 
                  class="input-field" 
                  required 
                />
              </div>

              <div class="form-group">
                <label>Slogan / Tagline Institusi</label>
                <input 
                  type="text" 
                  v-model="formTheme.institutionTagline" 
                  class="input-field" 
                  required 
                />
              </div>

              <div class="color-pickers-row">
                <div class="form-group color-field">
                  <label>Warna Primer (Hijau Islami)</label>
                  <div class="color-input-wrapper">
                    <input 
                      type="color" 
                      v-model="formTheme.primaryColor" 
                      class="color-picker" 
                    />
                    <input 
                      type="text" 
                      v-model="formTheme.primaryColor" 
                      class="input-field font-mono" 
                    />
                  </div>
                  <small class="text-muted">Rekomendasi Umuslim: #006633 atau #0B6623</small>
                </div>

                <div class="form-group color-field">
                  <label>Warna Aksen (Emas)</label>
                  <div class="color-input-wrapper">
                    <input 
                      type="color" 
                      v-model="formTheme.accentColor" 
                      class="color-picker" 
                    />
                    <input 
                      type="text" 
                      v-model="formTheme.accentColor" 
                      class="input-field font-mono" 
                    />
                  </div>
                  <small class="text-muted">Rekomendasi Umuslim: #D4AF37 atau #C5A059</small>
                </div>
              </div>

              <div class="form-group">
                <label>URL Logo Institusi (Opsional)</label>
                <input 
                  type="url" 
                  v-model="formTheme.logoUrl" 
                  class="input-field" 
                  placeholder="https://contoh.ac.id/logo-kampus.png" 
                />
                <small class="text-muted">Biarkan kosong untuk menggunakan lambang monogram default institusi.</small>
              </div>

              <div class="settings-actions">
                <button type="button" class="btn btn-secondary" @click="handleResetDefault">
                  <i class="fas fa-undo"></i>
                  <span>Kembalikan Default Umuslim</span>
                </button>
                <button type="submit" class="btn btn-primary">
                  <i class="fas fa-save"></i>
                  <span>Simpan Perubahan</span>
                </button>
              </div>
            </form>
          </section>

          <!-- Live Theme Preview -->
          <section class="flat-card preview-card">
            <h3 class="preview-title">Pratinjau Langsung (Live Preview)</h3>
            <p class="preview-desc">Contoh tampilan elemen antarmuka dengan palet yang dipilih:</p>

            <div class="preview-box">
              <div 
                class="preview-header-mock" 
                :style="{ backgroundColor: '#ffffff', borderTop: '3px solid ' + formTheme.accentColor }"
              >
                <div 
                  class="preview-logo-mock" 
                  :style="{ backgroundColor: formTheme.primaryColor, borderColor: formTheme.accentColor }"
                >
                  <i class="fas fa-university text-white"></i>
                </div>
                <div>
                  <div class="preview-name-mock">{{ formTheme.institutionName }}</div>
                  <div class="preview-tag-mock" :style="{ color: formTheme.primaryColor }">
                    {{ formTheme.institutionTagline }}
                  </div>
                </div>
              </div>

              <div class="preview-buttons-mock">
                <button 
                  type="button" 
                  class="btn" 
                  :style="{ backgroundColor: formTheme.primaryColor, color: '#ffffff', borderColor: formTheme.primaryColor }"
                >
                  <i class="fas fa-check"></i>
                  <span>Tombol Primer</span>
                </button>

                <button 
                  type="button" 
                  class="btn" 
                  :style="{ backgroundColor: formTheme.accentColor, color: '#1e293b', borderColor: '#bfa030' }"
                >
                  <i class="fas fa-star"></i>
                  <span>Tombol Aksen</span>
                </button>
              </div>
            </div>
          </section>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import Navbar from '../components/Navbar.vue'
import { useThemeStore } from '../stores/theme'

const themeStore = useThemeStore()
const activeTab = ref('monitoring')
const savedSuccess = ref(false)

const formTheme = ref({
  institutionName: themeStore.institutionName,
  institutionTagline: themeStore.institutionTagline,
  primaryColor: themeStore.primaryColor,
  accentColor: themeStore.accentColor,
  logoUrl: themeStore.logoUrl
})

function handleSaveTheme() {
  themeStore.saveSettings(formTheme.value)
  savedSuccess.value = true
  setTimeout(() => {
    savedSuccess.value = false
  }, 3500)
}

function handleResetDefault() {
  themeStore.resetDefault()
  formTheme.value = {
    institutionName: themeStore.institutionName,
    institutionTagline: themeStore.institutionTagline,
    primaryColor: themeStore.primaryColor,
    accentColor: themeStore.accentColor,
    logoUrl: themeStore.logoUrl
  }
  savedSuccess.value = true
  setTimeout(() => {
    savedSuccess.value = false
  }, 3500)
}

function refreshData() {
  window.location.reload()
}
</script>

<style scoped>
.admin-portal {
  min-height: 100vh;
  background-color: var(--bg-page);
}

.portal-main {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 24px 40px;
}

.tab-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 24px;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 12px;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border-radius: var(--radius-sm);
  background-color: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
}

.tab-btn.active {
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
  border-color: var(--brand-primary-border);
}

/* Tab 1: Stats */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-box {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  background-color: var(--bg-surface-secondary);
}

.text-primary {
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
}

.text-accent {
  color: #bfa030;
  background-color: var(--brand-accent-light);
}

.text-info {
  color: #1e40af;
  background-color: #dbeafe;
}

.text-success {
  color: #166534;
  background-color: #dcfce7;
}

.stat-details {
  display: flex;
  flex-direction: column;
}

.stat-title {
  font-size: 0.775rem;
  color: var(--text-muted);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-number {
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--text-main);
  margin: 2px 0;
}

.stat-sub {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.table-section {
  padding: 24px;
}

.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.section-title {
  font-size: 1.15rem;
  font-weight: 700;
}

.section-desc {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.btn-sm {
  padding: 7px 14px;
  font-size: 0.825rem;
}

.font-bold {
  font-weight: 700;
}

.font-semibold {
  font-weight: 600;
}

.font-mono {
  font-family: monospace;
}

.text-muted {
  color: var(--text-muted);
}

.text-sm {
  font-size: 0.8rem;
}

/* Tab 2: Theme Settings */
.theme-settings-grid {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 24px;
}

@media (max-width: 900px) {
  .theme-settings-grid {
    grid-template-columns: 1fr;
  }
}

.settings-card {
  padding: 28px;
}

.settings-header {
  margin-bottom: 20px;
}

.alert-success {
  display: flex;
  align-items: center;
  gap: 10px;
  background-color: var(--status-success-bg);
  border: 1px solid var(--status-success-border);
  color: var(--status-success-text);
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  margin-bottom: 20px;
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
}

.color-pickers-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.color-input-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
}

.color-picker {
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  width: 44px;
  height: 42px;
  background: transparent;
  border: 1px solid var(--border-medium);
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.color-picker::-webkit-color-swatch {
  border-radius: 4px;
  border: none;
}

.settings-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 10px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

/* Live Preview Card */
.preview-card {
  padding: 28px;
}

.preview-title {
  font-size: 1.05rem;
  font-weight: 700;
}

.preview-desc {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 2px;
  margin-bottom: 20px;
}

.preview-box {
  background-color: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 24px;
}

.preview-header-mock {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-flat);
  margin-bottom: 20px;
}

.preview-logo-mock {
  width: 38px;
  height: 38px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid;
}

.text-white {
  color: #ffffff;
}

.preview-name-mock {
  font-size: 0.95rem;
  font-weight: 800;
  color: var(--text-main);
}

.preview-tag-mock {
  font-size: 0.725rem;
  font-weight: 600;
}

.preview-buttons-mock {
  display: flex;
  gap: 12px;
}
</style>
