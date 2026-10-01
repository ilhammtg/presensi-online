<template>
  <div class="modal-backdrop" @click.self="$emit('close')">
    <div class="flat-modal profile-modal">
      <div class="modal-header">
        <div class="header-title-box">
          <i class="fas fa-user-circle modal-icon"></i>
          <div>
            <h3 class="modal-title">Profil & Pengaturan Akun</h3>
            <p class="modal-sub">Kelola foto profil dan kata sandi Anda</p>
          </div>
        </div>
        <button class="btn-close-modal" @click="$emit('close')">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <!-- Navigation Tabs in Modal -->
      <div class="modal-tabs">
        <button 
          class="modal-tab-btn" 
          :class="{ 'active': activeTab === 'info' }"
          @click="activeTab = 'info'"
        >
          <i class="fas fa-id-card"></i>
          <span>Informasi Akademik & Foto</span>
        </button>
        <button 
          class="modal-tab-btn" 
          :class="{ 'active': activeTab === 'password' }"
          @click="activeTab = 'password'"
        >
          <i class="fas fa-key"></i>
          <span>Ganti Kata Sandi</span>
        </button>
      </div>

      <div class="modal-body">
        <!-- Toast / Alerts -->
        <div v-if="successMsg" class="alert alert-success">
          <i class="fas fa-check-circle"></i>
          <span>{{ successMsg }}</span>
        </div>
        <div v-if="errorMsg" class="alert alert-danger">
          <i class="fas fa-exclamation-circle"></i>
          <span>{{ errorMsg }}</span>
        </div>

        <!-- TAB 1: INFORMASI AKADEMIK & FOTO -->
        <div v-if="activeTab === 'info'" class="tab-content">
          <!-- Photo Section -->
          <div class="photo-section">
            <div class="avatar-large-wrapper">
              <img v-if="previewPhoto || auth.user?.avatar_url" :src="previewPhoto || auth.user?.avatar_url" alt="Avatar" class="avatar-large" />
              <div v-else class="avatar-large default-avatar">
                <i class="fas fa-user-tie"></i>
              </div>
            </div>
            <div class="photo-actions">
              <label class="btn btn-outline-primary btn-sm upload-btn">
                <i class="fas fa-camera"></i>
                <span>Pilih Foto Baru</span>
                <input type="file" accept="image/*" class="file-input-hidden" @change="handleFileSelected" />
              </label>
              <button 
                v-if="previewPhoto" 
                class="btn btn-primary btn-sm" 
                @click="handleSavePhoto" 
                :disabled="isSavingPhoto"
              >
                <i v-if="isSavingPhoto" class="fas fa-circle-notch fa-spin"></i>
                <i v-else class="fas fa-save"></i>
                <span>Simpan Foto</span>
              </button>
              <button 
                v-if="previewPhoto" 
                class="btn btn-outline-dark btn-sm" 
                @click="cancelPhotoPreview"
              >
                Batal
              </button>
            </div>
          </div>

          <!-- Academic Information (Read-Only) -->
          <div class="academic-info-card">
            <div class="academic-badge-banner">
              <i class="fas fa-shield-alt"></i>
              <span>Data Resmi SIAKAD (Terkunci & Read-Only)</span>
            </div>

            <div class="info-grid">
              <div class="info-group">
                <label>Nama Lengkap</label>
                <div class="readonly-input">
                  <i class="fas fa-lock input-lock-icon"></i>
                  <span>{{ auth.user?.name || '-' }}</span>
                </div>
              </div>

              <div class="info-group">
                <label>{{ auth.isDosen ? 'NIDN / NIP' : 'NIM / Identitas' }}</label>
                <div class="readonly-input">
                  <i class="fas fa-lock input-lock-icon"></i>
                  <span class="font-mono">{{ auth.user?.external_id || '-' }}</span>
                </div>
              </div>

              <div class="info-group">
                <label>Email Kampus</label>
                <div class="readonly-input">
                  <i class="fas fa-lock input-lock-icon"></i>
                  <span>{{ auth.user?.email || '-' }}</span>
                </div>
              </div>

              <div class="info-group">
                <label>Peran Akun</label>
                <div class="readonly-input">
                  <i class="fas fa-lock input-lock-icon"></i>
                  <span>{{ auth.isDosen ? 'Dosen Pengampu' : 'Administrator' }}</span>
                </div>
              </div>

              <div class="info-group">
                <label>Program Studi</label>
                <div class="readonly-input">
                  <i class="fas fa-lock input-lock-icon"></i>
                  <span>{{ auth.user?.prodi_name || 'Teknik Informatika' }}</span>
                </div>
              </div>

              <div class="info-group">
                <label>Fakultas</label>
                <div class="readonly-input">
                  <i class="fas fa-lock input-lock-icon"></i>
                  <span>{{ auth.user?.faculty_name || 'Fakultas Ilmu Komputer' }}</span>
                </div>
              </div>
            </div>

            <p class="lock-explainer">
              <i class="fas fa-info-circle"></i> Seluruh data akademik disinkronisasikan langsung dari pangkalan data SIAKAD institusi dan tidak dapat diubah oleh pengguna.
            </p>
          </div>
        </div>

        <!-- TAB 2: GANTI KATA SANDI -->
        <div v-if="activeTab === 'password'" class="tab-content">
          <form @submit.prevent="handleChangePassword" class="password-form">
            <div class="form-group">
              <label>Kata Sandi Saat Ini <span class="required">*</span></label>
              <div class="input-with-icon">
                <input 
                  :type="showOldPass ? 'text' : 'password'" 
                  v-model="oldPassword" 
                  class="form-control" 
                  placeholder="Masukkan kata sandi saat ini" 
                  required 
                />
                <button type="button" class="btn-toggle-eye" @click="showOldPass = !showOldPass">
                  <i class="fas" :class="showOldPass ? 'fa-eye-slash' : 'fa-eye'"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <label>Kata Sandi Baru <span class="required">* (Minimal 6 Karakter)</span></label>
              <div class="input-with-icon">
                <input 
                  :type="showNewPass ? 'text' : 'password'" 
                  v-model="newPassword" 
                  class="form-control" 
                  placeholder="Masukkan kata sandi baru" 
                  minlength="6"
                  required 
                />
                <button type="button" class="btn-toggle-eye" @click="showNewPass = !showNewPass">
                  <i class="fas" :class="showNewPass ? 'fa-eye-slash' : 'fa-eye'"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <label>Konfirmasi Kata Sandi Baru <span class="required">*</span></label>
              <div class="input-with-icon">
                <input 
                  :type="showConfirmPass ? 'text' : 'password'" 
                  v-model="confirmPassword" 
                  class="form-control" 
                  placeholder="Ulangi kata sandi baru" 
                  minlength="6"
                  required 
                />
                <button type="button" class="btn-toggle-eye" @click="showConfirmPass = !showConfirmPass">
                  <i class="fas" :class="showConfirmPass ? 'fa-eye-slash' : 'fa-eye'"></i>
                </button>
              </div>
            </div>

            <div class="form-actions">
              <button type="submit" class="btn btn-primary" :disabled="isChangingPass">
                <i v-if="isChangingPass" class="fas fa-circle-notch fa-spin"></i>
                <i v-else class="fas fa-key"></i>
                <span>Perbarui Kata Sandi</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits(['close'])
