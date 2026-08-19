<template>
  <div class="form-page">
    <header class="page-header">
      <div>
        <button class="breadcrumb" type="button" @click="goBack">
          {{ t('pages.licenses.form.breadcrumb.licenseManagement') }} /
          {{ t('pages.licenses.form.breadcrumb.createLicense') }}
        </button>
        <h1>{{ t('pages.licenses.form.breadcrumb.createLicense') }}</h1>
        <p>{{ t('pages.licenses.form.description') }}</p>
      </div>
      <div class="header-actions">
        <el-button @click="cancelCreate">{{ t('pages.licenses.form.actions.cancel') }}</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">
          {{ t('pages.licenses.form.actions.create') }}
        </el-button>
      </div>
    </header>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="form-content">
      <section class="form-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.basicInfo') }}</h2>
        </div>
        <div class="field-grid two-columns">
          <el-form-item :label="t('pages.licenses.form.fields.customerName')" prop="customer_id">
            <el-select
              v-model="form.customer_id"
              :placeholder="t('pages.licenses.form.placeholders.selectCustomer')"
              :loading="customerLoading"
              filterable
              remote
              :remote-method="searchCustomers"
              @change="selectCustomer"
            >
              <el-option v-for="customer in customers" :key="customer.id" :label="customer.name" :value="customer.id" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.customerId')">
            <el-input v-model="form.customer_code" disabled />
          </el-form-item>
        </div>
        <el-form-item :label="t('pages.licenses.form.fields.description')" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit
            :placeholder="t('pages.licenses.form.placeholders.enterDescription')"
          />
        </el-form-item>
      </section>

      <section class="form-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.licenseConfig') }}</h2>
        </div>
        <div class="field-grid three-columns">
          <el-form-item :label="t('pages.licenses.form.fields.validityPeriod')" prop="validity_type">
            <el-select v-model="form.validity_type" @change="changeValidityType">
              <el-option :label="t('pages.licenses.form.validityTypes.limited')" value="limited" />
              <el-option :label="t('pages.licenses.form.validityTypes.permanent')" value="permanent" />
            </el-select>
          </el-form-item>
          <el-form-item
            v-if="form.validity_type === 'limited'"
            :label="t('pages.licenses.form.fields.dateRange')"
            prop="date_range"
            class="date-field"
          >
            <el-date-picker
              v-model="form.date_range"
              type="daterange"
              value-format="YYYY-MM-DD"
              format="YYYY-MM-DD"
              :start-placeholder="t('pages.licenses.form.placeholders.startDate')"
              :end-placeholder="t('pages.licenses.form.placeholders.endDate')"
              :disabled-date="disablePastDate"
              @change="calculateValidityDays"
            />
          </el-form-item>
          <el-form-item v-else :label="t('pages.licenses.form.fields.validityDays')" class="date-field">
            <el-input :model-value="t('pages.licenses.form.validityTypes.permanent')" disabled />
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.maxActivations')" prop="max_activations">
            <el-input-number v-model="form.max_activations" :min="1" :max="999999" controls-position="right" />
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.deploymentType')" prop="deployment_type">
            <el-select v-model="form.deployment_type" :loading="enumLoading">
              <el-option v-for="option in deploymentTypes" :key="option.key" :label="option.display" :value="option.key" />
            </el-select>
          </el-form-item>
          <el-form-item :label="t('pages.licenses.form.fields.encryptionType')" prop="encryption_type">
            <el-select v-model="form.encryption_type" :loading="enumLoading">
              <el-option v-for="option in encryptionTypes" :key="option.key" :label="option.display" :value="option.key" />
            </el-select>
          </el-form-item>
        </div>
      </section>

      <section class="form-card config-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.featureConfig') }}</h2>
        </div>
        <JsonEditor ref="featureEditor" v-model="featureConfig" />
      </section>

      <section class="form-card config-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.usageLimits') }}</h2>
        </div>
        <JsonEditor ref="limitEditor" v-model="usageLimits" />
      </section>

      <section class="form-card config-card">
        <div class="section-heading">
          <h2>{{ t('pages.licenses.form.sections.customParameters') }}</h2>
        </div>
        <JsonEditor ref="parameterEditor" v-model="customParameters" />
      </section>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { getCustomers } from '@/api/customer'
import { getEnumOptions, type RawEnumItem } from '@/api/enum'
import { createLicense, type AuthorizationCodeCreateRequest } from '@/api/license'
import JsonEditor from '@/components/common/JsonEditor.vue'
import { getErrorMessage } from '@/utils/error'

