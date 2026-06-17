<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  alertItems,
  classBaseFields,
  classMembers,
  cohortInsights,
  currentClassStage,
  displayPolicies,
  examRecords,
  importBatchMetrics,
  importIssues,
  navItems,
  rosterInsights,
  scoreEntries,
  scoreFlowSteps,
  scoreDecisions,
  studentTrends,
  subjectScopeClasses,
  summaryMetrics,
  syncAudienceCards,
  syncRecords,
  teacherAssignments,
  workContexts
} from './data/dashboard'
import {
  bindTeacherAccount,
  createStudent,
  createTeacher,
  fetchClassroomWorkspace,
  importStudents,
  importTeachers,
  syncTeacherPermission,
  updatePolicy,
  updateStudent,
  updateTeacher,
  type ClassroomWorkspace
} from './api/classroom'
import { fetchExamDetail, fetchExams, importExam, updateExamScore } from './api/score'
import { fetchInsightDashboard, publishLatestExam } from './api/insight'
import { parseAndValidateCsv } from './utils/csvImport'

const activeNav = ref('overview')
const activeContextId = ref(workContexts[0].id)
const subjectScopeMode = ref<'single' | 'overall'>('single')
const activeSubjectClassId = ref(subjectScopeClasses[0].id)
const selectedExamId = ref(examRecords[0].id)
const classPanelMode = ref<'student' | 'teacher' | 'policy'>('student')
const selectedStudentId = ref('')
const selectedTeacherId = ref(teacherAssignments[0].id)
const selectedPolicyId = ref(displayPolicies[0].id)
const actionPanel = ref<
  | 'none'
  | 'student-add'
  | 'student-import'
  | 'student-edit'
  | 'teacher-add'
  | 'teacher-import'
  | 'teacher-edit'
  | 'policy-edit'
  | 'score-edit'
  | 'score-upload'
>('none')

const students = ref(structuredClone(classMembers))
const teachers = ref(structuredClone(teacherAssignments))
const policies = ref(structuredClone(displayPolicies))
const classBaseFieldRows = ref(structuredClone(classBaseFields))
const rosterInsightRows = ref(structuredClone(rosterInsights))
const classStage = ref(structuredClone(currentClassStage))
const exams = ref(structuredClone(examRecords))
const issues = ref(structuredClone(importIssues))
const scoreRows = ref(structuredClone(scoreEntries))
const scoreLoading = ref(false)
const scoreError = ref('')
const classroomLoading = ref(false)
const classroomError = ref('')
const classroomActionMessage = ref('')
const insightLoading = ref(false)
const insightError = ref('')
const insightActionMessage = ref('')
const summaryMetricRows = ref(structuredClone(summaryMetrics))
const studentTrendRows = ref(structuredClone(studentTrends))
const cohortInsightRows = ref(structuredClone(cohortInsights))
const alertRows = ref(structuredClone(alertItems))
const syncAudienceRows = ref(structuredClone(syncAudienceCards))
const syncRecordRows = ref(structuredClone(syncRecords))
const latestInsightExamName = ref('')
const canPublishLatestExam = ref(false)
const publishBlockers = ref<string[]>([])
const studentImportInput = ref<HTMLInputElement | null>(null)
const teacherImportInput = ref<HTMLInputElement | null>(null)
const studentImportFile = ref<File | null>(null)
const teacherImportFile = ref<File | null>(null)
const studentImportFileName = ref('')
const teacherImportFileName = ref('')

const studentForm = ref({
  id: '',
  studentNo: '',
  name: '',
  gender: '男',
  combination: '',
  parentMobile: '',
  parentStatus: '待补充',
  selectionStatus: '待确认'
})

const teacherForm = ref({
  id: '',
  subject: '',
  teacher: '',
  mobile: '',
  classes: '',
})

const policyForm = ref({
  id: '',
  title: '',
  value: '',
  note: ''
})

const examForm = ref({
  name: '',
  type: '月考',
  date: '2026-06-20',
  subjects: ['语文', '数学', '英语']
})

const scoreUploadForm = ref({
  examId: examRecords[0].id,
  fileName: '',
  fileSelected: false,
  file: null as File | null
})
const scoreUploadStep = ref<'meta' | 'upload' | 'validate' | 'confirm'>('meta')
const scoreFileInput = ref<HTMLInputElement | null>(null)
const uploadValidationState = ref({
  loading: false,
  error: '',
  summary: [] as Array<{ field: string; result: string; note: string }>,
  headers: [] as string[],
  previewRows: [] as Array<Record<string, string>>
})

const scoreEditForm = ref({
  id: '',
  studentName: '',
  subjectScores: {} as Record<string, string>,
  total: ''
})

const activeTitle = computed(() => {
  const current = navItems.find((item) => item.id === activeNav.value)
  return current?.label ?? '工作台'
})

const activeContext = computed(() => {
  return workContexts.find((item) => item.id === activeContextId.value) ?? workContexts[0]
})

const isSubjectTeacherView = computed(() => activeContext.value.roleLabel === '任课老师')

const activeSubjectClass = computed(() => {
  return subjectScopeClasses.find((item) => item.id === activeSubjectClassId.value) ?? subjectScopeClasses[0]
})

const activeSelectionScenario = computed(() => classStage.value)

const selectedStudent = computed(() => {
  return students.value.find((item) => item.id === selectedStudentId.value) ?? students.value[0]
})

const selectedTeacher = computed(() => {
  return teachers.value.find((item) => item.id === selectedTeacherId.value) ?? teachers.value[0]
})

const selectedPolicy = computed(() => {
  return policies.value.find((item) => item.id === selectedPolicyId.value) ?? policies.value[0]
})

const selectedExam = computed(() => {
  return exams.value.find((item) => item.id === selectedExamId.value) ?? exams.value[0]
})

const selectedExamSubjects = computed(() => {
  const subjects: string[] = selectedExam.value?.subjects ?? []
  if (subjects.length > 0) return subjects
  return ['语文', '数学', '英语']
})

const scoreGridStyle = computed(() => ({
  gridTemplateColumns: `1.3fr repeat(${selectedExamSubjects.value.length}, minmax(86px, 0.75fr)) 0.85fr 1fr`
}))

const importBatchMetricsView = computed(() => {
  const pendingCount = issues.value.filter((item) => item.status !== '已修复').length

  return [
    importBatchMetrics[0],
    importBatchMetrics[1],
    {
      label: '异常条目',
      value: `${pendingCount}`,
      note: pendingCount === 0 ? '当前批次异常已全部修复' : '待人工修复后重新校验'
    }
  ]
})

const knownStudentIds = computed(() => students.value.map((item) => item.studentNo))
const knownStudentNames = computed(() => students.value.map((item) => item.name))

function applyClassroomWorkspace(workspace: ClassroomWorkspace) {
  classBaseFieldRows.value = workspace.baseFields
  rosterInsightRows.value = workspace.rosterInsights
  classStage.value = workspace.stage
  students.value = workspace.students
  teachers.value = workspace.teachers
  policies.value = workspace.policies

  if (students.value.length > 0 && !students.value.find((item) => item.id === selectedStudentId.value)) {
    selectedStudentId.value = students.value[0].id
  }
  if (teachers.value.length > 0 && !teachers.value.find((item) => item.id === selectedTeacherId.value)) {
    selectedTeacherId.value = teachers.value[0].id
  }
  if (policies.value.length > 0 && !policies.value.find((item) => item.id === selectedPolicyId.value)) {
    selectedPolicyId.value = policies.value[0].id
  }
}

async function loadClassroomWorkspace() {
  try {
    classroomLoading.value = true
    classroomError.value = ''
    applyClassroomWorkspace(await fetchClassroomWorkspace())
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '获取班级与师生数据失败'
  } finally {
    classroomLoading.value = false
  }
}

function openScoreImport() {
  activeNav.value = 'scores'
  actionPanel.value = 'score-upload'
  scoreUploadStep.value = 'meta'
  uploadValidationState.value = {
    loading: false,
    error: '',
    summary: [],
    headers: [],
    previewRows: []
  }
}

function openExamDetail(examId: string) {
  selectedExamId.value = examId
  activeNav.value = 'scores'
  actionPanel.value = 'none'
  void loadExamDetail(examId)
}

function resolveIssue(issueId: string) {
  const target = issues.value.find((item) => item.id === issueId)
  if (target) target.status = '已修复'
}

function openStudentDetail(studentId: string) {
  activeNav.value = 'classes'
  classPanelMode.value = 'student'
  selectedStudentId.value = studentId
  actionPanel.value = 'none'
}

function openTeacherDetail(teacherId: string) {
  activeNav.value = 'classes'
  classPanelMode.value = 'teacher'
  selectedTeacherId.value = teacherId
  actionPanel.value = 'none'
}

