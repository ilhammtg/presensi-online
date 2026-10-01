# UI/UX Design System — Aplikasi Absensi Mahasiswa Universitas Almuslim

**Dokumen:** Frontend UI/UX Specification  
**Platform:** Mobile-first application  
**Target:** Mahasiswa, Dosen, dan Admin/Operator (bila diperlukan)  
**Design language:** Material 3, adapted for Universitas Almuslim  
**Primary visual direction:** Hijau, modern, akademik, bersih, profesional, dan human-centered  
**Status:** Master guideline untuk AI/frontend agent

---

## 1. Tujuan Produk

Aplikasi ini adalah aplikasi absensi akademik berbasis QR Code untuk Universitas Almuslim.

UX harus membuat pengguna dapat:

### Mahasiswa
- Login dengan akun kampus.
- Melihat ringkasan aktivitas akademik.
- Melihat daftar mata kuliah.
- Membuka detail mata kuliah.
- Melakukan scan QR untuk absensi.
- Melihat status kehadiran.
- Melihat riwayat absensi.
- Melihat KTM digital.
- Mengelola profil.
- Menerima informasi penting terkait perkuliahan.

### Dosen
- Login dengan akun kampus.
- Melihat jadwal mengajar hari ini.
- Melihat daftar mata kuliah yang diampu.
- Membuat/mengaktifkan QR absensi.
- Melihat mahasiswa dalam kelas.
- Melihat status kehadiran.
- Melihat rekapitulasi absensi.
- Mengelola profil.

---

# 2. Prinsip Desain Utama

Gunakan Material 3 sebagai dasar visual, tetapi jangan membuat aplikasi terlihat seperti template Material generik.

Referensi utama:
https://m3.material.io/

Prinsip:

1. **Clarity over decoration**
2. **Content first**
3. **Consistent spacing**
4. **Accessible contrast**
5. **Large enough touch targets**
6. **Predictable navigation**
7. **Progressive disclosure**
8. **Clear feedback after every important action**
9. **Minimal visual noise**
10. **No unnecessary gradients**
11. **No excessive glassmorphism**
12. **No giant decorative illustrations**
13. **No AI-generated-looking dashboard**

Aplikasi harus terasa seperti produk kampus yang benar-benar digunakan setiap hari, bukan landing page AI.

---

# 3. Visual Personality

Karakter visual:

- Academic
- Trustworthy
- Fresh
- Calm
- Modern
- Structured
- Friendly
- Professional

Hindari:

- Neon green
- Gradient hijau-biru berlebihan
- Glassmorphism
- Card bertumpuk terlalu banyak
- Shadow terlalu tebal
- Rounded corner ekstrem
- Icon random
- Typography terlalu dekoratif
- Dashboard penuh statistik yang tidak penting

---

# 4. Brand Color System

Gunakan hijau sebagai identitas utama Universitas Almuslim.

## Primary

```text
Primary:        #0B6B3A
Primary Dark:   #07502B
Primary Light:  #D7F0E1
```

Primary digunakan untuk:

- CTA utama
- Active navigation
- Primary button
- Link penting
- QR action
- Highlight utama

Jangan menggunakan primary untuk semua elemen.

---

## Secondary

```text
Secondary:      #4F6357
Secondary Light:#DCE8DF
```

Digunakan untuk elemen pendukung.

---

## Background

```text
Background:     #F8FAF8
Surface:        #FFFFFF
Surface Variant:#EEF3EF
```

Background tidak boleh terlalu putih menyilaukan.

---

## Text

```text
Text Primary:   #17201A
Text Secondary: #536057
Text Muted:     #707A73
Text Disabled:  #A4ADA7
```

Jangan gunakan pure black `#000000` untuk body text.

---

## Status Colors

### Success / Hadir

```text
Success:        #2E7D32
Success Surface:#E3F3E4
```

### Warning / Terlambat

