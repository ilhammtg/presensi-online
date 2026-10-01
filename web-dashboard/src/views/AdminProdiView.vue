<template>
  <div class="admin-portal">
    <Navbar />

    <main class="portal-main">
      <!-- Loading State -->
      <div v-if="loading" class="loading-container">
        <i class="fas fa-circle-notch fa-spin"></i>
        <span>Memuat data monitoring program studi...</span>
      </div>

      <div v-else-if="error" class="alert-banner alert-danger">
        <i class="fas fa-exclamation-triangle"></i>
        <span>{{ error }}</span>
      </div>

      <template v-else>
        <!-- Prodi Header Banner -->
        <section class="prodi-banner flat-card">
          <div class="banner-left">
            <div class="prodi-badge">
              <i class="fas fa-university"></i>
              <span>Wilayah Program Studi</span>
            </div>
            <h1 class="prodi-title">{{ overview?.prodi_name || 'Program Studi' }}</h1>
            <p class="prodi-sub">
              <span>{{ overview?.faculty_name || 'Fakultas' }}</span>
              <span class="dot-separator">•</span>
              <span>Kode: {{ overview?.prodi_code || '-' }}</span>
              <span class="dot-separator">•</span>
              <span>Semester Ganjil 2026/2027</span>
            </p>
          </div>
          <div class="banner-right">
            <button class="btn btn-secondary btn-sm" @click="fetchData">
              <i class="fas fa-sync-alt" :class="{ 'fa-spin': refreshing }"></i>
              <span>Segarkan Data</span>
            </button>
          </div>
        </section>

        <!-- KPI Cards Grid (Skala Prodi) -->
        <section class="stats-grid">
          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-primary">
              <i class="fas fa-chart-line"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Rata-rata Kehadiran Prodi</span>
              <div class="stat-value-row">
                <span class="stat-number">{{ overview?.avg_attendance_rate || 0 }}%</span>
                <span 
                  class="badge"
                  :class="overview?.avg_attendance_rate >= 75 ? 'badge-success' : 'badge-warning'"
                >
                  {{ overview?.avg_attendance_rate >= 75 ? 'Memenuhi Standar' : 'Perlu Perhatian' }}
                </span>
              </div>
              <span class="stat-sub">Akumulasi seluruh mahasiswa prodi</span>
            </div>
          </div>

          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-accent">
              <i class="fas fa-chalkboard-teacher"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Kepatuhan Dosen Mengajar</span>
              <div class="stat-value-row">
                <span class="stat-number">{{ overview?.teaching_compliance_rate || 0 }}%</span>
                <span class="badge badge-info">Target Minggu Berjalan</span>
              </div>
              <span class="stat-sub">{{ overview?.total_sessions_held || 0 }} sesi perkuliahan terlaksana</span>
            </div>
          </div>

          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-info">
              <i class="fas fa-user-graduate"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Mahasiswa Aktif Terdaftar</span>
              <span class="stat-number">{{ overview?.total_active_students || 0 }} Mahasiswa</span>
              <span class="stat-sub">KRS aktif di wilayah prodi</span>
            </div>
          </div>

          <div class="flat-card stat-box">
            <div class="stat-icon-wrapper text-success">
              <i class="fas fa-calendar-day"></i>
            </div>
            <div class="stat-details">
              <span class="stat-title">Perkuliahan Hari Ini</span>
              <span class="stat-number">{{ overview?.today_classes_scheduled || 0 }} Terjadwal</span>
              <span class="stat-sub">{{ overview?.today_classes_completed || 0 }} sesi selesai diselenggarakan</span>
            </div>
          </div>
        </section>

        <!-- Navigation Tabs for Prodi Scale Monitoring -->
        <div class="tab-bar">
          <button 
            class="tab-btn" 
            :class="{ 'active': activeTab === 'live' }"
            @click="activeTab = 'live'"
          >
            <span class="live-dot-pulse"></span>
            <span>Monitoring Kelas Hari Ini (Live)</span>
            <span class="badge badge-success" style="font-size: 0.72rem; padding: 2px 7px;">{{ liveClasses.length }}</span>
          </button>

          <button 
            class="tab-btn" 
            :class="{ 'active': activeTab === 'lecturers' }"
            @click="activeTab = 'lecturers'"
          >
            <i class="fas fa-user-tie"></i>
            <span>Kepatuhan Perkuliahan Dosen ({{ lecturers.length }})</span>
          </button>

          <button 
            class="tab-btn" 
            :class="{ 'active': activeTab === 'at-risk' }"
            @click="activeTab = 'at-risk'"
          >
            <i class="fas fa-exclamation-circle text-danger"></i>
            <span>Peringatan Dini Mahasiswa Bermasalah ({{ atRiskStudents.length }})</span>
          </button>

          <button 
            class="tab-btn" 
            :class="{ 'active': activeTab === 'classes' }"
            @click="activeTab = 'classes'"
          >
            <i class="fas fa-book-open"></i>
            <span>Daftar Kelas & Jadwal Prodi ({{ classes.length }})</span>
          </button>
        </div>

        <!-- TAB 0: MONITORING KELAS HARI INI (LIVE) -->
        <section 
          v-if="activeTab === 'live'" 
          ref="liveContainerRef"
          class="flat-card table-section live-radar-container"
          :class="{ 'fullscreen-active': isFullscreen }"
        >
          <!-- Fullscreen Banner Bar (Only displayed in Fullscreen) -->
          <div v-if="isFullscreen" class="fullscreen-topbar">
            <div class="fs-brand">
              <span class="fs-live-badge"><span class="pulse-indicator"></span> LIVE BROADCAST</span>
              <h1 class="fs-title">{{ overview?.prodi_name }} — {{ overview?.faculty_name }}</h1>
              <span class="fs-sub font-mono">{{ currentDateStr }} • {{ currentTimeStr }} WIB</span>
            </div>
            <div class="fs-actions">
              <button class="btn btn-secondary btn-sm" @click="toggleFullscreen">
                <i class="fas fa-compress"></i>
                <span>Keluar Layar Penuh (ESC)</span>
              </button>
            </div>
          </div>

          <div class="section-heading live-section-header">
            <div>
              <div class="live-radar-badge">
                <span class="pulse-indicator"></span>
                <span>REAL-TIME CLASS RADAR & LEADERBOARD</span>
              </div>
              <h2 class="section-title">Monitoring Langsung Perkuliahan Hari Ini</h2>
              <p class="section-desc">
                Kelas yang aktif dan memiliki aktivitas presensi terbaru akan otomatis melesat naik ke posisi teratas papan peringkat secara langsung.
              </p>
            </div>

            <div class="live-header-actions">
              <div class="live-filter-group">
                <button 
                  class="pill-filter" 
                  :class="{ active: liveFilter === 'ALL' }" 
                  @click="liveFilter = 'ALL'"
                >
                  Semua ({{ liveClasses.length }})
                </button>
                <button 
                  class="pill-filter live-active" 
                  :class="{ active: liveFilter === 'SEDANG_BERLANGSUNG' }" 
                  @click="liveFilter = 'SEDANG_BERLANGSUNG'"
                >
                  Berlangsung ({{ countLiveStatus('SEDANG_BERLANGSUNG') }})
                </button>
                <button 
                  class="pill-filter live-danger" 
                  :class="{ active: liveFilter === 'TIDAK_MASUK' }" 
                  @click="liveFilter = 'TIDAK_MASUK'"
                >
                  Belum Hadir (>30m) ({{ countLiveStatus('TIDAK_MASUK') }})
                </button>
                <button 
                  class="pill-filter" 
                  :class="{ active: liveFilter === 'SELESAI' }" 
                  @click="liveFilter = 'SELESAI'"
                >
                  Selesai ({{ countLiveStatus('SELESAI') }})
                </button>
              </div>

              <!-- Fullscreen Button -->
              <button class="btn btn-secondary btn-sm btn-fs-toggle" @click="toggleFullscreen" :title="isFullscreen ? 'Keluar dari layar penuh' : 'Buka mode layar penuh untuk monitor ruang prodi'">
                <i :class="isFullscreen ? 'fas fa-compress' : 'fas fa-expand'"></i>
                <span>{{ isFullscreen ? 'Keluar Fullscreen' : 'Layar Penuh (TV Monitor)' }}</span>
              </button>
            </div>
          </div>

          <div v-if="filteredLiveClasses.length === 0" class="empty-state">
            <i class="fas fa-calendar-check text-muted" style="font-size: 2.5rem; margin-bottom: 12px;"></i>
            <p>Tidak ada perkuliahan yang cocok dengan filter saat ini.</p>
          </div>

          <div v-else class="table-responsive">
            <table class="academic-table live-race-table">
              <thead>
                <tr>
                  <th style="width: 70px; text-align: center;">Pos</th>
                  <th>Status & Waktu</th>
                  <th>Mata Kuliah & Ruang</th>
                  <th>Dosen Pengampu</th>
                  <th>Presensi Mahasiswa</th>
                  <th>Persentase Hadir</th>
                  <th>Aktivitas Radar</th>
                </tr>
              </thead>
              <TransitionGroup name="race-list" tag="tbody">
                <tr 
                  v-for="(c, idx) in filteredLiveClasses" 
                  :key="c.schedule_id"
                  :class="{
                    'highlight-live': c.status === 'SEDANG_BERLANGSUNG',
                    'highlight-danger': c.status === 'TIDAK_MASUK',
                    'row-just-bumped': c.justBumped
                  }"
                >
                  <td style="text-align: center;">
                    <div class="rank-badge-box">
                      <span v-if="idx === 0 && c.status === 'SEDANG_BERLANGSUNG'" class="medal-crown" title="Posisi Teratas">👑</span>
                      <span 
                        class="rank-number"
                        :class="{
                          'rank-first': idx === 0 && c.status === 'SEDANG_BERLANGSUNG',
                          'rank-top3': idx < 3 && c.status === 'SEDANG_BERLANGSUNG'
                        }"
                      >
                        #{{ idx + 1 }}
                      </span>
                    </div>
                  </td>
                  <td>
                    <div class="status-cell">
                      <span 
                        class="badge"
                        :class="{
                          'badge-success live-pulse-tag': c.status === 'SEDANG_BERLANGSUNG',
                          'badge-danger': c.status === 'TIDAK_MASUK',
                          'badge-info': c.status === 'SELESAI',
                          'badge-secondary': c.status === 'BELUM_DIMULAI'
                        }"
                      >
                        <i v-if="c.status === 'SEDANG_BERLANGSUNG'" class="fas fa-circle live-dot"></i>
                        <i v-else-if="c.status === 'TIDAK_MASUK'" class="fas fa-exclamation-triangle"></i>
                        <i v-else-if="c.status === 'SELESAI'" class="fas fa-check-circle"></i>
                        <i v-else class="fas fa-clock"></i>
                        {{ formatLiveStatus(c.status) }}
                      </span>
                      <span class="font-mono text-xs text-muted" style="margin-top: 4px;">
                        {{ c.start_time.substring(0, 5) }} - {{ c.end_time.substring(0, 5) }}
                      </span>
                    </div>
                  </td>
                  <td>
                    <div class="font-bold text-main">{{ c.course_name }}</div>
                    <div class="room-chip font-mono text-xs">
                      <i class="fas fa-map-marker-alt"></i> {{ c.room_name }} • {{ c.course_code }}
                    </div>
                  </td>
                  <td>
                    <div class="font-bold">{{ c.lecturer_name }}</div>
                    <div class="text-xs text-muted font-mono">NIDN: {{ c.lecturer_nidn }}</div>
                  </td>
                  <td>
                    <div class="counter-box">
                      <span class="count-val text-success font-bold">{{ c.total_hadir }}</span>
                      <span class="count-slash">/</span>
                      <span class="count-total">{{ c.total_enrolled }} Mhs</span>
                      <span v-if="c.total_izin_sakit > 0" class="badge badge-subtle text-xs" style="margin-left: 6px;">
                        {{ c.total_izin_sakit }} Izin
                      </span>
                      <span v-if="c.justBumped" class="badge badge-rank-up animate-pop">
                        <i class="fas fa-arrow-up"></i> NAIK!
                      </span>
                    </div>
                  </td>
                  <td>
                    <div class="compliance-bar-wrapper">
                      <div class="progress-bar">
                        <div 
                          class="progress-fill" 
                          :style="{ width: `${c.total_enrolled > 0 ? Math.round((c.total_hadir / c.total_enrolled) * 100) : 0}%` }"
                          :class="c.total_hadir / (c.total_enrolled || 1) >= 0.75 ? 'bg-success' : 'bg-warning'"
                        ></div>
                      </div>
                      <span class="text-xs font-bold font-mono">
                        {{ c.total_enrolled > 0 ? Math.round((c.total_hadir / c.total_enrolled) * 100) : 0 }}%
                      </span>
                    </div>
                  </td>
                  <td>
                    <span v-if="c.status === 'SEDANG_BERLANGSUNG'" class="text-success font-bold text-xs live-active-badge">
                      <i class="fas fa-wifi fa-fade"></i> Sesi Berjalan
                    </span>
                    <span v-else-if="c.status === 'TIDAK_MASUK'" class="text-danger font-bold text-xs">
                      <i class="fas fa-bell"></i> Belum Masuk (>30m)
                    </span>
                    <span v-else-if="c.status === 'SELESAI'" class="text-muted text-xs">
                      Selesai diselenggarakan
                    </span>
                    <span v-else class="text-muted text-xs">
                      Menunggu jadwal
                    </span>
                  </td>
                </tr>
              </TransitionGroup>
            </table>
          </div>
        </section>

        <!-- TAB 1: KEPATUHAN PERKULIAHAN DOSEN -->
        <section v-if="activeTab === 'lecturers'" class="flat-card table-section">
          <div class="section-heading">
            <div>
              <h2 class="section-title">Monitoring Kepatuhan Mengajar Dosen</h2>
              <p class="section-desc">
                Pelacakan kepatuhan dosen dalam menyelenggarakan pertemuan perkuliahan sesuai target kalender akademik (16 Pertemuan per MK).
              </p>
            </div>
          </div>

          <div v-if="lecturers.length === 0" class="empty-state">
            <i class="fas fa-info-circle"></i>
            <p>Belum ada dosen pengampu yang terdaftar di program studi ini.</p>
          </div>

          <div v-else class="table-responsive">
            <table class="academic-table">
              <thead>
                <tr>
                  <th>Dosen Pengampu</th>
                  <th>Mata Kuliah Diampu</th>
                  <th>Progres Pertemuan</th>
                  <th>Tingkat Kepatuhan</th>
                  <th>Status Perkuliahan</th>
                  <th>Pertemuan Terakhir</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="l in lecturers" :key="l.lecturer_id">
                  <td>
                    <div class="font-bold">{{ l.lecturer_name }}</div>
                    <div class="text-xs font-mono text-muted">NIDN: {{ l.nidn }}</div>
                  </td>
                  <td>
                    <div class="course-chips">
                      <span v-for="(c, idx) in l.course_names" :key="idx" class="badge badge-subtle">
                        {{ c }}
                      </span>
                    </div>
                  </td>
                  <td>
                    <div class="progress-cell">
                      <span class="font-bold">{{ l.total_sessions_held }}</span>
                      <span class="text-muted"> / {{ l.target_sessions }} Target</span>
                    </div>
                  </td>
                  <td>
                    <div class="compliance-bar-wrapper">
                      <div class="progress-bar">
                        <div 
                          class="progress-fill" 
                          :style="{ width: `${Math.min(l.compliance_percentage, 100)}%` }"
                          :class="l.compliance_percentage >= 75 ? 'bg-success' : 'bg-warning'"
                        ></div>
                      </div>
                      <span class="text-xs font-bold">{{ l.compliance_percentage }}%</span>
                    </div>
                  </td>
                  <td>
                    <span 
                      class="badge"
                      :class="{
                        'badge-success': l.status === 'lancar',
                        'badge-warning': l.status === 'perlu_perhatian',
                        'badge-danger': l.status === 'tertinggal'
                      }"
                    >
                      <i 
                        class="fas"
                        :class="{
                          'fa-check-circle': l.status === 'lancar',
                          'fa-exclamation-triangle': l.status === 'perlu_perhatian',
                          'fa-times-circle': l.status === 'tertinggal'
                        }"
                      ></i>
                      {{ l.status === 'lancar' ? 'Sesuai Jadwal' : l.status === 'perlu_perhatian' ? 'Perlu Perhatian' : 'Tertinggal' }}
                    </span>
                  </td>
                  <td>
                    <span v-if="l.last_session_date" class="text-sm">
                      {{ formatDate(l.last_session_date) }}
                    </span>
                    <span v-else class="text-muted text-sm">Belum ada sesi</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- TAB 2: EARLY WARNING MAHASISWA BERMASALAH (< 75%) -->
        <section v-else-if="activeTab === 'at-risk'" class="flat-card table-section">
          <div class="section-heading">
            <div>
              <h2 class="section-title">Peringatan Dini Mahasiswa Bermasalah (Kehadiran &lt; 75%)</h2>
              <p class="section-desc">
                Daftar mahasiswa dengan tingkat kehadiran di bawah ambang batas minimal 75% yang terancam tidak memenuhi syarat mengikuti Ujian Akhir Semester (UAS).
              </p>
            </div>
            <div class="search-box">
              <i class="fas fa-search"></i>
              <input 
                type="text" 
                v-model="searchAtRisk" 
                placeholder="Cari NIM atau nama mahasiswa..." 
                class="search-input"
              />
            </div>
          </div>

          <div class="alert-banner alert-warning" style="margin-bottom: 20px;">
            <i class="fas fa-bell"></i>
            <div>
              <strong>Perhatian Kaprodi / Penjaminan Mutu:</strong>
              <span> Terdapat {{ filteredAtRiskStudents.length }} mahasiswa di program studi ini yang memerlukan evaluasi akademik atau pemanggilan resmi.</span>
            </div>
          </div>

          <div v-if="filteredAtRiskStudents.length === 0" class="empty-state">
            <i class="fas fa-check-circle text-success" style="font-size: 2rem;"></i>
            <p class="font-bold" style="margin-top: 8px;">Kondisi Akademik Sangat Baik</p>
            <p class="text-muted text-sm">Tidak ditemukan mahasiswa dengan persentase kehadiran di bawah 75%.</p>
          </div>

          <div v-else class="table-responsive">
            <table class="academic-table">
              <thead>
                <tr>
                  <th>NIM & Mahasiswa</th>
                  <th>Mata Kuliah & Dosen</th>
                  <th>Sesi Kelas</th>
                  <th>Hadir</th>
                  <th>Alpa</th>
                  <th>Izin/Sakit</th>
                  <th>% Kehadiran</th>
                  <th>Rekomendasi Tindak Lanjut</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="s in filteredAtRiskStudents" :key="s.student_id + s.course_code">
                  <td>
                    <div class="font-bold">{{ s.student_name }}</div>
                    <div class="text-xs font-mono text-muted">{{ s.student_nim }}</div>
                  </td>
                  <td>
                    <div class="font-semibold">{{ s.course_name }}</div>
                    <div class="text-xs text-muted">{{ s.course_code }} • {{ s.lecturer_name }}</div>
                  </td>
                  <td class="text-center font-bold">{{ s.total_sessions_held }}</td>
                  <td class="text-center text-success font-bold">{{ s.hadir_count }}</td>
                  <td class="text-center text-danger font-bold">{{ s.alpa_count }}</td>
                  <td class="text-center text-info font-bold">{{ s.izin_sakit_count }}</td>
                  <td>
                    <div class="attendance-rate-pill" :class="s.attendance_rate < 50 ? 'rate-critical' : 'rate-warning'">
                      <i class="fas fa-percentage"></i>
                      <span>{{ s.attendance_rate }}%</span>
                    </div>
                  </td>
                  <td>
                    <span 
                      class="badge" 
                      :class="s.attendance_rate < 50 ? 'badge-danger' : 'badge-warning'"
                    >
                      <i class="fas" :class="s.attendance_rate < 50 ? 'fa-user-slash' : 'fa-exclamation-circle'"></i>
                      {{ s.recommendation }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- TAB 3: DAFTAR KELAS & JADWAL PRODI -->
        <section v-else-if="activeTab === 'classes'" class="flat-card table-section">
          <div class="section-heading">
            <div>
              <h2 class="section-title">Master Kelas & Jadwal Program Studi</h2>
              <p class="section-desc">Rekapitulasi seluruh kelas aktif, ruangan, dan statistik kehadiran perkuliahan di wilayah prodi.</p>
            </div>
          </div>

          <div v-if="classes.length === 0" class="empty-state">
            <i class="fas fa-info-circle"></i>
            <p>Belum ada kelas aktif di program studi ini.</p>
          </div>

          <div v-else class="table-responsive">
            <table class="academic-table">
              <thead>
                <tr>
                  <th>Mata Kuliah</th>
                  <th>Jadwal & Waktu</th>
                  <th>Ruangan & Gedung</th>
                  <th>Dosen Pengampu</th>
                  <th>Peserta KRS</th>
                  <th>Sesi Terlaksana</th>
                  <th>Rata-rata Hadir</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="c in classes" :key="c.schedule_id">
                  <td>
                    <div class="font-bold">{{ c.course_name }}</div>
                    <div class="text-xs font-mono text-muted">{{ c.course_code }}</div>
                  </td>
                  <td>
                    <span class="badge badge-info">{{ c.day_name }}</span>
                    <div class="text-xs text-muted" style="margin-top: 3px;">{{ c.time_range }}</div>
                  </td>
                  <td>
                    <div class="text-sm font-semibold">{{ c.room_name }}</div>
                  </td>
                  <td>
                    <div class="text-sm">{{ c.lecturer_name }}</div>
                    <div class="text-xs font-mono text-muted">NIDN: {{ c.lecturer_nidn }}</div>
                  </td>
                  <td class="text-center">
                    <span class="badge badge-success">{{ c.total_enrolled }} Mhs</span>
                  </td>
                  <td class="text-center font-bold">
                    {{ c.completed_sessions }} / 16
                  </td>
                  <td>
                    <span 
                      class="badge"
                      :class="c.avg_attendance_rate >= 75 ? 'badge-success' : 'badge-warning'"
                    >
                      {{ c.avg_attendance_rate }}%
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import Navbar from '../components/Navbar.vue'

const auth = useAuthStore()

const loading = ref(true)
const refreshing = ref(false)
const error = ref('')
const activeTab = ref('live')
const searchAtRisk = ref('')

const overview = ref(null)
const lecturers = ref([])
const atRiskStudents = ref([])
const classes = ref([])
const liveClasses = ref([])
const liveFilter = ref('ALL')
let livePollTimer = null
let clockTimer = null

// Fullscreen State
const liveContainerRef = ref(null)
const isFullscreen = ref(false)
const currentDateStr = ref('')
const currentTimeStr = ref('')

function updateDateTime() {
  const now = new Date()
  currentDateStr.value = now.toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
  const h = String(now.getHours()).padStart(2, '0')
  const m = String(now.getMinutes()).padStart(2, '0')
  const s = String(now.getSeconds()).padStart(2, '0')
  currentTimeStr.value = `${h}:${m}:${s}`
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    if (liveContainerRef.value?.requestFullscreen) {
      liveContainerRef.value.requestFullscreen().catch(err => {
        console.error('Error attempting to enable fullscreen:', err)
      })
    }
  } else {
    if (document.exitFullscreen) {
      document.exitFullscreen()
    }
  }
}

function handleFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

// Track previous state for Dynamic Race Leaderboard
const prevClassStateMap = new Map()

const filteredAtRiskStudents = computed(() => {
  if (!searchAtRisk.value) return atRiskStudents.value
  const q = searchAtRisk.value.toLowerCase()
  return atRiskStudents.value.filter(s => 
    s.student_name.toLowerCase().includes(q) ||
    s.student_nim.toLowerCase().includes(q) ||
    s.course_name.toLowerCase().includes(q)
  )
})

const filteredLiveClasses = computed(() => {
  if (liveFilter.value === 'ALL') return liveClasses.value
  return liveClasses.value.filter(c => c.status === liveFilter.value)
})

function countLiveStatus(status) {
  return liveClasses.value.filter(c => c.status === status).length
}

function formatLiveStatus(status) {
  switch (status) {
    case 'SEDANG_BERLANGSUNG': return 'Sedang Berlangsung'
    case 'TIDAK_MASUK': return 'Dosen Belum Hadir (>30m)'
    case 'SELESAI': return 'Selesai'
    case 'BELUM_DIMULAI': return 'Belum Dimulai'
    default: return status
  }
}

async function fetchLiveToday() {
  try {
    const headers = { Authorization: `Bearer ${auth.token}` }
    const res = await fetch('/v1/prodi/live-today', { headers })
    if (res.ok) {
      const data = await res.json()
      const rawClasses = data.classes || []
      const nowTs = Date.now()

      const processed = rawClasses.map(item => {
        const prev = prevClassStateMap.get(item.schedule_id)
        let activityTime = prev?.activityTime || 0
        let justBumped = false

        if (prev) {
          // If status changed OR attendance count increased: BUMP TO TOP!
          if (prev.status !== item.status || item.total_hadir > prev.total_hadir || item.total_izin_sakit !== prev.total_izin_sakit) {
            activityTime = nowTs
            justBumped = true
          }
        } else {
          // Initial load: give active classes a head start
          if (item.status === 'SEDANG_BERLANGSUNG') {
            activityTime = nowTs + (item.total_hadir * 1000)
          }
        }

        prevClassStateMap.set(item.schedule_id, {
          status: item.status,
          total_hadir: item.total_hadir,
          total_izin_sakit: item.total_izin_sakit,
          activityTime
        })

        return {
          ...item,
          activityTime,
          justBumped
        }
      })

      // Race Leaderboard Dynamic Sorting:
      // Priority 1: SEDANG_BERLANGSUNG with latest activityTime & highest attendance
      // Priority 2: TIDAK_MASUK (needs immediate prodi attention)
      // Priority 3: SELESAI
      // Priority 4: BELUM_DIMULAI
      processed.sort((a, b) => {
        const getScore = (x) => {
          if (x.status === 'SEDANG_BERLANGSUNG') {
            return 10000000000 + (x.activityTime || 0) + (x.total_hadir * 100000)
          }
          if (x.status === 'TIDAK_MASUK') {
            return 5000000000 + (x.activityTime || 0)
          }
          if (x.status === 'SELESAI') {
            return 2000000000 + (x.activityTime || 0)
          }
          return 1000000000 + (x.activityTime || 0)
        }
        return getScore(b) - getScore(a)
      })

      liveClasses.value = processed

      // Remove justBumped highlight after 4 seconds
      setTimeout(() => {
        liveClasses.value.forEach(c => { c.justBumped = false })
      }, 4000)
    }
  } catch (e) {
    console.error('Error fetching live classes:', e)
  }
}

