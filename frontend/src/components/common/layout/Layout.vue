<template>
  <div class="app-layout" :class="{ 'app-layout--collapsed': appStore.sidebarCollapsed }">
    <div
      v-if="appStore.isMobile && !appStore.sidebarCollapsed"
      class="layout-overlay"
      aria-hidden="true"
      @click="closeMobileSidebar"
    />
    <Sidebar :nav-items="navItems" @nav-click="handleNavClick" />
    <div class="layout-main">
      <NavContent />
      <main class="layout-content"><slot /></main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/store/modules/app'
import Sidebar from './Sidebar.vue'
import NavContent from './NavContent.vue'

interface NavItem {
  id: string
  label: string
  href: string
  icon?: string
  active?: boolean
}

defineProps<{ appName?: string; pageTitle?: string }>()

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const defaultNavItems = computed<NavItem[]>(() => [
  { id: 'dashboard', label: t('navigation.menu.dashboard'), href: '/dashboard', icon: 'dashboard' },
  { id: 'customers', label: t('navigation.menu.customers'), href: '/customers', icon: 'customers' },
  { id: 'enterprise-leads', label: t('navigation.menu.enterpriseLeads'), href: '/enterprise-leads', icon: 'enterprise-leads' },
  { id: 'licenses', label: t('navigation.menu.licenses'), href: '/licenses', icon: 'licenses' },
  { id: 'invoices', label: t('navigation.menu.invoices'), href: '/invoices', icon: 'invoices' },
  { id: 'packages', label: t('navigation.menu.packages'), href: '/packages', icon: 'packages' }
])

const navItems = computed(() => defaultNavItems.value.map(item => ({
  ...item,
  active: route.path === item.href || route.path.startsWith(`${item.href}/`)
})))

const closeMobileSidebar = () => appStore.setSidebarCollapsed(true)
const handleNavClick = async (item: NavItem) => {
  if (route.path !== item.href) await router.push(item.href)
  if (appStore.isMobile) closeMobileSidebar()
}

watch(
  () => appStore.isMobile && !appStore.sidebarCollapsed,
  isOpen => document.body.classList.toggle('sidebar-drawer-open', isOpen),
  { immediate: true }
)
onUnmounted(() => document.body.classList.remove('sidebar-drawer-open'))
</script>

<style lang="scss" scoped>
.app-layout {
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: var(--app-bg-color);
}
.layout-main {
  height: 100vh;
  margin-left: var(--layout-sidebar-width);
  display: flex;
  flex-direction: column;
  min-width: 0;
  transition: margin-left 0.2s ease;
}
.app-layout--collapsed .layout-main { margin-left: var(--layout-sidebar-collapsed-width); }
.layout-content {
  flex: 1;
  min-height: 0;
  padding-top: var(--layout-header-height);
  overflow: auto;
  background: var(--app-bg-color);
}
.layout-overlay {
  position: fixed;
  inset: 0;
  z-index: 1998;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(2px);
}
@media (min-width: 768px) and (max-width: 1023px) {
  .layout-main, .app-layout--collapsed .layout-main { margin-left: var(--layout-sidebar-collapsed-width); }
}
@media (max-width: 767px) {
  .layout-main, .app-layout--collapsed .layout-main { margin-left: 0; }
}
@media print {
  .layout-main { margin-left: 0 !important; }
  .layout-content { padding-top: 0; overflow: visible; }
}
</style>

<style>
body.sidebar-drawer-open { overflow: hidden; }
</style>