function openStudentEditor(studentId?: string) {
  if (studentId) {
    const target = students.value.find((item) => item.id === studentId)
    if (!target) return
    selectedStudentId.value = studentId
    studentForm.value = {
      id: target.id,
      studentNo: target.studentNo,
      name: target.name,
      gender: target.gender,
      combination: target.combination,
      parentMobile: target.parentMobile,
      parentStatus: target.parentStatus,
      selectionStatus: target.selectionStatus
    }
    actionPanel.value = 'student-edit'
  } else {
    studentForm.value = {
      id: '',
      studentNo: '',
      name: '',
      gender: '男',
      combination: '',
      parentMobile: '',
      parentStatus: '待补充',
      selectionStatus: '待确认'
    }
    actionPanel.value = 'student-add'
  }
}

async function saveStudentForm() {
  const payload = {
    studentNo: studentForm.value.studentNo,
    name: studentForm.value.name,
    gender: studentForm.value.gender,
    combination: studentForm.value.combination,
    parentMobile: studentForm.value.parentMobile,
    parentStatus: studentForm.value.parentStatus,
    selectionStatus: studentForm.value.selectionStatus
  }

  try {
    classroomError.value = ''
    const workspace = studentForm.value.id
      ? await updateStudent(studentForm.value.id, payload)
      : await createStudent(payload)
    applyClassroomWorkspace(workspace)
    const selected = studentForm.value.id
      ? studentForm.value.id
      : workspace.students.find((item) => item.studentNo === payload.studentNo && item.name === payload.name)?.id
    if (selected) selectedStudentId.value = selected
    classPanelMode.value = 'student'
    actionPanel.value = 'none'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '保存学生失败'
  }
}

function openTeacherEditor(teacherId?: string) {
  if (teacherId) {
    const target = teachers.value.find((item) => item.id === teacherId)
    if (!target) return
    selectedTeacherId.value = teacherId
    teacherForm.value = {
      id: target.id,
      subject: target.subject,
      teacher: target.teacher,
      mobile: target.mobile,
      classes: target.classes,
    }
    actionPanel.value = 'teacher-edit'
  } else {
    teacherForm.value = {
      id: '',
      subject: '',
      teacher: '',
      mobile: '',
      classes: '',
    }
    actionPanel.value = 'teacher-add'
  }
}

async function saveTeacherForm() {
  const payload = {
    subject: teacherForm.value.subject,
    teacher: teacherForm.value.teacher,
    mobile: teacherForm.value.mobile,
    classes: teacherForm.value.classes,
  }

  try {
    classroomError.value = ''
    const workspace = teacherForm.value.id
      ? await updateTeacher(teacherForm.value.id, payload)
      : await createTeacher(payload)
    applyClassroomWorkspace(workspace)
    const selected = teacherForm.value.id
      ? teacherForm.value.id
      : workspace.teachers.find((item) => item.subject === payload.subject && item.teacher === payload.teacher)?.id
    if (selected) selectedTeacherId.value = selected
    classPanelMode.value = 'teacher'
    actionPanel.value = 'none'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '保存任课老师失败'
  }
}

async function bindTeacher(teacherId: string) {
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    applyClassroomWorkspace(await bindTeacherAccount(teacherId))
    selectedTeacherId.value = teacherId
    classPanelMode.value = 'teacher'
    classroomActionMessage.value = '教师账号已绑定，可以继续同步权限。'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '绑定教师账号失败'
  }
}

async function syncTeacher(teacherId: string) {
  const target = teachers.value.find((item) => item.id === teacherId)
  if (target?.accountStatus !== 'bound') {
    classroomActionMessage.value = '请先绑定教师账号，再同步权限。'
    return
  }
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    applyClassroomWorkspace(await syncTeacherPermission(teacherId))
    selectedTeacherId.value = teacherId
    classPanelMode.value = 'teacher'
    classroomActionMessage.value = '权限已同步，任课老师可查看授权范围内数据。'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '同步教师权限失败'
  }
}

function handleStudentImportFileChange(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  studentImportFile.value = file
  studentImportFileName.value = file.name
}

function handleTeacherImportFileChange(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  teacherImportFile.value = file
  teacherImportFileName.value = file.name
}

async function submitStudentImport() {
  if (!studentImportFile.value) return
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    const result = await importStudents(studentImportFile.value)
    applyClassroomWorkspace(result.workspace)
    classroomActionMessage.value = `学生导入完成：新增 ${result.summary.created} 条，更新 ${result.summary.updated} 条，跳过 ${result.summary.skipped} 条。`
    studentImportFile.value = null
    studentImportFileName.value = ''
    actionPanel.value = 'none'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '导入学生档案失败'
  }
}

async function submitTeacherImport() {
  if (!teacherImportFile.value) return
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    const result = await importTeachers(teacherImportFile.value)
    applyClassroomWorkspace(result.workspace)
    classroomActionMessage.value = `任课老师导入完成：新增 ${result.summary.created} 条，更新 ${result.summary.updated} 条，跳过 ${result.summary.skipped} 条。`
    teacherImportFile.value = null
    teacherImportFileName.value = ''
    actionPanel.value = 'none'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '导入任课老师失败'
  }
}

function openPolicyEditor(policyId?: string) {
  const target = policies.value.find((item) => item.id === (policyId ?? selectedPolicyId.value))
  if (!target) return
  selectedPolicyId.value = target.id
  policyForm.value = {
    id: target.id,
    title: target.title,
    value: target.value,
    note: target.note
  }
  actionPanel.value = 'policy-edit'
}

async function savePolicyForm() {
  try {
    classroomError.value = ''
    applyClassroomWorkspace(await updatePolicy(policyForm.value.id, {
      value: policyForm.value.value,
      note: policyForm.value.note
    }))
    selectedPolicyId.value = policyForm.value.id
    classPanelMode.value = 'policy'
    actionPanel.value = 'none'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '保存策略失败'
  }
}

function saveScoreUpload() {
  importExam({
    name: examForm.value.name || '新考试成绩记录',
    type: examForm.value.type,
    date: examForm.value.date,
    subjects: examForm.value.subjects,
    subjectCoverage: examForm.value.subjects.join(' / '),
    fileName: scoreUploadForm.value.fileName,
    scores: scoreRows.value,
    issues: issues.value
  }).then(async (detail) => {
    await loadExamList()
    selectedExamId.value = detail.exam.id
    scoreRows.value = detail.scores
    issues.value = detail.issues
  })
  actionPanel.value = 'none'
  scoreUploadStep.value = 'meta'
  scoreUploadForm.value.fileName = ''
  scoreUploadForm.value.fileSelected = false
  scoreUploadForm.value.file = null
  uploadValidationState.value = {
    loading: false,
    error: '',
    summary: [],
    headers: [],
    previewRows: []
  }
}

function triggerScoreFilePick() {
  scoreFileInput.value?.click()
}

function handleScoreFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  scoreUploadForm.value.fileName = file.name
  scoreUploadForm.value.fileSelected = true
  scoreUploadForm.value.file = file
}

async function runImportValidation() {
  if (!scoreUploadForm.value.file) return

  uploadValidationState.value.loading = true
  uploadValidationState.value.error = ''

  const result = await parseAndValidateCsv({
    file: scoreUploadForm.value.file,
    selectedSubjects: examForm.value.subjects,
    stage: activeSelectionScenario.value.id as 'pre-selection' | 'post-selection',
    knownStudentIds: knownStudentIds.value,
    knownStudentNames: knownStudentNames.value
  })

  uploadValidationState.value.loading = false

  if (!result.ok && result.error) {
    uploadValidationState.value.error = result.error
    return
  }

  uploadValidationState.value.summary = result.validationSummary
  uploadValidationState.value.headers = result.headers
  uploadValidationState.value.previewRows = result.rows.slice(0, 5)
  issues.value = result.issues
  scoreRows.value = result.scoreRows
}

function openScoreEditor(scoreId: string) {
  const target = scoreRows.value.find((item) => item.id === scoreId)
  if (!target) return

  scoreEditForm.value = {
    id: target.id,
    studentName: target.studentName,
    subjectScores: Object.fromEntries(
      selectedExamSubjects.value.map((subject: string) => [subject, target.subjectScores?.[subject] ?? ''])
    ),
    total: target.total
  }

  actionPanel.value = 'score-edit'
}

function saveScoreEdit() {
  const subjectScores = scoreEditForm.value.subjectScores
  updateExamScore(selectedExamId.value, scoreEditForm.value.id, {
    chinese: subjectScores['语文'] ?? '',
    math: subjectScores['数学'] ?? '',
    english: subjectScores['英语'] ?? '',
    electiveLabel: selectedExamSubjects.value.filter((subject: string) => !['语文', '数学', '英语'].includes(subject)).join(''),
    electiveScore: '',
    subjectScores,
    total: scoreEditForm.value.total
  }).then((detail) => {
    scoreRows.value = detail.scores
    issues.value = detail.issues
  })
  actionPanel.value = 'none'
}

