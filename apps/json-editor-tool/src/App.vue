<template>
  <div class="app-container">
    <header class="header">
      <h1>JSON Editor Tool</h1>
      <div class="toolbar">
        <button @click="formatJson(2)" class="btn">格式化 (2空格)</button>
        <button @click="formatJson(4)" class="btn">格式化 (4空格)</button>
        <button @click="compressJson" class="btn">压缩</button>
        <button @click="validateJson" class="btn">校验</button>
        <button @click="toggleTreeView" class="btn">{{ showTree ? '代码视图' : '树形视图' }}</button>
        <select v-model="indentSize" class="select">
          <option :value="2">2空格缩进</option>
          <option :value="4">4空格缩进</option>
        </select>
      </div>
    </header>

    <main class="main-content">
      <div class="editor-panel" v-show="!showTree">
        <div ref="editorContainer" class="monaco-editor"></div>
      </div>
      
      <div class="tree-panel" v-show="showTree">
        <TreeView :data="parsedJson" @select="onTreeSelect" />
      </div>

      <aside class="sidebar">
        <div class="panel">
          <h3>JSON 路径查询</h3>
          <input 
            v-model="jsonPath" 
            placeholder="$.store.book[*].author"
            class="input"
            @keyup.enter="queryJsonPath"
          />
          <button @click="queryJsonPath" class="btn btn-sm">查询</button>
          <div v-if="queryResult" class="result">
            <pre>{{ queryResult }}</pre>
          </div>
        </div>

        <div class="panel">
          <h3>状态</h3>
          <div class="status">
            <span :class="['status-badge', isValid ? 'valid' : 'invalid']">
              {{ isValid ? '✓ 有效 JSON' : '✗ 无效 JSON' }}
            </span>
          </div>
          <div v-if="errorMessage" class="error-message">
            {{ errorMessage }}
          </div>
        </div>

        <div class="panel">
          <h3>统计</h3>
          <div class="stats">
            <p>字符数: {{ charCount }}</p>
            <p>行数: {{ lineCount }}</p>
            <p>键值对数: {{ keyCount }}</p>
          </div>
        </div>
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, computed } from 'vue'
import * as monaco from 'monaco-editor'
import { JSONPath } from 'jsonpath-plus'
import TreeView from './components/TreeView.vue'

const editorContainer = ref<HTMLElement>()
const editor = ref<monaco.editor.IStandaloneCodeEditor>()
const jsonContent = ref('')
const parsedJson = ref<any>(null)
const isValid = ref(true)
const errorMessage = ref('')
const showTree = ref(false)
const indentSize = ref(2)
const jsonPath = ref('')
const queryResult = ref('')

// Monaco Editor JSON 语言配置
monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
  validate: true,
  allowComments: false,
  schemas: [],
  enableSchemaRequest: false
})

// 计算统计信息
const charCount = computed(() => jsonContent.value.length)
const lineCount = computed(() => jsonContent.value.split('\n').length)
const keyCount = computed(() => {
  try {
    const obj = JSON.parse(jsonContent.value)
    return countKeys(obj)
  } catch {
    return 0
  }
})

function countKeys(obj: any): number {
  if (obj === null || typeof obj !== 'object') return 0
  let count = 0
  if (Array.isArray(obj)) {
    for (const item of obj) {
      count += countKeys(item)
    }
  } else {
    for (const key in obj) {
      count++
      count += countKeys(obj[key])
    }
  }
  return count
}

function initEditor() {
  if (!editorContainer.value) return

  editor.value = monaco.editor.create(editorContainer.value, {
    value: jsonContent.value,
    language: 'json',
    theme: 'vs-dark',
    automaticLayout: true,
    fontSize: 14,
    minimap: { enabled: false },
    formatOnPaste: false,
    formatOnType: false,
    lineNumbers: 'on',
    scrollBeyondLastLine: false,
    wordWrap: 'on',
    tabSize: indentSize.value
  })

  editor.value.onDidChangeModelContent(() => {
    jsonContent.value = editor.value?.getValue() || ''
    validateJson()
    updateTreeView()
  })
}

function formatJson(spaces: number) {
  try {
    const obj = JSON.parse(jsonContent.value)
    const formatted = JSON.stringify(obj, null, spaces)
    editor.value?.setValue(formatted)
    isValid.value = true
    errorMessage.value = ''
  } catch (e: any) {
    isValid.value = false
    errorMessage.value = e.message
  }
}