```text
Warning:        #A15C00
Warning Surface:#FFF0D6
```

### Error / Tidak Hadir

```text
Error:          #BA1A1A
Error Surface:  #FFEDEA
```

### Info / Izin

```text
Info:           #1769AA
Info Surface:   #E3F1FC
```

Status color harus selalu disertai teks/icon. Jangan mengandalkan warna saja.

---

# 5. Gradient Policy

Gradient bukan elemen utama.

Default:

```text
No gradient
```

Jika digunakan:

- hanya untuk hero/header tertentu
- sangat subtle
- maksimal 2 warna yang masih satu keluarga
- tidak digunakan pada body background
- tidak digunakan pada setiap card

Contoh yang diperbolehkan:

```text
#0B6B3A → #138A4B
```

Opacity visual tetap tenang.

---

# 6. Typography

Gunakan font modern yang mudah dibaca.

Prioritas:

```text
Inter
Roboto
System UI
```

Jika project telah menggunakan font kampus tertentu, gunakan font tersebut secara konsisten.

## Type Scale

Gunakan skala berbasis 4 px/8 px dan hindari ukuran font acak.

### Display

```text
Display Large:  32 px / 40 px
Display Medium: 28 px / 36 px
Display Small:  24 px / 32 px
```

### Headline

```text
Headline Large: 22 px / 28 px
Headline Medium:20 px / 26 px
Headline Small: 18 px / 24 px
```

### Title

```text
Title Large:    18 px / 24 px
Title Medium:   16 px / 22 px
Title Small:    14 px / 20 px
```

### Body

```text
Body Large:     16 px / 24 px
Body Medium:    14 px / 20 px
Body Small:     12 px / 16 px
```

### Label

```text
Label Large:    14 px / 20 px
Label Medium:   12 px / 16 px
Label Small:    11 px / 16 px
```

Body text utama minimal 14 px. Hindari body 11–12 px kecuali metadata.

---

# 7. Font Weight

```text
Regular:    400
Medium:     500
SemiBold:   600
Bold:       700
```

Default:

- Body → 400
- Label → 500
- Title → 500–600
- Heading → 600–700

Jangan membuat semua teks bold.

---

# 8. Spacing System

Gunakan basis 4 px.

```text
4   = micro spacing
8   = compact spacing
12  = small spacing
16  = standard spacing
20  = comfortable spacing
24  = section spacing
32  = major section spacing
40  = large separation
48  = screen-level separation
```

Default screen padding:

```text
16 px horizontal
```

Pada layar tablet/large:

```text
24–32 px
```

---

# 9. Border Radius

Gunakan radius moderat.

```text
4 px   = small controls
8 px   = input / compact element
12 px  = buttons / standard cards
16 px  = prominent cards
20 px  = modal / bottom sheet
28 px  = pill / fully rounded controls
```

Jangan menggunakan radius 24–32 px pada semua card.

---

# 10. Elevation & Shadow

Gunakan elevation secara hemat.

Default:

```text
Level 0:
flat surface

Level 1:
subtle shadow / border

Level 2:
floating card

Level 3:
modal / prominent surface
```

Card biasa sebaiknya memakai:

- surface color
- border tipis
- shadow sangat subtle

Hindari shadow besar dan blur berlebihan.

---

# 11. Iconography

Gunakan satu icon family.

Prioritas:

- Material Symbols
- Material Icons

Jangan mencampur:

- Lucide
- Font Awesome
- Material
- custom icons

dalam satu UI tanpa alasan yang jelas.

Ukuran:

```text
16 px = inline
20 px = compact
24 px = default
32 px = prominent
48 px = empty state
```

Icon harus membantu pemahaman, bukan dekorasi.

---

# 12. Touch Target

Interactive element minimal:

```text
44 × 44 px
```

Ideal untuk mobile:

```text
48 × 48 px
```

Jangan membuat icon button 24×24 px sebagai area klik.