async function loadExamList() {
  try {
    scoreError.value = ''
    exams.value = await fetchExams()
    if (exams.value.length > 0 && !exams.value.find((item) => item.id === selectedExamId.value)) {
      selectedExamId.value = exams.value[0].id
    }
  } catch (error) {
    scoreError.value = error instanceof Error ? error.message : '获取考试列表失败'
  }
}

async function loadExamDetail(examID: string) {
  try {
    scoreLoading.value = true
    scoreError.value = ''
    const result = await fetchExamDetail(examID)
    scoreRows.value = result.scores
    issues.value = result.issues
  } catch (error) {
    scoreError.value = error instanceof Error ? error.message : '获取考试详情失败'
  } finally {
    scoreLoading.value = false
  }
}

async function loadInsightDashboard() {
  try {
    insightLoading.value = true
    insightError.value = ''
    const dashboard = await fetchInsightDashboard()
    summaryMetricRows.value = dashboard.summaryMetrics
    studentTrendRows.value = dashboard.studentTrends
    cohortInsightRows.value = dashboard.cohortInsights
    alertRows.value = dashboard.alertItems
    syncAudienceRows.value = dashboard.syncAudienceCards
    syncRecordRows.value = dashboard.syncRecords
    latestInsightExamName.value = dashboard.latestExamName
    canPublishLatestExam.value = dashboard.canPublish
    publishBlockers.value = dashboard.publishBlockers
  } catch (error) {
    insightError.value = error instanceof Error ? error.message : '获取分析与同步数据失败'
  } finally {
    insightLoading.value = false
  }
}

async function publishLatestExamSync() {
  try {
    insightError.value = ''
    insightActionMessage.value = ''
    const dashboard = await publishLatestExam()
    summaryMetricRows.value = dashboard.summaryMetrics
    studentTrendRows.value = dashboard.studentTrends
    cohortInsightRows.value = dashboard.cohortInsights
    alertRows.value = dashboard.alertItems
    syncAudienceRows.value = dashboard.syncAudienceCards
    syncRecordRows.value = dashboard.syncRecords
    latestInsightExamName.value = dashboard.latestExamName
    canPublishLatestExam.value = dashboard.canPublish
    publishBlockers.value = dashboard.publishBlockers
    insightActionMessage.value = '已创建最近考试的家长、学生和任课老师同步任务。'
  } catch (error) {
    insightError.value = error instanceof Error ? error.message : '发布同步任务失败'
  }
}

onMounted(async () => {
  await loadClassroomWorkspace()
  await loadExamList()
  if (selectedExamId.value) {
    await loadExamDetail(selectedExamId.value)
  }
  await loadInsightDashboard()
})
</script>

