#!/bin/bash
# =============================================================================
# Script Fix: Flutter USB Debugging - Samsung Galaxy A56
# Universitas Almuslim - Presensi Online Project
# =============================================================================
# Jalankan script ini dengan: bash scripts/fix-flutter-usb-debug.sh
# =============================================================================

set -e

ANDROID_SDK_HOME="$HOME/Android/Sdk"
PLATFORM_TOOLS="$ANDROID_SDK_HOME/platform-tools"
JDK_PATH="$HOME/development/jdk-17"

echo "======================================================"
echo "  Fix Flutter USB Debugging untuk Samsung Galaxy A56"
echo "======================================================"
echo ""

# -----------------------------------------------------------------------
# 1. Buat udev rules untuk Samsung/Android
# -----------------------------------------------------------------------
echo "[1/5] Membuat udev rules untuk Samsung Android..."
sudo tee /etc/udev/rules.d/51-android.rules > /dev/null << 'UDEV_EOF'
# Samsung Electronics - Android devices (ADB + MTP + Fastboot)
SUBSYSTEM=="usb", ATTR{idVendor}=="04e8", MODE="0666", GROUP="plugdev", TAG+="uaccess"
SUBSYSTEM=="usb", ATTR{idVendor}=="04e8", ATTR{idProduct}=="6860", MODE="0666", SYMLINK+="android%n", TAG+="uaccess"
SUBSYSTEM=="usb", ATTR{idVendor}=="04e8", ATTR{idProduct}=="685e", MODE="0666", SYMLINK+="android%n", TAG+="uaccess"
SUBSYSTEM=="usb", ATTR{idVendor}=="04e8", ATTR{idProduct}=="6864", MODE="0666", SYMLINK+="android%n", TAG+="uaccess"
UDEV_EOF
echo "   ✓ Udev rules dibuat di /etc/udev/rules.d/51-android.rules"

# -----------------------------------------------------------------------
# 2. Tambahkan user ke group plugdev (jika belum)
# -----------------------------------------------------------------------
echo "[2/5] Menambahkan user ke group plugdev..."
if groups "$USER" | grep -q "plugdev"; then
    echo "   ✓ User sudah di group plugdev"
else
    sudo usermod -a -G plugdev "$USER"
    echo "   ✓ User ditambahkan ke group plugdev (perlu logout/login ulang)"
fi

# -----------------------------------------------------------------------
# 3. Reload udev rules
# -----------------------------------------------------------------------
echo "[3/5] Mereload udev rules..."
sudo udevadm control --reload-rules
sudo udevadm trigger
echo "   ✓ Udev rules di-reload"

# -----------------------------------------------------------------------
# 4. Fix PATH - tambahkan ADB dan JDK ke .bashrc
# -----------------------------------------------------------------------
echo "[4/5] Memperbaiki PATH di ~/.bashrc..."

BASHRC="$HOME/.bashrc"
NEED_UPDATE=false

# Cek apakah JAVA_HOME sudah di-set dengan benar
if ! grep -q "JAVA_HOME=\"\$HOME/development/jdk-17\"" "$BASHRC" 2>/dev/null; then
    NEED_UPDATE=true
fi

# Cek apakah adb sudah di PATH
if ! grep -q "platform-tools" "$BASHRC" 2>/dev/null; then
    NEED_UPDATE=true
fi

if [ "$NEED_UPDATE" = true ]; then
    # Hapus entri JAVA_HOME lama yang mungkin salah
    sed -i '/^export JAVA_HOME/d' "$BASHRC"
    sed -i '/^export ANDROID_HOME/d' "$BASHRC"
    sed -i '/^export ANDROID_SDK_ROOT/d' "$BASHRC"
    # Hapus baris PATH yang mengandung platform-tools (jika sudah ada)
    sed -i '/platform-tools/d' "$BASHRC"

    cat >> "$BASHRC" << 'BASHRC_EOF'