async function fetchData() {
  refreshing.value = true
  error.value = ''
  try {
    const headers = { Authorization: `Bearer ${auth.token}` }

    // Fetch Overview
    const resOverview = await fetch('/v1/prodi/overview', { headers })
    if (!resOverview.ok) {
      const err = await resOverview.json()
      throw new Error(err.error || 'Gagal memuat ringkasan program studi')
    }
    overview.value = await resOverview.json()

    // Fetch Lecturers Compliance
    const resLecturers = await fetch('/v1/prodi/lecturers-compliance', { headers })
    if (resLecturers.ok) {
      const data = await resLecturers.json()
      lecturers.value = data.lecturers || []
    }

    // Fetch Students At Risk
    const resAtRisk = await fetch('/v1/prodi/students-at-risk', { headers })
    if (resAtRisk.ok) {
      const data = await resAtRisk.json()
      atRiskStudents.value = data.students || []
    }

    // Fetch Classes Summary
    const resClasses = await fetch('/v1/prodi/classes', { headers })
    if (resClasses.ok) {
      const data = await resClasses.json()
      classes.value = data.classes || []
    }

    // Fetch Live Today Classes
    await fetchLiveToday()
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function formatDate(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  return d.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

onMounted(() => {
  fetchData()
  updateDateTime()
  clockTimer = setInterval(updateDateTime, 1000)

  // Listen to fullscreen changes (user pressing ESC or button)
  document.addEventListener('fullscreenchange', handleFullscreenChange)

  // Poll live classes every 10 seconds for real-time race feel
  livePollTimer = setInterval(() => {
    fetchLiveToday()
  }, 10000)
})

onUnmounted(() => {
  if (livePollTimer) clearInterval(livePollTimer)
  if (clockTimer) clearInterval(clockTimer)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
})
</script>

<style scoped>
.admin-portal {
  min-height: 100vh;
  background-color: var(--bg-page);
}

.portal-main {
  max-width: 1320px;
  margin: 0 auto;
  padding: 24px 24px 60px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px 0;
  color: var(--text-muted);
}

.loading-container i {
  font-size: 2rem;
  color: var(--brand-primary);
}

/* Alert Banner */
.alert-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 18px;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
}

.alert-danger {
  background: #fef2f2;
  border: 1px solid #fecaca;
  color: #991b1b;
}

.alert-warning {
  background: #fefce8;
  border: 1px solid #fef08a;
  color: #854d0e;
}

/* Prodi Header Banner */
.prodi-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 28px;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
  border-left: 5px solid var(--brand-primary);
}

.prodi-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: var(--brand-primary-light);
  color: var(--brand-primary);
  font-size: 0.75rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  margin-bottom: 8px;
}

