<template>
  <header class="topbar" :class="{ 'topbar--collapsed': appStore.sidebarCollapsed }">
    <div class="topbar__left">
      <button type="button" class="icon-button" :aria-label="t('navigation.toggleSidebar')" @click="appStore.toggleSidebar()">
        <NavIcon name="sidebar-toggle" size="large" />
      </button>
      <nav v-if="breadcrumbs.length" class="breadcrumb" :aria-label="t('navigation.breadcrumbLabel')">
        <template v-for="(item, index) in breadcrumbs" :key="item.path || index">
          <button v-if="item.path && index < breadcrumbs.length - 1" type="button" class="breadcrumb-link" @click="navigateTo(item)">
            {{ item.title }}
          </button>
          <span v-else class="breadcrumb-current">{{ item.title }}</span>
          <span v-if="index < breadcrumbs.length - 1" class="breadcrumb-separator">/</span>
        </template>
      </nav>
    </div>

    <div class="topbar__right">
      <div class="external-links">
        <button v-for="link in externalLinks" :key="link.key" type="button" class="text-button" @click="openExternal(link.url)">
          {{ t(`navigation.external.${link.key}`) }}
        </button>
      </div>
      <button
        type="button"
        class="icon-button"
        :title="t('navigation.tooltip.theme')"
        :aria-label="t('navigation.tooltip.theme')"
        @click="toggleTheme"
      >
        <NavIcon name="dark-mode" />
      </button>
      <el-dropdown trigger="click" @command="handleLanguageChange">
        <button type="button" class="language-button" :title="t('navigation.tooltip.language')" :aria-label="t('navigation.tooltip.language')">
          <NavIcon name="language" />
          <span>{{ currentLanguageLabel }}</span>
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="option in languageOptions" :key="option.code" :command="option.code" :class="{ active: option.code === currentLanguage }">
              {{ option.label }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <el-dropdown trigger="click" placement="bottom-end" @command="handleUserCommand">
        <button type="button" class="user-button" :title="userInfo?.username || t('navigation.tooltip.user')" :aria-label="t('navigation.tooltip.user')">
          <span class="avatar">{{ userInitial }}</span>
          <span class="user-name">{{ userInfo?.username || '--' }}</span>
        </button>
        <template #dropdown>
          <el-dropdown-menu>
            <div class="user-summary">
              <strong>{{ userInfo?.username || '--' }}</strong>
              <span>{{ userInfo?.role || '--' }}</span>
            </div>
            <el-dropdown-item command="logout">{{ t('userMenu.logout') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAppStore } from '@/store/modules/app'
import { useUserStore } from '@/store/modules/user'
import { useBreadcrumb } from '@/utils/breadcrumb'
import { changeLanguage, type SupportedLocale } from '@/utils/language'
import NavIcon from '@/components/common/icons/NavIcon.vue'

const { t, locale } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const userStore = useUserStore()
const { breadcrumbs, navigateTo } = useBreadcrumb()
const userInfo = computed(() => userStore.userInfo)
const userInitial = computed(() => userInfo.value?.username?.charAt(0).toUpperCase() || '?')
const externalLinks = [
  { key: 'docs', url: 'https://docs.lm.cedar-v.com/' },
  { key: 'cloud', url: 'https://cedar-v.com/products/cedar-license-cloud/index.html' }
] as const
const languageOptions: Array<{ code: SupportedLocale; label: string }> = [
  { code: 'zh', label: '中文' },
  { code: 'en', label: 'English' },
  { code: 'ja', label: '日本語' }
]
const currentLanguage = ref(locale.value as SupportedLocale)
const currentLanguageLabel = computed(() => languageOptions.find(option => option.code === currentLanguage.value)?.label || '')
watch(locale, value => { currentLanguage.value = value as SupportedLocale })

const openExternal = (url: string) => window.open(url, '_blank', 'noopener,noreferrer')
const toggleTheme = () => appStore.setTheme(appStore.isDark ? 'light' : 'dark')
const handleLanguageChange = (lang: SupportedLocale) => {
  if (lang === currentLanguage.value) return
  changeLanguage(lang)
  currentLanguage.value = lang
}
const handleUserCommand = async (command: string) => {
  if (command !== 'logout') return
  try {
    await ElMessageBox.confirm(t('userMenu.logoutConfirm'), t('userMenu.logoutTitle'), {
      confirmButtonText: t('userMenu.confirm'),
      cancelButtonText: t('userMenu.cancel'),
      type: 'warning'
    })
    userStore.logout()
    await router.push('/login')
    ElMessage.success(t('userMenu.logoutSuccess'))
  } catch (error) {
    if (error !== 'cancel') console.error('Logout failed:', error)
  }
}
</script>

<style scoped>
.topbar {
  --topbar-text: #40574e;
  --topbar-secondary: #687d74;
  position: fixed; top: 0; left: var(--layout-sidebar-width); right: 0; z-index: 1997;
  height: var(--layout-header-height); padding: 0 20px; display: flex; align-items: center;
  justify-content: space-between; gap: 16px; background: var(--app-nav-bg);
  border-bottom: 1px solid var(--app-border-light); color:var(--topbar-text); font-size:13px; transition: left 0.2s ease;
}
.topbar--collapsed { left: var(--layout-sidebar-collapsed-width); }
.topbar__left,.topbar__right,.breadcrumb,.user-button,.language-button { display:flex; align-items:center; }
.topbar__left,.topbar__right { min-width:0; gap:12px; }
.icon-button,.language-button,.user-button,.text-button,.breadcrumb-link { border:0; background:transparent; color:var(--topbar-text); font:inherit; cursor:pointer; }
.icon-button { width:36px; height:36px; display:flex; align-items:center; justify-content:center; border-radius:0; flex-shrink:0; }
.icon-button:hover,.language-button:hover,.user-button:hover { background:var(--el-color-primary-light-9); color:var(--el-color-primary); }
.breadcrumb { min-width:0; gap:8px; white-space:nowrap; }
.breadcrumb-link { padding:0; }
.breadcrumb-link:hover { color:var(--el-color-primary); }
.breadcrumb-current { color:#1f372e; font-size:13px; font-weight:650; overflow:hidden; text-overflow:ellipsis; }
.breadcrumb-separator { color:var(--topbar-secondary); }
.external-links { display:flex; align-items:center; gap:4px; }
.text-button { padding:6px 8px; border-radius:0; }
.text-button:hover { color:var(--el-color-primary); background:var(--el-color-primary-light-9); }
.language-button,.user-button { min-height:36px; padding:0 9px; gap:7px; border-radius:0; }
.avatar { width:28px; height:28px; display:inline-flex; align-items:center; justify-content:center; border-radius:50%; background:var(--el-color-primary); color:#fff; font-weight:600; }
.user-name { max-width:120px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
.user-summary { min-width:180px; padding:10px 16px; display:flex; flex-direction:column; gap:3px; color:var(--app-text-primary); }
.user-summary span { color:var(--app-text-secondary); font-size:12px; }
:global([data-theme="dark"]) .topbar { --topbar-text:#c1d0ca; --topbar-secondary:#93a69e; }
:global([data-theme="dark"]) .breadcrumb-current { color:var(--app-text-primary); }
@media (min-width:768px) and (max-width:1023px) { .topbar,.topbar--collapsed { left:var(--layout-sidebar-collapsed-width); } .external-links { display:none; } }
@media (max-width:767px) { .topbar,.topbar--collapsed { left:0; padding:0 12px; } .external-links,.language-button span,.user-name { display:none; } .topbar__right { gap:4px; } .breadcrumb { max-width:48vw; } }
</style>
