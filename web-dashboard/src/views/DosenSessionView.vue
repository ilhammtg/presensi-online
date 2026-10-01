<template>
  <div class="dosen-portal">
    <Navbar />

    <main class="portal-main">
      <!-- Error / Alert Banner -->
      <div v-if="pageError || session.error" class="flat-card alert-box">
        <i class="fas fa-exclamation-triangle alert-icon"></i>
        <div class="alert-content">
          <strong>Perhatian:</strong> {{ pageError || session.error }}
        </div>
        <button class="btn-alert-close" @click="pageError = ''; session.error = ''">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <!-- Success Notification -->
      <div v-if="successMsg" class="flat-card alert-box success-box">
        <i class="fas fa-check-circle alert-icon"></i>
        <div class="alert-content">{{ successMsg }}</div>
        <button class="btn-alert-close" @click="successMsg = ''">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <!-- ========================================================== -->
      <!-- VIEW 1: KATALOG DAFTAR MATA KULIAH & APPROVAL CENTER -->
      <!-- ========================================================== -->
      <section v-if="!selectedSchedule && !session.currentSession" class="catalogue-section">
        <div class="catalogue-header">
          <div>
            <h1 class="page-title">Jadwal & Mata Kuliah Diampu</h1>
            <p class="page-subtitle">
              Pilih mata kuliah untuk membuka presensi, melihat daftar mahasiswa presensi real-time, memberi izin/sakit, atau melihat rekapitulasi kehadiran.
            </p>
          </div>

          <!-- Filter Options: Hari Ini vs Semua -->
          <div class="filter-pills">
            <button 
              class="pill-btn" 
              :class="{ 'active': filterSchedule === 'today' }"
              @click="filterSchedule = 'today'"
            >
              <i class="fas fa-calendar-day"></i>
              <span>Mata Kuliah Hari Ini ({{ todayName }})</span>
              <span class="count-bubble">{{ todaySchedulesCount }}</span>
            </button>
            <button 
              class="pill-btn" 
              :class="{ 'active': filterSchedule === 'all' }"
              @click="filterSchedule = 'all'"
            >
              <i class="fas fa-th-large"></i>
              <span>Semua Mata Kuliah</span>
              <span class="count-bubble">{{ session.schedules.length }}</span>
            </button>
          </div>
        </div>

        <!-- Schedule Cards Grid -->
        <div v-if="displayedSchedules.length > 0" class="courses-grid">
          <div 
            v-for="sched in displayedSchedules" 
            :key="sched.id" 
            class="flat-card course-card"
            @click="handleSelectSchedule(sched)"
          >
            <div class="card-top">
              <div class="code-and-unit">
                <span class="course-badge">{{ sched.course_code }}</span>
                <span class="unit-tag">{{ sched.class_unit }}</span>
              </div>

              <div class="card-status-pill">
                <span v-if="sched.has_active_session" class="status-chip active-chip">
                  <i class="fas fa-bolt"></i> Sesi Berlangsung
                </span>
                <span v-else-if="sched.is_today" class="status-chip today-chip">
                  <i class="fas fa-calendar-check"></i> Hari Ini
                </span>
                <span v-else class="status-chip regular-chip">
                  <i class="fas fa-calendar-alt"></i> {{ sched.day_name }}
                </span>
              </div>
            </div>

            <h2 class="card-course-title">{{ sched.course_name }}</h2>

            <div class="card-meta-list">
              <div class="meta-item">
                <i class="fas fa-clock"></i>
                <span>{{ sched.day_name }}, {{ sched.start_time }} - {{ sched.end_time }} WIB</span>
              </div>
              <div class="meta-item">
                <i class="fas fa-map-marker-alt"></i>
                <span>{{ sched.room_name }} • {{ sched.building_name }}</span>
              </div>
              <div class="meta-item">
                <i class="fas fa-user-graduate"></i>
                <span>{{ sched.enrolled_count }} Mahasiswa Terdaftar</span>
              </div>
            </div>

            <div class="card-footer">
              <span class="card-prompt">
                <span v-if="sched.has_active_session" class="text-success font-semibold">
                  <i class="fas fa-door-open"></i> Masuk ke Sesi Presensi Aktif &rarr;
                </span>
                <span v-else>Buka Kelas & Presensi &rarr;</span>
              </span>
            </div>
          </div>
        </div>

        <!-- Empty State if no classes in filter -->
        <div v-else class="flat-card empty-catalogue">
          <i class="fas fa-calendar-times empty-icon"></i>
          <h3 class="empty-title">Tidak Ada Jadwal Kuliah Hari Ini</h3>
          <p class="empty-desc">
            Tidak ada jadwal perkuliahan yang terjadwal untuk hari {{ todayName }}. 
            Klik tombol <strong>"Semua Mata Kuliah"</strong> untuk melihat daftar lengkap mata kuliah yang Anda ampu di semester ini.
          </p>
          <button class="btn btn-secondary" @click="filterSchedule = 'all'" style="margin-top: 14px;">
            <i class="fas fa-th-large"></i> Lihat Semua Mata Kuliah
          </button>
        </div>
      </section>

      <!-- ========================================================== -->
      <!-- VIEW 2: RUANG KELAS TERPILIH (Presensi & Mahasiswa, Rekap) -->
      <!-- ========================================================== -->
      <section v-else class="class-detail-section">
        <!-- Back Navigation & Class Header Bar -->
        <div class="detail-nav-bar">
          <button class="btn btn-secondary btn-sm" @click="handleBackToCatalogue">
            <i class="fas fa-arrow-left"></i>
            <span>Kembali ke Daftar Mata Kuliah</span>
          </button>

          <div class="class-breadcrumb">
            <span class="bc-code">{{ currentCourse?.course_code }}</span>
            <span class="bc-name">{{ currentCourse?.course_name }} ({{ currentCourse?.class_unit || 'Unit 01' }})</span>
          </div>
        </div>

        <!-- Class Navigation Tabs (Presensi & Mahasiswa vs Rekapitulasi) -->
        <div class="detail-tab-bar">
          <button 
            class="detail-tab-btn" 
            :class="{ 'active': activeTab === 'presensi' }"
            @click="activeTab = 'presensi'"
          >
            <i class="fas fa-qrcode"></i>
            <span>Presensi & Peserta Kelas</span>
            <span v-if="session.currentSession" class="tab-live-badge">Live</span>
          </button>

          <button 
            class="detail-tab-btn" 
            :class="{ 'active': activeTab === 'rekap' }"
            @click="activeTab = 'rekap'"
          >
            <i class="fas fa-chart-bar"></i>
            <span>Rekapitulasi Presensi Semester</span>
          </button>
        </div>

        <!-- ---------------------------------------------------- -->
        <!-- TAB 1: PRESENSI & DAFTAR SELURUH PESERTA KELAS -->
        <!-- ---------------------------------------------------- -->
        <div v-if="activeTab === 'presensi'" class="tab-pane">
          <!-- Control Bar for Session Initiation / Closure -->
          <section class="flat-card control-bar">
            <div class="course-header">
              <div class="course-badge">{{ currentCourse?.course_code }}</div>
              <div class="course-info">
                <h1 class="course-name">{{ currentCourse?.course_name }}</h1>
                <div class="course-meta">
                  <span><i class="fas fa-building"></i> {{ currentCourse?.building_name }} • {{ currentCourse?.room_name }}</span>
                  <span>•</span>
                  <span><i class="fas fa-calendar-alt"></i> {{ currentCourse?.day_name }}, {{ currentCourse?.start_time }} - {{ currentCourse?.end_time }} WIB</span>
                  <span>•</span>
                  <span class="status-indicator" :class="{ 'is-active': session.currentSession }">
                    <i class="fas fa-circle status-dot"></i>
                    {{ session.currentSession ? `Pertemuan ${session.currentSession.meeting_no} Sedang Berlangsung` : 'Sesi Belum Dibuka' }}
                  </span>
                </div>
              </div>
            </div>

            <div class="control-actions">
              <!-- If Inactive: Automatically Locked Meeting Number Following Schedule -->
              <div v-if="!session.currentSession" class="action-group">
                <template v-if="currentCourse?.is_today">
                  <!-- Case 1: Today's meeting already completed -->
                  <div v-if="isMeetingCompletedToday" class="calendar-locked-card">
                    <i class="fas fa-calendar-check text-success"></i>
                    <div class="calendar-locked-text">
                      <strong>Pertemuan Perkuliahan Hari Ini Telah Selesai</strong>
                      <span>Pertemuan ke-{{ lastCompletedMeetingNoToday }} minggu ini telah diselenggarakan. Pertemuan ke-{{ session.nextMeetingNo }} dibuka pada jadwal kalender akademik minggu selanjutnya.</span>
                    </div>
                  </div>

                  <!-- Case 2: Meeting can be opened today -->
                  <div v-else class="open-session-panel">
                    <!-- Duration Selector Presets -->
                    <div class="duration-selector-box">
                      <span class="duration-label">
                        <i class="fas fa-stopwatch"></i> Durasi Presensi:
                      </span>
                      <div class="duration-presets">
                        <button 
                          type="button" 
                          v-for="d in durationPresets" 
                          :key="d.value" 
                          class="duration-chip" 
                          :class="{ 'active': selectedDuration === d.value }"
                          @click="selectedDuration = d.value"
                        >
                          {{ d.label }}
                        </button>
                      </div>
                    </div>

                    <div class="open-btn-row">
                      <div class="auto-meeting-badge">
                        <i class="fas fa-calendar-check"></i>
                        <span>Pertemuan Ke-{{ session.nextMeetingNo }}</span>
                      </div>
                      <button class="btn btn-primary" @click="handleOpenSession" :disabled="session.loading">
                        <i v-if="session.loading" class="fas fa-circle-notch fa-spin"></i>
                        <i v-else class="fas fa-play"></i>
                        <span>Buka Sesi ({{ selectedDuration }} Menit) & Tampilkan QR</span>
                      </button>
                    </div>
                  </div>
                </template>
                <div v-else class="not-today-warning">
                  <i class="fas fa-ban"></i>
                  <span>Buka presensi dikunci (Jadwal: {{ currentCourse?.day_name }}, bukan hari ini)</span>
                </div>
              </div>

              <!-- If Active: Close Button & Session Countdown -->
              <div v-else class="action-group active-session-action-group">
                <div class="session-timer-pill" v-if="sessionRemainingFormatted">
                  <i class="fas fa-hourglass-half"></i>
                  <span>Sisa Waktu: <strong>{{ sessionRemainingFormatted }}</strong></span>
                </div>
                <button class="btn btn-danger" @click="showCloseModal = true" :disabled="session.loading">
                  <i class="fas fa-stop"></i>
                  <span>Tutup Sesi</span>
                </button>
              </div>
            </div>
          </section>

          <!-- Day Mismatch Notification if viewing a class not on today's schedule -->
          <div v-if="!currentCourse?.is_today && !session.currentSession" class="flat-card schedule-notice-box">
            <i class="fas fa-info-circle notice-icon"></i>
            <div>
              <strong>Informasi Jadwal Perkuliahan:</strong>
              <p>Mata kuliah ini dijadwalkan pada hari <strong>{{ currentCourse?.day_name }}</strong>. Sesi presensi hanya dapat dibuka pada hari kuliah yang sesuai dengan jadwal kalender akademik kampus.</p>
            </div>
          </div>

          <!-- Active Session Layout: Projector on Left, Unified Attendee Table on Right -->
          <div class="presensi-workspace" :class="{ 'with-projector': session.currentSession }">
            <!-- Left: High-Contrast Projector Screen Box (Shown when session is active) -->
            <div v-if="session.currentSession" class="flat-card projector-card">
              <div class="projector-header">
                <div>
                  <h2 class="section-title">Layar Proyektor Mahasiswa</h2>
                  <p class="section-desc">Arahkan kamera aplikasi mahasiswa ke QR Code di bawah</p>
                </div>
              </div>

              <!-- QR Code Canvas Display -->
              <div class="qr-display-box" id="qrDisplayContainer">
                <qrcode-vue
                  v-if="session.qrToken"
                  :value="qrStringPayload"
                  :size="260"
                  level="H"
                  render-as="canvas"
                  class="qr-canvas-element"
                />
                <div v-else class="qr-loading">
                  <i class="fas fa-circle-notch fa-spin"></i>
                  <span>Memuat Token Presensi...</span>
                </div>
              </div>

              <!-- Dynamic Rolling QR Action Panel -->
              <div class="qr-action-panel">
                <div class="rolling-progress-wrap">
                  <div class="rolling-progress-bar" :style="{ width: ((session.qrExpiresIn || 15) / 15 * 100) + '%' }"></div>
                </div>
                <div class="rolling-info-line">
                  <span class="rolling-chip">
                    <span class="pulse-dot"></span> Dynamic Rolling QR (TOTP)
                  </span>
                  <span class="rolling-countdown">
                    Berganti dlm <strong>{{ session.qrExpiresIn || 15 }}s</strong>
                  </span>
                </div>

                <!-- Tombol Perbesar QR Layar Penuh (Mode Proyektor) -->
                <div class="qr-action-buttons">
                  <button class="btn btn-primary btn-sm qr-action-btn w-100" @click="showFullscreenQR = true" title="Tampilkan QR Code Layar Penuh untuk Proyektor">
                    <i class="fas fa-expand"></i>
                    <span>Perbesar QR (Layar Penuh Proyektor)</span>
                  </button>
                </div>
              </div>

              <!-- Security Badges -->
              <div class="security-flags">
                <span class="sec-item"><i class="fas fa-shield-alt text-success"></i> Rolling Hash 15s</span>
                <span class="sec-item"><i class="fas fa-map-marker-alt text-primary"></i> Radius 35m</span>
                <span class="sec-item"><i class="fas fa-mobile-alt text-accent"></i> Anti-Fraud Device</span>
              </div>
            </div>

            <!-- Right / Full-width: Daftar Seluruh Peserta Kelas & Status Presensi Realtime -->
            <div class="flat-card students-panel-card">
              <div class="panel-header">
                <div>
                  <h2 class="section-title">Peserta Kelas & Status Presensi</h2>
                  <p class="section-desc">
                    {{ session.currentSession ? 'Pantau presensi mahasiswa dan berikan izin/sakit langsung pada baris mahasiswa saat absensi berlangsung.' : 'Daftar mahasiswa yang mengambil mata kuliah ini pada kelas ini.' }}
                  </p>
                </div>

                <!-- Live Attendance Summary Chips -->
                <div class="status-summary-bar">
                  <span class="sum-chip total-chip">Total: <strong>{{ session.enrolledStudents.length }}</strong></span>
                  <span class="sum-chip hadir-chip"><i class="fas fa-check-circle"></i> Hadir: <strong>{{ countHadir }}</strong></span>
                  <span class="sum-chip telat-chip"><i class="fas fa-clock"></i> Telat: <strong>{{ countTelat }}</strong></span>
                  <span class="sum-chip izin-chip"><i class="fas fa-envelope-open-text"></i> Izin/Sakit: <strong>{{ countIzinSakit }}</strong></span>
                  <span class="sum-chip belum-chip"><i class="fas fa-hourglass-start"></i> Belum Hadir: <strong>{{ countBelum }}</strong></span>
                </div>
              </div>

              <!-- Contextual Session Tips -->
              <div v-if="session.currentSession" class="live-session-tip">
                <i class="fas fa-broadcast-tower text-success"></i>
                <div class="tip-text">
                  <strong>Proses Absensi Sedang Berlangsung:</strong>
                  <span>Untuk mahasiswa yang berhalangan atau mengabari izin/sakit, klik tombol <strong>[+ Izin]</strong> atau <strong>[+ Sakit]</strong> pada baris mahasiswa di bawah.</span>
                </div>
              </div>
              <div v-else class="inactive-session-tip">
                <i class="fas fa-info-circle text-primary"></i>
                <div class="tip-text">
                  <strong>Sesi Perkuliahan Belum Dibuka:</strong>
                  <span>Pencatatan izin atau sakit hanya dapat dilakukan saat proses absensi sedang berlangsung. Silakan klik tombol <strong>"Buka Presensi Kelas"</strong> di atas saat kelas dimulai.</span>
                </div>
              </div>

              <!-- Instant Search & Status Filter Toolbar -->
              <div class="search-toolbar">
                <div class="search-box-wrapper">
                  <i class="fas fa-search search-icon"></i>
                  <input 
                    type="text" 
                    v-model="searchStudentQuery" 
                    placeholder="Cari mahasiswa berdasarkan nama atau NIM..." 
                    class="input-field search-student-input"
                  />
                  <button v-if="searchStudentQuery" class="clear-search-btn" @click="searchStudentQuery = ''">
                    <i class="fas fa-times"></i>
                  </button>
                </div>

                <!-- Quick Status Filter Pills -->
                <div class="student-status-filter-pills">
                  <button 
                    class="s-pill-btn" 
                    :class="{ 'active': studentStatusFilter === 'all' }"
                    @click="studentStatusFilter = 'all'"
                  >
                    Semua ({{ session.enrolledStudents.length }})
                  </button>
                  <button 
                    class="s-pill-btn s-belum" 
                    :class="{ 'active': studentStatusFilter === 'belum' }"
                    @click="studentStatusFilter = 'belum'"
                  >
                    <i class="fas fa-hourglass-start"></i> Belum Hadir ({{ countBelum }})
                  </button>
                  <button 
                    class="s-pill-btn s-hadir" 
                    :class="{ 'active': studentStatusFilter === 'hadir' }"
                    @click="studentStatusFilter = 'hadir'"
                  >
                    <i class="fas fa-check-circle"></i> Hadir ({{ countHadir + countTelat }})
                  </button>
                  <button 
                    class="s-pill-btn s-izin" 
                    :class="{ 'active': studentStatusFilter === 'izin' }"
                    @click="studentStatusFilter = 'izin'"
                  >
                    <i class="fas fa-envelope-open-text"></i> Izin/Sakit ({{ countIzinSakit }})
                  </button>
                </div>
              </div>

              <!-- Unified Students Table -->
              <div class="table-container">
                <table class="academic-table" v-if="filteredStudents.length > 0">
                  <thead>
                    <tr>
                      <th>No</th>
                      <th>NIM</th>
                      <th>Nama Mahasiswa</th>
                      <th>Status Presensi</th>
                      <th>Waktu / Catatan</th>
                      <th>Aksi Izin & Kehadiran</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(student, idx) in filteredStudents" :key="student.student_id">
                      <td>{{ idx + 1 }}</td>
                      <td class="font-mono">{{ student.student_nim }}</td>
                      <td class="font-semibold">{{ student.student_name }}</td>
                      <td>
                        <!-- Dynamic status based on attendance record -->
                        <div class="status-cell-wrap">
                          <span 
                            class="badge" 
                            :class="getStudentStatusBadgeClass(student.student_id)"
                          >
                            <i :class="getStudentStatusIcon(student.student_id)"></i>
                            {{ getStudentStatusText(student.student_id) }}
                          </span>
                          <!-- Source indicator -->
                          <span v-if="attendanceByStudentId[student.student_id]?.submission_source === 'manual_lecturer'" class="source-sub-badge wa" title="Input izin oleh dosen saat sesi absensi">
                            <i class="fab fa-whatsapp"></i> Izin Dosen
                          </span>
                          <span v-else-if="attendanceByStudentId[student.student_id]?.submission_source === 'app_request'" class="source-sub-badge app" title="Pengajuan mandiri via aplikasi mahasiswa">
                            <i class="fas fa-mobile-alt"></i> Pengajuan App
                          </span>
                          <span v-else-if="attendanceByStudentId[student.student_id]?.submission_source === 'self_scan'" class="source-sub-badge scan" title="Scan mandiri QR Code + GPS">
                            <i class="fas fa-qrcode"></i> Scan QR
                          </span>
                        </div>
                      </td>
                      <td class="text-muted meta-cell">
                        {{ getStudentNotesOrTime(student.student_id) }}
                      </td>
                      <td>
                        <!-- When Attendance is ACTIVE: Provide direct Izin/Sakit buttons -->
                        <div v-if="session.currentSession" class="action-btn-row">
                          <!-- 1. If student submitted via mobile app and is pending -->
                          <template v-if="attendanceByStudentId[student.student_id]?.submission_source === 'app_request' && !attendanceByStudentId[student.student_id]?.verified_by_lecturer">
                            <button v-if="attendanceByStudentId[student.student_id]?.attachment_url" class="btn btn-outline-primary btn-xs" @click="openAttachmentPreview(attendanceByStudentId[student.student_id].attachment_url)" title="Lihat surat dokter/dispensasi">
                              <i class="fas fa-paperclip"></i> Bukti
                            </button>
                            <button class="btn btn-success btn-xs" @click="handleApprovalAction(attendanceByStudentId[student.student_id], 'approve')" :disabled="approvalActionLoading" title="Setujui permohonan izin">
                              <i class="fas fa-check"></i> Setujui
                            </button>
                            <button class="btn btn-danger btn-xs" @click="handleApprovalAction(attendanceByStudentId[student.student_id], 'reject')" :disabled="approvalActionLoading" title="Tolak permohonan (alpa)">
                              <i class="fas fa-times"></i> Tolak
                            </button>
                          </template>

                          <!-- 2. If student is already marked Izin or Sakit: Allow editing reason -->
                          <template v-else-if="attendanceByStudentId[student.student_id]?.status === 'izin' || attendanceByStudentId[student.student_id]?.status === 'sakit'">
                            <button class="btn btn-outline-secondary btn-xs" @click="openPermissionModal(student, attendanceByStudentId[student.student_id]?.status)" title="Ubah keterangan izin/sakit">
                              <i class="fas fa-edit"></i> Ubah Izin
                            </button>
                          </template>

                          <!-- 3. If student already scanned Hadir/Terlambat -->
                          <template v-else-if="attendanceByStudentId[student.student_id]?.status === 'hadir' || attendanceByStudentId[student.student_id]?.status === 'terlambat'">
                            <button class="btn btn-outline-secondary btn-xs text-muted" @click="openPermissionModal(student, 'izin')" title="Koreksi presensi mahasiswa ke Izin/Sakit">
                              <i class="fas fa-pen"></i> Koreksi
                            </button>
                          </template>

                          <!-- 4. Default: Student hasn't attended yet: Give direct Hadir, Izin and Sakit buttons -->
                          <template v-else>
                            <button 
                              class="btn btn-outline-success btn-xs btn-row-hadir"
                              @click="openPermissionModal(student, 'hadir')"
                              title="Beri status Hadir (Manual) jika mahasiswa mengalami kendala presensi"
                            >
                              <i class="fas fa-user-check"></i>
                              <span>+ Hadir</span>
                            </button>
                            <button 
                              class="btn btn-outline-warning btn-xs btn-row-izin"
                              @click="openPermissionModal(student, 'izin')"
                              title="Beri Izin Resmi dari dosen untuk mahasiswa ini"
                            >
                              <i class="fas fa-envelope-open-text"></i>
                              <span>+ Izin</span>
                            </button>
                            <button 
                              class="btn btn-outline-info btn-xs btn-row-sakit"
                              @click="openPermissionModal(student, 'sakit')"
                              title="Beri keterangan Sakit untuk mahasiswa ini"
                            >
                              <i class="fas fa-medkit"></i>
                              <span>+ Sakit</span>
                            </button>
                          </template>
                        </div>

                        <!-- When Attendance is INACTIVE: Display locked notice -->
                        <div v-else class="locked-action-cell">
                          <span class="badge badge-subtle" title="Izin hanya dapat diberikan saat proses absensi sedang berlangsung">
                            <i class="fas fa-lock"></i> Sesi Belum Dibuka
                          </span>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>

                <div v-else class="empty-state">
                  <i class="fas fa-search empty-icon"></i>
                  <p class="empty-title">Mahasiswa Tidak Ditemukan</p>
                  <p class="empty-desc">Tidak ada mahasiswa yang cocok dengan filter atau pencarian "{{ searchStudentQuery }}".</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ---------------------------------------------------- -->
        <!-- TAB 2: REKAPITULASI PRESENSI KELAS SEMESTER -->
        <!-- ---------------------------------------------------- -->
        <div v-if="activeTab === 'rekap'" class="tab-pane">
          <div class="flat-card recap-card">
            <div class="recap-header">
              <div>
                <h2 class="section-title">Rekapitulasi Kehadiran Kelas</h2>
                <p class="section-desc">Matriks absensi mahasiswa per pertemuan semester berjalan</p>
              </div>
              <div class="recap-btn-actions">
                <button class="btn btn-secondary btn-sm" @click="handleRefreshRecap" title="Segarkan data presensi">
                  <i class="fas fa-sync-alt"></i>
                  <span>Segarkan</span>
                </button>
                <button class="btn btn-outline-success btn-sm" @click="exportRecapToCSV" title="Unduh rekapitulasi ke format spreadsheet CSV / Excel">
                  <i class="fas fa-file-excel"></i>
                  <span>Ekspor Excel (.csv)</span>
                </button>
                <button class="btn btn-outline-primary btn-sm" @click="printOfficialRecap" title="Cetak atau simpan lembar presensi resmi dalam bentuk PDF">
                  <i class="fas fa-print"></i>
                  <span>Cetak / PDF Resmi</span>
                </button>
              </div>
            </div>

            <!-- Legend Badges -->
            <div class="recap-legend">
              <span class="legend-item"><span class="badge badge-success">H</span> Hadir Tepat Waktu</span>
              <span class="legend-item"><span class="badge badge-warning">T</span> Terlambat</span>
              <span class="legend-item"><span class="badge badge-info">I</span> Izin Resmi</span>
              <span class="legend-item"><span class="badge badge-info">S</span> Sakit</span>
              <span class="legend-item"><span class="badge badge-danger">A</span> Alpa / Tidak Hadir</span>
            </div>

            <div class="table-container">
              <table class="academic-table recap-table" v-if="session.classRecap && session.classRecap.students?.length > 0">
                <thead>
                  <tr>
                    <th rowspan="2">No</th>
                    <th rowspan="2">NIM</th>
                    <th rowspan="2">Nama Mahasiswa</th>
                    <th 
                      v-if="session.classRecap.meetings?.length > 0" 
                      :colspan="session.classRecap.meetings.length" 
                      class="text-center"
                    >
                      Pertemuan Kuliah
                    </th>
                    <th colspan="5" class="text-center">Total Akumulasi</th>
                    <th rowspan="2" class="text-center">% Hadir</th>
                  </tr>
                  <tr>
                    <!-- Meeting Columns P1..Pn -->
                    <th 
                      v-for="m in session.classRecap.meetings" 
                      :key="'col-p-'+m" 
                      class="meeting-th"
                    >
                      P{{ m }}
                    </th>
                    <th class="col-counter">H</th>
                    <th class="col-counter">T</th>
                    <th class="col-counter">I</th>
                    <th class="col-counter">S</th>
                    <th class="col-counter">A</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in session.classRecap.students" :key="row.student_id">
                    <td>{{ idx + 1 }}</td>
                    <td class="font-mono">{{ row.student_nim }}</td>
                    <td class="font-semibold">{{ row.student_name }}</td>

                    <!-- Meeting status cells -->
                    <td 
                      v-for="m in session.classRecap.meetings" 
                      :key="'cell-'+row.student_id+'-'+m" 
                      class="status-cell text-center"
                    >
                      <span 
                        class="recap-badge" 
                        :class="getRecapBadgeClass(row.meeting_statuses[m])"
                      >
                        {{ getStatusLetter(row.meeting_statuses[m]) }}
                      </span>
                    </td>

                    <!-- Counters -->
                    <td class="col-counter font-semibold text-success">{{ row.hadir_count }}</td>
                    <td class="col-counter font-semibold text-warning">{{ row.terlambat_count }}</td>
                    <td class="col-counter font-semibold text-info">{{ row.izin_count }}</td>
                    <td class="col-counter font-semibold text-info">{{ row.sakit_count }}</td>
                    <td class="col-counter font-semibold text-danger">{{ row.alpa_count }}</td>

                    <!-- Percentage -->
                    <td class="text-center">
                      <span class="pct-chip" :class="getPctClass(row.percentage)">
                        {{ row.percentage }}%
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>

              <div v-else class="empty-state">
                <i class="fas fa-chart-line empty-icon"></i>
                <p class="empty-title">Belum Ada Sesi Pertemuan yang Diselenggarakan</p>
                <p class="empty-desc">Saat pertemuan kuliah selesai dilaksanakan, rekapitulasi matriks absensi mahasiswa akan otomatis tampil di sini.</p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Modal Tutup Sesi & BAP (Enforces BR-DSN-03 Min 10 Chars) -->
    <div v-if="showCloseModal" class="modal-overlay">
      <div class="modal-box flat-card">
        <div class="modal-header">
          <h3 class="modal-title">Tutup Sesi Perkuliahan</h3>
          <button class="btn-close" @click="showCloseModal = false">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <p class="modal-intro">
            Masukkan Berita Acara Perkuliahan (BAP) ringkas. Mahasiswa yang tidak melakukan presensi dan tidak memiliki keterangan izin/sakit akan <strong>otomatis tercatat sebagai Alpa</strong>.
          </p>
          <div class="form-group" style="margin-top: 14px;">
            <div class="label-with-counter">
              <label for="bapInput">Topik & Materi Perkuliahan (BAP)</label>
              <span class="char-counter" :class="{ 'valid': bapTopic.trim().length >= 10, 'invalid': bapTopic.trim().length < 10 }">
                {{ bapTopic.trim().length }}/10 karakter minimum
              </span>
            </div>
            <textarea 
              id="bapInput"
              v-model="bapTopic" 
              class="input-field" 
              rows="4" 
              placeholder="Contoh: Pembahasan model konsistensi data dan implementasi RPC pada sistem terdistribusi..."
            ></textarea>
            <p v-if="bapTopic.trim().length < 10" class="field-hint text-warning">
              <i class="fas fa-exclamation-triangle"></i> Sesuai aturan BR-DSN-03, ringkasan materi/BAP wajib diisi minimal 10 karakter sebelum sesi dapat ditutup.
            </p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showCloseModal = false">Batal</button>
          <button class="btn btn-danger" @click="handleCloseSession" :disabled="session.loading || bapTopic.trim().length < 10">
            <i v-if="session.loading" class="fas fa-circle-notch fa-spin"></i>
            <i v-else class="fas fa-check"></i>
            <span>Simpan BAP & Tutup Sesi</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Catat Izin / Sakit Mahasiswa (Dual-Channel Quick Presets) -->
    <div v-if="showPermissionModal" class="modal-overlay">
      <div class="modal-box flat-card">
        <div class="modal-header">
          <h3 class="modal-title">Beri Keterangan Kehadiran Mahasiswa</h3>
          <button class="btn-close" @click="showPermissionModal = false">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="target-student-box">
            <div class="student-name">{{ targetPermStudent?.student_name }}</div>
            <div class="student-nim">NIM: {{ targetPermStudent?.student_nim }}</div>
          </div>

          <!-- Quick Preset Buttons -->
          <div class="preset-shortcuts">
            <span class="preset-label">Preset Cepat:</span>
            <button type="button" class="preset-tag" @click="permStatus = 'hadir'; permNotes = 'Kendala teknis / verifikasi tatap muka dosen'">
              <i class="fas fa-user-check text-success"></i> Hadir (Manual)
            </button>
            <button type="button" class="preset-tag" @click="permStatus = 'izin'; permNotes = 'Izin resmi disetujui dosen'">
              <i class="fas fa-envelope-open-text text-warning"></i> Izin Dosen
            </button>
            <button type="button" class="preset-tag" @click="permStatus = 'sakit'; permNotes = 'Sakit (pemberitahuan ke dosen)'">
              <i class="fas fa-medkit text-info"></i> Sakit
            </button>
          </div>

          <div class="form-group" style="margin-top: 16px;">
            <label for="permStatusSelect">Status Kehadiran</label>
            <select id="permStatusSelect" v-model="permStatus" class="input-field">
              <option value="hadir">Hadir (Manual) - Mahasiswa mengalami kendala presensi selama sesi berlangsung</option>
              <option value="izin">Izin - Mahasiswa mendapat izin resmi dari dosen</option>
              <option value="sakit">Sakit - Mahasiswa tidak hadir karena sakit</option>
              <option value="alpa">Alpa - Tanpa keterangan</option>
            </select>
          </div>

          <div class="form-group" style="margin-top: 14px;">
            <label for="permNotesInput">Keterangan / Catatan Dosen</label>
            <input 
              id="permNotesInput"
              type="text" 
              v-model="permNotes" 
              class="input-field" 
              placeholder="Contoh: Kendala jaringan / izin keperluan dinas / sakit demam"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showPermissionModal = false">Batal</button>
          <button class="btn btn-primary" @click="handleSavePermission" :disabled="session.loading">
            <i v-if="session.loading" class="fas fa-circle-notch fa-spin"></i>
            <i v-else class="fas fa-save"></i>
            <span>Simpan Keterangan Sah</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Preview Dokumen / Lampiran Bukti Surat -->
    <div v-if="previewAttachmentUrl" class="modal-overlay" @click.self="previewAttachmentUrl = ''">
      <div class="modal-box flat-card attachment-preview-modal" style="max-width: 600px;">
        <div class="modal-header">
          <h3 class="modal-title">Lampiran Bukti Surat / Dispensasi</h3>
          <button class="btn-close" @click="previewAttachmentUrl = ''">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body text-center">
          <img :src="previewAttachmentUrl" alt="Bukti Surat" class="img-attachment-preview" />
          <div style="margin-top: 14px;">
            <a :href="previewAttachmentUrl" target="_blank" class="btn btn-outline-primary btn-sm">
              <i class="fas fa-external-link-alt"></i> Buka Ukuran Penuh
            </a>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Layar Penuh (Projector Mode) QR Code -->
    <div v-if="showFullscreenQR" class="fullscreen-qr-overlay" @click.self="showFullscreenQR = false">
      <div class="fullscreen-qr-container">
        <!-- Close button top-right -->
        <button class="btn-fullscreen-close" @click="showFullscreenQR = false" title="Tutup Layar Penuh (ESC)">
          <i class="fas fa-times"></i>
          <span>Tutup (ESC)</span>
        </button>

        <div class="fullscreen-qr-header">
          <div class="univ-brand-tag">UNIVERSITAS ALMUSLIM</div>
          <h1 class="fullscreen-course-title">{{ currentCourse?.course_name }}</h1>
          <div class="fullscreen-course-meta">
            <span class="meta-tag">{{ currentCourse?.course_code }}</span>
            <span class="meta-separator">•</span>
            <span class="meta-tag">Unit {{ currentCourse?.class_unit || '01' }}</span>
            <span class="meta-separator">•</span>
            <span class="meta-tag">Pertemuan Ke-{{ session.currentSession?.meeting_no || session.nextMeetingNo || 1 }}</span>
            <span class="meta-separator">•</span>
            <span class="meta-tag">{{ currentCourse?.room_name }} ({{ currentCourse?.building_name }})</span>
          </div>
        </div>

        <div class="fullscreen-qr-canvas-box">
          <qrcode-vue
            v-if="session.qrToken"
            :value="qrStringPayload"
            :size="380"
            level="H"
            render-as="canvas"
            class="fullscreen-qr-element"
          />
          <div v-else class="qr-loading">
            <i class="fas fa-circle-notch fa-spin"></i>
            <span>Memuat Token Presensi...</span>
          </div>
        </div>

        <div class="fullscreen-qr-footer">
          <div class="fullscreen-rolling-bar">
            <div class="fullscreen-rolling-fill" :style="{ width: ((session.qrExpiresIn || 15) / 15 * 100) + '%' }"></div>
          </div>
          <div class="fullscreen-rolling-info">
            <span class="totp-badge"><i class="fas fa-sync-alt fa-spin"></i> Dynamic TOTP QR Code</span>
            <span class="totp-countdown">Berganti otomatis dalam <strong>{{ session.qrExpiresIn || 15 }} detik</strong></span>
          </div>
          <p class="fullscreen-instruction">
            Buka aplikasi Presensi Almuslim di ponsel &rarr; Pilih menu <strong>Absensi</strong> &rarr; Arahkan kamera ke QR Code di atas.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import QrcodeVue from 'qrcode.vue'
