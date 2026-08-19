<template>
  <div class="customer-view-page">
    <header class="subpage-header">
      <div>
        <p class="breadcrumb">{{ t('customers.breadcrumb.customerManagement') }} /</p>
        <h1>{{ t('customers.viewCustomer') }}</h1>
      </div>
      <el-button @click="emit('back')">{{ t('customers.view.back') }}</el-button>
    </header>

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
          <div v-for="item in section.items" :key="item.label" class="info-item" :class="{ 'full-width': item.fullWidth }">
            <span class="info-label">{{ item.label }}</span>
            <span class="info-value" :class="{ multiline: item.multiline }">{{ item.value || '-' }}</span>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getCustomerDetail, type Customer } from '@/api/customer'
import { formatDate } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'

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
      { label: t('customers.view.customerCode'), value: customer.value?.customer_code },
      { label: t('customers.form.customerName'), value: customer.value?.customer_name },
      { label: t('customers.form.customerType'), value: customer.value?.customer_type_display },
      { label: t('customers.form.customerLevel'), value: customer.value?.customer_level_display },
      { label: t('customers.form.status'), value: customer.value?.status_display, fullWidth: true }
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
  },
  {
    title: t('customers.view.statusInfo'),
    items: [
      { label: t('customers.view.creator'), value: customer.value?.created_by },
      { label: t('customers.view.createTime'), value: formatTimestamp(customer.value?.created_at) },
      { label: t('customers.view.updater'), value: customer.value?.updated_by },
      { label: t('customers.view.updateTime'), value: formatTimestamp(customer.value?.updated_at) }
    ]
  }
])
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
.customer-view-page,.detail-sections { display:flex; flex-direction:column; gap:16px; }
.subpage-header { display:flex; align-items:flex-end; justify-content:space-between; gap:20px; }
.breadcrumb { margin:0 0 4px; color:var(--app-text-secondary); font-size:13px; }
.subpage-header h1 { margin:0; color:var(--app-text-primary); font-size:24px; line-height:1.35; }
.detail-sections { min-height:240px; }
.info-card,.state-card { padding:20px; background:var(--app-content-bg); border:1px solid var(--app-border-color); border-radius:var(--app-card-radius); box-shadow:var(--app-card-shadow); }
.info-card h2 { margin:0 0 18px; color:var(--app-text-primary); font-size:16px; }
.info-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:18px 32px; }
.info-item { display:flex; min-width:0; gap:12px; }
.info-item.full-width { grid-column:1 / -1; }
.info-label { flex:0 0 110px; color:var(--app-text-secondary); font-size:13px; }
.info-value { min-width:0; color:var(--app-text-primary); word-break:break-word; }
.info-value.multiline { white-space:pre-wrap; line-height:1.65; }
.stats-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:12px; }
.stat-item { display:flex; flex-direction:column; gap:8px; padding:14px 16px; background:var(--app-bg-color); border-radius:8px; }
.stat-item span { color:var(--app-text-secondary); font-size:13px; }
.stat-item strong { color:var(--app-text-primary); font-size:24px; line-height:1.2; }

@media (max-width:1024px) { .stats-grid { grid-template-columns:repeat(2,minmax(0,1fr)); } }
@media (max-width:768px) {
  .subpage-header { align-items:stretch; flex-direction:column; }
  .info-grid,.stats-grid { grid-template-columns:1fr; }
  .info-item.full-width { grid-column:auto; }
  .info-item { flex-direction:column; gap:4px; }
  .info-label { flex-basis:auto; }
  .info-card { padding:16px; }
}
</style>