Icon boleh 24 px, tetapi container klik tetap minimal 44–48 px.

---

# 13. Navigation Architecture

Gunakan bottom navigation untuk mobile.

### Mahasiswa

```text
Home
Mata Kuliah
Absensi
KTM
Profil
```

Jika 5 item terasa terlalu padat, gunakan:

```text
Home
Mata Kuliah
Absensi
Profil
```

KTM dapat diakses dari Profil.

### Dosen

```text
Home
Jadwal
Kelas
Rekap
Profil
```

Navigation harus mempertahankan posisi pengguna ketika berpindah halaman.

---

# 14. Authentication — WAJIB

Saat aplikasi pertama kali dibuka, pengguna harus masuk ke halaman login.

Flow:

```text
Splash
   ↓
Authentication Check
   ↓
Belum login → Login
   ↓
Berhasil login
   ↓
Role Detection
   ↓
Mahasiswa / Dosen
   ↓
Home
```

Jika session masih valid:

```text
Splash
   ↓
Session Check
   ↓
Home
```

Jangan menampilkan login setiap aplikasi dibuka jika session masih valid.

---

# 15. Splash Screen

Splash sederhana.

Komponen:

- Logo Universitas Almuslim
- Nama aplikasi
- background hijau atau putih
- loading indicator hanya jika diperlukan

Durasi jangan dibuat panjang.

Hindari:

- animasi logo berlebihan
- particle effect
- gradient kompleks
- ilustrasi AI

---

# 16. Login Screen

Login harus menjadi halaman yang paling sederhana.

Layout:

```text
Logo
Nama aplikasi
Welcome message

Email / NIM / NIP
Password

[ Masuk ]

Lupa Password
```

Jika login menggunakan akun kampus:

```text
Masuk dengan Akun Universitas
```

Jangan menampilkan terlalu banyak informasi.

Password:

- show/hide button
- validasi inline
- error jelas

Contoh:

```text
Email/NIM tidak boleh kosong
Password minimal 8 karakter
```

Error harus muncul dekat field yang bermasalah.

---

# 17. Home Mahasiswa

Struktur:

```text
Greeting
Nama Mahasiswa

Today's Overview

[ Mata Kuliah Berikutnya ]

[ Status Kehadiran ]

Quick Actions
[ Scan QR ]

Recent Attendance

Upcoming Classes
```

Contoh greeting:

```text
Selamat pagi,
Ilham
```

Jangan menggunakan greeting terlalu besar.

---

# 18. Home Dosen

Struktur:

```text
Greeting

Hari ini
27 September 2026

Jadwal Hari Ini

[ Mata Kuliah ]

Quick Actions

[ Generate QR ]

Ringkasan
- Kelas hari ini
- Mahasiswa hadir
- Absensi menunggu

Upcoming Schedule
```

Dashboard tidak perlu memuat puluhan statistik.

---

# 19. Mata Kuliah — Mahasiswa

Setiap course card:

```text
[ Icon ]

Algoritma dan Pemrograman
IF203
Semester 3

Dosen:
Nama Dosen

Jadwal:
Senin · 08:00–09:40

Status:
Hadir 8/10

[ Lihat Detail ]
```

Card harus memiliki hierarchy jelas.

---

# 20. Mata Kuliah — Dosen

Course card:

```text
Algoritma dan Pemrograman
IF203

Senin · 08:00–09:40

32 Mahasiswa

[ Generate QR ]
[ Mahasiswa ]
[ Rekap ]
```

Jangan menaruh tiga button besar berdampingan di mobile.

Gunakan:

```text
Primary action
Secondary actions
```

atau menu bottom sheet.

---

# 21. Detail Mata Kuliah

Header:

```text
Nama Mata Kuliah
Kode
Dosen
Jadwal
```

Tab:

```text
Overview
Absensi
Mahasiswa
```

Untuk dosen:

```text
Overview
Absensi
Mahasiswa
Rekap
```