<template>
  <div class="teacher-shell">
    <aside class="sidebar">
      <div class="brand">
        <p class="brand-kicker">OPC</p>
        <h1>成绩洞察平台</h1>
        <p class="brand-copy">教师端可体验原型</p>
      </div>

      <nav class="nav-list">
        <button
          v-for="item in navItems"
          :key="item.id"
          :class="['nav-item', { active: activeNav === item.id }]"
          type="button"
          @click="activeNav = item.id"
        >
          <strong>{{ item.label }}</strong>
          <span>{{ item.hint }}</span>
        </button>
      </nav>
    </aside>

    <main class="main-area">
      <header class="topbar">
        <div>
          <p class="topbar-label">当前视角：{{ activeContext.roleLabel }}</p>
          <h2>{{ activeTitle }}</h2>
        </div>

        <div class="topbar-actions">
          <label class="topbar-select">
            <span>工作身份</span>
            <select v-model="activeContextId">
              <option v-for="context in workContexts" :key="context.id" :value="context.id">
                {{ context.roleLabel }} / {{ context.primaryLabel }} / {{ context.secondaryLabel }}
              </option>
            </select>
          </label>
          <label v-if="isSubjectTeacherView" class="topbar-select">
            <span>查看范围</span>
            <select v-model="subjectScopeMode">
              <option value="single">单班分析</option>
              <option value="overall">任课整体</option>
            </select>
          </label>
          <label v-if="isSubjectTeacherView && subjectScopeMode === 'single'" class="topbar-select">
            <span>当前班级</span>
            <select v-model="activeSubjectClassId">
              <option v-for="item in subjectScopeClasses" :key="item.id" :value="item.id">
                {{ item.label }}
              </option>
            </select>
          </label>
          <button type="button" class="solid-btn" @click="openScoreImport">录入考试成绩</button>
        </div>
      </header>

      <template v-if="activeNav === 'overview'">
        <section class="metrics-grid">
          <article v-for="metric in summaryMetricRows" :key="metric.label" class="metric-card">
            <p>{{ metric.label }}</p>
            <strong>{{ metric.value }}</strong>
            <span>{{ metric.note }}</span>
          </article>
        </section>

        <section class="content-grid">
          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">最近趋势</p>
                <h3>学生变化样例</h3>
              </div>
            </div>

            <div class="trend-grid">
              <div v-for="student in studentTrendRows" :key="student.name" class="trend-card">
                <strong>{{ student.name }}</strong>
                <div class="trend-score">{{ student.totalScore }}</div>
                <div class="trend-meta">
                  <span>{{ student.delta }}</span>
                  <p>{{ student.tag }}</p>
                </div>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">重点预警</p>
                <h3>需优先跟进</h3>
              </div>
            </div>

            <div class="alert-list">
              <div v-for="alert in alertRows.slice(0, 3)" :key="`${alert.student}-${alert.subject}`" class="alert-item">
                <div class="alert-top">
                  <strong>{{ alert.student }}</strong>
                  <span>{{ alert.level }}</span>
                </div>
                <p>{{ alert.subject }} · {{ alert.detail }}</p>
              </div>
            </div>
          </article>
        </section>
      </template>

      <template v-else-if="activeNav === 'classes'">
        <section class="content-grid">
          <article v-if="classroomError" class="panel panel-wide">
            <div class="task-item">
              <strong>班级与师生接口异常</strong>
              <p>{{ classroomError }}</p>
              <span>请确认 Go 后端已运行在 127.0.0.1:8088</span>
            </div>
          </article>

          <article v-if="classroomLoading" class="panel panel-wide">
            <div class="status-inline">正在加载班级、学生、任课老师和展示策略...</div>
          </article>

          <article v-if="classroomActionMessage" class="panel panel-wide">
            <div class="status-inline">{{ classroomActionMessage }}</div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">班级基础信息</p>
                <h3>班级档案与阶段状态</h3>
              </div>
              <div class="panel-actions">
                <button type="button" class="ghost-btn small" @click="classPanelMode = 'policy'; openPolicyEditor()">
                  编辑班级策略
                </button>
                <button type="button" class="ghost-btn small" @click="actionPanel = 'student-import'">
                  导入学生档案
                </button>
              </div>
            </div>

            <div class="status-inline">
              <span v-if="activeSelectionScenario.id === 'pre-selection'">
                当前阶段以行政班和通用科目为主，暂不启用选科组合与赋分分析。
              </span>
                <span v-else>
                  当前阶段已启用选科组合、教学班关系和选考科赋分分析。
                </span>
              </div>

            <div class="base-grid">
              <div v-for="field in classBaseFieldRows" :key="field.label" class="base-item">
                <span>{{ field.label }}</span>
                <strong>{{ field.value }}</strong>
              </div>
              <div class="base-item">
                <span>班级阶段</span>
                <strong>{{ activeSelectionScenario.label }}</strong>
              </div>
            </div>

            <div class="insight-grid compact-insights">
              <div v-for="item in rosterInsightRows" :key="item.title" class="insight-card">
                <strong>{{ item.title }}</strong>
                <div class="insight-count">{{ item.count }}</div>
                <p>{{ item.detail }}</p>
              </div>
            </div>

            <div class="table-list">
              <div class="table-row table-header five-cols">
                <span>学号</span>
                <span>姓名</span>
                <span>性别</span>
                <span>选科组合</span>
                <span>家长状态</span>
              </div>
              <div v-for="student in students" :key="student.studentNo" class="table-row five-cols">
                <span>{{ student.studentNo }}</span>
                <span>{{ student.name }}</span>
                <span>{{ student.gender }}</span>
                <span>{{ student.combination }}</span>
                <span>{{ student.status }}</span>
              </div>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">学生档案管理</p>
                <h3>学生信息完整度与维护</h3>
              </div>
              <div class="panel-actions">
                <button type="button" class="ghost-btn small" @click="openStudentEditor()">手动新增学生</button>
                <button type="button" class="ghost-btn small" @click="actionPanel = 'student-import'">导入学生档案</button>
              </div>
            </div>

            <div class="table-list">
              <div class="table-row table-header six-cols">
                <span>学号</span>
                <span>姓名</span>
                <span>家长状态</span>
                <span>选科状态</span>
                <span>档案状态</span>
                <span>操作</span>
              </div>
              <div
                v-for="student in students"
                :key="student.id"
                class="table-row six-cols"
              >
                <span>{{ student.studentNo }}</span>
                <span>{{ student.name }}</span>
                <span>{{ student.parentStatus }}</span>
                <span>{{ student.selectionStatus }}</span>
                <span>{{ student.status }}</span>
                <span class="row-actions">
                  <button type="button" class="ghost-btn tiny" @click="openStudentDetail(student.id)">查看</button>
                  <button type="button" class="ghost-btn tiny" @click="openStudentEditor(student.id)">编辑</button>
                </span>
              </div>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">任课老师维护</p>
                <h3>班主任维护学科老师并同步权限</h3>
              </div>
              <div class="panel-actions">
                <button type="button" class="ghost-btn small" @click="openTeacherEditor()">新增任课老师</button>
                <button type="button" class="ghost-btn small" @click="actionPanel = 'teacher-import'">导入任课老师信息</button>
              </div>
            </div>

            <div class="table-list">
              <div class="table-row table-header six-cols">
                <span>学科</span>
                <span>任课老师</span>
                <span>手机号</span>
                <span>账号状态</span>
                <span>权限状态</span>
                <span>操作</span>
              </div>
              <div
                v-for="assignment in teachers"
                :key="assignment.id"
                class="table-row six-cols"
              >
                <span>{{ assignment.subject }}</span>
                <span>{{ assignment.teacher }}</span>
                <span>{{ assignment.mobile || '未维护' }}</span>
                <span>{{ assignment.accountStatus === 'bound' ? '已绑定' : '待绑定' }}</span>
                <span>{{ assignment.permissionStatus === 'synced' ? '已同步' : '待同步' }}</span>
                <span class="row-actions">
                  <button type="button" class="ghost-btn tiny" @click="openTeacherEditor(assignment.id)">编辑</button>
                  <button
                    v-if="assignment.accountStatus !== 'bound'"
                    type="button"
                    class="ghost-btn tiny"
                    @click="bindTeacher(assignment.id)"
                  >
                    绑定账号
                  </button>
                  <button
                    v-else-if="assignment.permissionStatus !== 'synced'"
                    type="button"
                    class="ghost-btn tiny"
                    @click="syncTeacher(assignment.id)"
                  >
                    同步权限
                  </button>
                  <button v-else type="button" class="ghost-btn tiny" @click="openTeacherDetail(assignment.id)">
                    查看
                  </button>
                </span>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">维护动作</p>
                <h3>班主任本周操作建议</h3>
              </div>
            </div>

            <div class="task-list">
              <div class="task-item">
                <strong>补任课老师账号</strong>
                <p>优先完成化学老师绑定，避免无法同步单科分析。</p>
                <span>1 项待处理</span>
              </div>
              <div class="task-item">
                <strong>补家长手机号</strong>
                <p>当前仍有 2 位家长未完成联系方式维护。</p>
                <span>2 项待处理</span>
              </div>
              <div class="task-item">
                <strong>确认选科阶段</strong>
                <p>未选科阶段班级不启用赋分字段和组合分析。</p>
                <span>{{ activeSelectionScenario.label }}</span>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">班级展示策略</p>
                <h3>当前策略摘要</h3>
              </div>
              <button type="button" class="ghost-btn small" @click="openPolicyEditor()">调整策略</button>
            </div>

            <div class="task-list">
              <div v-for="policy in policies" :key="policy.id" class="task-item">
                <strong>{{ policy.title }}</strong>
                <p>{{ policy.note }}</p>
                <span>{{ policy.value }}</span>
                <div class="row-actions top-gap">
                  <button type="button" class="ghost-btn tiny" @click="openPolicyEditor(policy.id)">查看配置</button>
                </div>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">当前维护对象</p>
                <h3>
                  {{
                    classPanelMode === 'student'
                      ? '学生详情'
                      : classPanelMode === 'teacher'
                        ? '任课老师详情'
                        : '策略配置详情'
                  }}
                </h3>
              </div>
            </div>

            <div v-if="classPanelMode === 'student'" class="task-list">
              <div class="task-item">
                <strong>{{ selectedStudent.name }} / {{ selectedStudent.studentNo }}</strong>
                <p>{{ selectedStudent.gender }} · {{ selectedStudent.combination }}</p>
                <span>{{ selectedStudent.status }}</span>
              </div>
              <div class="task-item">
                <strong>家长信息</strong>
                <p>{{ selectedStudent.parentMobile }}</p>
                <span>{{ selectedStudent.parentStatus }}</span>
              </div>
              <div class="task-item">
                <strong>选科信息</strong>
                <p>{{ selectedStudent.combination }}</p>
                <span>{{ selectedStudent.selectionStatus }}</span>
              </div>
            </div>

            <div v-else-if="classPanelMode === 'teacher'" class="task-list">
              <div class="task-item">
                <strong>{{ selectedTeacher.teacher }} / {{ selectedTeacher.subject }}</strong>
                <p>{{ selectedTeacher.classes }}</p>
                <span>{{ selectedTeacher.mobile || '未维护手机号' }}</span>
              </div>
              <div class="task-item">
                <strong>账号状态</strong>
                <p>{{ selectedTeacher.accountStatus === 'bound' ? selectedTeacher.accountId : '待绑定教师账号' }}</p>
                <span>{{ selectedTeacher.accountBoundAt || '需先维护手机号并执行绑定，后续换手机号不改变账号 ID' }}</span>
              </div>
              <div class="task-item">
                <strong>权限状态</strong>
                <p>{{ selectedTeacher.permissionStatus === 'synced' ? '权限已同步' : '权限待同步' }}</p>
                <span>{{ selectedTeacher.permissionSyncedAt || selectedTeacher.syncStatus }}</span>
              </div>
              <div class="task-item">
                <strong>建议动作</strong>
                <p>先维护手机号并绑定教师账号，再同步授权范围；手机号变更只更新登录凭证，不重建账号。</p>
                <span>适用于跨班任课老师</span>
              </div>
            </div>

            <div v-else class="task-list">
              <div class="task-item">
                <strong>{{ selectedPolicy.title }}</strong>
                <p>{{ selectedPolicy.note }}</p>
                <span>{{ selectedPolicy.value }}</span>
              </div>
              <div class="task-item">
                <strong>当前控制方式</strong>
                <p>由学校上限控制，班主任在允许范围内进行班级级配置。</p>
                <span>同班统一生效</span>
              </div>
              <div class="task-item">
                <strong>建议动作</strong>
                <p>在成绩发布前确认是否展示班级位置与分数段。</p>
                <span>避免家长端理解偏差</span>
              </div>
            </div>
          </article>

        </section>
      </template>

      <template v-else-if="activeNav === 'scores'">
        <section class="content-grid">
          <article v-if="scoreError" class="panel panel-wide">
            <div class="task-item">
              <strong>考试与成绩接口异常</strong>
              <p>{{ scoreError }}</p>
              <span>请确认 Go 后端已运行在 127.0.0.1:8088</span>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">考试与成绩总览</p>
                <h3>所有已导入成绩记录</h3>
              </div>
              <button type="button" class="solid-btn small" @click="openScoreImport">导入成绩</button>
            </div>

            <div class="table-list">
              <div class="table-row table-header five-cols">
                <span>考试名称</span>
                <span>考试类型</span>
                <span>日期</span>
                <span>学科覆盖</span>
                <span>导入状态</span>
              </div>
              <button
                v-for="exam in exams"
                :key="exam.id"
                type="button"
                :class="['table-row', 'five-cols', 'clickable-row', { selected: selectedExamId === exam.id }]"
                @click="openExamDetail(exam.id)"
              >
                <span>{{ exam.name }}</span>
                <span>{{ exam.type }}</span>
                <span>{{ exam.date }}</span>
                <span>{{ exam.subjectCoverage }}</span>
                <span>{{ exam.importStatus }}</span>
              </button>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">当前选中考试</p>
                <h3>{{ selectedExam.name }}</h3>
              </div>
            </div>

            <div v-if="scoreLoading" class="status-inline">正在加载考试详情与成绩明细...</div>

            <div class="task-list">
              <div class="task-item">
                <strong>考试类型</strong>
                <p>{{ selectedExam.type }}</p>
                <span>{{ selectedExam.subjectCoverage }}</span>
              </div>
              <div class="task-item">
                <strong>导入状态</strong>
                <p>{{ selectedExam.importStatus }}</p>
                <span>点击列表任一考试即可切换当前详情</span>
              </div>
              <div class="task-item">
                <strong>录入口径</strong>
                <p>导入成绩时直接填写考试信息，不再单独拆出创建考试流程。</p>
                <span>符合老师一次性录分习惯</span>
              </div>
            </div>

            <div class="score-table-wrap top-gap">
              <div class="score-table-title">
                <strong>考试成绩明细</strong>
                <span>按本次考试学科动态展示，支持逐行修改成绩</span>
              </div>
              <div class="table-list score-table-list">
              <div class="table-row table-header score-cols-dynamic" :style="scoreGridStyle">
                <span>学生</span>
                <span v-for="subject in selectedExamSubjects" :key="subject">{{ subject }}</span>
                <span>总分</span>
                <span>操作</span>
              </div>

              <div
                v-for="row in scoreRows"
                :key="row.id"
                class="table-row score-cols-dynamic"
                :style="scoreGridStyle"
              >
                <span>{{ row.studentName }}</span>
                <span v-for="subject in selectedExamSubjects" :key="subject">{{ row.subjectScores?.[subject] || '-' }}</span>
                <span>{{ row.total }}</span>
                <span class="row-actions">
                  <button type="button" class="ghost-btn tiny" @click="openScoreEditor(row.id)">修改成绩</button>
                </span>
              </div>
              </div>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">当前批次状态</p>
                <h3>导入与校验摘要</h3>
              </div>
            </div>

            <div class="metrics-grid compact-metrics">
              <article v-for="metric in importBatchMetricsView" :key="metric.label" class="metric-card">
                <p>{{ metric.label }}</p>
                <strong>{{ metric.value }}</strong>
                <span>{{ metric.note }}</span>
              </article>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">录入流程</p>
                <h3>一次完整的考试成绩处理</h3>
              </div>
            </div>

            <div class="flow-grid">
              <div v-for="step in scoreFlowSteps" :key="step.title" class="flow-card">
                <strong>{{ step.title }}</strong>
                <p>{{ step.description }}</p>
                <span>{{ step.status }}</span>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">录入规则</p>
                <h3>当前口径说明</h3>
              </div>
            </div>

            <ul class="bullet-list">
              <li>导入成绩时直接填写考试信息，不单独拆出创建考试动作</li>
              <li>上传成绩文件后系统自动校验、清洗并计算排名</li>
              <li v-if="activeSelectionScenario.id === 'post-selection'">主科默认只要求实际分，选考科目允许补充赋分</li>
              <li v-else>当前阶段只要求通用科目实际分，不启用选考赋分导入</li>
            </ul>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">导入异常</p>
                <h3>当前批次问题项</h3>
              </div>
            </div>

            <div class="issue-list">
              <div v-for="issue in issues" :key="issue.id" class="issue-card">
                <div class="issue-head">
                  <strong>第 {{ issue.row }} 行 · {{ issue.student }}</strong>
                  <span class="issue-state">{{ issue.status }}</span>
                </div>
                <p>{{ issue.issue }}</p>
                <span>{{ issue.suggestion }}</span>
                <div class="issue-actions">
                  <button type="button" class="ghost-btn small" @click="resolveIssue(issue.id)">
                    标记已修复
                  </button>
                </div>
              </div>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">关键处理决策</p>
                <h3>本次成绩处理口径</h3>
              </div>
            </div>

            <div class="decision-grid">
              <div v-for="item in scoreDecisions" :key="item.title" class="decision-card">
                <strong>{{ item.title }}</strong>
                <div class="decision-result">{{ item.result }}</div>
                <p>{{ item.detail }}</p>
              </div>
            </div>
          </article>

        </section>
      </template>

      <template v-else-if="activeNav === 'analysis'">
        <article v-if="insightError" class="panel panel-wide">
          <div class="task-item">
            <strong>分析与同步接口异常</strong>
            <p>{{ insightError }}</p>
            <span>请确认 Go 后端已运行在 127.0.0.1:8088</span>
          </div>
        </article>

        <article v-if="insightLoading" class="panel panel-wide">
          <div class="status-inline">正在基于真实考试成绩生成分析...</div>
        </article>

        <section class="metrics-grid">
          <article v-for="metric in summaryMetricRows" :key="metric.label" class="metric-card">
            <p>{{ metric.label }}</p>
            <strong>{{ metric.value }}</strong>
            <span>{{ metric.note }}</span>
          </article>
        </section>

        <section class="content-grid">
          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">班级分析</p>
                <h3>
                  {{
                    isSubjectTeacherView && subjectScopeMode === 'overall'
                      ? '任课整体分析'
                      : isSubjectTeacherView
                        ? `${activeSubjectClass.label} 单班分析`
                        : '进退步与分层'
                  }}
                </h3>
              </div>
            </div>

            <div class="trend-grid">
              <div v-for="student in studentTrendRows" :key="student.name" class="trend-card">
                <strong>{{ student.name }}</strong>
                <div class="trend-score">{{ student.totalScore }}</div>
                <div class="trend-meta">
                  <span>{{ student.delta }}</span>
                  <p>{{ student.tag }}</p>
                </div>
              </div>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">分层洞察</p>
                <h3>班级人群结构</h3>
              </div>
            </div>

            <div class="insight-grid">
              <div v-for="item in cohortInsightRows" :key="item.title" class="insight-card">
                <strong>{{ item.title }}</strong>
                <div class="insight-count">{{ item.students }}</div>
                <p>{{ item.insight }}</p>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">预警概览</p>
                <h3>当前重点关注</h3>
              </div>
            </div>

            <div class="alert-list">
              <div v-for="alert in alertRows.slice(0, 3)" :key="`${alert.student}-${alert.subject}`" class="alert-item">
                <div class="alert-top">
                  <strong>{{ alert.student }}</strong>
                  <span>{{ alert.level }}</span>
                </div>
                <p>{{ alert.subject }} · {{ alert.detail }}</p>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">班主任建议</p>
                <h3>下一步动作建议</h3>
              </div>
            </div>

            <div class="task-list">
              <div class="task-item">
                <strong>先处理高风险预警</strong>
                <p>优先关注连续下滑和单科异常波动学生。</p>
                <span>建议今日完成</span>
              </div>
              <div class="task-item">
                <strong>再补齐同步条件</strong>
                <p>补家长手机号和任课老师账号，避免同步链路断点。</p>
                <span>2 个待补项</span>
              </div>
              <div class="task-item">
                <strong>确认是否发家长</strong>
                <p>若异常记录全部修复，可进入正式同步环节。</p>
                <span>待班主任确认</span>
              </div>
            </div>
          </article>

          <article v-if="isSubjectTeacherView && subjectScopeMode === 'overall'" class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">任课整体</p>
                <h3>跨班级统计</h3>
              </div>
            </div>

            <div class="task-list">
              <div class="task-item">
                <strong>覆盖班级</strong>
                <p>高二（3）班、高一（8）班</p>
                <span>共 96 人</span>
              </div>
              <div class="task-item">
                <strong>整体均分</strong>
                <p>数学学科总览</p>
                <span>110.8 / 150</span>
              </div>
              <div class="task-item">
                <strong>重点关注</strong>
                <p>低于学科预警线学生</p>
                <span>12 人</span>
              </div>
            </div>
          </article>

          <article v-else-if="isSubjectTeacherView" class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">当前班级</p>
                <h3>{{ activeSubjectClass.label }} 单班视角</h3>
              </div>
            </div>

            <div class="task-list">
              <div class="task-item">
                <strong>班级均分</strong>
                <p>{{ activeSubjectClass.label }} / 数学</p>
                <span>112.6 / 150</span>
              </div>
              <div class="task-item">
                <strong>薄弱学生</strong>
                <p>低于预警线学生</p>
                <span>5 人</span>
              </div>
              <div class="task-item">
                <strong>进步最快</strong>
                <p>单班本次提升幅度最高</p>
                <span>许一诺 +13</span>
              </div>
            </div>
          </article>
        </section>
      </template>

      <template v-else-if="activeNav === 'sync'">
        <section class="content-grid">
          <article v-if="insightError" class="panel panel-wide">
            <div class="task-item">
              <strong>同步接口异常</strong>
              <p>{{ insightError }}</p>
              <span>请确认 Go 后端已运行在 127.0.0.1:8088</span>
            </div>
          </article>

          <article v-if="insightActionMessage" class="panel panel-wide">
            <div class="status-inline">{{ insightActionMessage }}</div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">同步对象准备度</p>
                <h3>当前可触达状态</h3>
              </div>
            </div>

            <div class="insight-grid">
              <div v-for="item in syncAudienceRows" :key="item.audience" class="insight-card">
                <strong>{{ item.audience }}</strong>
                <div class="insight-count">{{ item.readiness }}</div>
                <p>{{ item.note }}</p>
              </div>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">同步记录</p>
                <h3>{{ latestInsightExamName }} 同步任务</h3>
              </div>
              <button
                type="button"
                class="solid-btn small"
                :disabled="!canPublishLatestExam"
                @click="publishLatestExamSync"
              >
                发布最近考试
              </button>
            </div>

            <div v-if="publishBlockers.length" class="task-item">
              <strong>发布前置条件</strong>
              <p>{{ publishBlockers.join('；') }}</p>
              <span>补齐后才能创建同步任务</span>
            </div>

            <div class="table-list">
              <div class="table-row table-header four-cols">
                <span>同步对象</span>
                <span>渠道</span>
                <span>状态</span>
                <span>时间</span>
              </div>
              <div v-for="record in syncRecordRows" :key="`${record.target}-${record.time}`" class="table-row four-cols">
                <span>{{ record.target }}</span>
                <span>{{ record.channel }}</span>
                <span>{{ record.status }}</span>
                <span>{{ record.time }}</span>
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">同步策略</p>
                <h3>当前执行规则</h3>
              </div>
            </div>

            <ul class="bullet-list">
              <li>家长和学生共用一个小程序端，身份由绑定关系自动识别</li>
              <li>班主任录入或维护任课老师后，任课老师可自动获得授权范围内的查看权限</li>
              <li>只有成绩分析完成后，系统才允许对家长和学生发起正式同步</li>
            </ul>
          </article>
        </section>
      </template>
    </main>

    <div v-if="actionPanel !== 'none'" class="drawer-backdrop" @click="actionPanel = 'none'"></div>
    <aside v-if="actionPanel !== 'none'" class="drawer-panel">
      <div class="drawer-head">
        <div>
          <p class="panel-label">操作面板</p>
          <h3>
            {{
              actionPanel === 'student-add' ? '新增学生' :
              actionPanel === 'student-import' ? '导入学生档案' :
              actionPanel === 'student-edit' ? '编辑学生' :
              actionPanel === 'teacher-add' ? '新增任课老师' :
              actionPanel === 'teacher-import' ? '导入任课老师' :
              actionPanel === 'teacher-edit' ? '编辑任课老师' :
              actionPanel === 'policy-edit' ? '调整展示策略' :
              actionPanel === 'score-edit' ? '修改成绩' :
              actionPanel === 'score-upload' ? '上传成绩文件' :
              '操作'
            }}
          </h3>
        </div>
        <button type="button" class="ghost-btn tiny" @click="actionPanel = 'none'">关闭</button>
      </div>

      <div v-if="actionPanel === 'student-add' || actionPanel === 'student-edit'" class="form-stack">
        <label>
          <span>学号</span>
          <input v-model="studentForm.studentNo" type="text" />
        </label>
        <label>
          <span>姓名</span>
          <input v-model="studentForm.name" type="text" />
        </label>
        <label>
          <span>性别</span>
          <select v-model="studentForm.gender">
            <option value="男">男</option>
            <option value="女">女</option>
          </select>
        </label>
        <label>
          <span>选科组合</span>
          <input v-model="studentForm.combination" type="text" />
        </label>
        <label>
          <span>家长手机号</span>
          <input v-model="studentForm.parentMobile" type="text" />
        </label>
        <label>
          <span>家长状态</span>
          <select v-model="studentForm.parentStatus">
            <option value="已绑定">已绑定</option>
            <option value="待补充">待补充</option>
          </select>
        </label>
        <label>
          <span>选科状态</span>
          <select v-model="studentForm.selectionStatus">
            <option value="已登记">已登记</option>
            <option value="待确认">待确认</option>
          </select>
        </label>
        <div class="row-actions">
          <button type="button" class="solid-btn small" @click="saveStudentForm">保存学生</button>
          <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
        </div>
      </div>

      <div v-else-if="actionPanel === 'student-import'" class="task-list">
        <input
          ref="studentImportInput"
          type="file"
          accept=".csv"
          class="hidden-file-input"
          @change="handleStudentImportFileChange"
        />
        <div class="task-item">
          <strong>导入模板字段</strong>
          <p>学号、姓名、性别、家长手机号、选科组合。</p>
          <span>当前支持 CSV，学号相同会更新已有学生。</span>
        </div>
        <div class="task-item">
          <strong>已选择文件</strong>
          <p>{{ studentImportFileName || '暂未选择 CSV 文件' }}</p>
          <span>确认后会写入后端并刷新班级档案。</span>
        </div>
        <div class="row-actions">
          <button type="button" class="ghost-btn small" @click="studentImportInput?.click()">选择文件</button>
          <button type="button" class="solid-btn small" :disabled="!studentImportFile" @click="submitStudentImport">确认导入</button>
          <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
        </div>
      </div>

      <div v-else-if="actionPanel === 'teacher-add' || actionPanel === 'teacher-edit'" class="form-stack">
        <label>
          <span>学科</span>
          <input v-model="teacherForm.subject" type="text" />
        </label>
        <label>
          <span>任课老师</span>
          <input v-model="teacherForm.teacher" type="text" />
        </label>
        <label>
          <span>教师手机号</span>
          <input v-model="teacherForm.mobile" type="text" />
        </label>
        <label>
          <span>授课范围</span>
          <input v-model="teacherForm.classes" type="text" />
        </label>
        <div class="task-item">
          <strong>账号与权限状态</strong>
          <p>账号绑定和权限同步由列表中的操作按钮执行，不能在编辑表单中手动改状态。</p>
          <span>手机号是账号登录/联系凭证，账号 ID 保持稳定。</span>
        </div>
        <div class="row-actions">
          <button type="button" class="solid-btn small" @click="saveTeacherForm">保存任课老师</button>
          <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
        </div>
      </div>

      <div v-else-if="actionPanel === 'teacher-import'" class="task-list">
        <input
          ref="teacherImportInput"
          type="file"
          accept=".csv"
          class="hidden-file-input"
          @change="handleTeacherImportFileChange"
        />
        <div class="task-item">
          <strong>导入模板字段</strong>
          <p>学科、老师姓名、授课范围、手机号。</p>
          <span>当前支持 CSV；手机号只作为待绑定凭证，不会自动生成账号。</span>
        </div>
        <div class="task-item">
          <strong>已选择文件</strong>
          <p>{{ teacherImportFileName || '暂未选择 CSV 文件' }}</p>
          <span>导入后仍需点击同步权限，授权才会生效。</span>
        </div>
        <div class="row-actions">
          <button type="button" class="ghost-btn small" @click="teacherImportInput?.click()">选择文件</button>
          <button type="button" class="solid-btn small" :disabled="!teacherImportFile" @click="submitTeacherImport">确认导入</button>
          <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
        </div>
      </div>

      <div v-else-if="actionPanel === 'policy-edit'" class="form-stack">
        <label>
          <span>策略名称</span>
          <input v-model="policyForm.title" type="text" disabled />
        </label>
        <label>
          <span>当前值</span>
          <input v-model="policyForm.value" type="text" />
        </label>
        <label>
          <span>说明</span>
          <textarea v-model="policyForm.note" rows="4"></textarea>
        </label>
        <div class="row-actions">
          <button type="button" class="solid-btn small" @click="savePolicyForm">保存策略</button>
          <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
        </div>
      </div>

      <div v-else-if="actionPanel === 'score-edit'" class="form-stack">
        <label>
          <span>学生</span>
          <input v-model="scoreEditForm.studentName" type="text" disabled />
        </label>
        <label v-for="subject in selectedExamSubjects" :key="subject">
          <span>{{ subject }}</span>
          <input v-model="scoreEditForm.subjectScores[subject]" type="text" />
        </label>
        <label>
          <span>总分</span>
          <input v-model="scoreEditForm.total" type="text" />
        </label>
        <div class="row-actions">
          <button type="button" class="solid-btn small" @click="saveScoreEdit">保存成绩</button>
          <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
        </div>
      </div>

      <div v-else-if="actionPanel === 'score-upload'" class="form-stack">
        <div class="step-pills">
          <span :class="{ active: scoreUploadStep === 'meta' }">1. 考试信息</span>
          <span :class="{ active: scoreUploadStep === 'upload' }">2. 上传文件</span>
          <span :class="{ active: scoreUploadStep === 'validate' }">3. 校验预览</span>
          <span :class="{ active: scoreUploadStep === 'confirm' }">4. 确认导入</span>
        </div>

        <div v-if="scoreUploadStep === 'meta'" class="form-stack">
          <label>
            <span>考试名称</span>
            <input v-model="examForm.name" type="text" />
          </label>
          <label>
            <span>考试类型</span>
            <select v-model="examForm.type">
              <option value="周测">周测</option>
              <option value="月考">月考</option>
              <option value="期中">期中</option>
              <option value="期末">期末</option>
            </select>
          </label>
          <label>
            <span>考试日期</span>
            <input v-model="examForm.date" type="date" />
          </label>
          <div class="subject-picker">
            <span>本次考试覆盖学科</span>
            <div class="subject-grid">
              <label class="subject-chip">
                <input v-model="examForm.subjects" type="checkbox" value="语文" />
                <span>语文</span>
              </label>
              <label class="subject-chip">
                <input v-model="examForm.subjects" type="checkbox" value="数学" />
                <span>数学</span>
              </label>
              <label class="subject-chip">
                <input v-model="examForm.subjects" type="checkbox" value="英语" />
                <span>英语</span>
              </label>
              <template v-if="activeSelectionScenario.id === 'post-selection'">
                <label class="subject-chip">
                  <input v-model="examForm.subjects" type="checkbox" value="物理" />
                  <span>物理</span>
                </label>
                <label class="subject-chip">
                  <input v-model="examForm.subjects" type="checkbox" value="化学" />
                  <span>化学</span>
                </label>
                <label class="subject-chip">
                  <input v-model="examForm.subjects" type="checkbox" value="生物" />
                  <span>生物</span>
                </label>
                <label class="subject-chip">
                  <input v-model="examForm.subjects" type="checkbox" value="历史" />
                  <span>历史</span>
                </label>
                <label class="subject-chip">
                  <input v-model="examForm.subjects" type="checkbox" value="地理" />
                  <span>地理</span>
                </label>
                <label class="subject-chip">
                  <input v-model="examForm.subjects" type="checkbox" value="政治" />
                  <span>政治</span>
                </label>
              </template>
            </div>
          </div>
          <div class="row-actions">
            <button type="button" class="solid-btn small" @click="scoreUploadStep = 'upload'">下一步</button>
          </div>
        </div>

        <div v-else-if="scoreUploadStep === 'upload'" class="form-stack">
          <input
            ref="scoreFileInput"
            type="file"
            accept=".xlsx,.xls,.csv"
            class="hidden-file-input"
            @change="handleScoreFileChange"
          />
          <label>
            <span>成绩文件</span>
            <div class="file-picker">
              <button type="button" class="ghost-btn small" @click="triggerScoreFilePick">选择文件</button>
              <span>{{ scoreUploadForm.fileName || '暂未选择文件' }}</span>
            </div>
          </label>
          <div class="task-item">
            <strong>当前考试</strong>
            <p>{{ examForm.name || '未填写考试名称' }} / {{ examForm.type }}</p>
            <span>{{ examForm.subjects.join(' / ') || '未选择学科' }}</span>
          </div>
          <div class="row-actions">
            <button type="button" class="ghost-btn small" @click="scoreUploadStep = 'meta'">上一步</button>
            <button
              type="button"
              class="solid-btn small"
              :disabled="!scoreUploadForm.fileSelected"
              @click="runImportValidation(); scoreUploadStep = 'validate'"
            >
              下一步
            </button>
          </div>
        </div>

        <div v-else-if="scoreUploadStep === 'validate'" class="task-list">
          <div v-if="uploadValidationState.error" class="task-item">
            <strong>校验失败</strong>
            <p>{{ uploadValidationState.error }}</p>
            <span>请检查文件格式、表头和编码后重新上传</span>
          </div>
          <div v-else-if="uploadValidationState.loading" class="task-item">
            <strong>正在校验</strong>
            <p>系统正在解析文件并检查学号、学科、字段匹配情况。</p>
            <span>请稍候</span>
          </div>
          <template v-else>
            <div v-for="item in uploadValidationState.summary" :key="item.field" class="task-item">
              <strong>{{ item.field }}</strong>
              <p>{{ item.note }}</p>
              <span>{{ item.result }}</span>
            </div>

            <div class="task-item">
              <strong>识别到的字段</strong>
              <p>系统已从文件中识别以下表头：</p>
              <span>{{ uploadValidationState.headers.join(' / ') }}</span>
            </div>

            <div class="panel panel-embedded">
              <div class="panel-head">
                <div>
                  <p class="panel-label">识别结果预览</p>
                  <h3>前 5 行数据</h3>
                </div>
              </div>

              <div class="table-list">
                <div class="table-row table-header preview-cols">
                  <span>学号</span>
                  <span>姓名</span>
                  <span v-for="subject in examForm.subjects" :key="subject">{{ subject }}</span>
                </div>
                <div
                  v-for="(row, index) in uploadValidationState.previewRows"
                  :key="`${row['学号']}-${index}`"
                  class="table-row preview-cols"
                >
                  <span>{{ row['学号'] || '-' }}</span>
                  <span>{{ row['姓名'] || '-' }}</span>
                  <span v-for="subject in examForm.subjects" :key="subject">{{ row[subject] || '-' }}</span>
                </div>
              </div>
            </div>
          </template>
          <div class="row-actions">
            <button type="button" class="ghost-btn small" @click="scoreUploadStep = 'upload'">上一步</button>
            <button type="button" class="solid-btn small" @click="scoreUploadStep = 'confirm'">进入确认</button>
          </div>
        </div>

        <div v-else class="task-list">
          <div class="task-item">
            <strong>考试信息</strong>
            <p>{{ examForm.name || '未填写考试名称' }} / {{ examForm.type }} / {{ examForm.date }}</p>
            <span>{{ examForm.subjects.join(' / ') || '未选择学科' }}</span>
          </div>
          <div class="task-item">
            <strong>导入文件</strong>
            <p>{{ scoreUploadForm.fileName }}</p>
            <span>确认导入后将生成新的考试成绩记录</span>
          </div>
          <div class="row-actions">
            <button type="button" class="ghost-btn small" @click="scoreUploadStep = 'validate'">上一步</button>
            <button type="button" class="solid-btn small" @click="saveScoreUpload">确认导入</button>
            <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
          </div>
        </div>
      </div>
    </aside>
  </div>
