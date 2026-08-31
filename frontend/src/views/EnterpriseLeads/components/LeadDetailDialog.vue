<template>
    <div class="dialog_box">
        <el-dialog v-model="visible" :title="t('enterpriseLeads.detail.title', { company: detailData?.company_name })"
            width="min(800px, calc(100vw - 32px))" class="lead-detail-dialog" destroy-on-close v-loading="loading">
            <div v-if="detailData" class="detail-content">
                <!-- 企业基本信息 -->
                <div class="detail-section">
                    <h3 class="section-title">{{ t('enterpriseLeads.detail.basicInfo') }}</h3>
                    <div class="info-grid">
                        <div class="info-item">
                            <span class="label">{{ t('enterpriseLeads.table.company') }}：</span>
                            <span class="value">{{ detailData.company_name }}</span>
                        </div>
                        <div class="info-item">
                            <span class="label">{{ t('enterpriseLeads.table.contact') }}：</span>
                            <span class="value">{{ detailData.contact_name }}</span>
                        </div>
                        <div class="info-item">
                            <span class="label">{{ t('enterpriseLeads.table.phone') }}：</span>
                            <span class="value">{{ detailData.contact_phone }}</span>
                        </div>
                    </div>
                </div>

                <!-- 需求信息 -->
                <div class="detail-section">
                    <h3 class="section-title">{{ t('enterpriseLeads.detail.requirementInfo') }}</h3>
                    <div class="info-list">
                        <div class="info-item">
                            <span class="label">{{ t('enterpriseLeads.detail.email') }}：</span>
                            <span class="value">{{ detailData.contact_email || '-' }}</span>
                        </div>
                        <div class="info-item block">
                            <span class="label">{{ t('enterpriseLeads.detail.description') }}：</span>
                            <span class="value">{{ detailData.requirement || '-' }}</span>
                        </div>
                        <div class="info-item block">
                            <span class="label">{{ t('enterpriseLeads.detail.otherInfo') }}：</span>
                            <span class="value">{{ detailData.extra_info || '-' }}</span>
                        </div>
                    </div>
                </div>

                <!-- 跟进信息 -->
                <div class="detail-section">
                    <h3 class="section-title">{{ t('enterpriseLeads.detail.followUpInfo') }}</h3>
                    <div class="info-grid">
                        <div class="info-item">
                            <span class="label">{{ t('enterpriseLeads.table.submittedAt') }}：</span>
                            <span class="value">{{ detailData.created_at }}</span>
                        </div>
                        <div class="info-item">
                            <span class="label">{{ t('enterpriseLeads.table.status') }}：</span>
                            <span class="status-value" :class="detailData.status">{{
                                t(`enterpriseLeads.status.${detailData.status}`) }}</span>
                        </div>
                        <div class="info-item">
                            <span class="label">{{ t('enterpriseLeads.detail.followUpDate') }}：</span>
                            <span class="value">{{ detailData.follow_up_date || '-' }}</span>
                        </div>
                    </div>
                    <div class="info-list mt-12">
                        <div class="info-item block">
                            <span class="label">{{ t('enterpriseLeads.detail.followUpRecord') }}：</span>
                            <span class="value">{{ detailData.follow_up_record || '-' }}</span>
                        </div>
                        <div class="info-item block">
                            <span class="label">{{ t('enterpriseLeads.detail.internalRemark') }}：</span>
                            <span class="value">{{ detailData.internal_note || '-' }}</span>
                        </div>
                    </div>
                </div>
            </div>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getLeadDetail, type Lead } from '@/api/lead'
import { ElMessage } from 'element-plus'
import { formatDateTime } from '@/utils/date'

const props = defineProps<{
    modelValue: boolean
    id: string | number | null
}>()

const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const visible = ref(props.modelValue)
const loading = ref(false)
const detailData = ref<Lead | null>(null)

const fetchDetail = async (id: string | number) => {
    loading.value = true
    try {
        const res = await getLeadDetail(id)
        if (res.code === '000000' && res.data) {
            const data = res.data
            detailData.value = {
                ...data,
                created_at: formatDateTime(data.created_at || ''),
                follow_up_date: data.follow_up_date ? formatDateTime(data.follow_up_date) : null
            }
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

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-section {
  overflow: hidden;
  border: 1px solid var(--app-border-color);
  border-radius: 0;
}

.section-title {
  margin: 0;
  padding: 11px 16px;
  background: var(--app-action-btn-bg);
  color: var(--app-text-primary);
  font-size: 14px;
  font-weight: 600;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 24px;
  padding: 16px;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
}

.info-item {
  display: flex;
  min-width: 0;
  font-size: 14px;
  line-height: 1.6;
}

.info-item .label {
  width: 100px;
  flex-shrink: 0;
  color: var(--app-text-secondary);
}

.info-item .value {
  min-width: 0;
  color: var(--app-text-primary);
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.status-value {
  font-weight: 600;
}

.status-value.contacting,
.status-value.completed {
  color: var(--el-color-primary);
}

.status-value.pending {
  color: var(--el-color-warning);
}

.status-value.rejected {
  color: var(--el-color-info);
}

.mt-12 {
  margin-top: 0;
  padding-top: 0;
}

@media (max-width: 640px) {
  :deep(.el-dialog__header),
  :deep(.el-dialog__body) {
    padding-right: 16px;
    padding-left: 16px;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .info-item {
    flex-direction: column;
    gap: 4px;
  }

  .info-item .label {
    width: auto;
  }
}
</style>
