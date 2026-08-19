<template>
  <main class="login-page">
    <section class="brand-panel">
      <div class="brand-mark">
        <svg viewBox="0 0 41 40" fill="none" aria-hidden="true">
          <path d="M26.3125 11.4814L22.25 19.5947V22.7148L27.1191 13.0576L29.7393 18.1777L18.7988 40H0L13.5938 22.8037H14L13.8125 23.1201L7.46875 33.541H17.8438V16.9111L7.1875 25.6475L18.0312 10.1406L11.625 14.4463V14.2588L20.4375 0L26.3125 11.4814Z" fill="currentColor"/>
          <path d="M34.5498 39.9996H28.75L24.5938 32.8864L27.125 27.6246L34.5498 39.9996ZM41 39.9996H36.2705L27.9346 25.941L30.7188 20.1559L41 39.9996Z" fill="currentColor" opacity=".7"/>
        </svg>
        <span>Cedar-V</span>
      </div>
      <div class="brand-copy">
        <p class="eyebrow">{{ t('login.communityEdition') }}</p>
        <h1>{{ t('login.heroTitleLine1') }}{{ t('login.heroTitleLine2') }}</h1>
        <p>{{ t('login.heroSubtitle') }}</p>
      </div>
      <p class="brand-note">{{ t('login.brandNote') }}</p>
    </section>

    <section class="login-panel">
      <div class="language-switcher">
        <el-select v-model="currentLanguage" size="small" @change="handleLanguageChange">
          <el-option label="中文" value="zh" />
          <el-option label="English" value="en" />
          <el-option label="日本語" value="ja" />
        </el-select>
      </div>
      <div class="login-card">
        <header>
          <h2>{{ t('login.welcomeTitle') }}</h2>
          <p>{{ t('login.welcomeSubtitle') }}</p>
        </header>
        <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" class="login-form" @submit.prevent="handleLogin">
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              :placeholder="t('login.usernamePlaceholder')"
              :prefix-icon="User"
              size="large"
              clearable
              autocomplete="username"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              :placeholder="t('login.passwordPlaceholder')"
              :prefix-icon="Lock"
              size="large"
              show-password
              autocomplete="current-password"
              @keyup.enter="handleLogin"
            />
          </el-form-item>
          <div class="form-options">
            <el-checkbox v-model="rememberMe">{{ t('login.remember') }}</el-checkbox>
          </div>
          <el-button type="primary" native-type="submit" size="large" class="login-button" :loading="loading">
            {{ t('login.submit') }}
          </el-button>
        </el-form>
        <footer>{{ t('login.copyright', { year: currentYear }) }}</footer>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { Lock, User } from '@element-plus/icons-vue'
import { Login, type LoginRequest } from '@/api/user'
import { useUserStore } from '@/store/modules/user'
import { changeLanguage, type SupportedLocale } from '@/utils/language'

const REMEMBER_KEY = 'loginInfo'
const { t, locale } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const loginFormRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(false)
const currentLanguage = ref(locale.value as SupportedLocale)
const currentYear = new Date().getFullYear()
const loginForm = reactive({ username: '', password: '' })
const loginRules: FormRules = {
  username: [
    { required: true, message: () => t('login.error.usernameRequired'), trigger: 'blur' },
    { min: 3, message: () => t('login.error.usernameMinLength'), trigger: 'blur' }
  ],
  password: [
    { required: true, message: () => t('login.error.passwordRequired'), trigger: 'blur' },
    { min: 6, message: () => t('login.error.passwordMinLength'), trigger: 'blur' }
  ]
}

const handleLanguageChange = (lang: SupportedLocale) => {
  changeLanguage(lang)
  currentLanguage.value = lang
}

const loadRememberedUsername = () => {
  const saved = localStorage.getItem(REMEMBER_KEY)
  if (!saved) return
  try {
    const value = JSON.parse(saved) as { username?: string; password?: string }
    if (value.username) {
      loginForm.username = value.username
      rememberMe.value = true
      localStorage.setItem(REMEMBER_KEY, JSON.stringify({ username: value.username }))
    } else {
      localStorage.removeItem(REMEMBER_KEY)
    }
  } catch {
    localStorage.removeItem(REMEMBER_KEY)
  }
}