</template>

<style>
:root {
  color-scheme: light;
  --bg: #f5f1ea;
  --panel: #fffdf9;
  --panel-alt: #f8f2e8;
  --ink: #201911;
  --muted: #736350;
  --line: rgba(74, 52, 24, 0.12);
  --accent: #9a5b26;
  --accent-soft: #efe0cb;
  --shadow: 0 24px 60px rgba(70, 49, 23, 0.12);
  --radius-xl: 28px;
  --radius-lg: 20px;
  --radius-md: 16px;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  min-height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.9), transparent 28%),
    linear-gradient(180deg, #faf6ef 0%, #f3ece2 100%);
  color: var(--ink);
  font-family: "Avenir Next", "PingFang SC", "Noto Sans SC", sans-serif;
}

button,
a,
input {
  font: inherit;
}

button {
  cursor: pointer;
}

a {
  color: var(--accent);
  text-decoration: none;
}

#app {
  min-height: 100vh;
}

.teacher-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
}

.sidebar {
  padding: 28px 20px;
  background: rgba(247, 241, 233, 0.86);
  border-right: 1px solid var(--line);
}

.brand {
  padding: 8px 12px 22px;
}

.brand-kicker,
.topbar-label,
.panel-label {
  margin: 0;
  color: var(--accent);
  letter-spacing: 0.14em;
  text-transform: uppercase;
  font-size: 0.76rem;
  font-weight: 700;
}