Gunakan tabs hanya jika konten memang berbeda.

---

# 22. QR Attendance — Mahasiswa

Flow:

```text
Mata Kuliah
 ↓
Scan QR
 ↓
Camera Permission
 ↓
Scanner
 ↓
QR detected
 ↓
Validation
 ↓
Attendance confirmation
 ↓
Success
```

Scanner screen:

- camera preview full area
- scan frame
- instruction
- flash button
- close button

Jangan menutupi kamera dengan terlalu banyak UI.

---

# 23. Attendance Confirmation

Setelah QR berhasil dibaca:

```text
Konfirmasi Absensi

Algoritma dan Pemrograman

Pertemuan 8
27 September 2026
08:00–09:40

Status:
Siap melakukan absensi

[ Konfirmasi Kehadiran ]
```

Jangan langsung submit jika tindakan dapat memiliki konsekuensi penting.

---

# 24. Attendance Success

Gunakan success state yang jelas:

```text
✓

Absensi berhasil

Algoritma dan Pemrograman
Pertemuan 8

Waktu:
08:17

Status:
Hadir
```

CTA:

```text
Kembali ke Mata Kuliah
```

Animasi success boleh digunakan, tetapi singkat.

---

# 25. Attendance Status

Gunakan kombinasi:

```text
Icon + Label + Color
```

### Hadir

```text
● Hadir
```

### Terlambat

```text
◷ Terlambat
```

### Izin

```text
ℹ Izin
```

### Sakit

```text
✚ Sakit
```

### Alpa

```text
! Tidak Hadir
```

Jangan menggunakan warna sebagai satu-satunya indikator.

---

# 26. Status Absensi dalam Card

Contoh:

```text
Pertemuan 8
27 September 2026

Hadir
08:17
```

Gunakan status chip/pill dengan warna surface yang lembut.

Contoh:

```text
Green surface + green text
```

bukan background hijau solid yang terlalu kuat.

---

# 27. Rekapitulasi Absensi

Dosen:

```text
Rekap Absensi

Algoritma dan Pemrograman

Hadir       28
Terlambat    2
Izin         1
Alpa         1

Attendance rate

████████████████░░ 90%

Daftar mahasiswa
```

Tambahkan filter:

```text
Semua
Hadir
Terlambat
Izin
Alpa
```

Jangan membuat chart kompleks jika tabel sudah cukup.

---

# 28. Student List

Mahasiswa dalam kelas:

```text
[Avatar] Ahmad Fauzan
        23.11.1234
        Hadir

[Avatar] Siti Rahma
        23.11.1235
        Terlambat
```

Avatar boleh menggunakan:

- foto mahasiswa
- initial

Jangan menggunakan avatar AI.

---

# 29. Generate QR — Dosen

Screen:

```text
Generate Absensi

Mata Kuliah
Algoritma dan Pemrograman

Pertemuan
8

Durasi QR
15 menit

[ Generate QR ]
```

Setelah generate:

```text
QR aktif

[        QR        ]

Scan oleh mahasiswa

Berakhir dalam
14:32

[ Perpanjang ]
[ Tutup Absensi ]
```

QR harus menjadi focal point.

---

# 30. QR Expired State

```text
QR Absensi Berakhir

Kode QR sudah tidak berlaku.

[ Generate QR Baru ]
```

Jangan menggunakan error red untuk kondisi normal seperti QR expired.

Gunakan neutral/warning state.

---

# 31. KTM Digital

Halaman KTM harus terlihat seperti kartu identitas resmi tetapi tidak mencoba menjadi replika dokumen fisik secara berlebihan.

Struktur:

```text
Logo Universitas Almuslim

KARTU TANDA MAHASISWA

Foto
Nama
NIM
Program Studi
Fakultas

QR / barcode bila diperlukan
```

Tambahkan tombol:

```text
Tampilkan KTM
```