function compressJson() {
  try {
    const obj = JSON.parse(jsonContent.value)
    const compressed = JSON.stringify(obj)
    editor.value?.setValue(compressed)
    isValid.value = true
    errorMessage.value = ''
  } catch (e: any) {
    isValid.value = false
    errorMessage.value = e.message
  }
}

function validateJson() {
  try {
    const value = jsonContent.value.trim()
    if (!value) {
      isValid.value = true
      errorMessage.value = ''
      parsedJson.value = null
      return
    }
    const obj = JSON.parse(value)
    isValid.value = true
    errorMessage.value = ''
    parsedJson.value = obj
  } catch (e: any) {
    isValid.value = false
    errorMessage.value = e.message
    parsedJson.value = null
  }
}

function updateTreeView() {
  if (isValid.value && jsonContent.value.trim()) {
    try {
      parsedJson.value = JSON.parse(jsonContent.value)
    } catch {
      parsedJson.value = null
    }
  } else {
    parsedJson.value = null
  }
}

function toggleTreeView() {
  showTree.value = !showTree.value
}

function queryJsonPath() {
  if (!jsonPath.value || !parsedJson.value) return
  try {
    const result = JSONPath({ path: jsonPath.value, json: parsedJson.value })
    queryResult.value = JSON.stringify(result, null, 2)
  } catch (e: any) {
    queryResult.value = `查询错误: ${e.message}`
  }
}

function onTreeSelect(path: string) {
  if (!editor.value) return
  
  const lines = jsonContent.value.split('\n')
  // 简单定位：查找键名所在行
  const key = path.split('.').pop()
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].includes(`"${key}"`)) {
      editor.value.revealLineInCenter(i + 1)
      editor.value.setPosition({ lineNumber: i + 1, column: 1 })
      editor.value.focus()
      break
    }
  }
}

watch(indentSize, (newSize) => {
  if (editor.value) {
    editor.value.getModel()?.updateOptions({ tabSize: newSize })
  }
})

onMounted(() => {
  jsonContent.value = '{\n  "example": "Hello JSON Editor"\n}'
  initEditor()
  validateJson()
})

onBeforeUnmount(() => {
  editor.value?.dispose()
})
</script>

<style scoped>
.app-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  color: #d4d4d4;
}

.header {
  padding: 16px 24px;
  background: #252526;
  border-bottom: 1px solid #3c3c3c;
}

.header h1 {
  margin: 0 0 12px 0;
  font-size: 20px;
  color: #569cd6;
}

.toolbar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.btn {
  padding: 8px 16px;
  background: #0e639c;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.btn:hover {
  background: #1177bb
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.select {
  padding: 8px 12px;
  background: #3c3c3c;
  color: #d4d4d4;
  border: 1px solid #555;
  border-radius: 4px;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.editor-panel, .tree-panel {
  flex: 1;
  overflow: hidden;
}

.monaco-editor {
  width: 100%;
  height: 100%;
}

.sidebar {
  width: 280px;
  background: #252526;
  border-left: 1px solid #3c3c3c;
  padding: 16px;
  overflow-y: auto;
}

.panel {
  margin-bottom: 20px;
}

.panel h3 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #9cdcfe;
}

.input {
  width: 100%;
  padding: 8px 12px;
  background: #3c3c3c;
  color: #d4d4d4;
  border: 1px solid #555;
  border-radius: 4px;
  margin-bottom: 8px;
  box-sizing: border-box;
}

.input:focus {
  outline: none;
  border-color: #0e639c;
}

.result {
  margin-top: 12px;
  padding: 8px;
  background: #1e1e1e;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
}

.result pre {
  margin: 0;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.status-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
}

.status-badge.valid {
  background: #4ec9b0;
  color: #1e1e1e;
}

.status-badge.invalid {
  background: #f14c4c;
  color: white;
}

.error-message {
  margin-top: 8px;
  padding: 8px;
  background: #5a1d1d;
  border-radius: 4px;
  font-size: 12px;
  color: #f14c4c;
}

.stats p {
  margin: 4px 0;
  font-size: 13px;
  color: #808080;
}
</style>
