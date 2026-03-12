<script setup lang="ts">
import { ref, computed } from 'vue'
import { parseJson, formatJson, minifyJson, validateJson } from '../utils/jsonUtils'

// 输入内容
const inputText = ref('')

// 输出内容
const outputText = ref('')

// 错误信息
const errorMessage = ref('')

// 验证结果
const validationResult = ref<any>(null)

// 当前模式
const mode = ref<'format' | 'minify'>('format')

// 缩进空格数
const indent = ref(2)

// 解析状态
const isParsed = ref(false)

// 执行解析/格式化
function handleParse() {
  if (!inputText.value.trim()) {
    errorMessage.value = '请输入 JSON 内容'
    validationResult.value = null
    outputText.value = ''
    isParsed.value = false
    return
  }
  
  // 先验证
  validationResult.value = validateJson(inputText.value)
  
  if (!validationResult.value.isValid) {
    errorMessage.value = validationResult.value.error || 'JSON 格式错误'
    outputText.value = ''
    isParsed.value = false
    return
  }
  
  // 解析
  const result = parseJson(inputText.value)
  
  if (result.success) {
    errorMessage.value = ''
    isParsed.value = true
    
    // 根据模式输出
    if (mode.value === 'format') {
      outputText.value = formatJson(result.data, indent.value)
    } else {
      outputText.value = minifyJson(result.data)
    }
  } else {
    errorMessage.value = result.error || '解析失败'
    outputText.value = ''
    isParsed.value = false
  }
}

// 切换模式时重新处理
function handleModeChange() {
  if (isParsed.value) {
    handleParse()
  }
}

// 复制输出
async function copyOutput() {
  if (outputText.value) {
    await navigator.clipboard.writeText(outputText.value)
    alert('已复制到剪贴板')
  }
}

// 清空
function clearAll() {
  inputText.value = ''
  outputText.value = ''
  errorMessage.value = ''
  validationResult.value = null
  isParsed.value = false
}

// 示例数据
const sampleJson = `{
  "name": "JSON Parser",
  "version": "1.0.0",
  "features": ["parse", "format", "validate"],
  "config": {
    "indent": 2,
    "theme": "light"
  }
}`

function loadSample() {
  inputText.value = sampleJson
  handleParse()
}

// 计算验证状态颜色
const validationStatusColor = computed(() => {
  if (!validationResult.value) return '#666'
  return validationResult.value.isValid ? '#52c41a' : '#ff4d4f'
})

// 计算验证状态文本
const validationStatusText = computed(() => {
  if (!validationResult.value) return '未验证'
  if (validationResult.value.isValid) {
    const type = validationResult.value.type
    const size = validationResult.value.size
    return `有效 JSON (${type}, ${size} 字符)`
  }
  return '无效 JSON'
})
</script>

<template>
  <div class="parser-container">
    <header class="header">
      <h1>🔧 JSON Parser Tool</h1>
      <p class="subtitle">解析 · 格式化 · 验证</p>
    </header>
    
    <main class="main-content">
      <!-- 输入区 -->
      <section class="input-section">
        <div class="section-header">
          <h2>📥 输入 JSON</h2>
          <div class="header-actions">
            <button @click="loadSample" class="btn btn-secondary">加载示例</button>
            <button @click="clearAll" class="btn btn-secondary">清空</button>
          </div>
        </div>
        <textarea
          v-model="inputText"
          class="json-input"
          placeholder='{"key": "value"}'
          spellcheck="false"
        ></textarea>
        
        <!-- 错误提示 -->
        <div v-if="errorMessage" class="error-box">
          <span class="error-icon">❌</span>
          <span>{{ errorMessage }}</span>
          <span v-if="validationResult?.errorLine" class="error-position">
            (行 {{ validationResult.errorLine }}, 列 {{ validationResult.errorColumn }})
          </span>
        </div>
        
        <!-- 操作栏 -->
        <div class="action-bar">
          <div class="mode-select">
            <label>
              <input type="radio" v-model="mode" value="format" @change="handleModeChange" />
              美化 (Format)
            </label>
            <label v-if="mode === 'format'">
              缩进:
              <select v-model="indent" @change="handleParse" class="indent-select">
                <option :value="2">2 空格</option>
                <option :value="4">4 空格</option>
              </select>
            </label>
            <label>
              <input type="radio" v-model="mode" value="minify" @change="handleModeChange" />
              压缩 (Minify)
            </label>
          </div>
          <button @click="handleParse" class="btn btn-primary">🔄 解析 & 转换</button>
        </div>
      </section>
      
      <!-- 验证状态 -->
      <div v-if="validationResult" class="validation-status" :style="{ color: validationStatusColor }">
        <span class="status-icon">{{ validationResult.isValid ? '✅' : '❌' }}</span>
        <span>{{ validationStatusText }}</span>
        <span v-if="validationResult.isValid && validationResult.keys">
          · {{ validationResult.keyCount }} 个键
        </span>
        <span v-if="validationResult.isValid && validationResult.length !== undefined">
          · {{ validationResult.length }} 个元素
        </span>
      </div>
      
      <!-- 输出区 -->
      <section class="output-section">
        <div class="section-header">
          <h2>📤 输出结果</h2>
          <button 
            @click="copyOutput" 
            class="btn btn-secondary"
            :disabled="!outputText"
          >
            📋 复制结果
          </button>
        </div>
        <textarea
          v-model="outputText"
          class="json-output"
          placeholder="格式化后的 JSON 将显示在这里..."
          readonly
          spellcheck="false"
        ></textarea>
      </section>
    </main>
  </div>
</template>

<style scoped>
.parser-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.header {
  text-align: center;
  padding: 30px 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 12px;
  margin-bottom: 24px;
}

.header h1 {
  font-size: 2rem;
  margin-bottom: 8px;
}

.subtitle {
  font-size: 1rem;
  opacity: 0.9;
}

.main-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-header h2 {
  font-size: 1.2rem;
  color: #333;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.json-input,
.json-output {
  width: 100%;
  min-height: 300px;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  font-size: 14px;
  line-height: 1.6;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  resize: vertical;
  background: #fafafa;
  transition: border-color 0.2s;
}

.json-input:focus,
.json-output:focus {
  outline: none;
  border-color: #667eea;
}

.json-output {
  background: #f0f0f0;
}

.error-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 6px;
  color: #ff4d4f;
  margin-top: 12px;
}

.error-icon {
  font-size: 16px;
}

.error-position {
  color: #999;
  font-size: 12px;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding: 12px 16px;
  background: #f9f9f9;
  border-radius: 8px;
}

.mode-select {
  display: flex;
  align-items: center;
  gap: 16px;
}

.mode-select label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  color: #333;
}

.indent-select {
  padding: 4px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-left: 4px;
}

.validation-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #f6ffed;
  border: 1px solid #b7eb8f;
  border-radius: 6px;
  font-size: 14px;
}

.status-icon {
  font-size: 16px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5a6fd6;
}

.btn-secondary {
  background: #f0f0f0;
  color: #333;
}

.btn-secondary:hover {
  background: #e0e0e0;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
