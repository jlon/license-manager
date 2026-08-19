<template>
  <Layout app-name="Cedar-V" :page-title="t('dashboard.title')">
    <div class="dashboard-page">
      <header class="page-header">
        <div>
          <p class="eyebrow">{{ t('dashboard.eyebrow') }}</p>
          <h1>{{ t('dashboard.heading') }}</h1>
          <p class="page-description">{{ t('dashboard.description') }}</p>
        </div>
        <div class="header-actions">
          <span v-if="lastUpdatedAt" class="updated-at">{{ t('dashboard.updatedAt', { time: lastUpdatedAt }) }}</span>
          <el-button :icon="Refresh" :loading="refreshing" @click="refreshAll">{{ t('dashboard.refresh') }}</el-button>
        </div>
      </header>

      <section class="overview-grid" :aria-label="t('dashboard.overviewTitle')">
        <article v-for="card in statCards" :key="card.key" class="metric-card" :class="`metric-card--${card.tone}`">
          <div class="metric-icon"><el-icon><component :is="card.icon" /></el-icon></div>
          <div class="metric-content">
            <span class="metric-label">{{ card.label }}</span>
            <el-skeleton v-if="statsLoading && !stats" animated :rows="1" />
            <strong v-else class="metric-value">{{ stats ? card.value : '--' }}</strong>
            <span class="metric-help">{{ stats ? card.help : statsError || t('dashboard.noData') }}</span>
          </div>
        </article>
      </section>

      <section class="panel trend-panel">
        <div class="panel-header trend-header">
          <div>
            <h2>{{ t('dashboard.trendTitle') }}</h2>
            <p>{{ t('dashboard.trendDescription') }}</p>
          </div>
          <div class="trend-controls">
            <el-radio-group v-model="trendType" size="small" @change="handleTrendTypeChange">
              <el-radio-button value="week">{{ t('chart.licenseTrend.quickOptions.week') }}</el-radio-button>
              <el-radio-button value="month">{{ t('chart.licenseTrend.quickOptions.month') }}</el-radio-button>
              <el-radio-button value="custom">{{ t('dashboard.customRange') }}</el-radio-button>
            </el-radio-group>
            <el-date-picker
              v-if="trendType === 'custom'"
              v-model="customRange"
              type="daterange"
              value-format="YYYY-MM-DD"
              :range-separator="t('dashboard.to')"
              :start-placeholder="t('dashboard.startDate')"
              :end-placeholder="t('dashboard.endDate')"
              :disabled-date="disableFutureDate"
              @change="loadTrend"
            />
          </div>
        </div>
        <div v-loading="trendLoading" class="panel-body chart-body">
          <div v-if="trendError" class="state-block">
            <el-icon><WarningFilled /></el-icon>
            <span>{{ trendError }}</span>
            <el-button link type="primary" @click="loadTrend">{{ t('dashboard.retry') }}</el-button>
          </div>
          <LicenseTrendChart v-else :data="trendData" :empty-text="t('dashboard.noTrendData')" />
        </div>
      </section>

      <section class="panel recent-panel">
        <div class="panel-header">
          <div>
            <h2>{{ t('dashboard.recentLicenses.title') }}</h2>
            <p>{{ t('dashboard.recentDescription') }}</p>
          </div>
          <el-button link type="primary" @click="router.push({ name: 'licenses-list' })">{{ t('dashboard.viewAll') }}</el-button>
        </div>
        <div class="panel-body table-wrap">
          <div v-if="recentError" class="state-block">
            <el-icon><WarningFilled /></el-icon>
            <span>{{ recentError }}</span>
            <el-button link type="primary" @click="loadRecent">{{ t('dashboard.retry') }}</el-button>
          </div>
          <el-table
            v-else
            v-loading="recentLoading"
            :data="recentData"
            stripe
            @row-click="openAuthorization"
          >
            <el-table-column prop="customer_name" :label="t('dashboard.recentLicenses.columns.customerName')" min-width="150" show-overflow-tooltip />
            <el-table-column prop="description" :label="t('dashboard.recentLicenses.columns.description')" min-width="180" show-overflow-tooltip />
            <el-table-column :label="t('dashboard.recentLicenses.columns.status')" width="110" align="center">
              <template #default="{ row }"><span class="status-tag" :class="`status-tag--${row.status}`">{{ row.status_display }}</span></template>
            </el-table-column>
            <el-table-column :label="t('dashboard.recentLicenses.columns.expiryTime')" min-width="140">
              <template #default="{ row }">{{ formatDate(row.end_date) }}</template>
            </el-table-column>
            <el-table-column :label="t('dashboard.recentLicenses.columns.createTime')" min-width="140">
              <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
            </el-table-column>
            <template #empty><el-empty :description="t('dashboard.noRecentData')" :image-size="64" /></template>
          </el-table>
        </div>
      </section>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { computed, markRaw, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Bell, CircleCheck, Clock, Key, Plus, Refresh, Warning, WarningFilled } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import LicenseTrendChart from '@/components/charts/LicenseTrendChart.vue'
