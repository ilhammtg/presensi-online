import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './lib/api'
import '@fortawesome/fontawesome-free/css/all.min.css'
import './style.css'
import { useThemeStore } from './stores/theme'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// Initialize and apply institutional theme
const themeStore = useThemeStore(pinia)
themeStore.applyTheme()

app.mount('#app')