.brand h1,
.topbar h2,
.panel-head h3 {
  margin: 10px 0 0;
}

.brand h1 {
  font-size: 1.7rem;
  line-height: 1.1;
}

.brand-copy {
  margin: 10px 0 0;
  color: var(--muted);
}

.nav-list {
  display: grid;
  gap: 10px;
}

.nav-item {
  width: 100%;
  padding: 14px;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  text-align: left;
}

.nav-item strong,
.nav-item span {
  display: block;
}

.nav-item span {
  margin-top: 6px;
  color: var(--muted);
  font-size: 0.9rem;
}

.nav-item.active {
  background: var(--panel);
  border-color: var(--line);
  box-shadow: var(--shadow);
}

.main-area {
  padding: 28px;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.topbar h2 {
  font-size: 2rem;
  letter-spacing: -0.04em;
}

.topbar-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.stage-banner {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  margin-top: 18px;
  padding: 18px 20px;
  border-radius: var(--radius-lg);
  background: var(--panel);
  border: 1px solid var(--line);
  box-shadow: var(--shadow);
}

.stage-banner h3 {
  margin: 10px 0 0;
  font-size: 1.18rem;
}

.stage-banner p:last-child {
  margin: 0;
  max-width: 560px;
  color: var(--muted);
  line-height: 1.7;
}

.topbar-select {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--panel);
  border: 1px solid var(--line);
}

