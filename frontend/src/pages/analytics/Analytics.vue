<script setup lang="ts">
import { LineChart as IconLineChart, PieChart as IconPieChart } from 'lucide-vue-next'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, PieChart, LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent, TitleComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { billApi } from '@/api'

use([
  CanvasRenderer,
  BarChart,
  LineChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  TitleComponent,
])

const router = useRouter()

const hasTrendData = ref(false)
const hasCategoryData = ref(false)

const trendOption = ref({
  tooltip: { trigger: 'axis', backgroundColor: 'rgba(255,255,255,0.9)' },
  grid: { left: '4%', right: '5%', bottom: '15%', containLabel: true },
  xAxis: { type: 'category', data: [], axisLine: { lineStyle: { color: '#bdc3c7' } }, boundaryGap: false },
  yAxis: { type: 'value', splitLine: { lineStyle: { type: 'dashed', color: '#f0ebe4' } } },
  series: [
    { 
      name: '支出', 
      type: 'line', 
      smooth: true,
      showSymbol: false,
      areaStyle: {
        color: {
          type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
          colorStops: [{ offset: 0, color: 'rgba(230, 126, 34, 0.4)' }, { offset: 1, color: 'rgba(230, 126, 34, 0.05)' }]
        }
      },
      itemStyle: { color: '#e67e22' },
      data: []
    }
  ]
})

const pieOption = ref({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0, padding: 0 },
  series: [
    {
      name: '支出分类',
      type: 'pie',
      radius: ['45%', '70%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: false, position: 'center' },
      emphasis: { label: { show: true, fontSize: 18, fontWeight: 'bold' } },
      labelLine: { show: false },
      data: []
    }
  ]
})

async function fetchCharts() {
  try {
    const [trendRes, catRes]: any = await Promise.all([
      billApi.getTrendStats(),
      billApi.getCategoryStats()
    ])

    if (trendRes.code === 0) {
      const dataList = trendRes.data || []
      trendOption.value.xAxis.data = dataList.map((d: any) => d.month)
      trendOption.value.series[0].data = dataList.map((d: any) => d.expense)
      hasTrendData.value = dataList.some((d: any) => d.expense > 0)
    }

    if (catRes.code === 0) {
      const dataList = catRes.data || []
      pieOption.value.series[0].data = dataList.map((d: any) => ({
        name: d.name,
        value: d.value
      }))
      hasCategoryData.value = dataList.length > 0
    }
  } catch (err) {
    console.error('Failed to load charts:', err)
  }
}

onMounted(() => {
  fetchCharts()
})
</script>

<template>
  <div class="analytics-page">
    <div class="page-header">
      <h1>月度报表</h1>
    </div>

    <div class="charts-grid">
      <div class="card chart-card">
        <div class="section-header">
          <h2>收支趋势分析</h2>
        </div>
        <div class="chart-empty" v-if="!hasTrendData">
          <IconLineChart :size="36" class="empty-icon" />
          <p>你的账本空空如也，缺少指引</p>
          <button class="btn btn-secondary btn-sm" @click="router.push('/settings')">
            配置账单拉取 →
          </button>
        </div>
        <v-chart class="chart" :option="trendOption" autoresize v-else />
      </div>

      <div class="card chart-card">
        <div class="section-header">
          <h2>分类支出雷达</h2>
        </div>
        <div class="chart-empty" v-if="!hasCategoryData">
          <IconPieChart :size="36" class="empty-icon" />
          <p>本月暂无记账数据</p>
          <button class="btn btn-ghost btn-sm" @click="router.push('/settings')">
            探索记账姿势
          </button>
        </div>
        <v-chart class="chart" :option="pieOption" autoresize v-else />
      </div>
    </div>
  </div>
</template>

<style scoped>
.analytics-page {
  padding: 16px;
  max-width: 600px;
  margin: 0 auto;
  padding-top: calc(24px + env(safe-area-inset-top));
}
.page-header h1 {
  font-size: 1.6rem;
  font-weight: 700;
  margin: 0 0 24px 0;
  color: var(--text-primary);
}
.charts-grid {
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.chart-card {
  padding: 20px;
  display: flex;
  flex-direction: column;
}
.section-header h2 {
  font-size: 1.15rem;
  font-weight: 600;
  margin: 0 0 16px 0;
}

.chart {
  width: 100%;
  height: 260px;
}

.chart-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 0;
  color: var(--text-secondary);
  background: transparent;
  min-height: 200px;
}
.empty-icon {
  margin-bottom: 12px;
  opacity: 0.6;
}
.chart-empty p {
  margin: 0 0 12px 0;
  font-size: 0.9rem;
}
.chart-empty button {
  font-size: 0.8rem;
  padding: 6px 14px;
}
</style>