import Navbar from '../components/Navbar.vue'
import { useSessionStore } from '../stores/session'

const route = useRoute()
const router = useRouter()
const session = useSessionStore()

const filterSchedule = ref('today')
const activeTab = ref('presensi')
const selectedSchedule = ref(null)
const searchStudentQuery = ref('')
const studentStatusFilter = ref('all') // 'all' | 'belum' | 'hadir' | 'izin'

const previewAttachmentUrl = ref('')
const approvalActionLoading = ref(false)

const showFullscreenQR = ref(false)
const showCloseModal = ref(false)
const bapTopic = ref('')
const pageError = ref('')
const successMsg = ref('')

// Session Duration & Presets
const selectedDuration = ref(30)
const nowTime = ref(Date.now())
let clockTicker = null

// Permission Modal
const showPermissionModal = ref(false)
const targetPermStudent = ref(null)
const permStatus = ref('izin')
const permNotes = ref('')

const dayNames = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']
const todayName = computed(() => {
  const d = new Date().getDay()
  return dayNames[d]
})

const displayedSchedules = computed(() => {
  if (filterSchedule.value === 'today') {
    return session.schedules.filter(s => s.is_today)
  }
  return session.schedules
})

const todaySchedulesCount = computed(() => {
  return session.schedules.filter(s => s.is_today).length
})

