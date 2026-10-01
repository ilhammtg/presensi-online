<template>
  <div class="admin-portal">
    <Navbar />

    <main class="portal-main">
      <!-- Superadmin Header Banner -->
      <section class="system-banner flat-card">
        <div class="banner-left">
          <div class="system-badge">
            <i class="fas fa-user-shield"></i>
            <span>Hak Akses Superadmin</span>
          </div>
          <h1 class="system-title">Administrasi Sistem & Pengaturan Institusi</h1>
          <p class="system-sub">
            Pusat konfigurasi global, manajemen hak akses, dan tata kelola identitas visual kampus.
          </p>
        </div>
        <div class="banner-right">
          <button class="btn btn-secondary btn-sm" @click="fetchSystemData">
            <i class="fas fa-sync-alt" :class="{ 'fa-spin': refreshing }"></i>
            <span>Segarkan Status Sistem</span>
          </button>
        </div>
      </section>

      <!-- Global System Stats Grid -->
      <section class="stats-grid" v-if="stats">
        <div class="flat-card stat-box">
          <div class="stat-icon-wrapper text-primary">
            <i class="fas fa-landmark"></i>
          </div>
          <div class="stat-details">
            <span class="stat-title">Fakultas & Prodi</span>
            <span class="stat-number">{{ stats.total_faculties }} Fak • {{ stats.total_study_programs }} Prodi</span>
            <span class="stat-sub">Terdaftar di institusi</span>
          </div>
        </div>

        <div class="flat-card stat-box">
          <div class="stat-icon-wrapper text-accent">
            <i class="fas fa-users-cog"></i>
          </div>
          <div class="stat-details">
            <span class="stat-title">Dosen & Mahasiswa</span>
            <span class="stat-number">{{ stats.total_lecturers }} Dosen • {{ stats.total_students }} Mhs</span>
            <span class="stat-sub">Pengguna aktif sistem</span>
          </div>
        </div>

        <div class="flat-card stat-box">
          <div class="stat-icon-wrapper text-info">
            <i class="fas fa-calendar-check"></i>
          </div>
          <div class="stat-details">
            <span class="stat-title">Total Jadwal & Sesi</span>
            <span class="stat-number">{{ stats.total_schedules }} Kelas • {{ stats.total_sessions_held }} Sesi</span>
            <span class="stat-sub">Semester berjalan</span>
          </div>
        </div>

        <div class="flat-card stat-box">
          <div class="stat-icon-wrapper text-success">
            <i class="fas fa-globe-asia"></i>
          </div>
          <div class="stat-details">
            <span class="stat-title">Rata-rata Kehadiran Global</span>
            <span class="stat-number">{{ stats.global_avg_attendance }}%</span>
            <span class="stat-sub">Seluruh program studi</span>
          </div>
        </div>
      </section>

      <!-- System Admin Tab Bar -->
      <div class="tab-bar">
        <button 
          class="tab-btn" 
          :class="{ 'active': activeTab === 'siakad' }"
          @click="activeTab = 'siakad'"
        >
          <i class="fas fa-network-wired"></i>
          <span>Integrasi API SIAKAD & Alur Data</span>
        </button>

        <button 
          class="tab-btn" 
          :class="{ 'active': activeTab === 'geofence' }"
          @click="activeTab = 'geofence'"
        >
          <i class="fas fa-map-marked-alt"></i>
          <span>Titik Geofence Kampus ({{ campusLocations.length }})</span>
        </button>

        <button 
          class="tab-btn" 
          :class="{ 'active': activeTab === 'theme' }"
          @click="activeTab = 'theme'"
        >
          <i class="fas fa-palette"></i>
          <span>Pengaturan Identitas & Tema</span>
        </button>

        <button 
          class="tab-btn" 
          :class="{ 'active': activeTab === 'users' }"
          @click="activeTab = 'users'"
        >
          <i class="fas fa-users"></i>
          <span>Manajemen Pengguna ({{ users.length }})</span>
        </button>

        <button 
          class="tab-btn" 
          :class="{ 'active': activeTab === 'programs' }"
          @click="activeTab = 'programs'"
        >
          <i class="fas fa-sitemap"></i>
          <span>Master Fakultas & Prodi ({{ studyPrograms.length }})</span>
        </button>
      </div>

      <!-- TAB SIAKAD: INTEGRASI API SIAKAD & ALUR DATA -->
      <div v-if="activeTab === 'siakad'" class="siakad-tab-container">
        <!-- Visual Architecture Flowchart -->
        <section class="flat-card architecture-card">
          <div class="card-top-header">
            <div>
              <div class="tech-tag font-mono">
                <i class="fas fa-project-diagram"></i> ARSITEKTUR INTEGRASI DATA KAMPUS
              </div>
              <h2 class="section-title">Visualisasi Alur Data Presensi & SIAKAD</h2>
              <p class="section-desc">
                Arsitektur terpisah antara Sistem Akademik (SIAKAD) Kampus dan Sistem Presensi Online yang terhubung melalui enkripsi AES-256-GCM.
              </p>
            </div>
            <button class="btn btn-primary btn-sm" @click="handleTriggerSync" :disabled="syncing">
              <i class="fas fa-play" :class="{ 'fa-spin': syncing }"></i>
              <span>{{ syncing ? 'Sinkronisasi Berjalan...' : 'Sinkronkan Sekarang' }}</span>
            </button>
          </div>

          <div v-if="syncMessage" class="alert-banner alert-success" style="margin-top: 16px;">
            <i class="fas fa-check-circle"></i>
            <span>{{ syncMessage }}</span>
          </div>

          <!-- Flowchart Nodes -->
          <div class="flowchart-container">
            <div class="flow-node">
              <div class="node-icon bg-blue-subtle text-primary">
                <i class="fas fa-server"></i>
              </div>
              <div class="node-content">
                <span class="node-badge">Sumber Data</span>
                <h4 class="node-title">SIAKAD Kampus</h4>
                <p class="node-desc">Fakultas, Prodi, Dosen, Mahasiswa, Jadwal, & KRS</p>
                <div class="node-meta font-mono text-xs">{{ apiConfig.campus_api_url }}</div>
              </div>
            </div>

            <div class="flow-connector">
              <div class="connector-line"></div>
              <div class="connector-badge">
                <i class="fas fa-lock text-accent"></i>
                <span>AES-256-GCM</span>
              </div>
            </div>

            <div class="flow-node">
              <div class="node-icon bg-yellow-subtle text-accent">
                <i class="fas fa-cogs"></i>
              </div>
              <div class="node-content">
                <span class="node-badge">Worker Daemon</span>
                <h4 class="node-title">Sync Worker Service</h4>
                <p class="node-desc">Decrypts payload & writes relational models</p>
                <div class="node-meta font-mono text-xs">Cron: {{ apiConfig.campus_sync_cron || '0 3 * * *' }}</div>
              </div>
            </div>

            <div class="flow-connector">
              <div class="connector-line"></div>
              <div class="connector-badge">
                <i class="fas fa-database text-success"></i>
                <span>Relational Ingestion</span>
              </div>
            </div>

            <div class="flow-node">
              <div class="node-icon bg-green-subtle text-success">
                <i class="fas fa-database"></i>
              </div>
              <div class="node-content">
                <span class="node-badge">Database Storage</span>
                <h4 class="node-title">PostgreSQL & Redis</h4>
                <p class="node-desc">Postgres relasional & Redis TOTP Rolling QR</p>
                <div class="node-meta font-mono text-xs">2 Fak • 2 Prodi • 217 Jadwal</div>
              </div>
            </div>

            <div class="flow-connector">
              <div class="connector-line"></div>
              <div class="connector-badge">
                <i class="fas fa-mobile-alt text-info"></i>
                <span>REST & WebSocket</span>
              </div>
            </div>

            <div class="flow-node">
              <div class="node-icon bg-purple-subtle text-info">
                <i class="fas fa-qrcode"></i>
              </div>
              <div class="node-content">
                <span class="node-badge">End User Apps</span>
                <h4 class="node-title">Dashboard & Mobile App</h4>
                <p class="node-desc">Dosen buka sesi QR & Mahasiswa scan via GPS</p>
                <div class="node-meta font-mono text-xs">Multi-Point Geofence Active</div>
              </div>
            </div>
          </div>
        </section>

        <!-- API Config Form Card -->
        <section class="flat-card settings-card" style="margin-top: 24px;">
          <div class="settings-header">
            <h2 class="section-title">Konfigurasi Endpoint SIAKAD Kampus</h2>
            <p class="section-desc">
              Pengaturan ini dapat disesuaikan tanpa perlu mengubah source code atau restart container.
            </p>
          </div>

          <div v-if="apiConfigSaved" class="alert-banner alert-success" style="margin-bottom: 20px;">
            <i class="fas fa-check-circle"></i>
            <span>Konfigurasi endpoint SIAKAD berhasil disimpan ke database.</span>
          </div>

          <form @submit.prevent="handleSaveApiConfig" class="settings-form">
            <div class="form-row-2">
              <div class="form-group">
                <label class="form-label">SIAKAD API Endpoint URL</label>
                <input 
                  type="text" 
                  v-model="apiConfig.campus_api_url" 
                  class="form-control font-mono"
                  placeholder="http://mock-campus-api:9001/campus/sync" 
                  required 
                />
                <span class="form-hint">Endpoint internal Docker atau server publik kampus</span>
              </div>

              <div class="form-group">
                <label class="form-label">API Secret Key (AES-256 Auth / Bearer)</label>
                <input 
                  type="password" 
                  v-model="apiConfig.campus_api_key" 
                  class="form-control font-mono"
                  placeholder="••••••••••••••••••••••••" 
                  required 
                />
                <span class="form-hint">Kunci simetris pertukaran data SIAKAD</span>
              </div>
            </div>

            <div class="form-row-3">
              <div class="form-group">
                <label class="form-label">Jadwal Cron Sinkronisasi Otomatis</label>
                <input 
                  type="text" 
                  v-model="apiConfig.campus_sync_cron" 
                  class="form-control font-mono"
                  placeholder="0 3 * * *" 
                  required 
                />
                <span class="form-hint">Format cron (contoh: 0 3 * * * = Tiap jam 03.00)</span>
              </div>

              <div class="form-group">
                <label class="form-label">Mode Validasi Geofence</label>
                <select v-model="apiConfig.geofence_mode" class="form-control">
                  <option value="CAMPUS_AND_ROOM">Multi-Point Kampus & Ruangan (Rekomendasi)</option>
                  <option value="ROOM_ONLY">Hanya Ruangan Tertentu (Radius Ketat)</option>
                  <option value="CAMPUS_ONLY">Hanya Wilayah Kampus</option>
                </select>
                <span class="form-hint">Aturan pengecekan radius GPS saat mahasiswa scan</span>
              </div>

              <div class="form-group">
                <label class="form-label">Toleransi Keterlambatan (Menit)</label>
                <input 
                  type="number" 
                  v-model.number="apiConfig.max_tolerance_minutes" 
                  class="form-control font-mono"
                  min="5" 
                  max="60" 
                  required 
                />
                <span class="form-hint">Lewat dari waktu ini dinilai Terlambat</span>
              </div>
            </div>

            <div class="settings-actions">
              <button type="submit" class="btn btn-primary" :disabled="savingConfig">
                <i class="fas fa-save"></i>
                <span>{{ savingConfig ? 'Menyimpan...' : 'Simpan Konfigurasi SIAKAD' }}</span>
              </button>
            </div>
          </form>
        </section>
      </div>

      <!-- TAB GEOFENCE: TITIK GEOFENCE KAMPUS (MULTI-POINT) -->
      <div v-if="activeTab === 'geofence'" class="geofence-tab-container">
        <section class="flat-card table-section">
          <div class="section-heading" style="display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 16px;">
            <div>
              <h2 class="section-title">Manajemen Titik Geofence Kampus</h2>
              <p class="section-desc">
                Titik koordinat resmi universitas yang diakui oleh sistem presensi. Mahasiswa dapat melakukan presensi jika berada di dalam salah satu radius lokasi aktif ini.
              </p>
            </div>
            <button class="btn btn-primary btn-sm" @click="openAddLocationModal">
              <i class="fas fa-plus"></i>
              <span>Tambah Titik Kampus</span>
            </button>
          </div>

          <div v-if="locationSavedMessage" class="alert-banner alert-success" style="margin-bottom: 20px;">
            <i class="fas fa-check-circle"></i>
            <span>{{ locationSavedMessage }}</span>
          </div>

          <div v-if="campusLocations.length === 0" class="empty-state">
            <i class="fas fa-map-marker-alt" style="font-size: 2.5rem; margin-bottom: 12px;"></i>
            <p>Belum ada titik lokasi kampus yang dikonfigurasi.</p>
          </div>

          <div v-else class="table-responsive">
            <table class="academic-table">
              <thead>
                <tr>
                  <th>Nama Lokasi & Kampus</th>
                  <th>Koordinat GPS (Lat, Long)</th>
                  <th>Radius Toleransi</th>
                  <th>Status Lokasi</th>
                  <th>Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="loc in campusLocations" :key="loc.id">
                  <td>
                    <div class="font-bold text-main">{{ loc.name }}</div>
                    <div class="text-xs text-muted">{{ loc.description || '-' }}</div>
                  </td>
                  <td>
                    <div class="font-mono text-xs">
                      <span class="badge badge-subtle">
                        <i class="fas fa-crosshairs"></i> {{ loc.latitude.toFixed(6) }}, {{ loc.longitude.toFixed(6) }}
                      </span>
                    </div>
                  </td>
                  <td>
                    <div class="font-bold font-mono">
                      {{ loc.radius_meters }} Meter
                    </div>
                  </td>
                  <td>
                    <span class="badge" :class="loc.is_active ? 'badge-success' : 'badge-secondary'">
                      {{ loc.is_active ? 'Aktif Digunakan' : 'Nonaktif' }}
                    </span>
                  </td>
                  <td>
                    <div class="table-actions">
                      <button class="btn-icon text-primary" title="Edit Lokasi" @click="openEditLocationModal(loc)">
                        <i class="fas fa-edit"></i>
                      </button>
                      <button class="btn-icon text-danger" title="Hapus Lokasi" @click="handleDeleteLocation(loc.id)">
                        <i class="fas fa-trash-alt"></i>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- Form Tambah / Edit Lokasi Kampus Modal / Box -->
        <section v-if="showLocationForm" class="flat-card settings-card" style="margin-top: 24px;">
          <div class="settings-header">
            <h3 class="section-title">{{ editingLocationId ? 'Edit Titik Lokasi Kampus' : 'Tambah Titik Lokasi Kampus Baru' }}</h3>
            <p class="section-desc">Masukkan koordinat GPS akurat dan radius toleransi geofence.</p>
          </div>

          <form @submit.prevent="handleSaveLocation" class="settings-form">
            <div class="form-row-2">
              <div class="form-group">
                <label class="form-label">Nama Titik / Kampus</label>
                <input 
                  type="text" 
                  v-model="locationForm.name" 
                  class="form-control" 
                  placeholder="Contoh: Kampus Induk Matang Glumpang Dua"
                  required 
                />
              </div>

              <div class="form-group">
                <label class="form-label">Keterangan / Alamat</label>
                <input 
                  type="text" 
                  v-model="locationForm.description" 
                  class="form-control" 
                  placeholder="Gedung Rektorat & Fakultas Ilmu Komputer"
                />
              </div>
            </div>

            <div class="form-row-3">
              <div class="form-group">
                <label class="form-label">Latitude</label>
                <input 
                  type="number" 
                  step="any"
                  v-model.number="locationForm.latitude" 
                  class="form-control font-mono" 
                  placeholder="5.193730"
                  required 
                />
              </div>

              <div class="form-group">
                <label class="form-label">Longitude</label>
                <input 
                  type="number" 
                  step="any"
                  v-model.number="locationForm.longitude" 
                  class="form-control font-mono" 
                  placeholder="96.787492"
                  required 
                />
              </div>

              <div class="form-group">
                <label class="form-label">Radius Presensi (Meter)</label>
                <input 
                  type="number" 
                  v-model.number="locationForm.radius_meters" 
                  class="form-control font-mono" 
                  min="10"
                  max="1000"
                  placeholder="80"
                  required 
                />
              </div>
            </div>

            <div class="form-group">
              <label class="checkbox-label" style="display: flex; align-items: center; gap: 8px; cursor: pointer;">
                <input type="checkbox" v-model="locationForm.is_active" />
                <span class="font-bold">Aktifkan titik lokasi ini untuk validasi presensi</span>
              </label>
            </div>

            <div class="settings-actions">
              <button type="submit" class="btn btn-primary" :disabled="savingLocation">
                <i class="fas fa-check"></i>
                <span>{{ savingLocation ? 'Menyimpan...' : 'Simpan Titik Lokasi' }}</span>
              </button>
              <button type="button" class="btn btn-secondary" @click="showLocationForm = false">
                Batal
              </button>
            </div>
          </form>
        </section>
      </div>

      <!-- TAB 1: PENGATURAN IDENTITAS & TEMA KAMPUS -->
      <div v-if="activeTab === 'theme'" class="theme-settings-grid">
        <section class="flat-card settings-card">
          <div class="settings-header">
            <h2 class="section-title">Konfigurasi Visual & Identitas Kampus</h2>
            <p class="section-desc">
              Pengaturan ini akan diterapkan ke seluruh portal perkuliahan dan dashboard institusi.
            </p>
          </div>

          <div v-if="themeSavedSuccess" class="alert-banner alert-success" style="margin-bottom: 20px;">
            <i class="fas fa-check-circle"></i>
            <span>Konfigurasi tema berhasil disimpan dan diterapkan ke seluruh sistem.</span>
          </div>

          <form @submit.prevent="handleSaveTheme" class="settings-form">
            <div class="form-group">
              <label class="form-label">Nama Perguruan Tinggi / Institusi</label>
              <input 
                type="text" 
                v-model="themeForm.campus_name" 
                class="form-control"
                placeholder="Universitas Almuslim" 
                required 
              />
            </div>

            <div class="form-group">
              <label class="form-label">Tagline / Sub-Identitas</label>
              <input 
                type="text" 
                v-model="themeForm.campus_tagline" 
                class="form-control"
                placeholder="Sistem Informasi Presensi & QR Dinamis" 
              />
            </div>

            <div class="color-pickers-row">
              <div class="form-group color-item">
                <label class="form-label">Warna Primer Kampus</label>
                <div class="color-input-wrapper">
                  <input type="color" v-model="themeForm.primary_color" class="color-picker" />
                  <input type="text" v-model="themeForm.primary_color" class="form-control font-mono" />
                </div>
                <span class="text-xs text-muted">Contoh Umuslim: #006633 atau #0B6623</span>
              </div>

              <div class="form-group color-item">
                <label class="form-label">Warna Aksen / Emas</label>
                <div class="color-input-wrapper">
                  <input type="color" v-model="themeForm.accent_color" class="color-picker" />
                  <input type="text" v-model="themeForm.accent_color" class="form-control font-mono" />
                </div>
                <span class="text-xs text-muted">Contoh Umuslim: #D4AF37 atau #C5A059</span>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">URL Logo Kampus (Opsional)</label>
              <input 
                type="text" 
                v-model="themeForm.logo_url" 
                class="form-control"
                placeholder="https://... atau path gambar" 
              />
            </div>

            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="savingTheme">
                <i v-if="savingTheme" class="fas fa-circle-notch fa-spin"></i>
                <i v-else class="fas fa-save"></i>
                <span>Simpan Konfigurasi Tema</span>
              </button>
            </div>
          </form>
        </section>

        <!-- Live Theme Preview Box -->
        <section class="flat-card preview-card">
          <div class="settings-header">
            <h2 class="section-title">Pratinjau Visual (Live Preview)</h2>
            <p class="section-desc">Tampilan antarmuka dengan palet warna yang dipilih.</p>
          </div>

          <div class="preview-box">
            <div class="mock-nav" :style="{ borderTopColor: themeForm.accent_color }">
              <div class="mock-brand">
                <div class="mock-icon" :style="{ backgroundColor: themeForm.primary_color }">
                  <i class="fas fa-university"></i>
                </div>
                <div>
                  <div class="mock-title">{{ themeForm.campus_name || 'Nama Kampus' }}</div>
                  <div class="mock-sub">{{ themeForm.campus_tagline || 'Tagline Kampus' }}</div>
                </div>
              </div>
            </div>

            <div class="mock-body">
              <div class="mock-card">
                <div class="mock-card-title">Mata Kuliah Pratinjau</div>
                <div class="mock-btn" :style="{ backgroundColor: themeForm.primary_color }">
                  <i class="fas fa-qrcode"></i> Buka Presensi
                </div>
                <div class="mock-badge" :style="{ backgroundColor: themeForm.accent_color + '22', color: themeForm.accent_color }">
                  Aksen Emas
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- TAB 2: MANAJEMEN PENGGUNA -->
      <section v-else-if="activeTab === 'users'" class="flat-card table-section">
        <div class="section-heading">
          <div>
            <h2 class="section-title">Manajemen Akun Pengguna</h2>
            <p class="section-desc">Daftar akun dosen, mahasiswa, admin prodi, dan administrator sistem.</p>
          </div>
          <div class="filters-row">
            <select v-model="userRoleFilter" class="form-select" @change="fetchUsers">
              <option value="">Semua Peran</option>
              <option value="dosen">Dosen</option>
              <option value="mahasiswa">Mahasiswa</option>
              <option value="admin_prodi">Admin Prodi</option>
              <option value="superadmin">Superadmin</option>
            </select>
          </div>
        </div>

        <div class="table-responsive">
          <table class="academic-table">
            <thead>
              <tr>
                <th>NIM / NIDN</th>
                <th>Nama Lengkap</th>
                <th>Email Kampus</th>
                <th>Peran</th>
                <th>Program Studi</th>
                <th>Status Perangkat</th>
                <th>Status Akun</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id">
                <td class="font-mono text-sm font-semibold">{{ u.external_id }}</td>
                <td class="font-bold">{{ u.name }}</td>
                <td class="text-sm text-muted">{{ u.email }}</td>
                <td>
                  <span 
                    class="badge"
                    :class="{
                      'badge-info': u.role === 'dosen',
                      'badge-subtle': u.role === 'mahasiswa',
                      'badge-warning': u.role === 'admin_prodi',
                      'badge-success': u.role === 'superadmin'
                    }"
                  >
                    {{ u.role.toUpperCase() }}
                  </span>
                </td>
                <td class="text-sm">{{ u.prodi_name || '-' }}</td>
                <td>
                  <span v-if="u.device_id" class="text-xs badge badge-success">
                    <i class="fas fa-mobile-alt"></i> Terdaftar
                  </span>
                  <span v-else class="text-xs text-muted">Belum Terdaftar</span>
                </td>
                <td>
                  <span class="badge" :class="u.is_active ? 'badge-success' : 'badge-danger'">
                    {{ u.is_active ? 'Aktif' : 'Non-Aktif' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- TAB 3: MASTER FAKULTAS & PROGRAM STUDI -->
      <section v-else-if="activeTab === 'programs'" class="flat-card table-section">
        <div class="section-heading">
          <div>
            <h2 class="section-title">Struktur Fakultas & Program Studi</h2>
            <p class="section-desc">Daftar unit akademik institusi beserta akses pemantauan khusus.</p>
          </div>
        </div>

        <div class="table-responsive">
          <table class="academic-table">
            <thead>
              <tr>
                <th>Kode Prodi</th>
                <th>Program Studi</th>
                <th>Fakultas</th>
                <th>ID Program Studi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="sp in studyPrograms" :key="sp.id">
                <td><span class="unit-tag">{{ sp.code }}</span></td>
                <td class="font-bold">{{ sp.name }}</td>
                <td>{{ sp.faculty_name }}</td>
                <td class="font-mono text-xs text-muted">{{ sp.id }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import Navbar from '../components/Navbar.vue'

const auth = useAuthStore()

const activeTab = ref('siakad')
const refreshing = ref(false)
const savingTheme = ref(false)
const themeSavedSuccess = ref(false)

const stats = ref(null)
const studyPrograms = ref([])
const users = ref([])
const userRoleFilter = ref('')

// SIAKAD API State
const apiConfig = ref({
  campus_api_url: 'http://mock-campus-api:9001/campus/sync',
  campus_api_key: 'rahasia-kampus-almuslim-2026',
  campus_sync_cron: '0 3 * * *',
  campus_name: 'Universitas Almuslim',
  campus_tagline: 'Sistem Informasi Presensi & QR Dinamis',
  geofence_mode: 'CAMPUS_AND_ROOM',
  max_tolerance_minutes: 15
})
const savingConfig = ref(false)
const apiConfigSaved = ref(false)
const syncing = ref(false)
const syncMessage = ref('')

// Campus Locations State
const campusLocations = ref([])
const showLocationForm = ref(false)
const editingLocationId = ref(null)
const savingLocation = ref(false)
const locationSavedMessage = ref('')
const locationForm = ref({
  name: '',
  description: '',
  latitude: 5.193730,
  longitude: 96.787492,
  radius_meters: 80,
  is_active: true
})

const themeForm = ref({
  campus_name: 'Universitas Almuslim',
  campus_tagline: 'Sistem Informasi Presensi & QR Dinamis',
  primary_color: '#006633',
  accent_color: '#D4AF37',
  logo_url: ''
})

async function fetchSystemData() {
  refreshing.value = true
  try {
    const headers = { Authorization: `Bearer ${auth.token}` }

    // Fetch Stats
    const resStats = await fetch('/v1/system/stats', { headers })
    if (resStats.ok) {
      stats.value = await resStats.json()
    }

    // Fetch Study Programs
    const resPrograms = await fetch('/v1/system/study-programs', { headers })
    if (resPrograms.ok) {
      const data = await resPrograms.json()
      studyPrograms.value = data.study_programs || []
    }

    // Fetch Theme
    const resTheme = await fetch('/v1/system/theme')
    if (resTheme.ok) {
      const t = await resTheme.json()
      themeForm.value = { ...themeForm.value, ...t }
    }

    // Fetch API Config
    await fetchApiConfig()

    // Fetch Campus Locations
    await fetchCampusLocations()

    await fetchUsers()
  } catch (err) {
    console.error('Error fetching system data:', err)
  } finally {
    refreshing.value = false
  }
}

async function fetchApiConfig() {
  try {
    const headers = { Authorization: `Bearer ${auth.token}` }
    const res = await fetch('/v1/system/config', { headers })
    if (res.ok) {
      const data = await res.json()
      if (data && data.campus_api_url) {
        apiConfig.value = { ...apiConfig.value, ...data }
      }
    }
  } catch (err) {
    console.error('Error fetching API config:', err)
  }
}

async function handleSaveApiConfig() {
  savingConfig.value = true
  apiConfigSaved.value = false
  try {
    const res = await fetch('/v1/system/config', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`
      },
      body: JSON.stringify(apiConfig.value)
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Gagal menyimpan konfigurasi SIAKAD')
    }
    apiConfigSaved.value = true
    setTimeout(() => {
      apiConfigSaved.value = false
    }, 4000)
  } catch (err) {
    alert(err.message)
  } finally {
    savingConfig.value = false
  }
}

async function handleTriggerSync() {
  syncing.value = true
  syncMessage.value = ''
  try {
    const res = await fetch('/v1/system/sync', {
      method: 'POST',
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    if (res.ok) {
      const data = await res.json()
      syncMessage.value = data.message || 'Sinkronisasi SIAKAD berhasil dipicu.'
      setTimeout(async () => {
        await fetchSystemData()
        syncMessage.value = ''
      }, 3500)
    }
  } catch (err) {
    alert('Gagal memicu sinkronisasi: ' + err.message)
  } finally {
    syncing.value = false
  }
}

async function fetchCampusLocations() {
  try {
    const headers = { Authorization: `Bearer ${auth.token}` }
    const res = await fetch('/v1/system/locations', { headers })
    if (res.ok) {
      const data = await res.json()
      campusLocations.value = data.locations || []
    }
  } catch (err) {
    console.error('Error fetching campus locations:', err)
  }
}

function openAddLocationModal() {
  editingLocationId.value = null
  locationForm.value = {
    name: '',
    description: '',
    latitude: 5.193730,
    longitude: 96.787492,
    radius_meters: 80,
    is_active: true
  }
  showLocationForm.value = true
}

function openEditLocationModal(loc) {
  editingLocationId.value = loc.id
  locationForm.value = {
    name: loc.name,
    description: loc.description,
    latitude: loc.latitude,
    longitude: loc.longitude,
    radius_meters: loc.radius_meters,
    is_active: loc.is_active
  }
  showLocationForm.value = true
}

async function handleSaveLocation() {
  savingLocation.value = true
  locationSavedMessage.value = ''
  try {
    const url = editingLocationId.value
      ? `/v1/system/locations/${editingLocationId.value}`
      : '/v1/system/locations'
    const method = editingLocationId.value ? 'PUT' : 'POST'

    const res = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`
      },
      body: JSON.stringify(locationForm.value)
    })

    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Gagal menyimpan titik lokasi')
    }

    locationSavedMessage.value = 'Titik geofence kampus berhasil disimpan.'
    showLocationForm.value = false
    await fetchCampusLocations()
    setTimeout(() => {
      locationSavedMessage.value = ''
    }, 4000)
  } catch (err) {
    alert(err.message)
  } finally {
    savingLocation.value = false
  }
}

