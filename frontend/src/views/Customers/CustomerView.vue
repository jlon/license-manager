<template>
  <div class="customer-view-page">
    <SubpageHeader
      :title="t('customers.viewCustomer')"
      :back-label="t('customers.actions.backToList')"
      @back="emit('back')"
    />

    <section v-if="error" class="state-card">
      <el-result icon="error" :title="t('customers.message.getDetailFailed')" :sub-title="error">
        <template #extra>
          <el-button type="primary" @click="loadCustomer">{{ t('customers.actions.retry') }}</el-button>
          <el-button @click="emit('back')">{{ t('customers.view.back') }}</el-button>
        </template>
      </el-result>
    </section>

    <div v-else v-loading="loading" class="detail-sections">
      <section v-for="section in detailSections" :key="section.title" class="info-card">
        <h2>{{ section.title }}</h2>
        <div class="info-grid">
          <div
            v-for="item in section.items"
            :key="item.label"
            class="info-item"
            :class="{ 'full-width': item.fullWidth, 'is-code': item.kind === 'code', 'is-status': item.kind === 'status' }"
          >
            <span class="info-label">{{ item.label }}</span>
            <span class="info-value" :class="{ multiline: item.multiline }">
              <el-rate
                v-if="item.kind === 'level'"
                :model-value="Number(item.value) || 0"
                disabled
                class="customer-level-rate"
              />
              <template v-else>{{ item.value || '-' }}</template>
            </span>
          </div>
        </div>
      </section>

      <section class="info-card">
        <h2>{{ t('customers.view.licenseStats') }}</h2>
        <div class="stats-grid">
          <div v-for="item in stats" :key="item.label" class="stat-item">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
          </div>
        </div>
      </section>

      <section class="info-card">
        <h2>{{ statusSection.title }}</h2>
        <div class="info-grid">
          <div v-for="item in statusSection.items" :key="item.label" class="info-item">
            <span class="info-label">{{ item.label }}</span>
            <span class="info-value">{{ item.value || '-' }}</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getCustomerDetail, type Customer } from '@/api/customer'
import { formatDate } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'
import SubpageHeader from '@/components/common/SubpageHeader.vue'

const props = defineProps<{ customerId: string }>()
const emit = defineEmits<{ back: [] }>()
const { t } = useI18n()
const loading = ref(false)
const error = ref('')
const customer = ref<Customer>()

interface DetailItem {
  label: string
  value?: string
  fullWidth?: boolean
  multiline?: boolean
  kind?: 'code' | 'status' | 'level'
}

interface DetailSection {
  title: string
  items: DetailItem[]
}

const formatTimestamp = (value?: string) => value ? formatDate(value) : '-'
const detailSections = computed<DetailSection[]>(() => [
  {
    title: t('customers.form.basicInfo'),
    items: [
      { label: t('customers.view.customerCode'), value: customer.value?.customer_code, kind: 'code' },
      { label: t('customers.form.customerName'), value: customer.value?.customer_name },
      { label: t('customers.form.customerType'), value: customer.value?.customer_type_display },
      { label: t('customers.form.customerLevel'), value: customer.value?.customer_level_display, kind: 'level' },
      { label: t('customers.form.status'), value: customer.value?.status_display, kind: 'status' }
    ]
  },
  {
    title: t('customers.form.contactInfo'),
    items: [
      { label: t('customers.form.contactPerson'), value: customer.value?.contact_person },
      { label: t('customers.form.phone'), value: customer.value?.phone },
      { label: t('customers.form.email'), value: customer.value?.email },
      { label: t('customers.form.address'), value: customer.value?.address, fullWidth: true }
    ]
  },
  {
    title: t('customers.form.businessInfo'),
    items: [
      { label: t('customers.form.companySize'), value: customer.value?.company_size_display },
      { label: t('customers.form.description'), value: customer.value?.description, fullWidth: true, multiline: true }
    ]
  }
])
const statusSection = computed<DetailSection>(() => ({
  title: t('customers.view.statusInfo'),
  items: [
    { label: t('customers.view.creator'), value: customer.value?.created_by },
    { label: t('customers.view.createTime'), value: formatTimestamp(customer.value?.created_at) },
    { label: t('customers.view.updater'), value: customer.value?.updated_by },
    { label: t('customers.view.updateTime'), value: formatTimestamp(customer.value?.updated_at) }
  ]
}))
const stats = computed(() => {
  const values = customer.value?.authorization_stats
  return [
    { label: t('customers.view.totalAuthCodes'), value: values?.total_auth_codes || 0 },
    { label: t('customers.view.expiredAuthCodes'), value: values?.expired_auth_codes || 0 },
    { label: t('customers.view.expiringAuthCodes'), value: values?.expiring_soon_auth_codes || 0 },
    { label: t('customers.view.totalLicenseCount'), value: values?.total_licenses || 0 },
    { label: t('customers.view.activeLicenseCount'), value: values?.active_licenses || 0 },
    { label: t('customers.view.inactiveLicenseCount'), value: values?.inactive_licenses || 0 },
    { label: t('customers.view.expiredLicenseCount'), value: values?.expired_licenses || 0 }
  ]
})

