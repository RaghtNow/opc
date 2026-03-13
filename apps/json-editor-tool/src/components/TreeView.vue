<template>
  <div class="tree-view">
    <div v-if="!data" class="empty">
      无有效 JSON 数据
    </div>
    <TreeNode 
      v-else 
      :data="data" 
      :name="'root'" 
      :path="'$'"
      :depth="0"
      @select="onSelect"
    />
  </div>
</template>

<script setup lang="ts">
import TreeNode from './TreeNode.vue'

defineProps<{
  data: any
}>()

const emit = defineEmits<{
  select: [path: string]
}>()

function onSelect(path: string) {
  emit('select', path)
}
</script>

<style scoped>
.tree-view {
  height: 100%;
  padding: 16px;
  overflow: auto;
  background: #1e1e1e;
  color: #d4d4d4;
  font-size: 14px;
  font-family: 'Consolas', 'Monaco', monospace;
}

.empty {
  color: #808080;
  text-align: center;
  padding: 40px;
}
</style>