const currentCourse = computed(() => {
  if (selectedSchedule.value) return selectedSchedule.value
  if (session.currentSession && session.schedules.length > 0) {
    return session.schedules.find(s => s.id === session.currentSession.schedule_id) || null
  }
  return null
})

// Duration Presets
const durationPresets = computed(() => [
  { label: '15 Menit', value: 15 },
  { label: '30 Menit (Standar)', value: 30 },
  { label: '45 Menit', value: 45 },
  { label: '60 Menit (Lab)', value: 60 },
  { label: 'Sampai Jam Selesai', value: calculateRemainingCourseMinutes() }
])

function calculateRemainingCourseMinutes() {
  if (!currentCourse.value?.end_time) return 90
  const parts = currentCourse.value.end_time.split(':').map(Number)
  const now = new Date()
  const end = new Date()
  end.setHours(parts[0], parts[1], 0, 0)
  const diffMinutes = Math.round((end - now) / 60000)
  return diffMinutes > 15 ? diffMinutes : 90
}

// Academic Calendar Rule: Check if a meeting has ALREADY been completed today for this schedule
const isMeetingCompletedToday = computed(() => {
  if (!session.scheduleSessions || session.scheduleSessions.length === 0) return false
  const now = new Date()
  const nowY = now.getFullYear()
  const nowM = now.getMonth()
  const nowD = now.getDate()
  return session.scheduleSessions.some(s => {
    if (s.is_open) return false // Active session is not completed
    if (!s.session_date) return false
    const d = new Date(s.session_date)
    return d.getFullYear() === nowY && d.getMonth() === nowM && d.getDate() === nowD
  })
})

