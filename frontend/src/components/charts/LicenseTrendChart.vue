<template>
  <div class="chart-container">
    <VChart v-if="data.length" class="trend-chart" :option="chartOption" autoresize />
    <el-empty v-else :description="emptyText" :image-size="64" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { useAppStore } from '@/store/modules/app'
import type { TrendDataItem } from '@/api/dashboard'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent])

const props = defineProps<{ data: TrendDataItem[]; emptyText: string }>()
const { t } = useI18n()
const appStore = useAppStore()

const chartOption = computed(() => {
  const values = props.data.map(item => item.total_authorizations)
  return {
    animationDuration: 350,
    grid: { left: '3%', right: '3%', top: 24, bottom: 12, containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.data.map(item => item.date.substring(5)),
      axisLine: { lineStyle: { color: appStore.isDark ? '#606266' : '#dcdfe6' } },
      axisTick: { show: false },
      axisLabel: { color: appStore.isDark ? '#cfd3dc' : '#606266', fontSize: 12, hideOverlap: true }
    },
    yAxis: {
      type: 'value',
      min: 0,
      minInterval: 1,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: appStore.isDark ? '#cfd3dc' : '#606266', fontSize: 12 },
      splitLine: { lineStyle: { color: appStore.isDark ? '#414243' : '#ebeef5' } }
    },
    tooltip: {
      trigger: 'axis',
      appendToBody: true,
      confine: true,
      backgroundColor: appStore.isDark ? '#2d2d2d' : '#ffffff',
      borderColor: appStore.isDark ? '#414243' : '#e4e7ed',
      textStyle: { color: appStore.isDark ? '#ffffff' : '#303133' },
      formatter: (params: Array<{ axisValue: string; value: number }>) => {
        const point = params[0]
        return `${point.axisValue}<br/>${t('chart.licenseTrend.tooltip.licenseCount')}: ${point.value}`
      }
    },
    series: [{
      type: 'line',
      data: values,
      smooth: true,
      showSymbol: values.length <= 31,
      symbolSize: 6,
      lineStyle: { color: '#019c7c', width: 3 },
      itemStyle: { color: '#019c7c' },
      areaStyle: { color: 'rgba(1, 156, 124, 0.10)' },
      emphasis: { focus: 'series' }
    }]
  }
})
</script>

<style scoped>
.chart-container { width: 100%; min-height: 220px; display: flex; align-items: center; justify-content: center; }
.trend-chart { width: 100%; height: 280px; }
@media (max-width: 1023px) { .trend-chart { height: 240px; } }
@media (max-width: 767px) { .chart-container { min-height: 200px; } .trend-chart { height: 220px; } }
</style>