const handleLogin = async () => {
  if (!loginFormRef.value || loading.value) return
  const valid = await loginFormRef.value.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    const loginData: LoginRequest = { username: loginForm.username, password: loginForm.password }
    const response = await Login(loginData)
    if (response.code !== '000000' || !response.data?.token || !response.data?.user_info) {
      ElMessage.error(response.message || t('login.error.general'))
      return
    }
    userStore.setLoginData(response.data.token, {
      username: response.data.user_info.username,
      role: response.data.user_info.role
    })
    if (rememberMe.value) localStorage.setItem(REMEMBER_KEY, JSON.stringify({ username: loginForm.username }))
    else localStorage.removeItem(REMEMBER_KEY)
    ElMessage.success(response.message || t('login.success'))
    await router.push('/dashboard')
  } catch (error: any) {
    ElMessage.error(error?.backendMessage || error?.response?.data?.message || error?.message || t('login.error.general'))
  } finally {
    loading.value = false
  }
}

onMounted(loadRememberedUsername)
</script>

<style lang="scss" scoped>
.login-page { min-height:100vh; display:grid; grid-template-columns:minmax(0,1.08fr) minmax(420px,.92fr); background:#f4f7f6; }
.brand-panel { position:relative; overflow:hidden; padding:48px clamp(40px,6vw,96px); display:flex; flex-direction:column; justify-content:space-between; color:#fff; background:linear-gradient(145deg,#0d6656 0%,#019c7c 58%,#33b796 100%); }
.brand-panel::before { content:''; position:absolute; inset:0; opacity:.2; background:radial-gradient(circle at 78% 18%,#fff 0,transparent 26%),linear-gradient(120deg,transparent 48%,rgba(255,255,255,.22) 49%,transparent 50%); }
.brand-panel>* { position:relative; z-index:1; }
.brand-mark { display:flex; align-items:center; gap:12px; font-family:'Swis721 BlkCn BT',sans-serif; font-size:28px; }
.brand-mark svg { width:40px; height:40px; }
.brand-copy { max-width:620px; }
.brand-copy .eyebrow { margin:0 0 10px; font-size:13px; font-weight:700; letter-spacing:.12em; text-transform:uppercase; opacity:.78; }
.brand-copy h1 { margin:0; max-width:560px; font-size:clamp(38px,5vw,68px); line-height:1.08; letter-spacing:-.03em; }
.brand-copy>p:last-child { max-width:560px; margin:24px 0 0; font-size:17px; line-height:1.8; opacity:.86; }
.brand-note { margin:0; font-size:13px; opacity:.72; }
.login-panel { position:relative; padding:64px 32px 32px; display:flex; align-items:center; justify-content:center; background:var(--app-content-bg); }
.language-switcher { position:absolute; top:20px; right:24px; width:112px; }
.login-card { width:min(100%,420px); }
.login-card header { margin-bottom:30px; }
.login-card h2 { margin:0; color:var(--app-text-primary); font-size:30px; line-height:1.3; }
.login-card header p { margin:8px 0 0; color:var(--app-text-secondary); font-size:14px; }
.login-form :deep(.el-form-item) { margin-bottom:22px; }
.login-form :deep(.el-input__wrapper) { min-height:48px; }
.form-options { margin:-4px 0 22px; display:flex; justify-content:space-between; }
.login-button { width:100%; min-height:48px; font-weight:600; }
.login-card footer { margin-top:36px; color:var(--app-text-secondary); font-size:12px; text-align:center; }
@media (max-width:1023px) {
  .login-page { grid-template-columns:1fr; }
  .brand-panel { display:none; }
  .login-panel { min-height:100vh; padding:72px 24px 32px; }
}
@media (max-width:480px) {
  .login-panel { padding-inline:18px; }
  .language-switcher { right:18px; }
  .login-card h2 { font-size:26px; }
}
</style>