const lastCompletedMeetingNoToday = computed(() => {
  if (!session.scheduleSessions) return 0
  const now = new Date()
  const nowY = now.getFullYear()
  const nowM = now.getMonth()
  const nowD = now.getDate()
  const match = session.scheduleSessions.find(s => {
    if (s.is_open || !s.session_date) return false
    const d = new Date(s.session_date)
    return d.getFullYear() === nowY && d.getMonth() === nowM && d.getDate() === nowD
  })
  return match ? match.meeting_no : 0
})

// Active Session Remaining Countdown
const sessionRemainingFormatted = computed(() => {
  if (!session.currentSession) return ''
  const sess = session.currentSession
  let expireTimestamp = 0
  if (sess.expires_at) {
    expireTimestamp = new Date(sess.expires_at).getTime()
  } else if (sess.opened_at) {
    const dur = (sess.duration_minutes || 30) * 60 * 1000
    expireTimestamp = new Date(sess.opened_at).getTime() + dur
  }
  if (!expireTimestamp) return ''
  const diffSec = Math.floor((expireTimestamp - nowTime.value) / 1000)
  if (diffSec <= 0) return 'Waktu Habis'
  const m = Math.floor(diffSec / 60)
  const s = diffSec % 60
  return `${m}m ${s.toString().padStart(2, '0')}s`
})