const loadCustomer = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await getCustomerDetail(props.customerId)
    if (!response.data) throw new Error(t('customers.view.customerNotExist'))
    customer.value = response.data
  } catch (loadError) {
    error.value = getErrorMessage(loadError, t('customers.message.getDetailFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(loadCustomer)
</script>

<style scoped>
.customer-view-page,.detail-sections { display:flex; flex-direction:column; gap:14px; }
.customer-view-page { min-width:0; padding:22px var(--layout-content-padding) 28px; box-sizing:border-box; }
.customer-view-page :deep(.subpage-header__title-row h1) { color:#20372f; font-size:22px; letter-spacing:-.015em; }
.customer-view-page :deep(.subpage-header__back) { color:#587066; font-size:13px; font-weight:500; }
.detail-sections { min-height:240px; }
.info-card,.state-card { overflow:hidden; background:#fff; border:1px solid #bfd3ca; border-radius:var(--app-card-radius); box-shadow:0 3px 12px rgba(29,72,57,.07); }
.info-card h2 { margin:0; padding:13px 18px; color:#294238; background:#f8faf9; border-bottom:1px solid #cbded6; border-left:3px solid var(--el-color-primary); font-size:15px; font-weight:650; line-height:1.4; }
.info-grid { padding:15px 18px; display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:12px 28px; }
.info-item { min-width:0; min-height:32px; display:grid; grid-template-columns:76px minmax(0,1fr); align-items:center; column-gap:12px; }
.info-item.full-width { grid-column:1 / -1; }
.info-label { color:#587067; font-size:13px; font-weight:500; line-height:1.5; white-space:nowrap; }
.info-value { min-width:0; color:#233a31; font-size:14px; font-weight:500; line-height:1.55; word-break:break-word; }
.info-value.multiline { white-space:pre-wrap; line-height:1.65; }
.customer-level-rate { height:22px; }
.customer-level-rate :deep(.el-rate__item) { margin-right:4px; }
.customer-level-rate :deep(.el-rate__icon) { margin-right:0; color:#f5a623; font-size:18px; }
.customer-level-rate :deep(.el-rate__icon:not(.is-active)) { color:#e8eeeb; }
.info-item.is-code .info-value { color:#008f72; font-family:ui-monospace,SFMono-Regular,Consolas,monospace; font-size:13px; font-weight:600; }
.info-item.is-status .info-value { width:max-content; padding:2px 9px; color:#167a5c; background:#e9f7f1; border:1px solid #d0ecdf; font-size:12px; font-weight:600; line-height:1.6; }
.stats-grid { padding:16px 18px 18px; display:grid; grid-template-columns:repeat(7,minmax(0,1fr)); gap:10px; background:#fff; }
.stat-item { position:relative; min-width:0; min-height:78px; padding:13px 15px; display:flex; flex-direction:column; justify-content:space-between; gap:8px; background:#f8fbfa; border:1px solid #d8e6e0; }
.stat-item::before { position:absolute; top:-1px; right:-1px; left:-1px; height:2px; background:#91cbbb; content:''; }
.stat-item span { overflow:hidden; color:#557067; font-size:13px; font-weight:500; line-height:1.45; text-overflow:ellipsis; white-space:nowrap; }
.stat-item strong { color:#087f66; font-size:22px; font-weight:700; font-variant-numeric:tabular-nums; line-height:1.15; }
.stat-item:nth-child(2)::before,
.stat-item:nth-child(7)::before { background:#e58a83; }
.stat-item:nth-child(2) strong,
.stat-item:nth-child(7) strong { color:#c94f48; }
.stat-item:nth-child(3)::before { background:#e6ad55; }
.stat-item:nth-child(3) strong { color:#b87318; }

[data-theme='dark'] .info-card h2,
[data-theme='dark'] .stat-item,
[data-theme='dark'] .stats-grid { background:color-mix(in srgb,var(--app-content-bg) 88%,#145848); }
[data-theme='dark'] .info-card,
[data-theme='dark'] .state-card { background:var(--app-content-bg); border-color:var(--app-border-color); }
[data-theme='dark'] .info-label { color:#93a69e; }
[data-theme='dark'] .info-value,
[data-theme='dark'] .stat-item strong { color:var(--app-text-primary); }

@media (max-width:1400px) {
  .stats-grid { grid-template-columns:repeat(4,minmax(0,1fr)); }
}
@media (max-width:1200px) { .info-grid { grid-template-columns:repeat(2,minmax(0,1fr)); } }
@media (max-width:768px) {
  .customer-view-page { padding:12px; }
  .info-grid,.stats-grid { grid-template-columns:1fr; }
  .info-item.full-width { grid-column:auto; }
  .info-grid { padding:14px 16px 4px; }
  .info-item { min-height:auto; grid-template-columns:72px minmax(0,1fr); }
  .stat-item { min-height:72px; border-right:0; }
}
</style>
