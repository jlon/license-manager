<template>
  <Layout app-name="Cedar-V" :page-title="t('dashboard.title')">
    <div class="dashboard-page">
      <section class="overview-panel">
        <header class="page-header">
          <h1>{{ t('dashboard.heading') }}</h1>
          <div class="header-actions">
            <span v-if="lastUpdatedAt" class="updated-at">{{ t('dashboard.updatedAt', { time: lastUpdatedAt }) }}</span>
            <el-button :icon="Refresh" :loading="refreshing" @click="refreshAll">{{ t('dashboard.refresh') }}</el-button>
          </div>
        </header>

        <section class="overview-grid" :aria-label="t('dashboard.overviewTitle')">
          <article v-for="card in statCards" :key="card.key" class="metric-card" :class="`metric-card--${card.tone}`">
            <div class="metric-content">
              <span class="metric-label">{{ card.label }}</span>
              <el-skeleton v-if="statsLoading && !stats" animated :rows="1" />
              <strong v-else class="metric-value">{{ stats ? card.value : '--' }}</strong>
              <span class="metric-help">{{ stats ? card.help : statsError || t('dashboard.noData') }}</span>
            </div>
          </article>
        </section>
      </section>

      <section class="panel trend-panel">
        <div class="panel-header trend-header">
          <div>
            <h2>{{ t('dashboard.trendTitle') }}</h2>
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
          <template v-else>
            <div class="trend-layout">
              <div class="trend-chart-area">
                <LicenseTrendChart :data="trendData" :empty-text="t('dashboard.noTrendData')" />
              </div>
              <aside v-if="trendSummary" class="trend-summary" :aria-label="t('dashboard.trendSummary.title')">
                <h3>{{ t('dashboard.trendSummary.title') }}</h3>
                <div v-for="item in trendSummaryItems" :key="item.key" class="trend-summary__item">
                  <span>{{ item.label }}</span>
                  <strong>{{ item.value }}</strong>
                </div>
              </aside>
            </div>
          </template>
        </div>
      </section>

      <section class="panel recent-panel">
        <div class="panel-header">
          <div>
            <h2>{{ t('dashboard.recentLicenses.title') }}</h2>
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
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Refresh, WarningFilled } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import LicenseTrendChart from '@/components/charts/LicenseTrendChart.vue'
import {
  getAuthorizationTrend,
  getOverviewStats,
  getRecentAuthorizations,
  type RecentAuthorizationItem,
  type StatsOverviewData,
  type TrendDataItem,
  type TrendSummary
} from '@/api/dashboard'
import { formatDate } from '@/utils/date'

const { t } = useI18n()
const router = useRouter()
const stats = ref<StatsOverviewData | null>(null)
const recentData = ref<RecentAuthorizationItem[]>([])
const trendData = ref<TrendDataItem[]>([])
const trendSummary = ref<TrendSummary | null>(null)
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
const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

const statCards = computed(() => {
  const data = stats.value
  return [
    { key: 'total', label: t('dashboard.stats.totalAuthCodes'), value: data?.total_auth_codes ?? 0, help: t('dashboard.stats.subText.monthNew', { n: data?.month_new_auth_codes ?? 0 }), tone: 'brand' },
    { key: 'active', label: t('dashboard.stats.activeLicenses'), value: data?.active_licenses ?? 0, help: t('dashboard.stats.subText.vsLastMonth', { rate: formatRate(data?.growth_rate?.licenses_mom) }), tone: 'success' },
    { key: 'today', label: t('dashboard.stats.todayNewLicenses'), value: data?.today_new_licenses ?? 0, help: t('dashboard.stats.subText.yesterday', { n: data?.yesterday_new_licenses ?? 0 }), tone: 'neutral' },
    { key: 'due7', label: t('dashboard.stats.expiringIn7Days'), value: data?.expiring_in_7days ?? 0, help: (data?.expiring_in_7days ?? 0) > 0 ? t('dashboard.stats.subText.urgent') : t('dashboard.stats.subText.noRisk'), tone: (data?.expiring_in_7days ?? 0) > 0 ? 'danger' : 'neutral' },
    { key: 'due30', label: t('dashboard.stats.expiringIn30Days'), value: data?.expiring_in_30days ?? 0, help: t('dashboard.stats.subText.expiryReminder'), tone: (data?.expiring_in_30days ?? 0) > 0 ? 'warning' : 'neutral' },
    { key: 'alerts', label: t('dashboard.stats.abnormalAlerts'), value: data?.abnormal_alerts ?? 0, help: t('dashboard.stats.subText.heartbeatTimeout'), tone: (data?.abnormal_alerts ?? 0) > 0 ? 'danger' : 'neutral' }
  ]
})
const trendSummaryItems = computed(() => {
  const data = trendSummary.value
  return [
    { key: 'total', label: t('dashboard.trendSummary.total'), value: data?.total_count ?? 0 },
    { key: 'new', label: t('dashboard.trendSummary.new'), value: data?.new_count ?? 0 },
    { key: 'expired', label: t('dashboard.trendSummary.expired'), value: data?.expired_count ?? 0 },
    { key: 'growth', label: t('dashboard.trendSummary.growth'), value: `${formatRate(data?.growth_rate)}%` }
  ]
})