// Map attendance records by student_id
const attendanceByStudentId = computed(() => {
  const map = {}
  if (session.attendees) {
    for (const a of session.attendees) {
      map[a.student_id] = a
    }
  }
  return map
})

// Filtered students by search query & quick status filter
const filteredStudents = computed(() => {
  if (!session.enrolledStudents) return []
  let list = session.enrolledStudents

  // Filter by status tab if selected
  if (studentStatusFilter.value === 'belum') {
    list = list.filter(s => {
      const att = attendanceByStudentId.value[s.student_id]
      return !att
    })
  } else if (studentStatusFilter.value === 'hadir') {
    list = list.filter(s => {
      const att = attendanceByStudentId.value[s.student_id]
      return att && (att.status === 'hadir' || att.status === 'terlambat')
    })
  } else if (studentStatusFilter.value === 'izin') {
    list = list.filter(s => {
      const att = attendanceByStudentId.value[s.student_id]
      return att && (att.status === 'izin' || att.status === 'sakit')
    })
  }

  const q = searchStudentQuery.value.trim().toLowerCase()
  if (!q) return list
  return list.filter(s => 
    (s.student_name && s.student_name.toLowerCase().includes(q)) ||
    (s.student_nim && s.student_nim.toLowerCase().includes(q))
  )
})

// Live Counters
const countHadir = computed(() => {
  return session.attendees.filter(a => a.status === 'hadir').length
})

const countTelat = computed(() => {
  return session.attendees.filter(a => a.status === 'terlambat').length
})

const countIzinSakit = computed(() => {
  return session.attendees.filter(a => a.status === 'izin' || a.status === 'sakit').length
})

const countBelum = computed(() => {
  const total = session.enrolledStudents.length
  const recorded = session.attendees.length
  return Math.max(0, total - recorded)
})

const qrStringPayload = computed(() => {
  if (!session.currentSession || !session.qrToken) return ''
  const sessId = session.currentSession.id || session.currentSession.ID
  return JSON.stringify({
    session_id: sessId,
    token: session.qrToken
  })
})

function handleKeyDown(e) {
  if (e.key === 'Escape' && showFullscreenQR.value) {
    showFullscreenQR.value = false
  }
}

onMounted(async () => {
  pageError.value = ''
  window.addEventListener('keydown', handleKeyDown)

  // 0. Pre-fetch permissions for counter badges
  session.fetchPermissions()

  // 1. Fetch lecturer schedules
  const scheds = await session.fetchDosenSchedules()

  // 2. Check if direct route has sessionId
  const paramSessionId = route.params.sessionId
  if (paramSessionId) {
    try {
      const sess = await session.loadSessionById(paramSessionId)
      const matched = scheds.find(s => s.id === sess.schedule_id)
      if (matched) {
        selectedSchedule.value = matched
        await session.fetchScheduleDetails(matched.id)
      }
    } catch (err) {
      pageError.value = err.message || 'Sesi perkuliahan tidak valid atau telah ditutup.'
      setTimeout(() => router.replace('/dosen'), 2500)
    }
    return
  }

  // 3. Check if any schedule has an active session currently running
  for (const s of scheds) {
    if (s.has_active_session && s.active_session_id) {
      selectedSchedule.value = s
      await session.fetchScheduleDetails(s.id)
      await session.loadSessionById(s.active_session_id)
      if (route.params.sessionId !== s.active_session_id) {
        router.replace(`/dosen/sesi/${s.active_session_id}`)
      }
      return
    }
  }

  // Start clock ticker for live remaining countdown
  clockTicker = setInterval(() => {
    nowTime.value = Date.now()
  }, 1000)

  // Default select today's schedule if available
  const todaySched = scheds.find(s => s.is_today)
  if (todaySched) {
    selectedSchedule.value = todaySched
    await session.fetchScheduleDetails(todaySched.id)
    if (todaySched.has_active_session && todaySched.active_session_id) {
      await session.loadSessionById(todaySched.active_session_id)
      if (route.params.sessionId !== todaySched.active_session_id) {
        router.replace(`/dosen/sesi/${todaySched.active_session_id}`)
      }
    }
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
  if (clockTicker) {
    clearInterval(clockTicker)
    clockTicker = null
  }
})

// Watch route param sessionId to handle browser navigation (back/forward)
watch(() => route.params.sessionId, async (newSessionId) => {
  if (newSessionId) {
    if (!session.currentSession || (session.currentSession.id !== newSessionId && session.currentSession.ID !== newSessionId)) {
      try {
        await session.loadSessionById(newSessionId)
      } catch (err) {
        console.error('Failed to load session from route change:', err)
      }
    }
  }
})

async function handleSelectSchedule(sched) {
  selectedSchedule.value = sched
  pageError.value = ''
  successMsg.value = ''
  searchStudentQuery.value = ''
  await session.fetchScheduleDetails(sched.id)

  // If this schedule already has an active session, restore it immediately
  if (sched.has_active_session && sched.active_session_id) {
    await session.loadSessionById(sched.active_session_id)
    if (route.params.sessionId !== sched.active_session_id) {
      router.push(`/dosen/sesi/${sched.active_session_id}`)
    }
  } else if (!session.currentSession || session.currentSession.schedule_id !== sched.id) {
    // If the selected schedule does not have an active session, reset current session
    session.currentSession = null
    session.qrToken = ''
    session.disconnectWebSocket()
    if (route.params.sessionId) {
      router.push('/dosen')
    }
  }
}

function handleBackToCatalogue() {
  selectedSchedule.value = null
  pageError.value = ''
  successMsg.value = ''
  searchStudentQuery.value = ''
  if (!session.currentSession) {
    router.push('/dosen')
  }
  session.fetchDosenSchedules()
}

async function handleOpenSession() {
  if (!selectedSchedule.value) return
  if (!selectedSchedule.value.is_today) {
    pageError.value = `Presensi hanya dapat dibuka pada hari jadwal perkuliahan (${selectedSchedule.value.day_name || 'hari ini'}).`
    return
  }
  pageError.value = ''
  successMsg.value = ''
  try {
    const meetingNum = session.nextMeetingNo || 1
    const sess = await session.openSession(selectedSchedule.value.id, meetingNum, selectedDuration.value)
    const activeId = sess.id || sess.ID
    router.push(`/dosen/sesi/${activeId}`)
    await session.fetchScheduleDetails(selectedSchedule.value.id)
    successMsg.value = `Sesi Pertemuan ${meetingNum} (${selectedDuration.value} Menit) berhasil dibuka.`
  } catch (err) {
    pageError.value = err.message
  }
}

async function handleCloseSession() {
  pageError.value = ''
  try {
    await session.closeSession(bapTopic.value || 'Perkuliahan selesai diselenggarakan.')
    showCloseModal.value = false
    bapTopic.value = ''
    successMsg.value = 'Sesi perkuliahan berhasil ditutup. Rekapitulasi absensi telah diperbarui.'
    if (selectedSchedule.value) {
      await session.fetchScheduleDetails(selectedSchedule.value.id)
    }
    router.push('/dosen')
  } catch (err) {
    pageError.value = err.message
  }
}

function openPermissionModal(student, defaultStatus = 'izin') {
  targetPermStudent.value = student
  const existing = attendanceByStudentId.value[student.student_id]
  permStatus.value = existing ? existing.status : defaultStatus
  permNotes.value = existing && existing.notes ? existing.notes : ''
  if (!permNotes.value) {
    if (permStatus.value === 'hadir') permNotes.value = 'Kendala teknis / verifikasi tatap muka dosen'
    if (permStatus.value === 'izin') permNotes.value = 'Izin resmi disetujui dosen'
    if (permStatus.value === 'sakit') permNotes.value = 'Sakit (pemberitahuan ke dosen)'
  }
  showPermissionModal.value = true
}

async function handleSavePermission() {
  if (!session.currentSession || !targetPermStudent.value) return
  const sessId = session.currentSession.id || session.currentSession.ID
  try {
    await session.markPermission(
      sessId,
      targetPermStudent.value.student_id,
      permStatus.value,
      permNotes.value
    )
    showPermissionModal.value = false
    successMsg.value = `Keterangan ${permStatus.value.toUpperCase()} untuk ${targetPermStudent.value.student_name} berhasil dicatat sah saat sesi absensi berlangsung.`
  } catch (err) {
    pageError.value = err.message
  }
}

async function handleRefreshRecap() {
  if (selectedSchedule.value) {
    await session.fetchScheduleDetails(selectedSchedule.value.id)
    successMsg.value = 'Data rekapitulasi presensi diperbarui.'
  }
}

async function handleApprovalAction(perm, action) {
  if (!perm) return
  approvalActionLoading.value = true
  pageError.value = ''
  try {
    const isApprove = action === 'approve'
    const attId = perm.attendance_id || perm.id
    await session.approvePermission(attId, action)
    const studentName = perm.student_name || 'Mahasiswa'
    successMsg.value = isApprove 
      ? `Permohonan izin untuk ${studentName} telah disetujui.`
      : `Permohonan izin untuk ${studentName} telah ditolak (dicatat Alpa).`
    
    // Refresh attendees if active session is currently running
    if (session.currentSession) {
      await session.fetchAttendees(session.currentSession.id || session.currentSession.ID)
    }
  } catch (err) {
    pageError.value = err.message
  } finally {
    approvalActionLoading.value = false
  }
}

function openAttachmentPreview(url) {
  previewAttachmentUrl.value = url
}

function exportRecapToCSV() {
  if (!session.classRecap || !session.classRecap.students?.length) {
    pageError.value = 'Belum ada data rekapitulasi untuk diekspor.'
    return
  }

  const recap = session.classRecap
  const meetings = recap.meetings || []
  const courseName = currentCourse.value?.course_name || 'Mata-Kuliah'
  const classUnit = currentCourse.value?.class_unit || 'Kelas'

  const rows = [
    [`REKAPITULASI PRESENSI MAHASISWA`],
    [`Mata Kuliah: ${courseName} (${classUnit})`],
    [`Tanggal Cetak: ${new Date().toLocaleDateString('id-ID')} ${new Date().toLocaleTimeString('id-ID')}`],
    [],
    [
      'No',
      'NIM',
      'Nama Mahasiswa',
      ...meetings.map(m => `P${m}`),
      'Hadir (H)',
      'Terlambat (T)',
      'Izin (I)',
      'Sakit (S)',
      'Alpa (A)',
      'Persentase Hadir (%)'
    ]
  ]

  recap.students.forEach((st, idx) => {
    const meetingStatuses = meetings.map(m => {
      const s = st.meeting_statuses[m]
      if (s === 'hadir') return 'H'
      if (s === 'terlambat') return 'T'
      if (s === 'izin') return 'I'
      if (s === 'sakit') return 'S'
      if (s === 'alpa') return 'A'
      return '-'
    })

    rows.push([
      idx + 1,
      `'${st.student_nim}`,
      `"${st.student_name.replace(/"/g, '""')}"`,
      ...meetingStatuses,
      st.hadir_count,
      st.terlambat_count,
      st.izin_count,
      st.sakit_count,
      st.alpa_count,
      `${st.percentage}%`
    ])
  })

  const csvContent = '\uFEFF' + rows.map(r => r.join(';')).join('\r\n')
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const filename = `Rekap-Presensi-${courseName.replace(/\s+/g, '_')}-${classUnit}-${new Date().toISOString().slice(0, 10)}.csv`

  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = filename
  link.click()
  URL.revokeObjectURL(link.href)

  successMsg.value = `Rekapitulasi berhasil diekspor ke ${filename}`
}

function printOfficialRecap() {
  window.print()
}

function getStudentStatusBadgeClass(studentId) {
  const att = attendanceByStudentId.value[studentId]
  if (!att) {
    return session.currentSession ? 'badge-muted' : 'badge-subtle'
  }
  if (att.status === 'hadir') return 'badge-success'
  if (att.status === 'terlambat') return 'badge-warning'
  if (att.status === 'izin' || att.status === 'sakit') return 'badge-info'
  return 'badge-danger'
}

function getStudentStatusIcon(studentId) {
  const att = attendanceByStudentId.value[studentId]
  if (!att) {
    return session.currentSession ? 'fas fa-hourglass-start' : 'fas fa-minus'
  }
  if (att.status === 'hadir') return 'fas fa-check-circle'
  if (att.status === 'terlambat') return 'fas fa-clock'
  if (att.status === 'izin') return 'fas fa-envelope-open-text'
  if (att.status === 'sakit') return 'fas fa-medkit'
  return 'fas fa-times-circle'
}

function getStudentStatusText(studentId) {
  const att = attendanceByStudentId.value[studentId]
  if (!att) {
    return session.currentSession ? 'Belum Presensi' : 'Belum Mulai'
  }
  if (att.status === 'hadir') return 'Hadir Tepat Waktu'
  if (att.status === 'terlambat') return 'Terlambat'
  if (att.status === 'izin') return 'Izin Resmi'
  if (att.status === 'sakit') return 'Sakit'
  return 'Alpa'
}

function getStudentNotesOrTime(studentId) {
  const att = attendanceByStudentId.value[studentId]
  if (!att) return '-'
  const timeStr = att.scanned_at ? `${formatTime(att.scanned_at)} WIB` : ''
  const dist = att.distance_meters ? ` (${Math.round(att.distance_meters)}m)` : ''

  if (timeStr && att.notes) {
    return `${timeStr} — ${att.notes}`
  }
  if (timeStr) {
    return `${timeStr}${dist}`
  }
  if (att.notes) {
    return att.notes
  }
  return '-'
}

function getRecapBadgeClass(status) {
  if (status === 'hadir') return 'recap-h'
  if (status === 'terlambat') return 'recap-t'
  if (status === 'izin') return 'recap-i'
  if (status === 'sakit') return 'recap-s'
  return 'recap-a'
}

function getStatusLetter(status) {
  if (status === 'hadir') return 'H'
  if (status === 'terlambat') return 'T'
  if (status === 'izin') return 'I'
  if (status === 'sakit') return 'S'
  return 'A'
}

function getPctClass(pct) {
  if (pct >= 80) return 'pct-high'
  if (pct >= 60) return 'pct-mid'
  return 'pct-low'
}

function formatTime(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<style scoped>
.dosen-portal {
  min-height: 100vh;
  background-color: var(--bg-page);
}

.portal-main {
  max-width: 1320px;
  margin: 0 auto;
  padding: 0 24px 40px;
}

/* Alert Boxes */
.alert-box {
  background-color: var(--status-danger-bg);
  border: 1px solid var(--status-danger-border);
  color: var(--status-danger-text);
  padding: 12px 18px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.9rem;
}

.alert-box.success-box {
  background-color: var(--status-success-bg);
  border-color: var(--status-success-border);
  color: var(--status-success-text);
}

.alert-icon {
  font-size: 1.1rem;
}

.alert-content {
  flex: 1;
}

.btn-alert-close {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 4px;
}

/* VIEW 1: CATALOGUE */
.catalogue-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.catalogue-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  flex-wrap: wrap;
  gap: 16px;
}

.page-title {
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--text-main);
}

