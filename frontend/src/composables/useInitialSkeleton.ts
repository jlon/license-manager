import { computed, onScopeDispose, ref, watch, type Ref } from 'vue'

export const useInitialSkeleton = (loading: Ref<boolean>, delay = 120) => {
  const initialized = ref(false)
  const skeletonVisible = ref(false)
  let timer: ReturnType<typeof setTimeout> | undefined

  const clearTimer = () => {
    if (timer) clearTimeout(timer)
    timer = undefined
  }

  watch(
    loading,
    isLoading => {
      clearTimer()
      if (isLoading && !initialized.value) {
        timer = setTimeout(() => {
          skeletonVisible.value = true
        }, delay)
      } else {
        skeletonVisible.value = false
      }
    },
    { immediate: true }
  )

  const markInitialized = () => {
    initialized.value = true
    skeletonVisible.value = false
    clearTimer()
  }

  onScopeDispose(clearTimer)

  return {
    initialized,
    skeletonVisible,
    isUpdating: computed(() => loading.value && initialized.value),
    markInitialized
  }
}