.topbar-select span {
  color: var(--muted);
  font-size: 0.84rem;
}

.topbar-select select {
  border: 0;
  background: transparent;
  color: var(--ink);
  outline: none;
  min-width: 180px;
}

.solid-btn,
.ghost-btn,
.tab-btn {
  border-radius: 999px;
  padding: 11px 18px;
  border: 1px solid transparent;
}

.solid-btn {
  background: var(--ink);
  color: #f9f4ea;
}

.ghost-btn {
  background: var(--panel);
  color: var(--ink);
  border-color: var(--line);
}

.ghost-btn.small,
.solid-btn.small {
  padding: 8px 14px;
  font-size: 0.92rem;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-top: 22px;
}

.compact-metrics {
  margin-top: 0;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.metric-card,
.panel {
  background: var(--panel);
  border: 1px solid var(--line);
  box-shadow: var(--shadow);
}

.metric-card {
  padding: 20px;
  border-radius: var(--radius-lg);
}

.metric-card p,
.metric-card span,
.table-row p,
.issue-card p,
.issue-card span,
.alert-item p,
.trend-meta p,
.bullet-list li {
  margin: 0;
  color: var(--muted);
}

.metric-card strong {
  display: block;
  margin: 12px 0 8px;
  font-size: 2rem;
  letter-spacing: -0.04em;
}

.content-grid {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: 16px;
  margin-top: 18px;
}

.panel {
  padding: 20px;
  border-radius: var(--radius-lg);
}

.panel-wide {
  grid-column: 1 / -1;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.panel-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.panel-head h3 {
  font-size: 1.35rem;
}

.inline-note {
  color: var(--accent);
  font-size: 0.9rem;
  font-weight: 600;
}

.trend-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.flow-grid,
.insight-grid {
  display: grid;
  gap: 12px;
}

.flow-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.insight-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.compact-insights {
  margin-bottom: 14px;
}

.base-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}

.base-item {
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--panel-alt);
}