const auth = useAuthStore()

const activeTab = ref('info')
const successMsg = ref('')
const errorMsg = ref('')

// Photo state
const previewPhoto = ref('')
const isSavingPhoto = ref(false)

// Password state
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const showOldPass = ref(false)
const showNewPass = ref(false)
const showConfirmPass = ref(false)
const isChangingPass = ref(false)

onMounted(async () => {
  await auth.fetchProfile()
})

function handleFileSelected(event) {
  const file = event.target.files?.[0]
  if (!file) return

  if (file.size > 2 * 1024 * 1024) {
    errorMsg.value = 'Ukuran foto maksimal 2MB.'
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    previewPhoto.value = e.target.result
    errorMsg.value = ''
  }
  reader.readAsDataURL(file)
}

function cancelPhotoPreview() {
  previewPhoto.value = ''
}

async function handleSavePhoto() {
  if (!previewPhoto.value) return
  isSavingPhoto.value = true
  errorMsg.value = ''
  successMsg.value = ''

  try {
    const msg = await auth.updateProfilePhoto(previewPhoto.value)
    previewPhoto.value = ''
    successMsg.value = msg || 'Foto profil berhasil diperbarui.'
    setTimeout(() => { successMsg.value = '' }, 4000)
  } catch (err) {
    errorMsg.value = err.message
  } finally {
    isSavingPhoto.value = false
  }
}

