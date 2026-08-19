<template>
  <Layout app-name="Cedar-V" :page-title="t('enterpriseLeads.title')">
    <div class="lead-page">
      <header class="page-header">
        <div>
          <h1>{{ t('enterpriseLeads.title') }}</h1>
          <p>{{ t('enterpriseLeads.description') }}</p>
        </div>
        <el-button type="primary" :icon="Refresh" :loading="loading || summaryLoading" @click="handleRefresh">
          {{ t('enterpriseLeads.actions.refresh') }}
        </el-button>
      </header>

      <section class="metrics-grid" v-loading="summaryLoading">
        <article v-for="stat in stats" :key="stat.key" class="metric-card" :class="`metric-card--${stat.tone}`">
          <div class="metric-icon">
            <el-icon><component :is="stat.icon" /></el-icon>
          </div>
          <div class="metric-content">
            <span class="metric-label">{{ t(`enterpriseLeads.stats.${stat.key}`) }}</span>
            <strong class="metric-value">{{ stat.value.toLocaleString() }}</strong>
          </div>
        </article>
      </section>

      <section class="filter-card">
        <div class="filter-grid">
          <el-input
            v-model="filters.keyword"
            :placeholder="t('enterpriseLeads.filter.searchPlaceholder')"
            clearable
            @keyup.enter="handleFilter"
          />
          <el-select v-model="filters.status" :placeholder="t('enterpriseLeads.filter.statusPlaceholder')" clearable>
            <el-option :label="t('enterpriseLeads.filter.allStatus')" value="" />
            <el-option :label="t('enterpriseLeads.status.pending')" value="pending" />
            <el-option :label="t('enterpriseLeads.status.contacting')" value="contacting" />
            <el-option :label="t('enterpriseLeads.status.completed')" value="completed" />
            <el-option :label="t('enterpriseLeads.status.rejected')" value="rejected" />
          </el-select>
          <div class="filter-actions">
            <el-button type="primary" :icon="Search" @click="handleFilter">{{ t('enterpriseLeads.actions.query') }}</el-button>
            <el-button @click="handleReset">{{ t('enterpriseLeads.actions.reset') }}</el-button>
          </div>
        </div>
      </section>

      <section class="table-card">
        <div class="table-heading">
          <div>
            <h2>{{ t('enterpriseLeads.table.title') }}</h2>
            <span>{{ total.toLocaleString() }} {{ t('enterpriseLeads.table.unit') }}</span>
          </div>
        </div>

        <el-alert v-if="listError" :title="listError" type="error" show-icon :closable="false" class="list-alert">
          <template #default>
            <el-button link type="primary" @click="fetchData">{{ t('enterpriseLeads.actions.retry') }}</el-button>
          </template>
        </el-alert>

        <div class="table-scroll">
          <el-table v-loading="loading" :data="tableData" stripe row-key="id">
            <el-table-column prop="id" :label="t('enterpriseLeads.table.id')" min-width="150" show-overflow-tooltip />
            <el-table-column prop="company_name" :label="t('enterpriseLeads.table.company')" min-width="180" show-overflow-tooltip />
            <el-table-column prop="contact_name" :label="t('enterpriseLeads.table.contact')" min-width="110" />
            <el-table-column prop="contact_phone" :label="t('enterpriseLeads.table.phone')" min-width="140" />
            <el-table-column prop="created_at" :label="t('enterpriseLeads.table.submittedAt')" min-width="170" />
            <el-table-column :label="t('enterpriseLeads.table.status')" width="110" align="center">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)" effect="light">
                  {{ t(`enterpriseLeads.status.${row.status}`) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="t('enterpriseLeads.table.actions')" width="210" fixed="right" align="center" class-name="operation-column">
              <template #default="{ row }">
                <div class="desktop-actions">
                  <el-button link type="primary" @click="handleView(row)">{{ t('enterpriseLeads.actions.view') }}</el-button>
                  <el-button link type="primary" @click="handleEdit(row)">{{ t('enterpriseLeads.actions.edit') }}</el-button>
                  <el-button link type="danger" @click="handleDelete(row)">{{ t('enterpriseLeads.actions.delete') }}</el-button>
                </div>
                <el-dropdown class="compact-actions" trigger="click" @command="handleRowCommand($event, row)">
                  <el-button link type="primary">{{ t('enterpriseLeads.actions.more') }}</el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="view">{{ t('enterpriseLeads.actions.view') }}</el-dropdown-item>
                      <el-dropdown-item command="edit">{{ t('enterpriseLeads.actions.edit') }}</el-dropdown-item>
                      <el-dropdown-item command="delete" divided>{{ t('enterpriseLeads.actions.delete') }}</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty :description="t('enterpriseLeads.table.empty')" :image-size="72" />
            </template>
          </el-table>
        </div>

        <div class="pagination-row">
          <el-pagination
            v-model:current-page="page"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            :pager-count="5"
            :total="total"
            @current-change="fetchData"
            @size-change="handleSizeChange"
          />
        </div>
      </section>
    </div>

    <LeadDetailDialog
      v-model="detailVisible"
      :id="selectedLead?.id ?? null"
    />

    <LeadEditDialog
      v-model="editVisible"
      :id="selectedLead?.id ?? null"
      @save="handleUpdate"
    />
  </Layout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Layout from '@/components/common/layout/Layout.vue'
import { Phone, CircleCheck, Refresh, User, OfficeBuilding, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import LeadDetailDialog from './components/LeadDetailDialog.vue'
import LeadEditDialog from './components/LeadEditDialog.vue'
import { getLeads, getLeadSummary, updateLead, deleteLead, type Lead } from '@/api/lead'
import { formatDateTime } from '@/utils/date'

const { t } = useI18n()

const stats = ref([
  { key: 'total', value: 0, icon: OfficeBuilding, field: 'total_count', tone: 'brand' },
  { key: 'pending', value: 0, icon: Phone, field: 'pending_count', tone: 'warning' },
  { key: 'contacting', value: 0, icon: CircleCheck, field: 'contacted_count', tone: 'info' },
  { key: 'completed', value: 0, icon: User, field: 'converted_count', tone: 'success' }
])

const filters = ref({
  keyword: '',
  status: ''
})

const tableData = ref<Lead[]>([])
const loading = ref(false)
const summaryLoading = ref(false)
const listError = ref('')
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

const detailVisible = ref(false)
const editVisible = ref(false)
const selectedLead = ref<Lead | null>(null)

const fetchSummary = async () => {
  summaryLoading.value = true
  try {
    const res = await getLeadSummary()
    if (res.code === '000000' && res.data) {
      stats.value.forEach(item => {
        if (res.data[item.field] !== undefined) {
          item.value = res.data[item.field]
        }
      })
    }
  } catch (error) {
    console.error('Fetch lead summary error:', error)
  } finally {
    summaryLoading.value = false
  }
}

const fetchData = async () => {
  loading.value = true
  listError.value = ''
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      search: filters.value.keyword,
      status: filters.value.status
    }
    const res = await getLeads(params)
    if (res.code === '000000' && res.data) {
      tableData.value = res.data.leads.map((item: any) => ({
        ...item,
        created_at: formatDateTime(item.created_at)
      }))
      total.value = res.data.total_count
    }
  } catch (error: any) {
    console.error('Fetch leads error:', error)
    listError.value = error.backendMessage || t('enterpriseLeads.messages.fetchError')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
  fetchSummary()
})

const handleFilter = () => {
  if (page.value === 1) fetchData()
  else page.value = 1
}

const handleReset = () => {
  filters.value.keyword = ''
  filters.value.status = ''
  handleFilter()
}

const handleSizeChange = () => {
  if (page.value === 1) fetchData()
  else page.value = 1
}

const handleRefresh = () => {
  fetchData()
  fetchSummary()
}

const handleView = (row: Lead) => {
  selectedLead.value = row
  detailVisible.value = true
}

const handleEdit = (row: Lead) => {
  selectedLead.value = row
  editVisible.value = true
}

const handleRowCommand = (command: string, row: Lead) => {
  if (command === 'view') handleView(row)
  else if (command === 'edit') handleEdit(row)
  else if (command === 'delete') handleDelete(row)
}

const statusTagType = (status: string): 'primary' | 'success' | 'warning' | 'info' => {
  if (status === 'completed') return 'success'
  if (status === 'pending') return 'warning'
  if (status === 'rejected') return 'info'
  return 'primary'
}

const handleDelete = (row: Lead) => {
  ElMessageBox.confirm(
    t('enterpriseLeads.messages.deleteConfirm'),
    t('common.confirm'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then(async () => {
    try {
      const res = await deleteLead(row.id)
      if (res.code === '000000') {
        ElMessage.success(t('enterpriseLeads.messages.deleteSuccess'))
        fetchData()
        fetchSummary()
      }
    } catch (error: any) {
      console.error('Delete lead error:', error)
      ElMessage.error(error.backendMessage || t('enterpriseLeads.messages.deleteError'))
    }
  }).catch(() => {
    // Cancelled
  })
}

const handleUpdate = async (updatedData: any) => {
  try {
    const { id, ...data } = updatedData
    // 格式化日期为 ISO 8601 格式 (2026-02-05T10:00:00Z)，去除毫秒
    if (data.follow_up_date) {
      data.follow_up_date = new Date(data.follow_up_date).toISOString().replace(/\.\d{3}/, '')
    }
    const res = await updateLead(id, data)
    if (res.code === '000000') {
      ElMessage.success(t('enterpriseLeads.messages.updateSuccess'))
      fetchData()
      fetchSummary()
    }
  } catch (error: any) {
    console.error('Update lead error:', error)
    ElMessage.error(error.backendMessage || t('enterpriseLeads.messages.updateError'))
  }
}
</script>

<style scoped>
.lead-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  padding: var(--layout-content-padding);
  box-sizing: border-box;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
}

.page-header h1,
.table-heading h2 {
  margin: 0;
  color: var(--app-text-primary);
}

.page-header h1 {
  font-size: 24px;
  line-height: 1.35;
}

.page-header p {
  margin: 4px 0 0;
  color: var(--app-text-secondary);
  font-size: 13px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  min-height: 96px;
}

.metric-card {
  --metric-color: var(--el-color-primary);
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 14px;
  padding: 16px;
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-color);
  border-left: 4px solid var(--metric-color);
  border-radius: var(--app-card-radius);
  box-shadow: var(--app-card-shadow);
}

.metric-card--warning { --metric-color: var(--el-color-warning); }
.metric-card--info { --metric-color: var(--el-color-info); }
.metric-card--success { --metric-color: var(--el-color-success); }

.metric-icon {
  display: flex;
  width: 42px;
  height: 42px;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: color-mix(in srgb, var(--metric-color) 12%, transparent);
  color: var(--metric-color);
  font-size: 22px;
}

.metric-content {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 3px;
}

.metric-label {
  overflow: hidden;
  color: var(--app-text-secondary);
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.metric-value {
  color: var(--app-text-primary);
  font-size: 26px;
  line-height: 1.2;
}

.filter-card,
.table-card {
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-color);
  border-radius: var(--app-card-radius);
  box-shadow: var(--app-card-shadow);
}

.filter-card {
  padding: 16px;
}

.filter-grid {
  display: grid;
  grid-template-columns: minmax(240px, 1fr) minmax(160px, 220px) auto;
  gap: 12px;
  align-items: center;
}

.filter-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.table-card {
  min-width: 0;
  overflow: hidden;
}

.table-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--app-border-color);
}

.table-heading > div {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.table-heading h2 {
  font-size: 16px;
}

.table-heading span {
  color: var(--app-text-secondary);
  font-size: 12px;
}

.list-alert {
  margin: 16px 16px 0;
}

.table-scroll {
  width: 100%;
  overflow-x: auto;
}

.desktop-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

.compact-actions {
  display: none;
}

.pagination-row {
  display: flex;
  justify-content: flex-end;
  padding: 16px;
  border-top: 1px solid var(--app-border-color);
  overflow-x: auto;
}

@media (max-width: 1200px) {
  .metrics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  :deep(.operation-column) {
    width: 100px !important;
  }

  .desktop-actions {
    display: none;
  }

  .compact-actions {
    display: inline-flex;
  }
}

@media (max-width: 768px) {
  .lead-page {
    padding: 12px;
  }

  .page-header {
    align-items: stretch;
    flex-direction: column;
  }

  .page-header .el-button {
    width: 100%;
  }

  .metrics-grid,
  .filter-grid {
    grid-template-columns: 1fr;
  }

  .filter-actions .el-button {
    flex: 1;
  }

  .pagination-row {
    justify-content: flex-start;
  }
}

@media (max-width: 480px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