const formatRate = (value?: number) => {
  const rate = value ?? 0
  return `${rate >= 0 ? '+' : ''}${rate.toFixed(2)}`
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
  try {
    const params: Parameters<typeof getAuthorizationTrend>[0] = { type: trendType.value, timezone }
    if (trendType.value === 'custom' && customRange.value) {
      const [start_date, end_date] = customRange.value
      Object.assign(params, { start_date, end_date })
    }
    const response = (await getAuthorizationTrend(params)).data
    trendData.value = response.trend_data
    trendSummary.value = response.summary
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
.dashboard-page {
  --dashboard-border:#cfe0d9;
  --dashboard-canvas:#eef6f2;
  --dashboard-surface:#fff;
  --dashboard-shadow:0 7px 22px rgba(20,88,70,.055);
  --dashboard-text:var(--app-text-primary);
  --dashboard-regular:#40574e;
  --dashboard-secondary:#6f827a;
  --metric-label-color:#40574e;
  --metric-help-color:#6f827a;
  --metric-number-color:#183129;
  min-height:calc(100vh - var(--layout-header-height));
  padding:24px;
  display:flex;
  flex-direction:column;
  gap:18px;
  background:var(--dashboard-canvas);
}
.overview-panel { overflow:hidden; background:var(--dashboard-surface); border:1px solid var(--dashboard-border); box-shadow:var(--dashboard-shadow); }
.page-header { min-height:58px; padding:10px 20px; display:flex; align-items:center; justify-content:space-between; gap:20px; border-bottom:1px solid var(--dashboard-border); }
.page-header h1 { margin:0; color:var(--dashboard-text); font-size:20px; font-weight:650; line-height:1.3; letter-spacing:-.015em; }
.header-actions { display:flex; align-items:center; gap:12px; flex-shrink:0; }
.updated-at { color:var(--dashboard-secondary); font-size:12px; }
.overview-grid { position:relative; display:grid; grid-template-columns:1.2fr repeat(5,minmax(0,1fr)); overflow:hidden; background:var(--dashboard-surface); }
.metric-card { --metric:var(--metric-number-color); min-width:0; min-height:118px; padding:19px 20px; display:flex; align-items:center; background:var(--dashboard-surface); border-right:1px solid var(--dashboard-border); }
.metric-card:last-child { border-right:0; }
.metric-card--brand { --metric:#008c70; background:color-mix(in srgb,var(--dashboard-surface) 82%,#bcebd9); }
.metric-card--success { --metric:#13765f; }
.metric-card--warning { --metric:#c67d0a; }
.metric-card--danger { --metric:#cf4545; }
.metric-content { min-width:0; display:flex; flex:1; flex-direction:column; gap:3px; }
.metric-label { color:var(--metric-label-color); font-size:14px; font-weight:500; letter-spacing:.01em; }
.metric-value { color:var(--metric); font-size:32px; font-weight:700; font-variant-numeric:tabular-nums; letter-spacing:-.025em; line-height:1.15; }
.metric-help { min-height:18px; overflow:hidden; color:var(--metric-help-color); font-size:12px; text-overflow:ellipsis; white-space:nowrap; }
.panel { overflow:hidden; background:var(--dashboard-surface); border:1px solid var(--dashboard-border); border-radius:0; box-shadow:var(--dashboard-shadow); }
.panel-header { min-height:64px; padding:14px 20px; display:flex; align-items:center; justify-content:space-between; gap:16px; border-bottom:1px solid var(--dashboard-border); }
.panel-header h2 { margin:0; color:var(--dashboard-text); font-size:17px; }
.trend-controls { display:flex; align-items:center; justify-content:flex-end; gap:10px; flex-wrap:wrap; }
.panel-body { min-height:120px; padding:0; }
.chart-body { min-height:310px; }
.trend-layout { min-height:310px; display:grid; grid-template-columns:minmax(0,3fr) minmax(220px,1fr); }
.trend-chart-area { min-width:0; padding:12px 18px 8px; }
.trend-summary { padding:22px 20px; display:flex; flex-direction:column; justify-content:center; background:color-mix(in srgb,var(--dashboard-surface) 94%,#bcebd9); border-left:1px solid var(--dashboard-border); }
.trend-summary h3 { margin:0 0 12px; color:var(--dashboard-regular); font-size:13px; font-weight:600; }
.trend-summary__item { min-width:0; padding:13px 0; display:flex; align-items:center; justify-content:space-between; gap:12px; border-bottom:1px solid var(--dashboard-border); }
.trend-summary__item:last-child { border-bottom:0; }
.trend-summary__item span { overflow:hidden; color:var(--dashboard-secondary); font-size:12px; text-overflow:ellipsis; white-space:nowrap; }
.trend-summary__item strong { color:var(--dashboard-text); font-size:18px; font-variant-numeric:tabular-nums; }
.state-block { min-height:220px; display:flex; align-items:center; justify-content:center; gap:8px; color:var(--dashboard-secondary); }
.table-wrap { padding:0 20px 12px; overflow-x:auto; }
.table-wrap :deep(.el-table) { border-radius:0; }
.table-wrap :deep(.el-table__row) { cursor:pointer; }
.table-wrap :deep(.el-table__header th.el-table__cell) { background:color-mix(in srgb,var(--dashboard-surface) 82%,#cdeee1); color:var(--dashboard-regular); }
.table-wrap :deep(.el-table td.el-table__cell) { border-bottom-color:var(--dashboard-border); }
.table-wrap :deep(.el-table__body .cell) { color:var(--dashboard-regular); }
.dashboard-page :deep(.el-empty__description p) { color:var(--dashboard-secondary); }
.dashboard-page :deep(.el-radio-button__inner) { color:var(--dashboard-regular); }
.dashboard-page :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) { color:#fff; }
.dashboard-page :deep(.el-radio-button.is-active .el-radio-button__inner) { color:#fff; }
.dashboard-page :deep(.el-button:not(.el-button--primary):not(.is-link)) { color:var(--dashboard-regular); }
.status-tag { display:inline-flex; min-width:54px; padding:3px 9px; justify-content:center; border:1px solid currentColor; border-radius:0; font-size:12px; line-height:18px; }
.status-tag--normal { color:#287a54; background:#eaf8f0; }
.status-tag--locked { color:#a56605; background:#fff5df; }
.status-tag--expired { color:#b83d3d; background:#ffeded; }
.dashboard-page :deep(.el-button),
.dashboard-page :deep(.el-input__wrapper),
.dashboard-page :deep(.el-range-editor.el-input__wrapper),
.dashboard-page :deep(.el-radio-button__inner),
.dashboard-page :deep(.el-loading-mask) {
  border-radius:0;
}
.dashboard-page :deep(.el-button:not(.is-link)) {
  box-shadow:none;
}
.dashboard-page :deep(.el-button:not(.is-link):hover) {
  transform:none;
  box-shadow:inset 0 -2px 0 rgba(15, 23, 42, 0.08);
}
.dashboard-page :deep(.el-radio-button:first-child .el-radio-button__inner),
.dashboard-page :deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius:0;
}
:global([data-theme="dark"]) .dashboard-page {
  --dashboard-border:var(--app-border-color);
  --dashboard-canvas:color-mix(in srgb,var(--app-bg-color) 86%,#145848);
  --dashboard-surface:var(--app-content-bg);
  --dashboard-shadow:none;
  --dashboard-regular:#c1d0ca;
  --dashboard-secondary:#93a69e;
  --metric-label-color:#c1d0ca;
  --metric-help-color:#93a69e;
  --metric-number-color:#f0f5f3;
}
@media (max-width:1199px) {
  .overview-grid { grid-template-columns:repeat(3,minmax(0,1fr)); }
  .metric-card:nth-child(3n) { border-right:0; }
  .metric-card:nth-child(-n+3) { border-bottom:1px solid var(--dashboard-border); }
}
@media (max-width:1023px) {
  .trend-layout { grid-template-columns:1fr; }
  .trend-summary { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); padding:14px 18px; border-top:1px solid var(--dashboard-border); border-left:0; }
  .trend-summary h3 { grid-column:1/-1; }
  .trend-summary__item { padding:10px 12px; }
}
@media (max-width:767px) {
  .dashboard-page { padding:14px; }
  .page-header { align-items:flex-start; flex-direction:column; }
  .header-actions { width:100%; justify-content:space-between; }
  .overview-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }
  .metric-card { min-height:108px; padding:16px; border-bottom:1px solid var(--dashboard-border); }
  .metric-card:nth-child(3n) { border-right:1px solid var(--dashboard-border); }
  .metric-card:nth-child(2n) { border-right:0; }
  .metric-card:nth-last-child(-n+2) { border-bottom:0; }
  .trend-header { align-items:flex-start; flex-direction:column; }
  .trend-controls { width:100%; justify-content:flex-start; }
  .trend-controls :deep(.el-date-editor) { width:100%; }
  .panel-header { padding:13px 14px; }
  .trend-summary { grid-template-columns:1fr 1fr; gap:8px; }
  .trend-summary__item { padding:9px 10px; }
  .table-wrap { padding:0 14px 10px; }
}
</style>