Jika QR KTM dapat digunakan untuk validasi, tampilkan status keamanan secara jelas.

---

# 32. Profile

Struktur:

```text
Avatar

Nama
NIM
Program Studi

Account
- Data pribadi
- KTM
- Password

Preferences
- Notification
- Appearance

Support
- Bantuan
- Tentang aplikasi

[ Keluar ]
```

Logout harus ditempatkan di bagian bawah.

---

# 33. Empty State

Jangan menggunakan ilustrasi AI.

Contoh:

```text
Belum ada jadwal

Tidak ada jadwal perkuliahan untuk hari ini.
```

CTA jika diperlukan:

```text
Lihat Semua Mata Kuliah
```

Gunakan icon sederhana dari Material Symbols.

---

# 34. Loading State

Gunakan skeleton loading.

Contoh:

```text
████████████
██████
████████████████
```

Jangan menggunakan spinner pada seluruh halaman jika skeleton dapat digunakan.

Spinner digunakan untuk action singkat:

```text
[ Memproses... ]
```

---

# 35. Error State

Error harus actionable.

Buruk:

```text
Something went wrong
```

Lebih baik:

```text
Data mata kuliah belum dapat dimuat.

Periksa koneksi internet dan coba lagi.

[ Coba Lagi ]
```

---

# 36. Snackbar

Gunakan snackbar untuk feedback ringan.

Contoh:

```text
Absensi berhasil disimpan
```

atau:

```text
QR berhasil dibuat
```

Jangan menggunakan snackbar untuk informasi kritis yang harus dibaca lama.

---

# 37. Dialog

Gunakan dialog hanya untuk:

- destructive action
- confirmation
- permission-sensitive action
- tindakan yang tidak boleh terjadi tanpa konfirmasi

Contoh logout:

```text
Keluar dari akun?

Anda harus login kembali untuk menggunakan aplikasi.

[ Batal ] [ Keluar ]
```

---

# 38. Bottom Sheet

Gunakan bottom sheet untuk pilihan contextual.

Contoh pada course card:

```text
Algoritma dan Pemrograman

Generate QR
Daftar Mahasiswa
Rekap Absensi
```

Ini lebih baik daripada menjejalkan banyak button ke dalam card.

---

# 39. Buttons

### Primary

Untuk tindakan utama:

```text
[ Generate QR ]
[ Scan QR ]
[ Konfirmasi Kehadiran ]
```

### Secondary

Untuk tindakan pendukung:

```text
[ Lihat Rekap ]
```

### Text button

Untuk tindakan ringan:

```text
Lihat semua
```

Jangan membuat semua button berwarna hijau solid.

---

# 40. Form Fields

Input:

- tinggi sekitar 48–56 px
- label jelas
- placeholder hanya sebagai contoh
- helper/error text
- icon hanya jika membantu

Jangan menggunakan placeholder sebagai satu-satunya label.

---

# 41. Accessibility

Target minimal:

- contrast WCAG AA
- interactive target minimal 44×44 px
- jangan mengandalkan warna saja
- text harus dapat dibaca pada ukuran mobile
- focus state harus terlihat
- error harus dapat dipahami
- status memiliki icon + text

Untuk teks normal, target contrast minimal sekitar 4.5:1.

---

# 42. Responsive Layout

Walaupun aplikasi utama mobile, komponen harus responsive.

Breakpoint konseptual:

```text
Mobile:
< 600 px

Tablet:
600–1024 px

Desktop:
> 1024 px
```

Pada mobile:

```text
single column
bottom navigation
full-width CTA
```

Pada tablet/desktop:

```text
content max-width
multi-column layout bila relevan
sidebar/navigation rail
```

Jangan sekadar memperbesar mobile layout ke desktop.

---

# 43. Motion & Animation

Gunakan motion untuk:

- page transition
- state change
- feedback
- bottom sheet
- modal
- loading

Hindari:

- bounce berlebihan
- spinning logo
- particle
- floating object
- infinite animation tanpa fungsi

Animasi harus terasa cepat dan natural.

Prinsip:

```text
Fast interaction:
~100–200 ms

Standard transition:
~200–300 ms

Complex transition:
~300–400 ms
```

Respect:

```text
prefers-reduced-motion
```

---

# 44. Login → Home Transition

Jangan membuat transisi sinematik.

Gunakan:

```text
Login success
→ short transition
→ Home
```

Fokus pada responsiveness.

---

# 45. QR Scanner Motion

Scanner boleh menggunakan animasi scanning line yang sangat subtle.

Tidak boleh:

- neon
- laser effect berlebihan
- glowing cyberpunk

Visual harus tetap akademik.

---

# 46. Dark Mode

Dark mode dapat disediakan, tetapi bukan prioritas pertama.

Jika dibuat:

- jangan membalik warna secara otomatis
- gunakan semantic color tokens
- hijau primary tetap terkontrol
- background harus benar-benar dark
- contrast tetap memenuhi accessibility

---

# 47. Component Architecture

Frontend agent harus membangun reusable components.

Contoh:

```text
AppShell
BottomNavigation
TopAppBar
CourseCard
AttendanceStatus
StatusChip
PrimaryButton
SecondaryButton
TextField
QrScanner
QrDisplay
StudentListItem
AttendanceSummary
EmptyState
ErrorState
LoadingSkeleton
ConfirmationDialog
BottomSheet
Avatar
KtmCard
```

Jangan membuat komponen khusus untuk setiap halaman jika pola visualnya sama.

---

# 48. Design Tokens

Implementasikan token, bukan hard-coded value di seluruh project.

Contoh:

```ts
colors.primary
colors.surface
colors.background
colors.textPrimary
colors.textSecondary
colors.success
colors.warning
colors.error

spacing.xs
spacing.sm
spacing.md
spacing.lg
spacing.xl

radius.sm
radius.md
radius.lg

typography.bodyMedium
typography.titleLarge
typography.headlineSmall
```

Dengan demikian theme dapat diubah tanpa mengubah seluruh UI.

---

# 49. Data States

Setiap halaman yang mengambil data harus memiliki minimal:

```text
Loading
Success
Empty
Error
```

Untuk mutation:

```text
Idle
Submitting
Success
Error
```

Jangan membuat UI yang hanya menangani happy path.

---

# 50. Permission UX

Untuk camera:

```text
Scan QR membutuhkan akses kamera.

Kamera digunakan hanya untuk membaca QR absensi.
```

Request permission pada saat context membutuhkan, bukan saat aplikasi pertama kali dibuka.

---

# 51. Offline / Network Feedback

Karena absensi bergantung pada data server, aplikasi harus memberi feedback jika jaringan bermasalah.

Contoh:

```text
Koneksi terputus

Absensi belum dapat dikirim.
Periksa koneksi dan coba lagi.

[ Coba Lagi ]
```

Jangan memberi kesan absensi berhasil sebelum server mengonfirmasi.

---

# 52. Security UX

Jangan menampilkan:

- token
- API key
- credential
- data teknis backend

kepada user.

QR attendance sebaiknya memiliki:

- expiration
- session identifier
- validasi server
- anti-replay mechanism

UI harus menampilkan status secara sederhana tanpa membocorkan mekanisme internal.

---

# 53. UX Flow Utama

## Mahasiswa

```text
Splash
 ↓
Session Check
 ↓
Login
 ↓
Home
 ↓
Mata Kuliah
 ↓
Detail Mata Kuliah
 ↓
Scan QR
 ↓
QR Validation
 ↓
Confirmation
 ↓
Attendance Success
 ↓
Attendance History
```

## Dosen

```text
Splash
 ↓
Session Check
 ↓
Login
 ↓
Home
 ↓
Jadwal
 ↓
Detail Mata Kuliah
 ↓
Generate QR
 ↓
Active QR
 ↓
Attendance Monitoring
 ↓
Rekap
```

