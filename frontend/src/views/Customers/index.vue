<template>
  <Layout app-name="Cedar-V" :page-title="pageTitle">
    <div v-if="viewMode === 'list'" class="customer-page">
      <header class="page-header">
        <div>
          <h1>{{ t('customers.title') }}</h1>
          <p>{{ t('customers.description') }}</p>
        </div>
        <el-button type="primary" @click="openCreate">
          {{ t('customers.actions.add') }}
        </el-button>
      </header>

      <section class="filter-card">
        <el-alert v-if="enumError" :title="enumError" type="warning" show-icon :closable="false">
          <template #default>
            <el-button link type="primary" @click="loadEnums">
              {{ t('customers.actions.retry') }}
            </el-button>
          </template>
        </el-alert>

        <div class="filter-grid">
          <el-select v-model="filters.customerType" :placeholder="t('customers.filter.customerType')" clearable>
            <el-option v-for="option in customerTypeOptions" :key="option.key" :label="option.display" :value="option.key" />
          </el-select>
          <el-select v-model="filters.customerLevel" :placeholder="t('customers.filter.customerLevel')" clearable>
            <el-option v-for="option in customerLevelOptions" :key="option.key" :label="option.display" :value="option.key" />
          </el-select>
          <el-select v-model="filters.status" :placeholder="t('customers.filter.status')" clearable>
            <el-option v-for="option in statusOptions" :key="option.key" :label="option.display" :value="option.key" />
          </el-select>
          <el-input
            v-model="filters.keyword"
            :placeholder="t('customers.search.placeholder')"
            clearable
            @keyup.enter="handleQuery"
          />
          <div class="filter-actions">
            <el-button type="primary" @click="handleQuery">{{ t('customers.actions.query') }}</el-button>
            <el-button @click="handleReset">{{ t('customers.actions.reset') }}</el-button>
          </div>
        </div>
      </section>

      <section class="table-card">
        <el-alert v-if="listError" :title="listError" type="error" show-icon :closable="false" class="list-alert">
          <template #default>
            <el-button link type="primary" @click="loadCustomers">
              {{ t('customers.actions.retry') }}
            </el-button>
          </template>
        </el-alert>

        <div class="table-scroll">
          <el-table
            v-loading="loading"
            :data="customers"
            :element-loading-text="t('customers.table.loading')"
            stripe
            row-key="id"
          >
            <el-table-column prop="customer_code" :label="t('customers.table.customerCode')" min-width="140" show-overflow-tooltip>
              <template #default="{ row }">
                <el-button link type="primary" @click="openDetail(row)">{{ row.customer_code }}</el-button>
              </template>
            </el-table-column>
            <el-table-column prop="customer_name" :label="t('customers.table.customerName')" min-width="180" show-overflow-tooltip />
            <el-table-column prop="customer_type_display" :label="t('customers.table.customerType')" min-width="120" />
            <el-table-column prop="contact_person" :label="t('customers.table.contactPerson')" min-width="120" show-overflow-tooltip />
            <el-table-column prop="customer_level_display" :label="t('customers.table.customerLevel')" min-width="120" />
            <el-table-column prop="status_display" :label="t('customers.table.status')" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" effect="light">{{ row.status_display }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('customers.table.createTime')" width="130" align="center">
              <template #default="{ row }">{{ formatDateShort(row.created_at) }}</template>
            </el-table-column>
            <el-table-column :label="t('customers.table.operation')" fixed="right" width="300" align="center" class-name="operation-column">
              <template #default="{ row }">
                <div class="desktop-actions">
                  <el-button link type="primary" @click="openLicenses(row)">{{ t('customers.actions.viewLicense') }}</el-button>
                  <el-button link type="primary" @click="openEdit(row)">{{ t('customers.actions.edit') }}</el-button>
                  <el-button
                    link
                    :type="row.status === 'active' ? 'warning' : 'success'"
                    :loading="actionId === row.id"
                    @click="toggleStatus(row)"
                  >
                    {{ row.status === 'active' ? t('customers.actions.disable') : t('customers.actions.enable') }}
                  </el-button>
                  <el-button link type="danger" :loading="actionId === row.id" @click="removeCustomer(row)">
                    {{ t('customers.actions.delete') }}
                  </el-button>
                </div>
                <el-dropdown class="compact-actions" trigger="click" @command="handleRowCommand($event, row)">
                  <el-button link type="primary">{{ t('customers.actions.more') }}</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="licenses">{{ t('customers.actions.viewLicense') }}</el-dropdown-item>
                      <el-dropdown-item command="edit">{{ t('customers.actions.edit') }}</el-dropdown-item>
                      <el-dropdown-item command="status">
                        {{ row.status === 'active' ? t('customers.actions.disable') : t('customers.actions.enable') }}
                      </el-dropdown-item>
                      <el-dropdown-item command="delete" divided>{{ t('customers.actions.delete') }}</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty :description="t('customers.table.empty')" :image-size="72" />
            </template>
          </el-table>
        </div>

        <div class="pagination-row">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :page-sizes="[16, 32, 50, 100]"
            :total="pagination.total"
            layout="total, sizes, prev, pager, next, jumper"
            :pager-count="5"
            @size-change="handleSizeChange"
            @current-change="loadCustomers"
          />
        </div>
      </section>
    </div>

    <CustomerForm
      v-else-if="viewMode === 'form'"
      :customer-id="isEditMode ? currentCustomerId : undefined"
      :is-edit="isEditMode"
      @save="handleFormSaved"
      @cancel="showList"
    />
    <CustomerView v-else :customer-id="currentCustomerId" @back="showList" />
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import Layout from '@/components/common/layout/Layout.vue'
import CustomerForm from './CustomerForm.vue'
import CustomerView from './CustomerView.vue'
import {
  deleteCustomer,
  getCustomers,
  toggleCustomerStatus,
  type Customer,
  type CustomerQueryRequest
} from '@/api/customer'
import {
  getCustomerLevelEnums,
  getCustomerTypeEnums,
  getStatusEnums,
  type RawEnumItem
} from '@/api/enum'
import { formatDateShort } from '@/utils/date'
import { getErrorMessage } from '@/utils/error'