.page-subtitle {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin-top: 4px;
}

.filter-pills {
  display: flex;
  gap: 8px;
  background-color: var(--bg-surface-secondary);
  padding: 4px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-subtle);
}

.pill-btn {
  background: transparent;
  border: none;
  padding: 6px 14px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  border-radius: var(--radius-full);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: all 0.15s ease;
}

.pill-btn.active {
  background-color: var(--brand-primary);
  color: #ffffff;
}

.count-bubble {
  background-color: rgba(255, 255, 255, 0.25);
  font-size: 0.75rem;
  padding: 1px 7px;
  border-radius: var(--radius-full);
}

.pill-btn:not(.active) .count-bubble {
  background-color: var(--border-subtle);
  color: var(--text-main);
}

/* Courses Grid */
.courses-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
}

.course-card {
  padding: 22px 24px;
  cursor: pointer;
  transition: transform 0.15s ease, border-color 0.15s ease;
  display: flex;
  flex-direction: column;
}

.course-card:hover {
  transform: translateY(-2px);
  border-color: var(--brand-primary);
}

.card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.code-and-unit {
  display: flex;
  align-items: center;
  gap: 8px;
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

.status-chip {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: var(--radius-full);
}

.status-chip.active-chip {
  background-color: var(--brand-accent-light);
  color: #854d0e;
  border: 1px solid var(--brand-accent);
}

.status-chip.today-chip {
  background-color: var(--brand-primary-light);
  color: var(--brand-primary);
  border: 1px solid var(--brand-primary-border);
}

.status-chip.regular-chip {
  background-color: var(--bg-surface-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
}

.card-course-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.35;
  margin-bottom: 14px;
  min-height: 48px;
}

.card-meta-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 18px;
  border-top: 1px solid var(--border-subtle);
  padding-top: 14px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.meta-item i {
  color: var(--brand-primary);
  width: 14px;
}

.card-footer {
  margin-top: auto;
  border-top: 1px solid var(--border-subtle);
  padding-top: 12px;
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--brand-primary);
  text-align: right;
}

.empty-catalogue {
  text-align: center;
  padding: 60px 24px;
}

/* VIEW 2: CLASS DETAIL */
.class-detail-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.detail-nav-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 14px;
}