---

# 54. Screen Inventory

## Shared

- Splash
- Login
- Forgot Password
- Profile
- Settings
- Notifications
- Help
- About
- Error
- Offline
- Session Expired

## Mahasiswa

- Home
- Mata Kuliah
- Detail Mata Kuliah
- Scan QR
- Confirm Attendance
- Attendance Success
- Attendance History
- KTM
- Notification

## Dosen

- Home
- Jadwal
- Mata Kuliah
- Detail Mata Kuliah
- Generate QR
- Active QR
- Student List
- Attendance Monitoring
- Attendance Recap

---

# 55. UI Quality Checklist

Sebelum menyelesaikan frontend, agent WAJIB mengecek:

### Visual

- [ ] Tidak ada gradient berlebihan
- [ ] Tidak ada glassmorphism
- [ ] Tidak ada shadow berlebihan
- [ ] Tidak ada card terlalu rounded
- [ ] Typography konsisten
- [ ] Spacing mengikuti 4 px system
- [ ] Primary green digunakan secara terkontrol
- [ ] Icon berasal dari satu icon family

### UX

- [ ] Login tersedia
- [ ] Session persistence tersedia
- [ ] Role-based navigation
- [ ] Loading state
- [ ] Empty state
- [ ] Error state
- [ ] Success state
- [ ] Permission state
- [ ] Confirmation untuk tindakan penting

### Accessibility

- [ ] Contrast memenuhi WCAG AA
- [ ] Touch target minimal 44 px
- [ ] Status tidak hanya dibedakan melalui warna
- [ ] Form error jelas
- [ ] Focus state tersedia
- [ ] Reduced motion diperhatikan

### Responsive

- [ ] Mobile
- [ ] Tablet
- [ ] Desktop bila diperlukan
- [ ] Tidak ada horizontal overflow
- [ ] Navigation tetap usable

---

# 56. Instruksi Khusus untuk AI Frontend Agent

**Jangan menghasilkan UI berdasarkan pola dashboard AI generik.**

Jangan:

```text
Huge gradient hero
+
glass cards
+
random blobs
+
giant statistics
+
neon green
+
AI illustrations
```

Sebaliknya:

```text
Material 3 foundation
+
Almuslim green identity
+
clear hierarchy
+
consistent spacing
+
restrained elevation
+
semantic status colors
+
accessible interaction
+
real mobile UX
```

Prioritaskan:

1. Information hierarchy
2. Usability
3. Accessibility
4. Consistency
5. Performance
6. Visual polish

Visual polish tidak boleh mengorbankan usability.

---

# 57. Definition of Done

Frontend dianggap selesai hanya jika:

- Login flow tersedia.
- Role mahasiswa/dosen berjalan.
- Navigation berbeda sesuai role.
- Semua screen utama tersedia.
- QR scan flow tersedia secara UI.
- QR generation flow tersedia secara UI.
- Attendance status tersedia.
- Attendance recap tersedia.
- KTM tersedia.
- Profile tersedia.
- Loading/empty/error/success state tersedia.
- Responsive.
- Accessible.
- Tidak terlihat seperti template AI.
- Tidak menggunakan visual dekoratif yang tidak memiliki fungsi.
- Design token digunakan secara konsisten.
- Material 3 digunakan sebagai foundation.

---

# 58. Final Design Direction

Bayangkan produk ini sebagai:

> **Aplikasi akademik resmi Universitas Almuslim yang modern, tenang, profesional, dan mudah digunakan mahasiswa setiap hari.**

Bukan:

> dashboard startup AI.

Semua keputusan desain harus menjawab pertanyaan:

**“Apakah elemen ini membantu mahasiswa atau dosen menyelesaikan tugasnya dengan lebih cepat dan lebih jelas?”**

Jika tidak, hapus.