type ViewMode = 'list' | 'form' | 'detail'
type RowCommand = 'licenses' | 'edit' | 'status' | 'delete'

const { t } = useI18n()
const router = useRouter()
const viewMode = ref<ViewMode>('list')
const isEditMode = ref(false)
const currentCustomerId = ref('')
const customers = ref<Customer[]>([])
const loading = ref(false)
const listError = ref('')
const enumError = ref('')
const actionId = ref('')
const filters = reactive({ customerType: '', customerLevel: '', status: '', keyword: '' })
const pagination = reactive({ page: 1, pageSize: 16, total: 0 })
const customerTypeOptions = ref<RawEnumItem[]>([])
const customerLevelOptions = ref<RawEnumItem[]>([])
const statusOptions = ref<RawEnumItem[]>([])

const pageTitle = computed(() => {
  if (viewMode.value === 'detail') return t('customers.viewCustomer')
  if (viewMode.value === 'form') return isEditMode.value ? t('customers.editCustomer') : t('customers.addCustomer')
  return t('customers.title')
})

const loadEnums = async () => {
  enumError.value = ''
  try {
    const [typeResponse, levelResponse, statusResponse] = await Promise.all([
      getCustomerTypeEnums(),
      getCustomerLevelEnums(),
      getStatusEnums()
    ])
    customerTypeOptions.value = typeResponse.data.items
    customerLevelOptions.value = levelResponse.data.items
    statusOptions.value = statusResponse.data.items
  } catch (error) {
    enumError.value = getErrorMessage(error, t('customers.message.loadEnumError'))
  }
}

const loadCustomers = async () => {
  loading.value = true
  listError.value = ''
  const params: CustomerQueryRequest = {
    page: pagination.page,
    page_size: pagination.pageSize,
    search: filters.keyword.trim() || undefined,
    customer_type: (filters.customerType || undefined) as CustomerQueryRequest['customer_type'],
    customer_level: filters.customerLevel || undefined,
    status: (filters.status || undefined) as CustomerQueryRequest['status']
  }
  try {
    const response = await getCustomers(params)
    customers.value = response.data.list
    pagination.total = response.data.total
  } catch (error) {
    listError.value = getErrorMessage(error, t('customers.message.loadError'))
  } finally {
    loading.value = false
  }
}

const handleQuery = () => {
  pagination.page = 1
  loadCustomers()
}
const handleReset = () => {
  Object.assign(filters, { customerType: '', customerLevel: '', status: '', keyword: '' })
  pagination.page = 1
  loadCustomers()
}
const handleSizeChange = () => {
  pagination.page = 1
  loadCustomers()
}

const openCreate = () => {
  isEditMode.value = false
  currentCustomerId.value = ''
  viewMode.value = 'form'
}

