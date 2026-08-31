<template>
  <header class="subpage-header">
    <div class="subpage-header__main">
      <button type="button" class="subpage-header__back" :disabled="backDisabled" @click="emit('back')">
        <el-icon><ArrowLeft /></el-icon>
        <span>{{ backLabel }}</span>
      </button>
      <div class="subpage-header__title-row">
        <h1>{{ title }}</h1>
        <slot name="title-meta" />
      </div>
      <p v-if="description" class="subpage-header__description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="subpage-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup lang="ts">
import { ArrowLeft } from '@element-plus/icons-vue'

defineProps<{
  title: string
  backLabel: string
  description?: string
  backDisabled?: boolean
}>()

const emit = defineEmits<{ back: [] }>()
</script>

<style scoped>
.subpage-header { min-height:64px; display:flex; align-items:flex-end; justify-content:space-between; gap:24px; }
.subpage-header__main { min-width:0; }
.subpage-header__back { margin:0 0 7px; padding:0; display:inline-flex; align-items:center; gap:5px; border:0; background:transparent; color:#526a60; font:inherit; font-size:13px; cursor:pointer; }
.subpage-header__back:hover { color:var(--el-color-primary); }
.subpage-header__back:disabled { cursor:not-allowed; opacity:.45; }
.subpage-header__title-row { min-width:0; display:flex; align-items:center; gap:10px; }
.subpage-header__title-row h1 { margin:0; overflow:hidden; color:var(--app-text-primary); font-size:24px; font-weight:650; line-height:1.35; text-overflow:ellipsis; white-space:nowrap; }
.subpage-header__description { margin:5px 0 0; color:#667b72; font-size:13px; }
.subpage-header__actions { display:flex; align-items:center; justify-content:flex-end; gap:8px; flex-shrink:0; }
@media (max-width:768px) {
  .subpage-header { min-height:auto; align-items:stretch; flex-direction:column; gap:12px; }
  .subpage-header__actions { justify-content:flex-start; overflow-x:auto; }
}
</style>