import {
  getAuthorizationTrend,
  getOverviewStats,
  getRecentAuthorizations,
  type RecentAuthorizationItem,
  type StatsOverviewData,
  type TrendDataItem
} from '@/api/dashboard'
import { formatDate } from '@/utils/date'

const { t } = useI18n()
const router = useRouter()
const stats = ref<StatsOverviewData | null>(null)
const recentData = ref<RecentAuthorizationItem[]>([])
const trendData = ref<TrendDataItem[]>([])
const statsLoading = ref(false)
const recentLoading = ref(false)
const trendLoading = ref(false)
const refreshing = ref(false)
const statsError = ref('')
const recentError = ref('')
const trendError = ref('')
const lastUpdatedAt = ref('')
const trendType = ref<'week' | 'month' | 'custom'>('month')
const customRange = ref<[string, string] | null>(null)

const statCards = computed(() => {
  const data = stats.value
  return [
    { key: 'total', label: t('dashboard.stats.totalAuthCodes'), value: data?.total_auth_codes ?? 0, help: t('dashboard.stats.subText.monthNew', { n: data?.month_new_auth_codes ?? 0 }), tone: 'brand', icon: markRaw(Key) },
    { key: 'active', label: t('dashboard.stats.activeLicenses'), value: data?.active_licenses ?? 0, help: t('dashboard.stats.subText.vsLastMonth', { rate: formatRate(data?.growth_rate?.licenses_mom) }), tone: 'success', icon: markRaw(CircleCheck) },
    { key: 'today', label: t('dashboard.stats.todayNewLicenses'), value: data?.today_new_licenses ?? 0, help: t('dashboard.stats.subText.yesterday', { n: data?.yesterday_new_licenses ?? 0 }), tone: 'neutral', icon: markRaw(Plus) },
    { key: 'due7', label: t('dashboard.stats.expiringIn7Days'), value: data?.expiring_in_7days ?? 0, help: (data?.expiring_in_7days ?? 0) > 0 ? t('dashboard.stats.subText.urgent') : t('dashboard.stats.subText.noRisk'), tone: 'danger', icon: markRaw(Warning) },
    { key: 'due30', label: t('dashboard.stats.expiringIn30Days'), value: data?.expiring_in_30days ?? 0, help: t('dashboard.stats.subText.expiryReminder'), tone: 'warning', icon: markRaw(Clock) },
    { key: 'alerts', label: t('dashboard.stats.abnormalAlerts'), value: data?.abnormal_alerts ?? 0, help: t('dashboard.stats.subText.heartbeatTimeout'), tone: 'danger', icon: markRaw(Bell) }
  ]
})

const formatRate = (value?: number) => {
  const rate = value ?? 0
  return `${rate >= 0 ? '+' : ''}${rate.toFixed(2)}`
}
const localDate = (date: Date) => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
const trendRange = () => {
  if (trendType.value === 'custom' && customRange.value) return customRange.value
  const today = new Date()
  if (trendType.value === 'week') {
    const weekday = today.getDay() || 7
    const start = new Date(today)
    start.setDate(today.getDate() - weekday + 1)
    return [localDate(start), localDate(today)] as [string, string]
  }
  return [localDate(new Date(today.getFullYear(), today.getMonth(), 1)), localDate(today)] as [string, string]
}
const errorText = (error: any) => error?.backendMessage || error?.response?.data?.message || t('dashboard.loadFailed')

const loadStats = async () => {
  statsLoading.value = true
  statsError.value = ''
  try { stats.value = (await getOverviewStats()).data } catch (error) { statsError.value = errorText(error) } finally { statsLoading.value = false }
}
const loadRecent = async () => {
  recentLoading.value = true
  recentError.value = ''
  try { recentData.value = (await getRecentAuthorizations({ limit: 8 })).data.list } catch (error) { recentError.value = errorText(error) } finally { recentLoading.value = false }
}
const loadTrend = async () => {
  if (trendType.value === 'custom' && !customRange.value) return
  trendLoading.value = true
  trendError.value = ''
  const [start_date, end_date] = trendRange()
  try {
    trendData.value = (await getAuthorizationTrend({ type: trendType.value, start_date, end_date })).data.trend_data
  } catch (error) { trendError.value = errorText(error) } finally { trendLoading.value = false }
}
const refreshAll = async () => {
  refreshing.value = true
  await Promise.allSettled([loadStats(), loadTrend(), loadRecent()])
  lastUpdatedAt.value = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  refreshing.value = false
}
const handleTrendTypeChange = () => {
  if (trendType.value !== 'custom' || customRange.value) loadTrend()
}
const disableFutureDate = (date: Date) => date.getTime() > Date.now()
const openAuthorization = (row: RecentAuthorizationItem) => router.push({ name: 'licenses-view', params: { id: row.id } })
onMounted(refreshAll)
</script>