const openEdit = (customer: Customer) => {
  isEditMode.value = true
  currentCustomerId.value = customer.id
  viewMode.value = 'form'
}

const openDetail = (customer: Customer) => {
  currentCustomerId.value = customer.id
  viewMode.value = 'detail'
}

const openLicenses = (customer: Customer) => {
  router.push({
    path: '/licenses/list',
    query: { customerId: customer.id, customerName: customer.customer_name }
  })
}

const showList = () => {
  viewMode.value = 'list'
}

const handleFormSaved = async () => {
  showList()
  await loadCustomers()
}

const toggleStatus = async (customer: Customer) => {
  const nextStatus = customer.status === 'active' ? 'disabled' : 'active'
  const isDisabling = nextStatus === 'disabled'
  try {
    await ElMessageBox.confirm(
      t(isDisabling ? 'customers.confirm.disableMessage' : 'customers.confirm.enableMessage', { name: customer.customer_name }),
      t(isDisabling ? 'customers.confirm.disableTitle' : 'customers.confirm.enableTitle'),
      { confirmButtonText: t('customers.confirm.confirm'), cancelButtonText: t('customers.confirm.cancel'), type: 'warning' }
    )
  } catch { return }

  actionId.value = customer.id
  try {
    const response = await toggleCustomerStatus(customer.id, nextStatus)
    ElMessage.success(response.message)
    await loadCustomers()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('customers.message.statusError')))
  } finally {
    actionId.value = ''
  }
}

const removeCustomer = async (customer: Customer) => {
  try {
    await ElMessageBox.confirm(
      t('customers.confirm.deleteMessage', { name: customer.customer_name }),
      t('customers.confirm.deleteTitle'),
      { confirmButtonText: t('customers.confirm.confirm'), cancelButtonText: t('customers.confirm.cancel'), type: 'warning' }
    )
  } catch { return }

  actionId.value = customer.id
  try {
    const response = await deleteCustomer(customer.id)
    ElMessage.success(response.message)
    if (customers.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadCustomers()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('customers.message.deleteError')))
  } finally {
    actionId.value = ''
  }
}

const handleRowCommand = (command: RowCommand, customer: Customer) => {
  if (command === 'licenses') openLicenses(customer)
  if (command === 'edit') openEdit(customer)
  if (command === 'status') toggleStatus(customer)
  if (command === 'delete') removeCustomer(customer)
}

onMounted(() => {
  loadEnums()
  loadCustomers()
})
</script>

<style scoped>
.customer-page { display:flex; flex-direction:column; gap:16px; min-width:0; padding:var(--layout-content-padding); box-sizing:border-box; }
.page-header { display:flex; align-items:flex-end; justify-content:space-between; gap:20px; }
.page-header h1 { margin:0; color:var(--app-text-primary); font-size:24px; line-height:1.35; }
.page-header p { margin:4px 0 0; color:var(--app-text-secondary); font-size:13px; }
.filter-card,.table-card { background:var(--app-content-bg); border:1px solid var(--app-border-color); border-radius:var(--app-card-radius); box-shadow:var(--app-card-shadow); }
.filter-card { display:flex; flex-direction:column; gap:12px; padding:16px; }
.filter-grid { display:grid; grid-template-columns:repeat(3,minmax(140px,180px)) minmax(220px,1fr) auto; gap:12px; align-items:center; }
.filter-actions { display:flex; justify-content:flex-end; gap:8px; }
.table-card { min-width:0; overflow:hidden; }
.list-alert { margin:16px 16px 0; }
.table-scroll { width:100%; overflow-x:auto; }
.desktop-actions { display:flex; align-items:center; justify-content:center; white-space:nowrap; }
.compact-actions { display:none; }
.pagination-row { display:flex; justify-content:flex-end; padding:16px; border-top:1px solid var(--app-border-color); overflow-x:auto; }

@media (max-width:1200px) {
  .filter-grid { grid-template-columns:repeat(2,minmax(0,1fr)); }
  .filter-actions { justify-content:flex-start; }
  :deep(.operation-column) { width:100px !important; }
  .desktop-actions { display:none; }
  .compact-actions { display:inline-flex; }
}

@media (max-width:768px) {
  .customer-page { padding:12px; }
  .page-header { align-items:stretch; flex-direction:column; }
  .page-header .el-button { width:100%; }
  .filter-grid { grid-template-columns:1fr; }
  .filter-actions .el-button { flex:1; }
  .pagination-row { justify-content:flex-start; }
}
</style>