.prodi-title {
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--text-main);
  margin: 0;
}

.prodi-sub {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin: 6px 0 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.dot-separator {
  color: var(--border-subtle);
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
  margin-bottom: 28px;
}

.stat-box {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 20px;
}

.stat-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  background-color: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.stat-icon-wrapper.text-primary {
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
}

.stat-icon-wrapper.text-accent {
  color: #b45309;
  background-color: #fef3c7;
}

.stat-icon-wrapper.text-info {
  color: #0284c7;
  background-color: #e0f2fe;
}

.stat-icon-wrapper.text-success {
  color: #16a34a;
  background-color: #dcfce7;
}

.stat-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-title {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.stat-value-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-number {
  font-size: 1.4rem;
  font-weight: 800;
  color: var(--text-main);
}

.stat-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
}

/* Tab Bar */
.tab-bar {
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 24px;
  overflow-x: auto;
}

.tab-btn {
  background: transparent;
  border: none;
  padding: 12px 18px;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border-bottom: 3px solid transparent;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.tab-btn:hover {
  color: var(--text-main);
}

.tab-btn.active {
  color: var(--brand-primary);
  border-bottom-color: var(--brand-primary);
}

/* Tables Section */
.table-section {
  padding: 24px;
}

.section-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
  gap: 16px;
  flex-wrap: wrap;
}

.section-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

.section-desc {
  font-size: 0.825rem;
  color: var(--text-muted);
  margin: 4px 0 0;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
}

.search-box i {
  color: var(--text-muted);
  font-size: 0.85rem;
}

.search-input {
  border: none;
  background: transparent;
  outline: none;
  font-size: 0.85rem;
  width: 220px;
}

.academic-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.academic-table th {
  background-color: var(--bg-surface-secondary);
  color: var(--text-muted);
  font-weight: 700;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid var(--border-subtle);
}

.academic-table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-subtle);
  color: var(--text-main);
  vertical-align: middle;
}