async function handleChangePassword() {
  errorMsg.value = ''
  successMsg.value = ''

  if (newPassword.value !== confirmPassword.value) {
    errorMsg.value = 'Konfirmasi kata sandi baru tidak cocok.'
    return
  }

  if (newPassword.value.length < 6) {
    errorMsg.value = 'Kata sandi minimal 6 karakter.'
    return
  }

  isChangingPass.value = true
  try {
    const msg = await auth.changePassword(oldPassword.value, newPassword.value, confirmPassword.value)
    successMsg.value = msg || 'Kata sandi berhasil diubah.'
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    setTimeout(() => { successMsg.value = '' }, 4000)
  } catch (err) {
    errorMsg.value = err.message
  } finally {
    isChangingPass.value = false
  }
}
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(15, 23, 42, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
  backdrop-filter: blur(2px);
}

.profile-modal {
  background-color: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  width: 100%;
  max-width: 640px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow: hidden;
  animation: modalScale 0.2s ease-out;
}

@keyframes modalScale {
  from { opacity: 0; transform: scale(0.96); }
  to { opacity: 1; transform: scale(1); }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
  background-color: var(--bg-surface-secondary);
}

.header-title-box {
  display: flex;
  align-items: center;
  gap: 12px;
}

.modal-icon {
  font-size: 1.8rem;
  color: var(--brand-primary);
}

.modal-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-dark);
  margin: 0;
}

.modal-sub {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 2px 0 0 0;
}

.btn-close-modal {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 1.1rem;
  cursor: pointer;
  padding: 6px;
  border-radius: 4px;
}

.btn-close-modal:hover {
  color: var(--text-dark);
  background-color: rgba(0, 0, 0, 0.05);
}

.modal-tabs {
  display: flex;
  border-bottom: 1px solid var(--border-subtle);
  background-color: #fafbfc;
}

.modal-tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-muted);
  background: none;
  border: none;
  border-bottom: 3px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.modal-tab-btn.active {
  color: var(--brand-primary);
  border-bottom-color: var(--brand-primary);
  background-color: var(--bg-surface);
}

.modal-body {
  padding: 20px;
  max-height: 75vh;
  overflow-y: auto;
}

.photo-section {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 20px;
}

.avatar-large-wrapper {
  flex-shrink: 0;
}

.avatar-large {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid var(--brand-accent);
}

.default-avatar {
  background-color: var(--brand-primary-light);
  color: var(--brand-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2.2rem;
}

.photo-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.upload-btn {
  position: relative;
  cursor: pointer;
}

.file-input-hidden {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.academic-info-card {
  background-color: var(--bg-surface-secondary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  padding: 14px;
}

.academic-badge-banner {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--brand-primary);
  background-color: var(--brand-primary-light);
  padding: 4px 10px;
  border-radius: var(--radius-full);
  margin-bottom: 14px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

@media (max-width: 600px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}

.info-group label {
  display: block;
  font-size: 0.76rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.readonly-input {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #f1f5f9;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 0.88rem;
  color: #334155;
  font-weight: 600;
}

.input-lock-icon {
  font-size: 0.75rem;
  color: #94a3b8;
}

.lock-explainer {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 12px;
  margin-bottom: 0;
  line-height: 1.4;
}

/* Password Form */
.password-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group label {
  display: block;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-dark);
  margin-bottom: 6px;
}

.required {
  color: #dc2626;
  font-size: 0.75rem;
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.input-with-icon .form-control {
  padding-right: 40px;
}

.btn-toggle-eye {
  position: absolute;
  right: 10px;
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 6px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}
</style>