async function handleDeleteLocation(id) {
  if (!confirm('Apakah Anda yakin ingin menghapus titik geofence ini?')) return
  try {
    const res = await fetch(`/v1/system/locations/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${auth.token}` }
    })
    if (res.ok) {
      await fetchCampusLocations()
    }
  } catch (err) {
    alert('Gagal menghapus lokasi: ' + err.message)
  }
}

async function fetchUsers() {
  try {
    const headers = { Authorization: `Bearer ${auth.token}` }
    let url = '/v1/system/users'
    if (userRoleFilter.value) {
      url += `?role=${userRoleFilter.value}`
    }
    const res = await fetch(url, { headers })
    if (res.ok) {
      const data = await res.json()
      users.value = data.users || []
    }
  } catch (err) {
    console.error('Error fetching users:', err)
  }
}

async function handleSaveTheme() {
  savingTheme.value = true
  themeSavedSuccess.value = false
  try {
    const res = await fetch('/v1/system/theme', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`
      },
      body: JSON.stringify(themeForm.value)
    })

    if (!res.ok) {
      throw new Error('Gagal menyimpan tema')
    }

    // Apply CSS variables dynamically to document
    document.documentElement.style.setProperty('--brand-primary', themeForm.value.primary_color)
    document.documentElement.style.setProperty('--brand-accent', themeForm.value.accent_color)

    themeSavedSuccess.value = true
    setTimeout(() => {
      themeSavedSuccess.value = false
    }, 4000)
  } catch (err) {
    alert(err.message)
  } finally {
    savingTheme.value = false
  }
}

onMounted(() => {
  fetchSystemData()
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

/* System Header Banner */
.system-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 28px;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
  border-left: 5px solid #0284c7;
}

.system-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background-color: #e0f2fe;
  color: #0369a1;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  margin-bottom: 8px;
}

.system-title {
  font-size: 1.5rem;
  font-weight: 800;
  color: var(--text-main);
  margin: 0;
}

.system-sub {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin: 6px 0 0;
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

.stat-number {
  font-size: 1.25rem;
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

/* Theme Settings Grid */
.theme-settings-grid {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 24px;
}

@media (max-width: 992px) {
  .theme-settings-grid {
    grid-template-columns: 1fr;
  }
}

.settings-card, .preview-card {
  padding: 24px;
}

.settings-header {
  margin-bottom: 20px;
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

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 0.825rem;
  font-weight: 700;
  color: var(--text-main);
}

.form-control, .form-select {
  padding: 9px 12px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  color: var(--text-main);
  background-color: #ffffff;
}

.color-pickers-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.color-input-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-picker {
  width: 42px;
  height: 38px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  cursor: pointer;
  padding: 2px;
}

.form-actions {
  margin-top: 12px;
}

/* Mock Preview */
.preview-box {
  background: #f8fafc;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.mock-nav {
  background: #ffffff;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  border-top: 4px solid transparent;
}

.mock-brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.mock-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 0.85rem;
}

.mock-title {
  font-weight: 700;
  font-size: 0.85rem;
  color: #0f172a;
}

.mock-sub {
  font-size: 0.7rem;
  color: #64748b;
}

.mock-body {
  padding: 16px;
}

.mock-card {
  background: #ffffff;
  padding: 14px;
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.mock-card-title {
  font-weight: 700;
  font-size: 0.8rem;
  color: #1e293b;
}

.mock-btn {
  color: #ffffff;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  width: fit-content;
}

.mock-badge {
  padding: 2px 8px;
  border-radius: 9999px;
  font-size: 0.7rem;
  font-weight: 700;
  width: fit-content;
}

/* Table Section */
.table-section {
  padding: 24px;
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

.unit-tag {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
  padding: 3px 8px;
  border-radius: 4px;
  border: 1px solid var(--brand-primary-border);
}

/* SIAKAD Architecture Flowchart Styles */
.architecture-card {
  padding: 24px;
}

.card-top-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 24px;
}

.tech-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background-color: #f1f5f9;
  color: #475569;
  border: 1px solid #cbd5e1;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 700;
  margin-bottom: 6px;
}

.flowchart-container {
  display: grid;
  grid-template-columns: 1fr auto 1fr auto 1fr auto 1fr;
  align-items: center;
  gap: 12px;
  padding: 24px;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow-x: auto;
}

@media (max-width: 1024px) {
  .flowchart-container {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  .flow-connector {
    transform: rotate(90deg);
    margin: 8px 0;
  }
}

.flow-node {
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-md);
  padding: 18px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.flow-node:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.08);
}

.node-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.bg-blue-subtle {
  background-color: #eff6ff;
  color: #2563eb;
}

.bg-yellow-subtle {
  background-color: #fefce8;
  color: #ca8a04;
}

.bg-green-subtle {
  background-color: #f0fdf4;
  color: #16a34a;
}

.bg-purple-subtle {
  background-color: #faf5ff;
  color: #9333ea;
}

.node-badge {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.node-title {
  margin: 2px 0 6px;
  font-size: 0.95rem;
  font-weight: 800;
  color: var(--text-main);
}

.node-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
  margin: 0 0 8px;
}

.node-meta {
  padding: 4px 8px;
  background-color: #f8fafc;
  border-radius: 4px;
  border: 1px solid #e2e8f0;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.flow-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-width: 80px;
}

.connector-line {
  width: 100%;
  height: 2px;
  background-color: #cbd5e1;
}

.connector-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background-color: #ffffff;
  border: 1px solid #cbd5e1;
  border-radius: 12px;
  font-size: 0.65rem;
  font-weight: 700;
  color: #475569;
  white-space: nowrap;
}

.form-row-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.form-row-3 {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 20px;
}

@media (max-width: 768px) {
  .form-row-2, .form-row-3 {
    grid-template-columns: 1fr;
  }
}

.form-hint {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 4px;
}

.settings-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.table-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-icon {
  background: none;
  border: none;
  padding: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.btn-icon:hover {
  background-color: rgba(0, 0, 0, 0.05);
}
</style>

