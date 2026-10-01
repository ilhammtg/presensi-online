import { defineStore } from 'pinia'
import { ref } from 'vue'

const DEFAULT_THEME = {
  primaryColor: '#006633',     // Hijau Tua Islami Umuslim
  accentColor: '#D4AF37',      // Emas Umuslim
  institutionName: 'Universitas Almuslim',
  institutionTagline: 'Sistem Informasi Presensi & QR Dinamis',
  logoUrl: ''
}

export const useThemeStore = defineStore('theme', () => {
  const saved = JSON.parse(localStorage.getItem('umuslim_theme') || 'null') || DEFAULT_THEME

  const primaryColor = ref(saved.primaryColor || DEFAULT_THEME.primaryColor)
  const accentColor = ref(saved.accentColor || DEFAULT_THEME.accentColor)
  const institutionName = ref(saved.institutionName || DEFAULT_THEME.institutionName)
  const institutionTagline = ref(saved.institutionTagline || DEFAULT_THEME.institutionTagline)
  const logoUrl = ref(saved.logoUrl || '')

  function applyTheme() {
    const root = document.documentElement

    // Primary Colors (Hijau)
    root.style.setProperty('--brand-primary', primaryColor.value)
    root.style.setProperty('--brand-primary-hover', adjustColorBrightness(primaryColor.value, -15))
    root.style.setProperty('--brand-primary-light', hexToRgba(primaryColor.value, 0.08))
    root.style.setProperty('--brand-primary-border', hexToRgba(primaryColor.value, 0.25))

    // Accent Colors (Emas)
    root.style.setProperty('--brand-accent', accentColor.value)
    root.style.setProperty('--brand-accent-hover', adjustColorBrightness(accentColor.value, -15))
    root.style.setProperty('--brand-accent-light', hexToRgba(accentColor.value, 0.12))
    root.style.setProperty('--brand-accent-border', hexToRgba(accentColor.value, 0.35))
  }

  function saveSettings(settings) {
    if (settings.primaryColor) primaryColor.value = settings.primaryColor
    if (settings.accentColor) accentColor.value = settings.accentColor
    if (settings.institutionName) institutionName.value = settings.institutionName
    if (settings.institutionTagline) institutionTagline.value = settings.institutionTagline
    if (settings.logoUrl !== undefined) logoUrl.value = settings.logoUrl

    const payload = {
      primaryColor: primaryColor.value,
      accentColor: accentColor.value,
      institutionName: institutionName.value,
      institutionTagline: institutionTagline.value,
      logoUrl: logoUrl.value
    }

    localStorage.setItem('umuslim_theme', JSON.stringify(payload))
    applyTheme()
  }

  function resetDefault() {
    saveSettings(DEFAULT_THEME)
  }

  // Helper color calculation
  function hexToRgba(hex, alpha) {
    hex = hex.replace('#', '')
    if (hex.length === 3) {
      hex = hex.split('').map(c => c + c).join('')
    }
    const r = parseInt(hex.substring(0, 2), 16) || 0
    const g = parseInt(hex.substring(2, 4), 16) || 0
    const b = parseInt(hex.substring(4, 6), 16) || 0
    return `rgba(${r}, ${g}, ${b}, ${alpha})`
  }

  function adjustColorBrightness(hex, percent) {
    hex = hex.replace('#', '')
    if (hex.length === 3) {
      hex = hex.split('').map(c => c + c).join('')
    }
    let r = parseInt(hex.substring(0, 2), 16)
    let g = parseInt(hex.substring(2, 4), 16)
    let b = parseInt(hex.substring(4, 6), 16)

    r = Math.min(255, Math.max(0, r + Math.round((percent / 100) * 255)))
    g = Math.min(255, Math.max(0, g + Math.round((percent / 100) * 255)))
    b = Math.min(255, Math.max(0, b + Math.round((percent / 100) * 255)))

    const toHex = (n) => n.toString(16).padStart(2, '0')
    return `#${toHex(r)}${toHex(g)}${toHex(b)}`
  }

  return {
    primaryColor,
    accentColor,
    institutionName,
    institutionTagline,
    logoUrl,
    applyTheme,
    saveSettings,
    resetDefault
  }
})
