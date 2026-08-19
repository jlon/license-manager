<template>
  <div class="detail-page">
    <header class="page-header">
      <div class="header-main">
        <button class="breadcrumb" type="button" @click="goBack">
          {{ t('pages.licenses.detail.breadcrumb.licenseManagement') }} /
          {{ t('pages.licenses.detail.breadcrumb.licenseDetail') }}
        </button>
        <div class="title-row">
          <h1>{{ licenseData?.code || t('pages.licenses.detail.breadcrumb.licenseDetail') }}</h1>
          <el-tag v-if="licenseData" :type="statusType(licenseData.status)">
            {{ licenseData.status_display || t(`pages.licenses.list.status.${licenseData.status}`) }}
          </el-tag>
        </div>
      </div>
      <div v-if="licenseData" class="header-actions">
        <el-button @click="copyCode">{{ t('pages.licenses.detail.actions.copyCode') }}</el-button>
        <el-button @click="openUpdateDialog">{{ t('pages.licenses.detail.actions.updateLicense') }}</el-button>
        <el-button @click="openValidityDialog">{{ t('pages.licenses.detail.actions.changeValidity') }}</el-button>
        <el-button type="primary" :loading="downloading" @click="downloadCertificate">
          {{ t('pages.licenses.detail.actions.downloadCertificate') }}
        </el-button>
      </div>
    </header>

    <section v-if="loadError" class="state-card">
      <el-empty :description="loadError">
        <el-button type="primary" @click="loadLicense">{{ t('pages.licenses.list.actions.retry') }}</el-button>
      </el-empty>
    </section>

    <section v-else class="detail-card" v-loading="loading">
      <el-tabs v-model="activeTab">
        <el-tab-pane :label="t('pages.licenses.detail.tabs.basic')" name="basic">
          <BasicInfo :license-data="licenseData" />
        </el-tab-pane>
        <el-tab-pane :label="t('pages.licenses.detail.tabs.authorization')" name="authorization">
          <AuthorizationInfo :license-data="licenseData" />
        </el-tab-pane>
        <el-tab-pane :label="t('pages.licenses.detail.tabs.license')" name="license">
          <LicenseInfo :license-data="licenseData" @license-revoked="loadLicense" />
        </el-tab-pane>
        <el-tab-pane :label="t('pages.licenses.detail.tabs.history')" name="history">
          <ChangeHistory :license-data="licenseData" />
        </el-tab-pane>
      </el-tabs>
    </section>

    <el-dialog
      v-model="validityDialogVisible"
      :title="t('pages.licenses.detail.dialogs.changeValidityTitle')"
      width="520px"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item :label="t('pages.licenses.detail.dialogs.validityType')">
          <el-radio-group v-model="validityForm.type">
            <el-radio value="limited">{{ t('pages.licenses.detail.dialogs.limited') }}</el-radio>
            <el-radio value="permanent">{{ t('pages.licenses.detail.dialogs.permanent') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="validityForm.type === 'limited'" :label="t('pages.licenses.detail.dialogs.dateRange')">
          <el-date-picker
            v-model="validityForm.range"
            type="daterange"
            value-format="YYYY-MM-DD"
            format="YYYY-MM-DD"
            :disabled-date="disablePastDate"
          />
        </el-form-item>
        <el-form-item :label="t('pages.licenses.detail.dialogs.reason')">
          <el-input
            v-model="validityForm.reason"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit
            :placeholder="t('pages.licenses.detail.dialogs.reasonPlaceholder')"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="validityDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="validitySubmitting" @click="submitValidityChange">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="updateDialogVisible"
      :title="t('pages.licenses.detail.updateDialog.title')"
      width="720px"
      destroy-on-close
    >
      <el-form label-position="top">
        <div class="dialog-grid">
          <el-form-item :label="t('pages.licenses.detail.updateDialog.fields.changeType')">
            <el-select v-model="updateForm.changeType">
              <el-option v-for="option in changeTypes" :key="option.key" :label="option.display" :value="option.key" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('pages.licenses.detail.updateDialog.fields.maxActivations')">
            <el-input-number v-model="updateForm.maxActivations" :min="1" :max="999999" controls-position="right" />
          </el-form-item>
        </div>
        <el-form-item :label="t('pages.licenses.detail.updateDialog.fields.featureConfig')">
          <JsonEditor ref="featureEditor" v-model="featureConfig" />
        </el-form-item>
        <el-form-item :label="t('pages.licenses.detail.updateDialog.fields.usageLimits')">
          <JsonEditor ref="limitEditor" v-model="usageLimits" />
        </el-form-item>
        <el-form-item :label="t('pages.licenses.detail.updateDialog.fields.customParameters')">
          <JsonEditor ref="parameterEditor" v-model="customParameters" />
        </el-form-item>
        <el-form-item :label="t('pages.licenses.detail.updateDialog.fields.reason')">
          <el-input v-model="updateForm.reason" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="updateDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="updateSubmitting" @click="submitUpdate">
          {{ t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { getEnumOptions, type RawEnumItem } from '@/api/enum'
import {
  downloadAuthorizationFile,
  getLicenseDetail,
  updateLicense,
  type AuthorizationCode,
  type LicenseUpdateRequest
} from '@/api/license'
import JsonEditor from '@/components/common/JsonEditor.vue'
import { getErrorMessage } from '@/utils/error'
import AuthorizationInfo from './components/AuthorizationInfo.vue'
import BasicInfo from './components/BasicInfo.vue'
import ChangeHistory from './components/ChangeHistory.vue'
import LicenseInfo from './components/LicenseInfo.vue'

type JsonData = Record<string, unknown>

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const activeTab = ref('basic')
const loading = ref(false)
const downloading = ref(false)
const loadError = ref('')
const licenseData = ref<AuthorizationCode | null>(null)

const validityDialogVisible = ref(false)
const validitySubmitting = ref(false)
const validityForm = reactive<{
  type: 'limited' | 'permanent'
  range: [string, string] | null
  reason: string
}>({ type: 'limited', range: null, reason: '' })

const updateDialogVisible = ref(false)
const updateSubmitting = ref(false)
const updateForm = reactive({ changeType: 'other', maxActivations: 1, reason: '' })
const changeTypes = ref<RawEnumItem[]>([])
const featureConfig = ref<JsonData | null>({})
const usageLimits = ref<JsonData | null>({})
const customParameters = ref<JsonData | null>({})
const featureEditor = ref<InstanceType<typeof JsonEditor>>()
const limitEditor = ref<InstanceType<typeof JsonEditor>>()
const parameterEditor = ref<InstanceType<typeof JsonEditor>>()

const statusType = (status: AuthorizationCode['status']) => {
  if (status === 'normal') return 'success'
  if (status === 'locked') return 'warning'
  return 'danger'
}

const goBack = () => {
  router.push({
    name: 'licenses-list',
    query: route.query.customerId
      ? { customerId: route.query.customerId, customerName: route.query.customerName }
      : undefined
  })
}

const copyCode = async () => {
  if (!licenseData.value?.code) {
    ElMessage.warning(t('pages.licenses.detail.messages.codeEmpty'))
    return
  }
  try {
    await navigator.clipboard.writeText(licenseData.value.code)
    ElMessage.success(t('pages.licenses.detail.messages.copySuccess'))
  } catch {
    ElMessage.error(t('pages.licenses.detail.messages.copyError'))
  }
}

const disablePastDate = (date: Date) => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return date.getTime() < today.getTime()
}

const toDateOnly = (value?: string) => value?.slice(0, 10) || ''

const openValidityDialog = () => {
  const start = toDateOnly(licenseData.value?.start_date)
  const end = toDateOnly(licenseData.value?.end_date)
  if (start && end) {
    const days = Math.ceil((new Date(`${end}T00:00:00`).getTime() - new Date(`${start}T00:00:00`).getTime()) / 86400000)
    validityForm.type = days >= 365000 ? 'permanent' : 'limited'
    validityForm.range = validityForm.type === 'limited' ? [start, end] : null
  } else {
    validityForm.type = 'limited'
    validityForm.range = null
  }
  validityForm.reason = ''
  validityDialogVisible.value = true
}

const permanentRange = (): [string, string] => {
  const start = new Date()
  const end = new Date(start)
  end.setDate(end.getDate() + 365000 - 1)
  const format = (date: Date) => {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }
  return [format(start), format(end)]
}

const submitValidityChange = async () => {
  if (!licenseData.value?.id) return
  if (!validityForm.reason.trim()) {
    ElMessage.warning(t('pages.licenses.detail.messages.changeValidityReasonRequired'))
    return
  }
  const range = validityForm.type === 'permanent' ? permanentRange() : validityForm.range
  if (!range?.[0] || !range?.[1]) {
    ElMessage.warning(t('pages.licenses.form.validation.dateRangeRequired'))
    return
  }

  validitySubmitting.value = true
  try {
    await updateLicense(licenseData.value.id, {
      start_date: range[0],
      end_date: range[1],
      change_type: 'renewal',
      reason: validityForm.reason.trim()
    })
    ElMessage.success(t('pages.licenses.detail.messages.changeValiditySuccess'))
    validityDialogVisible.value = false
    await loadLicense()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.detail.messages.changeValidityError')))
  } finally {
    validitySubmitting.value = false
  }
}

const parseJsonData = (value: unknown): JsonData => {
  if (!value) return {}
  if (typeof value === 'object') return value as JsonData
  if (typeof value !== 'string' || !value.trim()) return {}
  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}

const loadChangeTypes = async () => {
  if (changeTypes.value.length) return
  try {
    const response = await getEnumOptions('authorization_change_type')
    changeTypes.value = response.data.items
  } catch {
    changeTypes.value = [{ key: 'other', display: 'other' }]
  }
}

const openUpdateDialog = async () => {
  if (!licenseData.value) return
  updateForm.changeType = 'other'
  updateForm.maxActivations = licenseData.value.max_activations || 1
  updateForm.reason = ''
  featureConfig.value = parseJsonData(licenseData.value.feature_config)
  usageLimits.value = parseJsonData(licenseData.value.usage_limits)
  customParameters.value = parseJsonData(licenseData.value.custom_parameters)
  await loadChangeTypes()
  if (!changeTypes.value.some(item => item.key === updateForm.changeType)) {
    updateForm.changeType = changeTypes.value[0]?.key || 'other'
  }
  updateDialogVisible.value = true
}

const editorsAreValid = () => [featureEditor, limitEditor, parameterEditor]
  .every(editor => editor.value?.validate() ?? true)

const submitUpdate = async () => {
  if (!licenseData.value?.id) return
  if (!updateForm.changeType) {
    ElMessage.warning(t('pages.licenses.detail.updateDialog.validation.changeTypeRequired'))
    return
  }
  if (!updateForm.reason.trim()) {
    ElMessage.warning(t('pages.licenses.detail.updateDialog.validation.reasonRequired'))
    return
  }
  if (!editorsAreValid()) {
    ElMessage.error(t('pages.licenses.form.keyValue.validationFailed'))
    return
  }

  const payload: LicenseUpdateRequest = {
    max_activations: updateForm.maxActivations,
    feature_config: featureConfig.value || {},
    usage_limits: JSON.stringify(usageLimits.value || {}),
    custom_parameters: JSON.stringify(customParameters.value || {}),
    change_type: updateForm.changeType,
    reason: updateForm.reason.trim()
  }
  updateSubmitting.value = true
  try {
    await updateLicense(licenseData.value.id, payload)
    ElMessage.success(t('pages.licenses.detail.updateDialog.messages.updateSuccess'))
    updateDialogVisible.value = false
    await loadLicense()
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.detail.updateDialog.messages.updateError')))
  } finally {
    updateSubmitting.value = false
  }
}

const fileNameFromDisposition = (disposition?: string) => {
  if (!disposition) return 'authorization_package.zip'
  const utf8Name = disposition.match(/filename\*=UTF-8''([^;]+)/i)?.[1]
  if (utf8Name) {
    try { return decodeURIComponent(utf8Name) } catch { return utf8Name }
  }
  return disposition.match(/filename="?([^";]+)"?/i)?.[1] || 'authorization_package.zip'
}

const downloadErrorText = async (error: unknown) => {
  const data = (error as { response?: { data?: unknown } })?.response?.data
  if (data instanceof Blob) {
    try {
      const body = JSON.parse(await data.text()) as { message?: unknown }
      if (typeof body.message === 'string' && body.message.trim()) return body.message
    } catch {
      // Use the standard request error below.
    }
  }
  return getErrorMessage(error, t('pages.licenses.detail.messages.downloadError'))
}

const downloadCertificate = async () => {
  if (!licenseData.value?.id) return
  downloading.value = true
  try {
    const response = await downloadAuthorizationFile(licenseData.value.id)
    const url = URL.createObjectURL(new Blob([response.data], { type: 'application/zip' }))
    const link = document.createElement('a')
    link.href = url
    link.download = fileNameFromDisposition(response.headers['content-disposition'])
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success(t('pages.licenses.detail.messages.downloadSuccess'))
  } catch (error) {
    ElMessage.error(await downloadErrorText(error))
  } finally {
    downloading.value = false
  }
}

const loadLicense = async () => {
  const id = String(route.params.id || '')
  if (!id) {
    loadError.value = t('pages.licenses.detail.messages.missingId')
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    const response = await getLicenseDetail(id)
    if (!response.data) throw new Error(response.message)
    licenseData.value = response.data
  } catch (error) {
    loadError.value = getErrorMessage(error, t('pages.licenses.detail.messages.loadError'))
  } finally {
    loading.value = false
  }
}

onMounted(loadLicense)
</script>

<style scoped lang="scss">
.detail-page {
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

.breadcrumb {
  padding: 0;
  border: 0;
  background: none;
  color: var(--app-text-secondary);
  cursor: pointer;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 6px;

  h1 {
    max-width: 720px;
    margin: 0;
    overflow: hidden;
    color: var(--app-text-primary);
    font-size: 24px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.header-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.detail-card,
.state-card {
  min-height: 360px;
  padding: 20px;
  overflow: hidden;
  border: 1px solid var(--app-border-color);
  border-radius: var(--app-card-radius);
  background: var(--app-content-bg);
  box-shadow: var(--app-card-shadow);
}

.detail-card :deep(.el-tabs__content) {
  overflow: visible;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.dialog-grid :deep(.el-select),
.dialog-grid :deep(.el-input-number),
.detail-page :deep(.el-date-editor) {
  width: 100%;
}

@media (max-width: 900px) {
  .detail-page {
    padding: 12px;
  }

  .page-header {
    align-items: stretch;
    flex-direction: column;
  }

  .header-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 640px) {
  .dialog-grid {
    grid-template-columns: 1fr;
  }
}
</style>
