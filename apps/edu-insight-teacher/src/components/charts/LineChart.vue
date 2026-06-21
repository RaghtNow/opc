<script setup lang="ts">
import { computed } from 'vue'

export type ChartPoint = {
  label: string
  value: number
  note?: string
}

const props = withDefaults(defineProps<{
  points: ChartPoint[]
  height?: number
  min?: number
  max?: number
  invert?: boolean
  valueSuffix?: string
}>(), {
  height: 180,
  min: undefined,
  max: undefined,
  invert: false,
  valueSuffix: ''
})

const width = 1000
const top = 28
const right = 34
const bottom = 40
const left = 46

const plotHeight = computed(() => Math.max(60, props.height - top - bottom))
const plotWidth = width - left - right

const values = computed(() => props.points.map((point) => point.value).filter(Number.isFinite))

const domain = computed(() => {
  const rawMin = props.min ?? Math.min(...values.value, 0)
  const rawMax = props.max ?? Math.max(...values.value, 1)
  const range = Math.max(1, rawMax - rawMin)
  return {
    min: rawMin - range * 0.08,
    max: rawMax + range * 0.08
  }
})

const chartPoints = computed(() => {
  const count = props.points.length
  const range = Math.max(1, domain.value.max - domain.value.min)
  return props.points.map((point, index) => {
    const x = left + (count <= 1 ? plotWidth / 2 : (index / (count - 1)) * plotWidth)
    const ratio = (point.value - domain.value.min) / range
    const y = props.invert
      ? top + ratio * plotHeight.value
      : top + (1 - ratio) * plotHeight.value
    return { ...point, x, y }
  })
})

const linePath = computed(() => {
  if (chartPoints.value.length === 0) return ''
  return chartPoints.value
    .map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x.toFixed(2)} ${point.y.toFixed(2)}`)
    .join(' ')
})

const areaPath = computed(() => {
  if (chartPoints.value.length === 0) return ''
  const baseline = top + plotHeight.value
  const first = chartPoints.value[0]
  const last = chartPoints.value[chartPoints.value.length - 1]
  return `${linePath.value} L ${last.x.toFixed(2)} ${baseline} L ${first.x.toFixed(2)} ${baseline} Z`
})

function formatValue(value: number) {
  return Number.isInteger(value) ? `${value}${props.valueSuffix}` : `${value.toFixed(1)}${props.valueSuffix}`
}
</script>

<template>
  <div class="line-chart" :style="{ height: `${height}px` }">
    <svg :viewBox="`0 0 ${width} ${height}`" role="img">
      <path class="line-chart-area" :d="areaPath" />
      <line
        v-for="step in 4"
        :key="step"
        class="line-chart-grid"
        :x1="left"
        :x2="width - right"
        :y1="top + (step - 1) * (plotHeight / 3)"
        :y2="top + (step - 1) * (plotHeight / 3)"
      />
      <path class="line-chart-path" :d="linePath" />
      <g v-for="point in chartPoints" :key="`${point.label}-${point.value}`">
        <circle class="line-chart-dot" :cx="point.x" :cy="point.y" r="8" />
        <text class="line-chart-value" :x="point.x" :y="point.y - 14" text-anchor="middle">
          {{ formatValue(point.value) }}
        </text>
        <text class="line-chart-label" :x="point.x" :y="height - 12" text-anchor="middle">
          {{ point.label }}
        </text>
      </g>
    </svg>
  </div>
</template>

<style scoped>
.line-chart {
  width: 100%;
  overflow: hidden;
  border-radius: 18px;
  background:
    linear-gradient(to top, rgba(74, 52, 24, 0.08) 1px, transparent 1px) 0 0 / 100% 25%,
    linear-gradient(180deg, rgba(255, 253, 249, 0.82), rgba(248, 242, 232, 0.28));
}

.line-chart svg {
  display: block;
  width: 100%;
  height: 100%;
}

.line-chart-grid {
  stroke: rgba(74, 52, 24, 0.1);
  stroke-width: 1;
}

.line-chart-area {
  fill: rgba(154, 91, 38, 0.12);
}

.line-chart-path {
  fill: none;
  stroke: var(--accent);
  stroke-width: 5;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.line-chart-dot {
  fill: var(--panel);
  stroke: var(--accent);
  stroke-width: 5;
}

.line-chart-value {
  fill: var(--ink);
  font-size: 30px;
  font-weight: 700;
}

.line-chart-label {
  fill: var(--muted);
  font-size: 25px;
  font-weight: 600;
}
</style>
