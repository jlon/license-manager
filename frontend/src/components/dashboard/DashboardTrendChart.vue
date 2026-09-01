<template>
  <div class="dashboard-trend-chart">
    <VChart v-if="data.length" class="dashboard-trend-chart__canvas" :option="chartOption" autoresize />
    <el-empty v-else :description="emptyText" :image-size="58" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { useAppStore } from '@/store/modules/app'
import type { DashboardTrendPoint } from '@/api/dashboard'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent])

const props = defineProps<{
  data: DashboardTrendPoint[]
  emptyText: string
  color: string
}>()
const appStore = useAppStore()

const chartOption = computed(() => ({
  animationDuration: 350,
  color: [props.color],
  grid: { left: 18, right: 18, top: 18, bottom: 14, containLabel: true },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: props.data.map(item => item.date.substring(5)),
    axisLine: { lineStyle: { color: appStore.isDark ? '#606266' : '#dceae5' } },
    axisTick: { show: false },
    axisLabel: { color: appStore.isDark ? '#c1d0ca' : '#52685f', fontSize: 11, hideOverlap: true }
  },
  yAxis: {
    type: 'value', min: 0, minInterval: 1,
    axisLine: { show: false }, axisTick: { show: false },
    axisLabel: { color: appStore.isDark ? '#c1d0ca' : '#52685f', fontSize: 11 },
    splitLine: { lineStyle: { color: appStore.isDark ? '#414243' : '#e7f0ec' } }
  },
  tooltip: {
    trigger: 'axis', appendToBody: true, confine: true,
    backgroundColor: appStore.isDark ? '#2d2d2d' : '#ffffff',
    borderColor: appStore.isDark ? '#414243' : '#e4e7ed',
    textStyle: { color: appStore.isDark ? '#ffffff' : '#294138' }
  },
  series: [{
    type: 'line', data: props.data.map(item => item.count), smooth: true,
    showSymbol: props.data.length <= 31, symbolSize: 6,
    lineStyle: { width: 3 }, areaStyle: { opacity: 0.1 }, emphasis: { focus: 'series' }
  }]
}))
</script>

<style scoped>
.dashboard-trend-chart { min-height: 240px; display: flex; align-items: center; justify-content: center; }
.dashboard-trend-chart__canvas { width: 100%; height: 240px; }
@media (max-width: 767px) { .dashboard-trend-chart, .dashboard-trend-chart__canvas { min-height: 210px; height: 210px; } }
</style>