<style lang="scss" scoped>
.dashboard-page { padding: var(--layout-content-padding); display:flex; flex-direction:column; gap:var(--layout-content-gap); }
.page-header { display:flex; align-items:flex-end; justify-content:space-between; gap:20px; }
.eyebrow { margin:0 0 4px; color:var(--el-color-primary); font-size:12px; font-weight:700; letter-spacing:.08em; text-transform:uppercase; }
.page-header h1 { margin:0; color:var(--app-text-primary); font-size:24px; line-height:1.35; }
.page-description,.panel-header p { margin:4px 0 0; color:var(--app-text-secondary); font-size:13px; }
.header-actions { display:flex; align-items:center; gap:12px; flex-shrink:0; }
.updated-at { color:var(--app-text-secondary); font-size:12px; }
.overview-grid { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:12px; }
.metric-card { --metric:#64748b; min-height:112px; padding:18px; display:flex; align-items:center; gap:14px; background:var(--app-content-bg); border:1px solid var(--app-border-light); border-left:4px solid var(--metric); border-radius:var(--app-card-radius); box-shadow:var(--app-card-shadow); }
.metric-card--brand{--metric:#019c7c}.metric-card--success{--metric:#39a56b}.metric-card--warning{--metric:#e89b24}.metric-card--danger{--metric:#e45b5b}
.metric-icon { width:42px; height:42px; display:flex; align-items:center; justify-content:center; flex-shrink:0; border-radius:10px; background:color-mix(in srgb,var(--metric) 12%,transparent); color:var(--metric); font-size:22px; }
.metric-content { min-width:0; display:flex; flex:1; flex-direction:column; gap:3px; }
.metric-label { color:var(--app-text-regular); font-size:13px; }
.metric-value { color:var(--app-text-primary); font-size:28px; line-height:1.2; }
.metric-help { min-height:18px; overflow:hidden; color:var(--app-text-secondary); font-size:12px; text-overflow:ellipsis; white-space:nowrap; }
.panel { overflow:hidden; background:var(--app-content-bg); border:1px solid var(--app-border-light); border-radius:var(--app-card-radius); box-shadow:var(--app-card-shadow); }
.panel-header { min-height:66px; padding:14px 18px; display:flex; align-items:center; justify-content:space-between; gap:16px; border-bottom:1px solid var(--app-border-light); }
.panel-header h2 { margin:0; color:var(--app-text-primary); font-size:16px; }
.trend-controls { display:flex; align-items:center; justify-content:flex-end; gap:10px; flex-wrap:wrap; }
.panel-body { min-height:120px; padding:12px 18px 18px; }
.chart-body { min-height:280px; }
.state-block { min-height:220px; display:flex; align-items:center; justify-content:center; gap:8px; color:var(--app-text-secondary); }
.table-wrap { overflow-x:auto; }
.table-wrap :deep(.el-table__row) { cursor:pointer; }
.status-tag { display:inline-flex; min-width:54px; padding:3px 9px; justify-content:center; border-radius:999px; font-size:12px; }
.status-tag--normal { color:#287a54; background:#eaf8f0; }
.status-tag--locked { color:#a56605; background:#fff5df; }
.status-tag--expired { color:#b83d3d; background:#ffeded; }
@media (max-width:1199px) { .overview-grid { grid-template-columns:repeat(2,minmax(0,1fr)); } }
@media (max-width:767px) {
  .dashboard-page { padding:12px; }
  .page-header { align-items:flex-start; flex-direction:column; }
  .header-actions { width:100%; justify-content:space-between; }
  .overview-grid { grid-template-columns:1fr; }
  .metric-card { min-height:96px; }
  .trend-header { align-items:flex-start; flex-direction:column; }
  .trend-controls { width:100%; justify-content:flex-start; }
  .trend-controls :deep(.el-date-editor) { width:100%; }
  .panel-header { padding:13px 14px; }
  .panel-body { padding:10px 14px 14px; }
}
</style>
