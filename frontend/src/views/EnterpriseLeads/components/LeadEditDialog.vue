<template>
 <div>
     <el-dialog
    v-model="visible"
     :title="t('enterpriseLeads.edit.title', { company: detailData?.company_name })"
     width="min(800px, calc(100vw - 32px))"
     class="lead-edit-dialog"
     destroy-on-close
     :close-on-click-modal="false"
     v-loading="loading"
   >
     <el-form
       ref="formRef"
       :model="form"
       :rules="rules"
       label-position="top"
       class="edit-form"
     >
      <div class="form-row">
        <el-form-item :label="t('enterpriseLeads.table.company')" prop="company_name" class="flex-1">
          <el-input v-model="form.company_name" :placeholder="t('enterpriseLeads.table.company')" />
        </el-form-item>
        <el-form-item :label="t('enterpriseLeads.table.contact')" prop="contact_name" class="flex-1">
          <el-input v-model="form.contact_name" :placeholder="t('enterpriseLeads.table.contact')" />
        </el-form-item>
      </div>

      <div class="form-row">
        <el-form-item :label="t('enterpriseLeads.table.phone')" prop="contact_phone" class="flex-1">
          <el-input v-model="form.contact_phone" :placeholder="t('enterpriseLeads.table.phone')" />
        </el-form-item>
        <el-form-item :label="t('enterpriseLeads.detail.email')" prop="contact_email" class="flex-1">
          <el-input v-model="form.contact_email" :placeholder="t('enterpriseLeads.detail.email')" />
        </el-form-item>
      </div>

      <el-form-item :label="t('enterpriseLeads.detail.description')" prop="requirement">
        <el-input
          v-model="form.requirement"
          type="textarea"
          :rows="3"
          :placeholder="t('enterpriseLeads.detail.description')"
        />
      </el-form-item>

      <el-form-item :label="t('enterpriseLeads.detail.otherInfo')" prop="extra_info">
        <el-input
          v-model="form.extra_info"
          type="textarea"
          :rows="2"
          :placeholder="t('enterpriseLeads.detail.otherInfo')"
        />
      </el-form-item>

      <div class="form-row">
        <el-form-item :label="t('enterpriseLeads.table.status')" prop="status" class="flex-1">
          <el-select v-model="form.status" class="w-full" :placeholder="t('enterpriseLeads.filter.statusPlaceholder')">
            <el-option :label="t('enterpriseLeads.status.pending')" value="pending" />
            <el-option :label="t('enterpriseLeads.status.contacting')" value="contacting" />
            <el-option :label="t('enterpriseLeads.status.completed')" value="completed" />
            <el-option :label="t('enterpriseLeads.status.rejected')" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('enterpriseLeads.detail.followUpDate')" prop="follow_up_date" class="flex-1">
          <el-date-picker
            v-model="form.follow_up_date"
            type="date"
            class="w-full"
            :placeholder="t('enterpriseLeads.edit.datePlaceholder')"
            format="YYYY/MM/DD"
            value-format="YYYY-MM-DD"
          />
        </el-form-item>
      </div>
 
       <el-form-item :label="t('enterpriseLeads.detail.followUpRecord')" prop="follow_up_record">
         <el-input
           v-model="form.follow_up_record"
           type="textarea"
           :rows="4"
           maxlength="500"
           show-word-limit
           :placeholder="t('enterpriseLeads.edit.recordPlaceholder')"
         />
       </el-form-item>
 
       <el-form-item :label="t('enterpriseLeads.detail.internalRemark')" prop="internal_note">
         <el-input
           v-model="form.internal_note"
           type="textarea"
           :rows="4"
           maxlength="500"
           show-word-limit
           :placeholder="t('enterpriseLeads.edit.remarkPlaceholder')"
         />
       </el-form-item>
     </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" class="btn-save" @click="handleSave">
          {{ t('enterpriseLeads.edit.save') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
 </div>
</template>

<script setup lang="ts">
import { ref, watch, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import type { FormInstance, FormRules } from 'element-plus'
import { getLeadDetail, type Lead } from '@/api/lead'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  modelValue: boolean
  id: string | number | null
}>()

const emit = defineEmits(['update:modelValue', 'save'])

const { t } = useI18n()
const visible = ref(props.modelValue)
const formRef = ref<FormInstance>()
const loading = ref(false)
const detailData = ref<Lead | null>(null)

const form = reactive({
  company_name: '',
  contact_email: '',
  contact_name: '',
  contact_phone: '',
  extra_info: '',
  requirement: '',
  status: '',
  follow_up_date: '',
  follow_up_record: '',
  internal_note: ''
})

const rules = reactive<FormRules>({
  status: [{ required: true, message: t('enterpriseLeads.edit.statusRequired'), trigger: 'change' }],
  company_name: [{ required: true, message: t('enterpriseLeads.validation.nameRequired'), trigger: 'blur' }],
  contact_name: [{ required: true, message: t('enterpriseLeads.validation.contactRequired'), trigger: 'blur' }]
})

const fetchDetail = async (id: string | number) => {
  loading.value = true
  try {
    const res = await getLeadDetail(id)
    if (res.code === '000000' && res.data) {
      const data = res.data
      detailData.value = data
      form.company_name = data.company_name || ''
      form.contact_email = data.contact_email || ''
      form.contact_name = data.contact_name || ''
      form.contact_phone = data.contact_phone || ''
      form.extra_info = data.extra_info || ''
      form.requirement = data.requirement || ''
      form.status = data.status
      form.follow_up_date = data.follow_up_date || ''
      form.follow_up_record = data.follow_up_record || ''
      form.internal_note = data.internal_note || ''
    }
  } catch (error: any) {
    console.error('Fetch lead detail error:', error)
    ElMessage.error(error.backendMessage || t('enterpriseLeads.messages.fetchDetailError'))
  } finally {
    loading.value = false
  }
}

watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val && props.id) {
    fetchDetail(props.id)
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate((valid) => {
    if (valid) {
      emit('save', { id: props.id, ...form })
      visible.value = false
    }
  })
}
</script>

<style scoped>
:deep(.el-dialog) {
  overflow: hidden;
  padding: 0;
  background: var(--app-content-bg);
  border: 1px solid var(--app-border-color);
  border-radius: var(--app-card-radius);
}

:deep(.el-dialog__header) {
  display: flex;
  align-items: center;
  margin: 0;
  padding: 18px 24px;
  background: linear-gradient(120deg, var(--el-color-primary-dark-2), var(--el-color-primary));
}

:deep(.el-dialog__title),
:deep(.el-dialog__close) {
  color: var(--el-color-white);
}

:deep(.el-dialog__title) {
  font-size: 18px;
  font-weight: 600;
}

:deep(.el-dialog__headerbtn) {
  top: 14px;
  right: 14px;
}

:deep(.el-dialog__body) {
  max-height: min(70vh, 680px);
  padding: 24px;
  overflow-y: auto;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px;
  background: var(--app-action-btn-bg);
  border-top: 1px solid var(--app-border-color);
}

.edit-form .form-row {
  display: flex;
  gap: 20px;
}

.edit-form .flex-1 {
  min-width: 0;
  flex: 1;
}

.edit-form :deep(.el-form-item__label) {
  color: var(--app-text-regular);
  font-weight: 500;
}

.w-full {
  width: 100%;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 640px) {
  :deep(.el-dialog__header),
  :deep(.el-dialog__body),
  :deep(.el-dialog__footer) {
    padding-right: 16px;
    padding-left: 16px;
  }

  .edit-form .form-row {
    flex-direction: column;
    gap: 0;
  }

  .dialog-footer .el-button {
    flex: 1;
  }
}
</style>