interface CustomerOption {
  id: string
  name: string
  code: string
}

interface LicenseFormModel {
  customer_id: string
  customer_code: string
  description: string
  validity_type: 'limited' | 'permanent'
  validity_days: number
  date_range: [string, string] | null
  max_activations: number
  deployment_type: string
  encryption_type: string
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const formRef = ref<FormInstance>()
const featureEditor = ref<InstanceType<typeof JsonEditor>>()
const limitEditor = ref<InstanceType<typeof JsonEditor>>()
const parameterEditor = ref<InstanceType<typeof JsonEditor>>()
const submitting = ref(false)
const customerLoading = ref(false)
const enumLoading = ref(false)
const customers = ref<CustomerOption[]>([])
const deploymentTypes = ref<RawEnumItem[]>([])
const encryptionTypes = ref<RawEnumItem[]>([])
const featureConfig = ref<Record<string, unknown> | null>({})
const usageLimits = ref<Record<string, unknown> | null>({})
const customParameters = ref<Record<string, unknown> | null>({})

const form = reactive<LicenseFormModel>({
  customer_id: String(route.query.customerId || ''),
  customer_code: String(route.query.customerId || ''),
  description: '',
  validity_type: 'limited',
  validity_days: 365,
  date_range: null,
  max_activations: 1,
  deployment_type: 'standalone',
  encryption_type: 'standard'
})

const validateDateRange = (_rule: unknown, value: [string, string] | null, callback: (error?: Error) => void) => {
  if (form.validity_type === 'permanent') return callback()
  if (!value?.[0] || !value?.[1]) return callback(new Error(t('pages.licenses.form.validation.dateRangeRequired')))

  const start = new Date(`${value[0]}T00:00:00`)
  const end = new Date(`${value[1]}T00:00:00`)
  if (end <= start) return callback(new Error(t('pages.licenses.form.validation.endDateAfterStart')))
  if ((end.getTime() - start.getTime()) / 86400000 > 3650) {
    return callback(new Error(t('pages.licenses.form.validation.validityPeriodTooLong')))
  }
  callback()
}

const rules: FormRules<LicenseFormModel> = {
  customer_id: [{ required: true, message: t('pages.licenses.form.validation.customerRequired'), trigger: 'change' }],
  description: [
    { required: true, message: t('pages.licenses.form.validation.descriptionRequired'), trigger: 'blur' },
    { min: 1, max: 500, message: t('pages.licenses.form.validation.descriptionLength'), trigger: 'blur' }
  ],
  validity_type: [{ required: true, message: t('pages.licenses.form.validation.validityTypeRequired'), trigger: 'change' }],
  date_range: [{ validator: validateDateRange, trigger: 'change' }],
  max_activations: [
    { required: true, message: t('pages.licenses.form.validation.maxActivationsRequired'), trigger: 'change' },
    { type: 'number', min: 1, max: 999999, message: t('pages.licenses.form.validation.maxActivationsRange'), trigger: 'change' }
  ],
  deployment_type: [{ required: true, message: t('pages.licenses.form.validation.deploymentTypeRequired'), trigger: 'change' }],
  encryption_type: [{ required: true, message: t('pages.licenses.form.validation.encryptionTypeRequired'), trigger: 'change' }]
}

const disablePastDate = (date: Date) => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return date.getTime() < today.getTime()
}

const calculateValidityDays = (range: [string, string] | null) => {
  if (!range?.[0] || !range?.[1]) return
  const start = new Date(`${range[0]}T00:00:00`)
  const end = new Date(`${range[1]}T00:00:00`)
  form.validity_days = Math.ceil((end.getTime() - start.getTime()) / 86400000)
}

const changeValidityType = () => {
  if (form.validity_type === 'permanent') {
    form.validity_days = 365000
    form.date_range = null
    formRef.value?.clearValidate('date_range')
  } else {
    form.validity_days = 365
  }
}

const selectCustomer = (id: string) => {
  form.customer_code = customers.value.find(customer => customer.id === id)?.code || id
}

const ensureRouteCustomer = () => {
  const id = String(route.query.customerId || '')
  const name = String(route.query.customerName || '')
  if (id && name && !customers.value.some(customer => customer.id === id)) {
    customers.value.unshift({ id, name, code: id })
  }
}

