<script setup lang="ts">
import { computed } from 'vue'

export type BarChartRow = {
  label: string
  value: number
  max?: number
  meta?: string
  note?: string
  tone?: 'strong' | 'steady' | 'risk'
}

const props = defineProps<{
  rows: BarChartRow[]
}>()

const normalizedRows = computed(() => {
  const fallbackMax = Math.max(...props.rows.map((row) => row.value), 1)
  return props.rows.map((row) => {
    const max = Math.max(row.max ?? fallbackMax, 1)
    return {
      ...row,
      width: Math.max(5, Math.min(100, (row.value / max) * 100))
    }
  })
})
</script>

<template>
  <div class="horizontal-bars">
    <div v-for="row in normalizedRows" :key="row.label" class="horizontal-bar-row">
      <div class="horizontal-bar-head">
        <strong>{{ row.label }}</strong>
        <span>{{ row.meta ?? row.value }}</span>
      </div>
      <div class="horizontal-bar-track">
        <div :class="['horizontal-bar-fill', row.tone ? `tone-${row.tone}` : '']" :style="{ width: `${row.width}%` }"></div>
      </div>
      <p v-if="row.note">{{ row.note }}</p>
    </div>
  </div>
</template>

<style scoped>
.horizontal-bars {
  display: grid;
  gap: 12px;
}

.horizontal-bar-row {
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--panel-alt);
}

.horizontal-bar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.horizontal-bar-head span,
.horizontal-bar-row p {
  color: var(--muted);
}

.horizontal-bar-row p {
  margin: 8px 0 0;
  font-size: 0.88rem;
}

.horizontal-bar-track {
  height: 10px;
  margin-top: 12px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(74, 52, 24, 0.1);
}

.horizontal-bar-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #9a5b26, #d99a50);
}

.horizontal-bar-fill.tone-strong {
  background: linear-gradient(90deg, #2f6f4e, #7dbb7d);
}

.horizontal-bar-fill.tone-steady {
  background: linear-gradient(90deg, #9a5b26, #d99a50);
}

.horizontal-bar-fill.tone-risk {
  background: linear-gradient(90deg, #a13d2d, #e28a69);
}
</style>
