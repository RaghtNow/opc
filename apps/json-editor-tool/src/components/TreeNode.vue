<template>
  <div class="tree-node">
    <div 
      class="node-header" 
      :style="{ paddingLeft: depth * 16 + 'px' }"
      @click="toggle"
    >
      <span class="toggle" :class="{ expanded: isExpanded }">
        {{ isExpandable ? (isExpanded ? '▼' : '▶') : ' ' }}
      </span>
      <span class="key" :class="valueType">{{ name }}</span>
      <span class="colon">: </span>
      <span class="value" :class="valueType">
        {{ displayValue }}
      </span>
    </div>
    <div v-if="isExpanded && isExpandable" class="children">
      <TreeNode
        v-for="(item, key) in children"
        :key="key"
        :data="item"
        :name="Array.isArray(data) ? String(key) : String(key)"
        :path="currentPath + '.' + key"
        :depth="depth + 1"
        @select="onSelect"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  data: any
  name: string
  path: string
  depth: number
}>()

const emit = defineEmits<{
  select: [path: string]
}>()

const isExpanded = ref(props.depth < 2)

const valueType = computed(() => {
  if (props.data === null) return 'null'
  if (typeof props.data === 'boolean') return 'boolean'
  if (typeof props.data === 'number') return 'number'
  if (typeof props.data === 'string') return 'string'
  if (Array.isArray(props.data)) return 'array'
  if (typeof props.data === 'object') return 'object'
  return 'unknown'
})

const isExpandable = computed(() => {
  return typeof props.data === 'object' && props.data !== null
})

const displayValue = computed(() => {
  if (props.data === null) return 'null'
  if (typeof props.data === 'boolean') return String(props.data)
  if (typeof props.data === 'number') return String(props.data)
  if (typeof props.data === 'string') return `"${props.data}"`
  if (Array.isArray(props.data)) return `Array(${props.data.length})`
  if (typeof props.data === 'object') return `Object(${Object.keys(props.data).length})`
  return String(props.data)
})

const children = computed(() => {
  if (Array.isArray(props.data)) return props.data
  if (typeof props.data === 'object' && props.data !== null) return props.data
  return {}
})

const currentPath = computed(() => props.path)

function toggle() {
  if (isExpandable.value) {
    isExpanded.value = !isExpanded.value
  }
}

function onSelect(path: string) {
  emit('select', path)
}
</script>

<style scoped>
.tree-node {
  user-select: none;
}

.node-header {
  display: flex;
  align-items: center;
  padding: 2px 0;
  cursor: pointer;
}

.node-header:hover {
  background: #2a2d2e;
}

.toggle {
  width: 16px;
  font-size: 10px;
  color: #808080;
}

.toggle.expanded {
  color: #d4d4d4;
}

.key {
  color: #9cdcfe;
}

.key.string::after {
  content: ':';
  color: #d4d4d4;
}

.colon {
  color: #d4d4d4;
}

.value {
  margin-left: 4px;
}

.value.string {
  color: #ce9178;
}

.value.number {
  color: #b5cea8;
}

.value.boolean {
  color: #569cd6;
}

.value.null {
  color: #569cd6;
}

.value.array,
.value.object {
  color: #808080;
  font-style: italic;
}

.children {
  margin-left: 0;
}
</style>
