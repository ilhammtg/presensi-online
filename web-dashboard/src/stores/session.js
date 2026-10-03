import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'

export const useSessionStore = defineStore('session', () => {
  const authStore = useAuthStore()

  const schedules = ref([])
  const selectedSchedule = ref(null)
  const enrolledStudents = ref([])
  const scheduleSessions = ref([])
  const classRecap = ref(null)
  const nextMeetingNo = ref(1)

  const currentSession = ref(null)
  const qrToken = ref('')
  const qrExpiresIn = ref(15)
  const timerInterval = ref(null)
  const attendees = ref([])
  const wsConnected = ref(false)
  const loading = ref(false)
  const error = ref('')

  let ws = null

  async function fetchDosenSchedules() {
    loading.value = true
    try {
      const res = await fetch('/v1/dosen/schedules', {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (!res.ok) throw new Error('Gagal mengambil daftar mata kuliah')
      const data = await res.json()
      schedules.value = data.schedules || []
      return schedules.value
    } catch (err) {
      error.value = err.message
      return []
    } finally {
      loading.value = false
    }
  }

  async function fetchScheduleDetails(scheduleId) {
    if (!scheduleId) return
    loading.value = true
    try {
      // 1. Fetch enrolled students
      const studRes = await fetch(`/v1/schedules/${scheduleId}/students`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (studRes.ok) {
        const studData = await studRes.json()
        enrolledStudents.value = studData.students || []
      }

      // 2. Fetch sessions history & next meeting no
      const sessRes = await fetch(`/v1/schedules/${scheduleId}/sessions`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (sessRes.ok) {
        const sessData = await sessRes.json()
        scheduleSessions.value = sessData.sessions || []
        nextMeetingNo.value = sessData.next_meeting_no || 1

        // If this schedule has an open session, automatically restore it
        const activeSess = (sessData.sessions || []).find(s => s.is_open)
        if (activeSess) {
          const sessId = activeSess.session_id
          if (!currentSession.value || (currentSession.value.id !== sessId && currentSession.value.ID !== sessId)) {
            await loadSessionById(sessId)
          }
        }
      }

      // 3. Fetch recap
      const recapRes = await fetch(`/v1/schedules/${scheduleId}/recap`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (recapRes.ok) {
        classRecap.value = await recapRes.json()
      }
    } catch (err) {
      console.error('Failed to fetch schedule details:', err)
    } finally {
      loading.value = false
    }
  }

  async function markPermission(sessionId, studentId, status, notes = '') {
    loading.value = true
    try {
      const res = await fetch(`/v1/sessions/${sessionId}/permission`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          student_id: studentId,
          status,
          notes
        })
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || 'Gagal menandai izin/sakit')

      // Refresh attendees & recap
      await fetchAttendees(sessionId)
      if (selectedSchedule.value) {
        await fetchScheduleDetails(selectedSchedule.value.id)
      }
      return true
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  // Start internal countdown ticker for smooth UI progress bar
  function startLocalCountdown(initialSeconds = 15) {
    qrExpiresIn.value = initialSeconds
    if (timerInterval.value) clearInterval(timerInterval.value)

    timerInterval.value = setInterval(() => {
      if (qrExpiresIn.value > 0) {
        qrExpiresIn.value -= 1
      }
    }, 1000)
  }

  async function openSession(scheduleId, meetingNo, durationMinutes = 30) {
    loading.value = true
    error.value = ''
    try {
      const res = await fetch('/v1/sessions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authStore.token}`
        },
        body: JSON.stringify({
          schedule_id: scheduleId,
          meeting_no: meetingNo,
          duration_minutes: durationMinutes
        })
      })

      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.error || 'Gagal membuka sesi')
      }

      currentSession.value = data.session
      qrToken.value = data.current_qr_token
      startLocalCountdown(data.expires_in_seconds || 15)

      // Connect WebSocket for live updates
      const sessionId = data.session.id || data.session.ID
      connectWebSocket(sessionId)
      await fetchAttendees(sessionId)

      return data.session
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function checkActiveSession(scheduleId) {
    try {
      const res = await fetch(`/v1/sessions/active?schedule_id=${scheduleId}`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (!res.ok) return null
      const data = await res.json()
      if (data.session) {
        currentSession.value = data.session
        qrToken.value = data.current_qr_token || ''
        startLocalCountdown(data.expires_in_seconds || 15)
        const sessId = data.session.id || data.session.ID
        connectWebSocket(sessId)
        await fetchAttendees(sessId)
        return data.session
      }
      return null
    } catch (err) {
      console.error('Failed to check active session:', err)
      return null
    }
  }

  async function loadSessionById(sessionId) {
    loading.value = true
    error.value = ''
    try {
      const res = await fetch(`/v1/sessions/${sessionId}`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (!res.ok) throw new Error('Sesi tidak ditemukan atau telah berakhir')
      const sess = await res.json()
      const isOpen = sess.is_open !== undefined ? sess.is_open : sess.IsOpen
      if (!isOpen) throw new Error('Sesi ini telah ditutup')

      // Fetch current QR token
      const qrRes = await fetch(`/v1/sessions/${sessionId}/qr`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      const qrData = await qrRes.json()

      currentSession.value = sess
      qrToken.value = qrData.token || ''
      startLocalCountdown(qrData.expires_in_seconds || 15)

      connectWebSocket(sessionId)
      await fetchAttendees(sessionId)
      return sess
    } catch (err) {
      error.value = err.message
      currentSession.value = null
      throw err
    } finally {
      loading.value = false
    }
  }

  async function closeSession(bapTopic) {
    if (!currentSession.value) return
    const sessionId = currentSession.value.id || currentSession.value.ID
    loading.value = true
    try {
      const res = await fetch(`/v1/sessions/${sessionId}/close`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authStore.token}`
        },
        body: JSON.stringify({ bap_topic: bapTopic })
      })

      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.error || 'Gagal menutup sesi')
      }

      currentSession.value = null
      qrToken.value = ''
      disconnectWebSocket()
      if (timerInterval.value) clearInterval(timerInterval.value)

      return true
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchAttendees(sessionId) {
    if (!sessionId && currentSession.value) {
      sessionId = currentSession.value.id || currentSession.value.ID
    }
    if (!sessionId) return

    try {
      const res = await fetch(`/v1/sessions/${sessionId}/attendees`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      const data = await res.json()
      if (res.ok) {
        attendees.value = data.attendees || []
      }
    } catch (err) {
      console.error('Failed to fetch attendees:', err)
    }
  }

  async function overrideAttendance(attendanceId, status, notes = '') {
    try {
      const res = await fetch(`/v1/attendances/${attendanceId}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authStore.token}`
        },
        body: JSON.stringify({ status, notes })
      })
      const data = await res.json()
      if (!res.ok) throw new Error(data.error || 'Gagal mengubah status')

      // Update locally
      const item = attendees.value.find(a => a.attendance_id === attendanceId)
      if (item) {
        item.status = status
        item.notes = notes
      }
      return true
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  function connectWebSocket(sessionId) {
    disconnectWebSocket()

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/v1/ws?token=${authStore.token}&session_id=${sessionId}`

    try {
      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        wsConnected.value = true
        console.log('⚡ WebSocket Connected to session:', sessionId)
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          handleWebSocketMessage(msg)
        } catch (e) {
          console.error('Invalid WS JSON:', e)
        }
      }

      ws.onclose = () => {
        wsConnected.value = false
        console.log('WebSocket closed')
      }

      ws.onerror = (err) => {
        console.error('WebSocket Error:', err)
      }
    } catch (err) {
      console.error('WebSocket connection failed:', err)
    }
  }

  function handleWebSocketMessage(msg) {
    switch (msg.type) {
      case 'QR_REFRESHED':
        qrToken.value = msg.payload.token
        startLocalCountdown(msg.payload.expires_in_seconds || 15)
        break

      case 'ATTENDANCE_RECORDED':
        // Re-fetch full attendee details to get student name and NIM
        if (currentSession.value) {
          fetchAttendees(currentSession.value.id || currentSession.value.ID)
        }
        break

      case 'SESSION_CLOSED':
        currentSession.value = null
        qrToken.value = ''
        disconnectWebSocket()
        break
    }
  }


  async function fetchScheduleStudents(scheduleId) {
    if (!scheduleId) return []
    try {
      const res = await fetch(`/v1/schedules/${scheduleId}/students`, {
        headers: { Authorization: `Bearer ${authStore.token}` }
      })
      if (res.ok) {
        const data = await res.json()
        return data.students || []
      }
      return []
    } catch (err) {
      console.error('Failed to fetch students:', err)
      return []
    }
  }

  function disconnectWebSocket() {
    if (ws) {
      ws.close()
      ws = null
    }
    wsConnected.value = false
  }

  return {
    schedules,
    selectedSchedule,
    enrolledStudents,
    scheduleSessions,
    classRecap,
    nextMeetingNo,
    fetchDosenSchedules,
    fetchScheduleDetails,
    fetchScheduleStudents,
    markPermission,
    currentSession,
    qrToken,
    qrExpiresIn,
    attendees,
    wsConnected,
    loading,
    error,
    openSession,
    checkActiveSession,
    loadSessionById,
    closeSession,
    fetchAttendees,
    overrideAttendance,
    connectWebSocket,
    disconnectWebSocket
  }
})