.class-breadcrumb {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bc-code {
  font-weight: 800;
  background-color: var(--brand-primary);
  color: #ffffff;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.8rem;
}

.bc-name {
  font-weight: 700;
  color: var(--text-main);
  font-size: 1rem;
}

.detail-tab-bar {
  display: flex;
  border-bottom: 2px solid var(--border-subtle);
  gap: 24px;
}

.detail-tab-btn {
  background: transparent;
  border: none;
  padding: 12px 4px;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border-bottom: 3px solid transparent;
  margin-bottom: -2px;
  transition: all 0.15s ease;
}

.detail-tab-btn.active {
  color: var(--brand-primary);
  border-bottom-color: var(--brand-primary);
}

.tab-live-badge {
  background-color: #dc2626;
  color: #ffffff;
  font-size: 0.65rem;
  font-weight: 800;
  padding: 2px 6px;
  border-radius: var(--radius-full);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Control Bar */
.control-bar {
  padding: 20px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  gap: 20px;
  flex-wrap: wrap;
}

.course-header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.course-badge {
  background-color: var(--brand-primary);
  color: #ffffff;
  font-weight: 800;
  font-size: 0.95rem;
  padding: 8px 14px;
  border-radius: var(--radius-sm);
}

.course-name {
  font-size: 1.25rem;
  color: var(--text-main);
  line-height: 1.3;
}

.course-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.825rem;
  color: var(--text-muted);
  margin-top: 4px;
  flex-wrap: wrap;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: #dc2626;
}

.status-indicator.is-active {
  color: var(--brand-primary);
}

.status-dot {
  font-size: 0.5rem;
}

.action-group {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.auto-meeting-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background-color: var(--brand-primary-light);
  color: var(--brand-primary);
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 700;
  border: 1px solid var(--brand-primary-border);
}

.not-today-warning {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background-color: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  font-weight: 600;
}

.schedule-notice-box {
  padding: 16px 20px;
  margin-bottom: 20px;
  background-color: var(--brand-primary-light);
  border: 1px solid var(--brand-primary-border);
  display: flex;
  gap: 14px;
  align-items: flex-start;
  font-size: 0.875rem;
  color: var(--brand-primary);
}

.notice-icon {
  font-size: 1.25rem;
  margin-top: 2px;
}

/* Workspace layout */
.presensi-workspace {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.presensi-workspace.with-projector {
  display: grid;
  grid-template-columns: 460px 1fr;
  align-items: start;
}

@media (max-width: 1024px) {
  .presensi-workspace.with-projector {
    grid-template-columns: 1fr;
  }
}

/* Projector Card */
.projector-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.projector-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  text-align: left;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 700;
}

.section-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  background-color: var(--status-danger-bg);
  color: var(--status-danger-text);
  border: 1px solid var(--status-danger-border);
}

.status-pill.connected {
  background-color: var(--status-success-bg);
  color: var(--status-success-text);
  border-color: var(--status-success-border);
}

.qr-display-box {
  background-color: #ffffff;
  padding: 20px;
  border-radius: var(--radius-sm);
  border: 2px solid var(--border-subtle);
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-loading {
  width: 260px;
  height: 260px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-muted);
}

.qr-action-panel {
  width: 100%;
  background-color: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  margin-bottom: 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.static-token-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}

.qr-secure-hint {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.session-active-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--brand-primary);
}

.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--brand-primary);
  display: inline-block;
  box-shadow: 0 0 0 0 rgba(0, 102, 51, 0.4);
  animation: pulseGreen 2s infinite;
}

@keyframes pulseGreen {
  0% { box-shadow: 0 0 0 0 rgba(0, 102, 51, 0.7); }
  70% { box-shadow: 0 0 0 8px rgba(0, 102, 51, 0); }
  100% { box-shadow: 0 0 0 0 rgba(0, 102, 51, 0); }
}

.qr-action-buttons {
  display: flex;
  gap: 8px;
}

.qr-action-btn {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 7px 12px;
  font-size: 0.85rem;
  font-weight: 600;
}

.share-toast-msg {
  background-color: var(--brand-primary-light);
  color: var(--brand-primary);
  border: 1px solid var(--brand-primary);
  border-radius: 4px;
  padding: 6px 10px;
  font-size: 0.8rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
  animation: fadeIn 0.2s ease-in;
}

.security-flags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.sec-item {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  background-color: var(--bg-surface-secondary);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-subtle);
}

/* Students Panel Card */
.students-panel-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 16px;
}

.status-summary-bar {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.sum-chip {
  font-size: 0.775rem;
  padding: 4px 10px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-subtle);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.sum-chip.total-chip { background-color: var(--bg-surface-secondary); color: var(--text-main); font-weight: 600; }
.sum-chip.hadir-chip { background-color: var(--status-success-bg); color: var(--status-success-text); border-color: var(--status-success-border); }
.sum-chip.telat-chip { background-color: var(--status-warning-bg); color: var(--status-warning-text); border-color: var(--status-warning-border); }
.sum-chip.izin-chip { background-color: #eff6ff; color: #1e40af; border-color: #bfdbfe; }
.sum-chip.belum-chip { background-color: #f8fafc; color: #64748b; border-color: #e2e8f0; }

/* Search Toolbar */
.search-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
  gap: 16px;
  flex-wrap: wrap;
}

.search-box-wrapper {
  position: relative;
  flex: 1;
  max-width: 420px;
}

.search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  font-size: 0.85rem;
}

.search-student-input {
  width: 100%;
  padding-left: 36px;
  padding-right: 32px;
  font-size: 0.85rem;
}

.clear-search-btn {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
}

.search-hint {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.table-container {
  flex: 1;
  overflow-y: auto;
  max-height: 560px;
}

.font-mono {
  font-family: monospace;
  font-size: 0.85rem;
}

.font-semibold {
  font-weight: 600;
  color: var(--text-main);
}

.meta-cell {
  font-size: 0.825rem;
}

.btn-outline-primary {
  background: transparent;
  border: 1px solid var(--brand-primary);
  color: var(--brand-primary);
  padding: 5px 12px;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s ease;
}

.btn-outline-primary:hover:not(:disabled) {
  background-color: var(--brand-primary);
  color: #ffffff;
}

.btn-outline-primary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Badge variants */
.badge-subtle {
  background-color: var(--bg-surface-secondary);
  color: var(--text-muted);
  border: 1px solid var(--border-subtle);
}

/* Recap Table & Badges */
.recap-card {
  padding: 24px;
}

.recap-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
  flex-wrap: wrap;
  gap: 12px;
}

.recap-legend {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  padding: 10px 14px;
  background-color: var(--bg-surface-secondary);
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  color: var(--text-muted);
  flex-wrap: wrap;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.recap-table th {
  font-size: 0.8rem;
}

.meeting-th {
  min-width: 42px;
  text-align: center;
}

.col-counter {
  width: 36px;
  text-align: center;
}

.recap-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 800;
}

.recap-h { background-color: var(--status-success-bg); color: var(--status-success-text); border: 1px solid var(--status-success-border); }
.recap-t { background-color: var(--status-warning-bg); color: var(--status-warning-text); border: 1px solid var(--status-warning-border); }
.recap-i { background-color: #eff6ff; color: #1e40af; border: 1px solid #bfdbfe; }
.recap-s { background-color: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }
.recap-a { background-color: var(--status-danger-bg); color: var(--status-danger-text); border: 1px solid var(--status-danger-border); }

.pct-chip {
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-weight: 800;
  font-size: 0.8rem;
}

.pct-high { background-color: var(--status-success-bg); color: var(--status-success-text); }
.pct-mid { background-color: var(--status-warning-bg); color: var(--status-warning-text); }
.pct-low { background-color: var(--status-danger-bg); color: var(--status-danger-text); }

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 3rem;
  color: var(--border-medium);
  margin-bottom: 14px;
}

.empty-title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-main);
}

.empty-desc {
  max-width: 440px;
  margin: 6px auto 0;
  font-size: 0.825rem;
}

/* Modals */
.modal-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(15, 23, 42, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  z-index: 1000;
}

.modal-box {
  width: 100%;
  max-width: 480px;
  background-color: #ffffff;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-dropdown);
  overflow: hidden;
}

.modal-header {
  padding: 18px 24px;
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-main);
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1rem;
  cursor: pointer;
  padding: 4px;
}

.btn-close:hover {
  color: #dc2626;
}

.modal-body {
  padding: 20px 24px;
}

.modal-intro {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.target-student-box {
  background-color: var(--bg-surface-secondary);
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.student-name {
  font-weight: 700;
  color: var(--text-main);
}

.student-nim {
  font-size: 0.8rem;
  color: var(--text-muted);
  font-family: monospace;
}

.modal-footer {
  padding: 14px 24px;
  background-color: var(--bg-surface-secondary);
  border-top: 1px solid var(--border-subtle);
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* Open Session & Duration Controls */
.open-session-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: var(--bg-surface-secondary);
  padding: 14px 18px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.duration-selector-box {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.duration-label {
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--text-main);
  display: flex;
  align-items: center;
  gap: 6px;
}

.duration-presets {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.duration-chip {
  background: #ffffff;
  border: 1px solid var(--border-subtle);
  color: var(--text-main);
  padding: 5px 14px;
  font-size: 0.825rem;
  font-weight: 600;
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: all 0.15s ease;
}

.duration-chip:hover {
  border-color: var(--brand-primary);
  color: var(--brand-primary);
}

.duration-chip.active {
  background-color: var(--brand-primary);
  border-color: var(--brand-primary);
  color: #ffffff;
}

.open-btn-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

/* Calendar Locked Card (Academic Rule) */
.calendar-locked-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  padding: 12px 18px;
  border-radius: var(--radius-sm);
  max-width: 600px;
}

.calendar-locked-card i {
  font-size: 1.25rem;
  color: var(--brand-primary);
  margin-top: 2px;
}

.calendar-locked-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.calendar-locked-text strong {
  font-size: 0.9rem;
  color: #166534;
}

.calendar-locked-text span {
  font-size: 0.825rem;
  color: #374151;
  line-height: 1.4;
}

/* Session Timer Pill */
.active-session-action-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.session-timer-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fefce8;
  border: 1px solid #fde047;
  color: #854d0e;
  padding: 8px 14px;
  border-radius: var(--radius-full);
  font-size: 0.85rem;
  font-weight: 600;
}

.session-timer-pill i {
  color: #ca8a04;
}

/* Portal Switcher & Approval Counter */
.portal-switch-pills {
  display: flex;
  gap: 8px;
  background-color: var(--bg-surface-secondary);
  padding: 4px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-subtle);
}

.approval-pill {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.pending-bubble {
  background-color: #ef4444;
  color: #ffffff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: var(--radius-full);
  line-height: 1.2;
}

.tab-badge {
  background-color: #ef4444;
  color: #ffffff;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: var(--radius-full);
  margin-left: 6px;
}

/* Approval Center Cards & Layout */
.approval-card {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.approval-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.approval-header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.approval-stats-bar {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.stat-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  color: var(--text-muted);
}

.stat-pill strong {
  font-size: 1rem;
  color: var(--text-main);
}

.stat-pill.pending strong {
  color: #d97706;
}

.stat-pill.approved strong {
  color: #16a34a;
}

.stat-pill.rejected strong {
  color: #dc2626;
}

.approval-tabs-row {
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 8px;
}

.tab-filter-btn {
  background: transparent;
  border: none;
  padding: 6px 14px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: var(--radius-full);
  transition: all 0.15s ease;
}

.tab-filter-btn:hover {
  background-color: var(--bg-surface-secondary);
  color: var(--text-main);
}

.tab-filter-btn.active {
  background-color: var(--brand-primary);
  color: #ffffff;
}

/* Source Channels Badges */
.source-tag-wa {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #16a34a;
  background-color: #dcfce7;
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.source-tag-app {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #2563eb;
  background-color: #dbeafe;
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.source-tag-scan {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #475569;
  background-color: #f1f5f9;
  padding: 2px 8px;
  border-radius: var(--radius-full);
}

.source-sub-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 0.7rem;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  margin-top: 3px;
  width: fit-content;
}

.source-sub-badge.wa {
  background-color: #dcfce7;
  color: #15803d;
}

.source-sub-badge.app {
  background-color: #dbeafe;
  color: #1d4ed8;
}

.source-sub-badge.scan {
  background-color: #f1f5f9;
  color: #475569;
}

.status-cell-wrap {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* Rolling QR Progress Bar */
.rolling-progress-wrap {
  width: 100%;
  height: 6px;
  background-color: var(--bg-surface-secondary);
  border-radius: var(--radius-full);
  overflow: hidden;
  margin-bottom: 8px;
}

.rolling-progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #10b981, #3b82f6);
  transition: width 1s linear;
}

.rolling-info-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  font-size: 0.8rem;
  margin-bottom: 12px;
}

.rolling-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: var(--text-main);
}

.pulse-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: #10b981;
  box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  animation: pulseLive 1.5s infinite;
}

@keyframes pulseLive {
  0% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7);
  }
  70% {
    transform: scale(1);
    box-shadow: 0 0 0 6px rgba(16, 185, 129, 0);
  }
  100% {
    transform: scale(0.95);
    box-shadow: 0 0 0 0 rgba(16, 185, 129, 0);
  }
}

