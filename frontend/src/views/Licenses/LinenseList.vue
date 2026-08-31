<template>
  <div class="license-page">
    <header class="page-header">
      <div>
        <h1>{{ pageTitle }}</h1>
        <p>{{ t('pages.licenses.list.description') }}</p>
      </div>
      <el-button type="primary" @click="openCreate">
        {{ t('pages.licenses.list.createLicense') }}
      </el-button>
    </header>

    <section class="filter-card">
      <el-select
        v-model="filters.customerId"
        class="filter-control customer-filter"
        :placeholder="t('pages.licenses.list.filter.customerPlaceholder')"
        :loading="customerLoading"
        clearable
        filterable
      >
        <el-option
          v-for="customer in customerOptions"
          :key="customer.id"
          :label="customer.name"
          :value="customer.id"
        />
      </el-select>
      <el-input
        v-model="filters.code"
        class="filter-control code-filter"
        :placeholder="t('pages.licenses.list.filter.codePlaceholder')"
        clearable
        @keyup.enter="queryLicenses"
      />
      <el-select
        v-model="filters.status"
        class="filter-control status-filter"
        :placeholder="t('pages.licenses.list.filter.statusPlaceholder')"
        clearable
      >
        <el-option
          v-for="option in statusOptions"
          :key="option.key"
          :label="option.display"
          :value="option.key"
        />
      </el-select>
      <div class="filter-actions">
        <el-button type="primary" @click="queryLicenses">
          {{ t('pages.licenses.list.filter.query') }}
        </el-button>
        <el-button @click="resetFilters">
          {{ t('pages.licenses.list.filter.reset') }}
        </el-button>
      </div>
    </section>

    <section class="list-card">
      <div v-if="loadError" class="load-error">
        <span>{{ loadError }}</span>
        <el-button link type="primary" @click="loadLicenses">
          {{ t('pages.licenses.list.actions.retry') }}
        </el-button>
      </div>

      <el-table
        :data="licenses"
        v-loading="loading"
        :element-loading-text="t('pages.licenses.list.table.loading')"
        empty-text=" "
        stripe
      >
        <el-table-column prop="code" :label="t('pages.licenses.list.table.code')" min-width="210">
          <template #default="{ row }">
            <button class="code-link" type="button" @click="openDetail(row)">
              {{ row.code }}
            </button>
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.customer')" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.customer_name || row.customer_info?.customer_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.status')" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" effect="light">
              {{ statusText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.activationProgress')" min-width="170">
          <template #default="{ row }">
            <div class="activation-cell">
              <span>{{ activatedCount(row) }}/{{ row.max_activations }}</span>
              <el-progress :percentage="activationPercent(row)" :show-text="false" :stroke-width="6" />
            </div>
          </template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.endDate')" width="170">
          <template #default="{ row }">
            {{ row.end_date ? formatDate(row.end_date) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="description" :label="t('pages.licenses.list.table.description')" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column :label="t('pages.licenses.list.table.operation')" width="220" fixed="right" align="center">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button link type="primary" @click="openDetail(row)">
                {{ t('pages.licenses.list.actions.detail') }}
              </el-button>
              <el-button
                link
                :type="isLocked(row) ? 'success' : 'warning'"
                :loading="actionId === row.id"
                @click="toggleLock(row)"
              >
                {{ t(isLocked(row) ? 'pages.licenses.list.actions.unlock' : 'pages.licenses.list.actions.lock') }}
              </el-button>
              <el-button link type="danger" :disabled="actionId === row.id" @click="removeLicense(row)">
                {{ t('pages.licenses.list.actions.delete') }}
              </el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty :description="t('pages.licenses.list.table.empty')" />
        </template>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[16, 32, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="changePageSize"
          @current-change="loadLicenses"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { getCustomers } from '@/api/customer'
import { getAuthorizationStatusEnums, type RawEnumItem } from '@/api/enum'
import {
  deleteLicense,
  getLicenses,
  lockAuthorizationCode,
  type AuthorizationCode,
  type LicenseQueryRequest
} from '@/api/license'
import { formatDate } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'

interface CustomerOption {
  id: string
  name: string
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const routeCustomerId = String(route.query.customerId || '')
const routeCustomerName = String(route.query.customerName || '')
const filters = reactive({ customerId: routeCustomerId, code: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 16, total: 0 })
const licenses = ref<AuthorizationCode[]>([])
const customerOptions = ref<CustomerOption[]>([])
const statusOptions = ref<RawEnumItem[]>([])
const customerLoading = ref(false)
const loading = ref(false)
const loadError = ref('')
const actionId = ref('')

const selectedCustomerName = computed(() =>
  customerOptions.value.find(customer => customer.id === filters.customerId)?.name || routeCustomerName
)
const pageTitle = computed(() =>
  selectedCustomerName.value
    ? t('pages.licenses.list.customerTitle', { name: selectedCustomerName.value })
    : t('pages.licenses.list.title')
)

const activatedCount = (row: AuthorizationCode) =>
  row.current_activations ?? row.activated_licenses_count ?? 0

const activationPercent = (row: AuthorizationCode) => {
  if (!row.max_activations) return 0
  return Math.min(100, Math.round((activatedCount(row) / row.max_activations) * 100))
}

const isLocked = (row: AuthorizationCode) => row.is_locked === true || row.status === 'locked'

const statusText = (row: AuthorizationCode) =>
  row.status_display || t(`pages.licenses.list.status.${row.status}`)

const statusTagType = (status: AuthorizationCode['status']) => {
  if (status === 'normal') return 'success'
  if (status === 'locked') return 'warning'
  return 'danger'
}

const syncRouteCustomer = async () => {
  const query = { ...route.query }
  if (filters.customerId) {
    query.customerId = filters.customerId
    query.customerName = selectedCustomerName.value
  } else {
    delete query.customerId
    delete query.customerName
  }
  await router.replace({ query })
}

const openCreate = () => {
  router.push({
    name: 'licenses-create',
    query: filters.customerId
      ? { customerId: filters.customerId, customerName: selectedCustomerName.value }
      : undefined
  })
}

const openDetail = (row: AuthorizationCode) => {
  router.push({
    name: 'licenses-view',
    params: { id: row.id },
    query: filters.customerId
      ? { customerId: filters.customerId, customerName: selectedCustomerName.value }
      : undefined
  })
}

const queryLicenses = async () => {
  pagination.page = 1
  await syncRouteCustomer()
  await loadLicenses()
}

const resetFilters = async () => {
  filters.customerId = ''
  filters.code = ''
  filters.status = ''
  await queryLicenses()
}

const changePageSize = () => {
  pagination.page = 1
  loadLicenses()
}

const loadCustomers = async () => {
  customerLoading.value = true
  try {
    const response = await getCustomers({ page: 1, page_size: 100, status: 'active', sort: 'customer_name', order: 'asc' })
    customerOptions.value = response.data.list.map(customer => ({ id: customer.id, name: customer.customer_name }))
    if (routeCustomerId && routeCustomerName && !customerOptions.value.some(item => item.id === routeCustomerId)) {
      customerOptions.value.unshift({ id: routeCustomerId, name: routeCustomerName })
    }
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.list.message.customerLoadError')))
  } finally {
    customerLoading.value = false
  }
}

const loadStatuses = async () => {
  try {
    const response = await getAuthorizationStatusEnums()
    statusOptions.value = response.data.items
  } catch {
    statusOptions.value = (['normal', 'locked', 'expired'] as const).map(key => ({
      key,
      display: t(`pages.licenses.list.status.${key}`)
    }))
  }
}

const loadLicenses = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const params: LicenseQueryRequest = {
      page: pagination.page,
      page_size: pagination.pageSize,
      sort: 'created_at',
      order: 'desc'
    }
    if (filters.customerId) params.customer_id = filters.customerId
    if (filters.code.trim()) params.code = filters.code.trim()
    if (filters.status) params.status = filters.status as LicenseQueryRequest['status']

    const response = await getLicenses(params)
    licenses.value = response.data.list
    pagination.total = response.data.total
  } catch (error) {
    loadError.value = getErrorMessage(error, t('pages.licenses.list.message.loadError'))
  } finally {
    loading.value = false
  }
}

const isDialogCancel = (error: unknown) => error === 'cancel' || error === 'close'

const toggleLock = async (row: AuthorizationCode) => {
  const unlocking = isLocked(row)
  try {
    await ElMessageBox.confirm(
      t(unlocking ? 'pages.licenses.list.confirm.unlockMessage' : 'pages.licenses.list.confirm.lockMessage'),
      t(unlocking ? 'pages.licenses.list.confirm.unlockTitle' : 'pages.licenses.list.confirm.lockTitle'),
      {
        confirmButtonText: t('pages.licenses.list.confirm.confirm'),
        cancelButtonText: t('pages.licenses.list.confirm.cancel'),
        type: 'warning'
      }
    )
    actionId.value = row.id
    await lockAuthorizationCode(row.id, { is_locked: !unlocking })
    ElMessage.success(t(unlocking ? 'pages.licenses.list.message.unlockSuccess' : 'pages.licenses.list.message.lockSuccess'))
    await loadLicenses()
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(getErrorMessage(error, t(unlocking ? 'pages.licenses.list.message.unlockError' : 'pages.licenses.list.message.lockError')))
    }
  } finally {
    actionId.value = ''
  }
}

const removeLicense = async (row: AuthorizationCode) => {
  try {
    await ElMessageBox.confirm(
      t('pages.licenses.list.confirm.deleteMessage'),
      t('pages.licenses.list.confirm.deleteTitle'),
      {
        confirmButtonText: t('pages.licenses.list.confirm.deleteConfirm'),
        cancelButtonText: t('pages.licenses.list.confirm.cancel'),
        type: 'warning'
      }
    )
    actionId.value = row.id
    await deleteLicense(row.id)
    ElMessage.success(t('pages.licenses.list.message.deleteSuccess'))
    if (licenses.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadLicenses()
  } catch (error) {
    if (!isDialogCancel(error)) {
      ElMessage.error(getErrorMessage(error, t('pages.licenses.list.message.deleteError')))
    }
  } finally {
    actionId.value = ''
  }
}

onMounted(async () => {
  await Promise.all([loadCustomers(), loadStatuses()])
  await loadLicenses()
})
</script>

<style scoped lang="scss">
.license-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  padding: var(--layout-content-padding);
  box-sizing: border-box;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;

  h1 {
    margin: 0;
    color: var(--app-text-primary);
    font-size: 24px;
  }

  p {
    margin: 6px 0 0;
    color: var(--app-text-secondary);
    font-size: 14px;
  }
}

.filter-card,
.list-card {
  border: 1px solid var(--app-border-color);
  border-radius: 0;
  background: var(--app-content-bg);
}

.filter-card {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 16px;
}

.filter-control {
  width: 180px;
}

.customer-filter {
  width: 240px;
}

.code-filter {
  width: 220px;
}

.filter-actions {
  display: flex;
}

.list-card {
  min-width: 0;
  overflow: hidden;
}

.load-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 16px 16px 0;
  padding: 10px 12px;
  border-radius: 0;
  background: var(--el-color-danger-light-9);
  color: var(--el-color-danger);
}

.code-link {
  max-width: 100%;
  padding: 0;
  overflow: hidden;
  border: 0;
  background: none;
  color: var(--el-color-primary);
  font: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.activation-cell {
  display: grid;
  grid-template-columns: 52px minmax(72px, 1fr);
  align-items: center;
  gap: 8px;
  font-variant-numeric: tabular-nums;
}

.row-actions {
  display: flex;
  justify-content: center;
  white-space: nowrap;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  padding: 16px;
  overflow-x: auto;
}

@media (max-width: 768px) {
  .license-page {
    padding: 12px;
  }

  .page-header {
    align-items: stretch;
    flex-direction: column;
  }

  .filter-card,
  .filter-control,
  .customer-filter,
  .code-filter,
  .filter-actions,
  .filter-actions :deep(.el-button) {
    width: 100%;
  }

  .pagination {
    justify-content: flex-start;
  }
}
</style>
