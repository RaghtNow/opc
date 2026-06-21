<script setup lang="ts">
export type DistributionRow = {
  label: string
  range: string
  count: number
  percent: number
  width: number
  tone: string
}

defineProps<{
  rows: DistributionRow[]
}>()
</script>

<template>
  <div class="distribution-bars">
    <div v-for="row in rows" :key="row.label" class="distribution-row">
      <div class="distribution-head">
        <strong>{{ row.label }}</strong>
        <span>{{ row.count }} 人 · {{ row.percent }}%</span>
      </div>
      <div class="distribution-track">
        <div :class="['distribution-fill', `tone-${row.tone}`]" :style="{ width: `${row.width}%` }"></div>
      </div>
      <p>{{ row.range }}</p>
    </div>
  </div>
</template>

<style scoped>
.distribution-bars {
  display: grid;
  gap: 12px;
}

.distribution-row {
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--panel-alt);
}

.distribution-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.distribution-head span,
.distribution-row p {
  color: var(--muted);
}

.distribution-row p {
  margin: 10px 0 0;
  line-height: 1.6;
}

.distribution-track {
  height: 10px;
  margin-top: 12px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(74, 52, 24, 0.1);
}

.distribution-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #9a5b26, #d99a50);
}

.distribution-fill.tone-strong {
  background: linear-gradient(90deg, #2f6f4e, #7dbb7d);
}

.distribution-fill.tone-steady {
  background: linear-gradient(90deg, #9a5b26, #d99a50);
}

.distribution-fill.tone-risk {
  background: linear-gradient(90deg, #a13d2d, #e28a69);
}
</style>