.rolling-countdown strong {
  color: #2563eb;
  font-weight: 700;
}

/* Form Helpers & Character Counter */
.label-with-counter {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.char-counter {
  font-size: 0.75rem;
  font-weight: 600;
}

.char-counter.valid {
  color: #16a34a;
}

.char-counter.invalid {
  color: #d97706;
}

/* Preset Shortcuts */
.preset-shortcuts {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 10px;
}

.preset-label {
  font-size: 0.8rem;
  color: var(--text-muted);
  font-weight: 600;
}

.preset-tag {
  background: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-full);
  padding: 4px 10px;
  font-size: 0.775rem;
  font-weight: 600;
  cursor: pointer;
  color: var(--text-main);
  display: inline-flex;
  align-items: center;
  gap: 5px;
  transition: all 0.15s ease;
}

.preset-tag:hover {
  border-color: var(--brand-primary);
  background-color: #ffffff;
  color: var(--brand-primary);
}

/* Live & Inactive Session Tip Banners */
.session-tip-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  margin-bottom: 14px;
}

.live-session-tip {
  background-color: #ecfdf5;
  border: 1px solid #a7f3d0;
  color: #065f46;
}

.live-session-tip .pulse-icon {
  color: #059669;
  font-size: 1.1rem;
  animation: pulseLive 1.8s infinite;
}

.inactive-session-tip {
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #64748b;
}

.inactive-session-tip i {
  color: #94a3b8;
  font-size: 1.1rem;
}

.tip-text {
  flex: 1;
  line-height: 1.45;
}

/* Status Filter Pills Toolbar */
.student-status-filter-pills {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.s-pill-btn {
  background: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-full);
  padding: 5px 12px;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.s-pill-btn:hover {
  background-color: #ffffff;
  color: var(--text-main);
  border-color: var(--border-medium);
}

.s-pill-btn.active {
  background-color: var(--brand-primary);
  color: #ffffff;
  border-color: var(--brand-primary);
}

.s-pill-btn.s-belum.active {
  background-color: #475569;
  color: #ffffff;
  border-color: #475569;
}

.s-pill-btn.s-hadir.active {
  background-color: #16a34a;
  color: #ffffff;
  border-color: #16a34a;
}

.s-pill-btn.s-izin.active {
  background-color: #d97706;
  color: #ffffff;
  border-color: #d97706;
}

/* Action Buttons & Cell */
.btn-row-izin {
  color: #d97706;
  border-color: #fcd34d;
  background-color: #fffbeb;
}

.btn-row-izin:hover:not(:disabled) {
  background-color: #d97706;
  color: #ffffff;
  border-color: #d97706;
}

.btn-row-sakit {
  color: #2563eb;
  border-color: #bfdbfe;
  background-color: #eff6ff;
}

.btn-row-sakit:hover:not(:disabled) {
  background-color: #2563eb;
  color: #ffffff;
  border-color: #2563eb;
}

.btn-row-edit {
  color: #475569;
  border-color: #cbd5e1;
  background-color: #f8fafc;
}

.btn-row-edit:hover:not(:disabled) {
  background-color: #475569;
  color: #ffffff;
  border-color: #475569;
}

.locked-action-cell {
  font-size: 0.775rem;
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
  gap: 5px;
  opacity: 0.8;
}

.btn-group-row {
  display: flex;
  gap: 6px;
}

.btn-xs {
  padding: 3px 8px;
  font-size: 0.75rem;
}

.img-attachment-preview {
  max-width: 100%;
  max-height: 400px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
  object-fit: contain;
}

/* Official Print Stylesheet (BR-DSN Section 6.2 #3) */
@media print {
  body {
    background: #ffffff !important;
    color: #000000 !important;
  }

  .navbar,
  .catalogue-section,
  .back-bar,
  .view-tabs,
  .recap-btn-actions,
  .recap-legend,
  .open-session-panel,
  .projector-card,
  .students-panel-card,
  .modal-overlay,
  .alert-box {
    display: none !important;
  }

  .dosen-portal,
  .portal-main,
  .tab-pane,
  .recap-card {
    padding: 0 !important;
    margin: 0 !important;
    max-width: 100% !important;
    box-shadow: none !important;
    border: none !important;
    background: transparent !important;
  }

  .recap-table {
    width: 100% !important;
    border-collapse: collapse !important;
  }

  .recap-table th,
  .recap-table td {
    border: 1px solid #333333 !important;
    color: #000000 !important;
    padding: 6px 8px !important;
    font-size: 10pt !important;
  }

  .recap-badge {
    background: transparent !important;
    color: #000000 !important;
    border: 1px solid #666666 !important;
  }
}

/* Hadir Manual Action Button in Table */
.btn-row-hadir {
  color: #166534;
  border-color: #86efac;
  background-color: #f0fdf4;
}

.btn-row-hadir:hover {
  background-color: #16a34a;
  color: #ffffff;
  border-color: #16a34a;
}

/* ========================================================== */
/* FULLSCREEN PROJECTOR QR OVERLAY                           */
/* ========================================================== */
.fullscreen-qr-overlay {
  position: fixed;
  inset: 0;
  z-index: 99999;
  background-color: rgba(5, 10, 18, 0.96);
  backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  animation: fadeIn 0.25s ease-out;
}

.fullscreen-qr-container {
  position: relative;
  max-width: 720px;
  width: 100%;
  background: #0f172a;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 28px;
  padding: 36px 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  box-shadow: 0 25px 60px -15px rgba(0, 0, 0, 0.7), 0 0 50px rgba(14, 165, 233, 0.15);
  text-align: center;
}

.btn-fullscreen-close {
  position: absolute;
  top: 20px;
  right: 20px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #e2e8f0;
  padding: 8px 16px;
  border-radius: 9999px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
}

.btn-fullscreen-close:hover {
  background: #ef4444;
  border-color: #ef4444;
  color: #ffffff;
  transform: scale(1.04);
}

.fullscreen-qr-header {
  margin-bottom: 24px;
}

.univ-brand-tag {
  display: inline-block;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 1.5px;
  color: #38bdf8;
  text-transform: uppercase;
  background: rgba(56, 189, 248, 0.12);
  padding: 4px 14px;
  border-radius: 9999px;
  margin-bottom: 12px;
  border: 1px solid rgba(56, 189, 248, 0.25);
}

.fullscreen-course-title {
  font-size: 26px;
  font-weight: 800;
  color: #ffffff;
  margin: 0 0 10px 0;
  line-height: 1.3;
}

.fullscreen-course-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 14px;
  color: #94a3b8;
}

.meta-separator {
  color: #475569;
}

.fullscreen-qr-canvas-box {
  background: #ffffff;
  padding: 24px;
  border-radius: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 12px 35px rgba(0, 0, 0, 0.4);
  margin-bottom: 24px;
}

.fullscreen-qr-canvas-box canvas {
  display: block;
  max-width: 100%;
  height: auto;
}

.fullscreen-qr-footer {
  width: 100%;
  max-width: 480px;
}

.fullscreen-rolling-bar {
  width: 100%;
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 9999px;
  overflow: hidden;
  margin-bottom: 12px;
}

.fullscreen-rolling-fill {
  height: 100%;
  background: linear-gradient(90deg, #10b981, #06b6d4);
  border-radius: 9999px;
  transition: width 1s linear;
}

.fullscreen-rolling-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: #cbd5e1;
  margin-bottom: 14px;
}

.totp-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #34d399;
  font-weight: 600;
}

.totp-countdown strong {
  color: #38bdf8;
  font-size: 15px;
}

.fullscreen-instruction {
  font-size: 13.5px;
  color: #94a3b8;
  margin: 0;
  line-height: 1.5;
}

.fullscreen-instruction strong {
  color: #ffffff;
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.98); }
  to { opacity: 1; transform: scale(1); }
}
</style>