.academic-table tr:hover td {
  background-color: rgba(0, 0, 0, 0.015);
}

.course-chips {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.progress-cell {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.compliance-bar-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.progress-bar {
  width: 90px;
  height: 7px;
  background-color: var(--border-subtle);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: var(--radius-full);
}

.bg-success {
  background-color: #16a34a;
}

.bg-warning {
  background-color: #eab308;
}

.attendance-rate-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  font-size: 0.8rem;
  font-weight: 700;
}

.rate-critical {
  background-color: #fee2e2;
  color: #dc2626;
  border: 1px solid #fca5a5;
}

.rate-warning {
  background-color: #fef9c3;
  color: #854d0e;
  border: 1px solid #fde047;
}

.empty-state {
  text-align: center;
  padding: 48px 16px;
  color: var(--text-muted);
}

/* Live Class Monitoring Styles */
.live-dot-pulse {
  display: inline-block;
  width: 9px;
  height: 9px;
  background-color: #22c55e;
  border-radius: 50%;
  box-shadow: 0 0 0 rgba(34, 197, 94, 0.4);
  animation: pulse-green 1.5s infinite;
  margin-right: 6px;
}

@keyframes pulse-green {
  0% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.7);
  }
  70% {
    transform: scale(1);
    box-shadow: 0 0 0 6px rgba(34, 197, 94, 0);
  }
  100% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(34, 197, 94, 0);
  }
}