const searchCustomers = async (query = '') => {
  customerLoading.value = true
  try {
    const response = await getCustomers({
      page: 1,
      page_size: 100,
      customer_name: query.trim() || undefined,
      status: 'active',
      sort: 'customer_name',
      order: 'asc'
    })
    customers.value = response.data.list.map(customer => ({
      id: customer.id,
      name: customer.customer_name,
      code: customer.customer_code || customer.id
    }))
    ensureRouteCustomer()
    if (form.customer_id) selectCustomer(form.customer_id)
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.form.messages.searchCustomerErrorRetry')))
  } finally {
    customerLoading.value = false
  }
}

const loadEnums = async () => {
  enumLoading.value = true
  try {
    const [deploymentResponse, encryptionResponse] = await Promise.all([
      getEnumOptions('deployment_type'),
      getEnumOptions('encryption_type')
    ])
    deploymentTypes.value = deploymentResponse.data.items
    encryptionTypes.value = encryptionResponse.data.items
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.form.messages.loadEnumErrorRetry')))
  } finally {
    enumLoading.value = false
  }
}

const editorsAreValid = () => [featureEditor, limitEditor, parameterEditor]
  .every(editor => editor.value?.validate() ?? true)

const submitForm = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  if (!editorsAreValid()) {
    ElMessage.error(t('pages.licenses.form.keyValue.validationFailed'))
    return
  }

  const payload: AuthorizationCodeCreateRequest = {
    customer_id: form.customer_id,
    description: form.description.trim(),
    validity_days: form.validity_days,
    deployment_type: form.deployment_type,
    encryption_type: form.encryption_type,
    max_activations: form.max_activations,
    feature_config: featureConfig.value || {},
    usage_limits: JSON.stringify(usageLimits.value || {}),
    custom_parameters: JSON.stringify(customParameters.value || {})
  }

  submitting.value = true
  try {
    const response = await createLicense(payload)
    ElMessage.success(response.message || t('pages.licenses.form.messages.createSuccess'))
    await router.push({
      name: 'licenses-list',
      query: { customerId: form.customer_id, customerName: customers.value.find(item => item.id === form.customer_id)?.name || '' }
    })
  } catch (error) {
    ElMessage.error(getErrorMessage(error, t('pages.licenses.form.messages.submitErrorRetry')))
  } finally {
    submitting.value = false
  }
}

const goBack = () => router.back()

const cancelCreate = async () => {
  try {
    await ElMessageBox.confirm(
      t('pages.licenses.form.messages.cancelConfirm'),
      t('pages.licenses.form.messages.cancelTitle'),
      {
        confirmButtonText: t('pages.licenses.form.messages.cancelConfirmButton'),
        cancelButtonText: t('pages.licenses.form.messages.cancelCancelButton'),
        type: 'warning'
      }
    )
    goBack()
  } catch {
    // Continue editing.
  }
}

onMounted(async () => {
  ensureRouteCustomer()
  await Promise.all([searchCustomers(), loadEnums()])
})
</script>

<style scoped lang="scss">
.form-page {
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
  gap: 20px;

  h1 {
    margin: 6px 0 0;
    color: var(--app-text-primary);
    font-size: 24px;
  }

  p {
    margin: 6px 0 0;
    color: var(--app-text-secondary);
    font-size: 14px;
  }
}

.breadcrumb {
  padding: 0;
  border: 0;
  background: none;
  color: var(--app-text-secondary);
  cursor: pointer;
}

.header-actions {
  display: flex;
  flex-shrink: 0;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-card {
  padding: 20px;
  border: 1px solid var(--app-border-color);
  border-radius: var(--app-card-radius);
  background: var(--app-content-bg);
  box-shadow: var(--app-card-shadow);
}

.section-heading {
  margin-bottom: 18px;

  h2 {
    margin: 0;
    color: var(--app-text-primary);
    font-size: 17px;
  }
}

.field-grid {
  display: grid;
  gap: 0 20px;
}

.two-columns {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.three-columns {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.form-card :deep(.el-select),
.form-card :deep(.el-input-number),
.form-card :deep(.el-date-editor) {
  width: 100%;
}

.date-field {
  grid-column: span 2;
}

.config-card {
  overflow: hidden;
}

@media (max-width: 900px) {
  .form-page {
    padding: 12px;
  }

  .three-columns {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .page-header {
    align-items: stretch;
    flex-direction: column;
  }

  .header-actions :deep(.el-button) {
    flex: 1;
  }

  .two-columns,
  .three-columns {
    grid-template-columns: 1fr;
  }

  .date-field {
    grid-column: auto;
  }
}
</style>
