<template>
  <Layout app-name="Cedar-V" :page-title="t('pages.devices.list.title')">
    <div class="license-page data-list-page">
      <header class="data-list-header">
        <div>
          <h1>{{ t('pages.devices.list.title') }}</h1>
          <p>{{ t('pages.devices.list.description') }}</p>
        </div>
      </header>

      <section class="data-list-workspace">
        <div class="data-list-toolbar">
          <el-button :icon="Refresh" :loading="loading" @click="loadDevices">
            {{ t('pages.devices.list.actions.refresh') }}
          </el-button>
          <span class="data-list-toolbar__divider" aria-hidden="true"></span>
          <div class="filter-grid data-list-filter__grid data-list-toolbar__filters">
            <el-select
              v-model="filters.status"
              class="filter-control status-filter"
              :placeholder="t('pages.devices.list.filter.statusPlaceholder')"
              clearable
            >
              <el-option
                v-for="option in statusOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
            <el-select
              v-model="filters.isOnline"
              class="filter-control status-filter"
              :placeholder="t('pages.devices.list.filter.onlinePlaceholder')"
              clearable
            >
              <el-option :label="t('pages.devices.list.online.online')" value="true" />
              <el-option :label="t('pages.devices.list.online.offline')" value="false" />
            </el-select>
            <div class="filter-actions data-list-filter__actions">
              <el-button plain type="primary" :loading="isUpdating" @click="queryDevices">
                {{ t('pages.devices.list.filter.query') }}
              </el-button>
              <el-button @click="resetFilters">
                {{ t('pages.devices.list.filter.reset') }}
              </el-button>
            </div>
          </div>
        </div>

        <div class="list-card data-list-table">
          <div v-if="loadError" class="load-error data-list-alert">
            <span>{{ loadError }}</span>
            <el-button link type="primary" @click="loadDevices">
              {{ t('pages.devices.list.actions.retry') }}
            </el-button>
          </div>

          <div class="data-list-table__scroll">
            <el-table v-loading="skeletonVisible" :data="devices" row-key="id">
              <el-table-column :label="t('pages.devices.list.table.licenseKey')" width="290">
                <template #default="{ row }">
                  <div class="device-mono">
                    <span class="device-mono__text">{{ shortenKey(row.license_key) }}</span>
                    <el-button
                      link
                      :icon="CopyDocument"
                      :aria-label="t('pages.devices.list.actions.copy')"
                      @click="copyText(row.license_key)"
                    />
                  </div>
                </template>
              </el-table-column>
              <el-table-column
                prop="hardware_fingerprint"
                :label="t('pages.devices.list.table.fingerprint')"
                min-width="230"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <span class="device-mono__text">{{ row.hardware_fingerprint }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="customer_name"
                :label="t('pages.devices.list.table.customer')"
                width="150"
                show-overflow-tooltip
              >
                <template #default="{ row }">{{ row.customer_name || '-' }}</template>
              </el-table-column>
              <el-table-column
                prop="authorization_code"
                :label="t('pages.devices.list.table.authorizationCode')"
                min-width="250"
                show-overflow-tooltip
              >
                <template #default="{ row }">
                  <span class="device-mono__text">{{ row.authorization_code }}</span>
                </template>
              </el-table-column>
              <el-table-column :label="t('pages.devices.list.table.status')" width="110" align="center">
                <template #default="{ row }">
                  <el-tag :type="statusTagType(row.status)" effect="light">{{ statusText(row) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('pages.devices.list.table.online')" width="100" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.is_online ? 'success' : 'info'" effect="light">
                    {{ row.is_online ? t('pages.devices.list.online.online') : t('pages.devices.list.online.offline') }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column :label="t('pages.devices.list.table.activatedAt')" width="170">
                <template #default="{ row }">{{ formatDateTime(row.activated_at) }}</template>
              </el-table-column>
              <el-table-column :label="t('pages.devices.list.table.lastHeartbeat')" width="170">
                <template #default="{ row }">{{ formatDateTime(row.last_heartbeat) }}</template>
              </el-table-column>
              <el-table-column
                prop="last_online_ip"
                :label="t('pages.devices.list.table.lastOnlineIp')"
                width="150"
                show-overflow-tooltip
              >
                <template #default="{ row }">{{ row.last_online_ip || '-' }}</template>
              </el-table-column>
              <el-table-column
                :label="t('pages.devices.list.table.action')"
                width="120"
                fixed="right"
                align="center"
              >
                <template #default="{ row }">
                  <el-button
                    link
                    type="primary"
                    :disabled="!row.authorization_code_id"
                    @click="openDetail(row)"
                  >
                    {{ t('pages.devices.list.actions.viewDetail') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div class="pagination data-list-pagination">
            <el-pagination
              v-model:current-page="pagination.page"
              v-model:page-size="pagination.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="pagination.total"
              :disabled="loading"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="changePageSize"
              @current-change="loadDevices"
            />
          </div>
        </div>
      </section>
    </div>
  </Layout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { CopyDocument, Refresh } from '@element-plus/icons-vue'
import Layout from '@/components/common/layout/Layout.vue'
import {
  getLicenseDevices,
  type LicenseDevice,
  type LicenseListQueryRequest
} from '@/api/license'
import { getErrorMessage } from '@/utils/error'
import { useInitialSkeleton } from '@/composables/useInitialSkeleton'

const router = useRouter()
const { t, locale } = useI18n()

const filters = reactive({
  status: '' as '' | LicenseDevice['status'],
  isOnline: '' as '' | 'true' | 'false'
})
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
const devices = ref<LicenseDevice[]>([])
const loading = ref(false)
const { skeletonVisible, isUpdating, markInitialized } = useInitialSkeleton(loading)
const loadError = ref('')

const statusOptions = computed(() =>
  (['active', 'inactive', 'revoked'] as const).map(value => ({
    value,
    label: t(`pages.devices.list.status.${value}`)
  }))
)

const statusText = (row: LicenseDevice) =>
  row.status_display || t(`pages.devices.list.status.${row.status}`)

const statusTagType = (status: LicenseDevice['status']) => {
  if (status === 'active') return 'success'
  if (status === 'inactive') return 'warning'
  return 'danger'
}

const shortenKey = (key: string) =>
  key && key.length > 28 ? `${key.slice(0, 24)}…${key.slice(-4)}` : key

const formatDateTime = (value?: string) => {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(locale.value, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const copyText = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(t('pages.devices.list.message.copySuccess'))
  } catch {
    ElMessage.error(t('pages.devices.list.message.copyFailure'))
  }
}

const openDetail = (row: LicenseDevice) => {
  if (!row.authorization_code_id) return
  router.push({ name: 'licenses-view', params: { id: row.authorization_code_id } })
}

const loadDevices = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const params: LicenseListQueryRequest = {
      page: pagination.page,
      page_size: pagination.pageSize,
      sort: 'last_heartbeat',
      order: 'desc'
    }
    if (filters.status) params.status = filters.status
    if (filters.isOnline) params.is_online = filters.isOnline
    const response = await getLicenseDevices(params)
    devices.value = response.data.list
    pagination.total = response.data.total
  } catch (error) {
    loadError.value = getErrorMessage(error, t('pages.devices.list.message.loadError'))
  } finally {
    loading.value = false
    markInitialized()
  }
}

const queryDevices = async () => {
  pagination.page = 1
  await loadDevices()
}

const resetFilters = async () => {
  filters.status = ''
  filters.isOnline = ''
  await queryDevices()
}

const changePageSize = () => {
  pagination.page = 1
  loadDevices()
}

onMounted(loadDevices)
</script>

<style lang="scss" scoped>
.filter-grid {
  grid-template-columns: repeat(2, minmax(200px, 1fr));
}
.filter-control {
  width: 100%;
}
.device-mono {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
}
.device-mono__text {
  font-family: 'JetBrains Mono', 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
  color: var(--app-text-primary, #1f2d27);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