# Android SDK & Java - Flutter Development
export JAVA_HOME="$HOME/development/jdk-17"
export ANDROID_HOME="$HOME/Android/Sdk"
export ANDROID_SDK_ROOT="$HOME/Android/Sdk"
export PATH="$JAVA_HOME/bin:$ANDROID_HOME/platform-tools:$ANDROID_HOME/tools:$ANDROID_HOME/tools/bin:$PATH"
BASHRC_EOF
    echo "   ✓ JAVA_HOME, ANDROID_HOME, dan platform-tools ditambahkan ke PATH"
else
    echo "   ✓ PATH sudah dikonfigurasi dengan benar"
fi

# -----------------------------------------------------------------------
# 5. Set JAVA_HOME di Flutter config
# -----------------------------------------------------------------------
echo "[5/5] Mengkonfigurasi Flutter untuk menggunakan JDK yang benar..."
if [ -d "$JDK_PATH" ]; then
    "$HOME/development/flutter/bin/flutter" config --jdk-dir="$JDK_PATH" 2>/dev/null || true
    echo "   ✓ Flutter dikonfigurasi menggunakan JDK di: $JDK_PATH"
else
    echo "   ✗ JDK tidak ditemukan di $JDK_PATH"
fi

# -----------------------------------------------------------------------
# 6. Restart ADB server
# -----------------------------------------------------------------------
echo ""
echo "[Bonus] Merestart ADB server..."
"$PLATFORM_TOOLS/adb" kill-server 2>/dev/null || true
"$PLATFORM_TOOLS/adb" start-server 2>/dev/null || true
echo "   ✓ ADB server di-restart"

# -----------------------------------------------------------------------
# Tampilkan instruksi untuk HP
# -----------------------------------------------------------------------
echo ""
echo "======================================================"
echo "  LANGKAH SELANJUTNYA - LAKUKAN DI HP SAMSUNG"
echo "======================================================"
echo ""
echo "  1. Cabut kabel USB dari laptop"
echo ""
echo "  2. Di HP Samsung Galaxy A56:"
echo "     a. Buka Pengaturan → Tentang Ponsel"
echo "     b. Ketuk 'Nomor Build' sebanyak 7 kali"
echo "     c. Masukkan PIN/pola jika diminta"
echo "     d. Kembali ke Pengaturan → Opsi Pengembang"
echo "     e. Aktifkan 'USB Debugging'"
echo ""
echo "  3. Colok kembali kabel USB"
echo "     → Pilih MODE: 'Transfer File (MTP)' di HP"
echo ""
echo "  4. Di HP akan muncul dialog 'Izinkan debugging USB?'"
echo "     → Centang 'Selalu izinkan dari komputer ini'"
echo "     → Ketuk 'Izinkan'"
echo ""
echo "  5. Jalankan perintah ini untuk verifikasi:"
echo "     source ~/.bashrc"
echo "     adb devices"
echo "     flutter devices"
echo ""
echo "  6. Jalankan app Flutter:"
echo "     cd presensi-online/mobile"
echo "     flutter run"
echo ""
echo "======================================================"
echo "  CATATAN PENTING"
echo "======================================================"
echo ""
echo "  • Jika setelah step 4 perangkat masih tidak muncul:"
echo "    - Logout dan login ulang ke Linux (untuk plugdev group)"
echo "    - Atau: exec sudo -u \$USER bash"
echo ""
echo "  • Jika muncul 'unauthorized' di adb devices:"
echo "    - Periksa HP → ada notifikasi 'Izinkan USB Debugging?'"
echo "    - Ketuk 'Izinkan'"
echo ""
echo "  • Jika masalah JAVA_HOME masih muncul:"
echo "    - Jalankan: source ~/.bashrc"
echo "    - Lalu coba lagi: flutter run"
echo ""
echo "✓ Script selesai! Ikuti langkah di atas untuk menyelesaikan setup."