.live-section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 24px;
}

.live-radar-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  background-color: #f0fdf4;
  color: #15803d;
  border: 1px solid #bbf7d0;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  margin-bottom: 6px;
}

.pulse-indicator {
  width: 6px;
  height: 6px;
  background-color: #16a34a;
  border-radius: 50%;
  animation: pulse-green 1.2s infinite;
}

.live-filter-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pill-filter {
  padding: 6px 14px;
  border-radius: 20px;
  border: 1px solid #e2e8f0;
  background-color: #ffffff;
  font-size: 0.82rem;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s ease;
}

.pill-filter:hover {
  background-color: #f8fafc;
  color: #1e293b;
}

.pill-filter.active {
  background-color: #0f172a;
  color: #ffffff;
  border-color: #0f172a;
}

.pill-filter.live-active.active {
  background-color: #16a34a;
  border-color: #16a34a;
  color: #ffffff;
}

.pill-filter.live-danger.active {
  background-color: #dc2626;
  border-color: #dc2626;
  color: #ffffff;
}

.status-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.live-pulse-tag {
  background-color: #dcfce7;
  color: #15803d;
  border: 1px solid #86efac;
}

.live-dot {
  font-size: 0.5rem;
  margin-right: 4px;
  animation: pulse-green 1.5s infinite;
}