.base-item span {
  display: block;
  color: var(--muted);
  font-size: 0.84rem;
}

.base-item strong {
  display: block;
  margin-top: 10px;
  font-size: 1rem;
}

.decision-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.trend-card,
.flow-card,
.insight-card,
.decision-card,
.alert-item,
.issue-card,
.table-row {
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--panel-alt);
}

.flow-card span,
.flow-card p,
.insight-card p,
.decision-card p {
  color: var(--muted);
}

.flow-card p,
.insight-card p,
.decision-card p {
  margin: 10px 0 0;
  line-height: 1.6;
}

.flow-card span {
  display: inline-block;
  margin-top: 12px;
  color: var(--accent);
  font-size: 0.86rem;
  font-weight: 700;
}

.insight-count {
  margin-top: 14px;
  font-size: 1.8rem;
  font-weight: 700;
  letter-spacing: -0.04em;
}

.decision-result {
  margin-top: 14px;
  color: var(--accent);
  font-size: 1.05rem;
  font-weight: 700;
}

.trend-score {
  margin-top: 14px;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: -0.04em;
}

.trend-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 10px;
}

.trend-meta span {
  color: var(--accent);
  font-weight: 700;
}

.table-list,
.alert-list,
.issue-list,
.alert-grid {
  display: grid;
  gap: 10px;
}

.table-row {
  display: grid;
  gap: 12px;
  align-items: center;
}

.clickable-row {
  width: 100%;
  text-align: left;
  border: 0;
}

.clickable-row.selected {
  background: #fff5e8;
  box-shadow: inset 0 0 0 1px rgba(154, 91, 38, 0.22);
}

.table-header {
  color: var(--muted);
  font-size: 0.9rem;
}

.four-cols {
  grid-template-columns: 1.5fr 1fr 1fr 1fr;
}

.five-cols {
  grid-template-columns: 1.2fr 1fr 0.8fr 1fr 1fr;
}

.six-cols {
  grid-template-columns: 1fr 0.9fr 1fr 1fr 1fr 1.1fr;
}

.score-cols {
  grid-template-columns: 1.2fr repeat(5, 0.8fr) 1fr;
}

.score-cols-pre {
  grid-template-columns: 1.3fr repeat(4, 0.9fr) 1fr;
}

.score-cols-post {
  grid-template-columns: 1.2fr repeat(3, 0.75fr) 1fr 0.9fr 0.8fr 1fr;
}

.preview-cols {
  grid-template-columns: 1fr 1fr repeat(6, 0.8fr);
}

.score-table-wrap {
  padding: 16px;
  border-radius: var(--radius-md);
  background: #fcf7ef;
  border: 1px solid rgba(74, 52, 24, 0.08);
}

.score-table-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.score-table-title span {
  color: var(--muted);
  font-size: 0.88rem;
}

.score-table-list {
  overflow-x: auto;
}

.issue-card span {
  display: inline-block;
  margin-top: 10px;
}

.issue-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.issue-state {
  color: var(--accent);
  font-size: 0.84rem;
  font-weight: 700;
}

.issue-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.row-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.top-gap {
  margin-top: 10px;
}

.form-stack {
  display: grid;
  gap: 12px;
}

.form-stack label {
  display: grid;
  gap: 6px;
}

.subject-picker {
  display: grid;
  gap: 10px;
}

.subject-picker > span {
  color: var(--muted);
  font-size: 0.86rem;
}

.subject-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.subject-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--panel-alt);
  border: 1px solid var(--line);
}

.step-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 16px;
}

.step-pills span {
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--panel-alt);
  color: var(--muted);
  border: 1px solid var(--line);
}

.step-pills span.active {
  background: #fff3e4;
  color: var(--accent);
  border-color: rgba(154, 91, 38, 0.24);
}

.hidden-file-input {
  display: none;
}

.file-picker {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--panel-alt);
  border: 1px solid var(--line);
}

.file-picker span {
  color: var(--muted);
  font-size: 0.9rem;
}

.form-stack label span {
  color: var(--muted);
  font-size: 0.86rem;
}

.form-stack input,
.form-stack select,
.form-stack textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: 12px;
  background: var(--panel-alt);
  color: var(--ink);
  outline: none;
}

.empty-tip {
  margin: 0;
  color: var(--muted);
  line-height: 1.7;
}

.drawer-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(32, 25, 17, 0.18);
}

.drawer-panel {
  position: fixed;
  top: 0;
  right: 0;
  width: min(720px, 100%);
  height: 100vh;
  padding: 22px;
  overflow-y: auto;
  background: var(--panel);
  border-left: 1px solid var(--line);
  box-shadow: var(--shadow);
  z-index: 10;
}

.drawer-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.drawer-head h3 {
  margin: 10px 0 0;
  font-size: 1.35rem;
}

.panel-embedded {
  padding: 16px;
  box-shadow: none;
}

.ghost-btn.tiny {
  padding: 6px 10px;
  font-size: 0.82rem;
}

.status-inline {
  margin-bottom: 14px;
  padding: 12px 14px;
  border-radius: 14px;
  background: #fcf7ef;
  color: var(--muted);
  border: 1px solid rgba(154, 91, 38, 0.12);
  line-height: 1.6;
}

.alert-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.alert-top span {
  padding: 4px 8px;
  border-radius: 999px;
  background: var(--accent-soft);
  color: var(--accent);
  font-size: 0.8rem;
  font-weight: 700;
}

.bullet-list {
  padding-left: 20px;
  margin: 0;
  display: grid;
  gap: 10px;
}

@media (max-width: 1200px) {
  .teacher-shell {
    grid-template-columns: 1fr;
  }

  .metrics-grid,
  .base-grid,
  .trend-grid,
  .flow-grid,
  .insight-grid,
  .decision-grid,
  .content-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .compact-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .panel-wide {
    grid-column: auto;
  }
}

@media (max-width: 760px) {
  .main-area,
  .sidebar {
    padding: 18px;
  }

  .topbar,
  .stage-banner,
  .metrics-grid,
  .base-grid,
  .trend-grid,
  .flow-grid,
  .insight-grid,
  .decision-grid,
  .compact-metrics,
  .content-grid,
  .table-row,
  .four-cols,
  .five-cols,
  .six-cols,
  .score-cols,
  .score-cols-pre,
  .score-cols-post,
  .preview-cols {
    grid-template-columns: 1fr;
  }

  .subject-grid {
    grid-template-columns: 1fr;
  }

  .topbar {
    display: grid;
  }

  .topbar-actions,
  .panel-head,
  .panel-actions {
    flex-wrap: wrap;
  }
}
</style>
