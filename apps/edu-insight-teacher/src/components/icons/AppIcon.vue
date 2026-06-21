<script setup lang="ts">
import { computed } from 'vue'
import { iconRegistry, type IconName } from './iconRegistry'

const props = withDefaults(defineProps<{
  name: IconName
  size?: number | string
  strokeWidth?: number
  title?: string
}>(), {
  size: 20,
  strokeWidth: 1.9,
  title: ''
})

const icon = computed(() => iconRegistry[props.name])
const sizeValue = computed(() => typeof props.size === 'number' ? `${props.size}px` : props.size)
</script>

<template>
  <svg
    class="app-icon"
    :width="sizeValue"
    :height="sizeValue"
    viewBox="0 0 24 24"
    fill="none"
    aria-hidden="true"
    focusable="false"
  >
    <title v-if="title">{{ title }}</title>
    <path
      v-for="path in icon.paths"
      :key="path"
      :d="path"
      stroke="currentColor"
      :stroke-width="strokeWidth"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  </svg>
</template>