.room-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
  margin-top: 3px;
}

.counter-box {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.9rem;
}

.count-slash {
  color: var(--text-muted);
}

.highlight-live {
  background-color: rgba(34, 197, 94, 0.04);
}

.highlight-danger {
  background-color: rgba(239, 68, 68, 0.04);
}

/* Fullscreen Mode & Header Actions */
.live-header-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.btn-fs-toggle {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background-color: #0f172a;
  color: #ffffff;
  border-color: #0f172a;
  font-weight: 700;
  transition: all 0.2s ease;
}

.btn-fs-toggle:hover {
  background-color: var(--brand-primary);
  border-color: var(--brand-primary);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* Fullscreen Active Container */
.live-radar-container.fullscreen-active {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  z-index: 99999;
  border-radius: 0;
  margin: 0;
  padding: 32px 48px;
  background: #ffffff;
  overflow-y: auto;
  box-sizing: border-box;
}

.fullscreen-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 2px solid var(--border-subtle);
}

.fs-brand {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.fs-live-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  width: fit-content;
  padding: 4px 12px;
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.fs-title {
  font-size: 1.85rem;
  font-weight: 900;
  color: var(--text-main);
  margin: 0;
  letter-spacing: -0.02em;
}

.fs-sub {
  font-size: 0.95rem;
  color: var(--brand-primary);
  font-weight: 700;
}

/* Race Leaderboard Table & Badges */
.live-race-table {
  position: relative;
}

.rank-badge-box {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.medal-crown {
  position: absolute;
  top: -14px;
  font-size: 0.9rem;
  animation: float-crown 2s infinite ease-in-out;
}

@keyframes float-crown {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

.rank-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  border-radius: 50%;
  font-size: 0.85rem;
  font-weight: 800;
  font-family: monospace;
  background-color: var(--bg-surface-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
}

.rank-first {
  background: linear-gradient(135deg, #fef08a 0%, #eab308 100%);
  color: #713f12;
  border-color: #ca8a04;
  box-shadow: 0 0 12px rgba(234, 179, 8, 0.4);
  font-size: 0.95rem;
}

.rank-top3 {
  background: linear-gradient(135deg, #dcfce7 0%, #22c55e 100%);
  color: #064e3b;
  border-color: #16a34a;
}

/* FLIP Transition: Moving rows smoothly up when status/attendance changes */
.race-list-move {
  transition: transform 0.85s cubic-bezier(0.25, 1, 0.5, 1);
}

.race-list-enter-active,
.race-list-leave-active {
  transition: all 0.6s ease;
}

.race-list-enter-from,
.race-list-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

/* Rank-up Bump animation */
.row-just-bumped {
  animation: flash-row-bump 3.5s ease-out;
}

@keyframes flash-row-bump {
  0% {
    background-color: rgba(234, 179, 8, 0.28);
    box-shadow: inset 0 0 16px rgba(234, 179, 8, 0.5);
  }
  30% {
    background-color: rgba(34, 197, 94, 0.2);
  }
  100% {
    background-color: transparent;
  }
}

.badge-rank-up {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background-color: #16a34a;
  color: #ffffff;
  font-size: 0.7rem;
  font-weight: 800;
  padding: 2px 7px;
  border-radius: 12px;
  box-shadow: 0 0 10px rgba(22, 163, 74, 0.5);
}

.animate-pop {
  animation: pop-in 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes pop-in {
  0% { transform: scale(0.6); opacity: 0; }
  100% { transform: scale(1); opacity: 1; }
}

.live-active-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
</style>


