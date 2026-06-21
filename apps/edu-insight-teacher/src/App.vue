<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
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
  createClass,
  createStudent,
  createTeacher,
  deleteStudent,
  deleteTeacher,
  fetchClassroomWorkspace,
  importStudents,
  importTeachers,
  syncTeacherPermission,
  updatePolicy,
  updateStudent,
  updateTeacher,
  type ClassOption,
  type ClassroomWorkspace
} from './api/classroom'
import {
  fetchCurrentUser,
  loginWithSMS,
  sendSMSCode,
  type CurrentUser,
  type WorkIdentity
} from './api/auth'
import { fetchExamDetail, fetchExams, importExam, updateExamScore } from './api/score'
import { fetchInsightDashboard, publishLatestExam, type AnalysisDashboard } from './api/insight'
import { parseAndValidateCsv } from './utils/csvImport'
import AppIcon from './components/icons/AppIcon.vue'
import type { IconName } from './components/icons/iconRegistry'
import HorizontalBarChart, { type BarChartRow } from './components/charts/HorizontalBarChart.vue'
import DistributionBars, { type DistributionRow } from './components/charts/DistributionBars.vue'
import LineChart, { type ChartPoint } from './components/charts/LineChart.vue'

const AUTH_TOKEN_KEY = 'edu-insight-teacher-token'

const activeNav = ref('overview')
const activeContextId = ref(workContexts[0].id)
const scopeMode = ref<'single' | 'overall'>('single')
const analysisMode = ref<'report' | 'transcript' | 'distribution' | 'compare' | 'trend' | 'student' | 'warning'>('report')
const overallAnalysisView = ref<'merged' | 'compare'>('merged')
const activeSubjectClassId = ref(subjectScopeClasses[0].id)
const availableWorkContexts = ref(structuredClone(workContexts))
const availableSubjectClasses = ref(structuredClone(subjectScopeClasses))
const managedClassOptions = ref<ClassOption[]>([])
const activeHomeroomClassId = ref(subjectScopeClasses[0].id)
const selectedExamId = ref(examRecords[0].id)
const classPanelMode = ref<'student' | 'teacher' | 'policy'>('student')
const selectedStudentId = ref('')
const selectedTeacherId = ref(teacherAssignments[0].id)
const selectedPolicyId = ref(displayPolicies[0].id)
const selectedAnalysisStudentId = ref('')
const actionPanel = ref<
  | 'none'
  | 'class-create'
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
const analysisDashboard = ref<AnalysisDashboard | null>(null)
const studentImportInput = ref<HTMLInputElement | null>(null)
const teacherImportInput = ref<HTMLInputElement | null>(null)
const studentImportFile = ref<File | null>(null)
const teacherImportFile = ref<File | null>(null)
const studentImportFileName = ref('')
const teacherImportFileName = ref('')
const authLoading = ref(true)
const authSubmitting = ref(false)
const authError = ref('')
const authMessage = ref('')
const authToken = ref(localStorage.getItem(AUTH_TOKEN_KEY) ?? '')
const currentUser = ref<CurrentUser | null>(null)
const loginForm = ref({
  mobile: '13800001003',
  code: ''
})

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

const classForm = ref({
  schoolName: '星河高级中学',
  gradeName: '高一',
  className: '',
  homeroomTeacher: '李老师',
  academicYear: '2025-2026 学年',
  stageId: 'pre-selection'
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

const navIconMap: Record<string, IconName> = {
  overview: 'dashboard',
  classes: 'classroom',
  scores: 'scores',
  analysis: 'analysis',
  sync: 'sync'
}

const analysisModeOptions: Array<{
  id: 'report' | 'transcript' | 'distribution' | 'compare' | 'trend' | 'student' | 'warning'
  label: string
  description: string
  icon: IconName
}> = [
  { id: 'report', label: '考试报告', description: '单次考试总览与诊断结论', icon: 'analysis' },
  { id: 'transcript', label: '成绩单', description: '学生成绩、排名和薄弱科目', icon: 'scores' },
  { id: 'distribution', label: '水平分布', description: '分数段、分层和离散度', icon: 'barChart' },
  { id: 'compare', label: '班级对比', description: '单班与整体下的横向比较', icon: 'layer' },
  { id: 'trend', label: '多次趋势', description: '历次考试和学科趋势', icon: 'lineChart' },
  { id: 'student', label: '学生分析', description: '单个学生的长期变化', icon: 'student' },
  { id: 'warning', label: '预警建议', description: '重点学生和教学动作', icon: 'warning' }
]

function setAnalysisMode(mode: typeof analysisMode.value) {
  analysisMode.value = mode
  if (mode === 'report') overallAnalysisView.value = 'merged'
  if (mode === 'compare') overallAnalysisView.value = 'compare'
}

const activeTitle = computed(() => {
  const current = navItems.find((item) => item.id === activeNav.value)
  return current?.label ?? '工作台'
})

const activeContext = computed(() => {
  return availableWorkContexts.value.find((item) => item.id === activeContextId.value) ?? availableWorkContexts.value[0]
})

const isSubjectTeacherView = computed(() => activeContext.value.roleLabel === '任课老师')

const activeSubjectClass = computed(() => {
  return availableSubjectClasses.value.find((item) => item.id === activeSubjectClassId.value) ?? availableSubjectClasses.value[0]
})

const activeHomeroomClass = computed(() => {
  return availableSubjectClasses.value.find((item) => item.id === activeHomeroomClassId.value) ?? availableSubjectClasses.value[0]
})

const activeScopeLabel = computed(() => {
  if (scopeMode.value === 'overall') {
    return isSubjectTeacherView.value ? '任课班级整体' : '所带班级整体'
  }
  return isSubjectTeacherView.value ? activeSubjectClass.value?.label : activeHomeroomClass.value?.label
})

const currentClassId = computed(() => {
  if (isSubjectTeacherView.value) {
    return activeSubjectClassId.value
  }
  return activeHomeroomClassId.value
})

const isOverallScope = computed(() => scopeMode.value === 'overall')

const canMaintainClassData = computed(() => !isOverallScope.value)

const insightScopeClassIds = computed(() => {
  if (!isOverallScope.value) return [currentClassId.value]
  return availableSubjectClasses.value.map((item) => item.id)
})

const insightScopePayload = computed(() => ({
  mode: scopeMode.value,
  classId: currentClassId.value,
  classIds: insightScopeClassIds.value,
  examId: isOverallScope.value ? undefined : selectedExamId.value
}))

const analysisScopeNote = computed(() => {
  if (!isOverallScope.value) {
    return `当前分析数据源：${activeScopeLabel.value}`
  }
  return `当前分析数据源：${activeScopeLabel.value}，聚合 ${insightScopeClassIds.value.length} 个班级`
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

const isAuthenticated = computed(() => Boolean(currentUser.value && authToken.value))

const currentUserName = computed(() => currentUser.value?.user.name ?? '未登录')

const currentUserMobile = computed(() => currentUser.value?.user.mobile ?? '')

const selectedExamSubjects = computed(() => {
  const subjects: string[] = selectedExam.value?.subjects ?? []
  if (subjects.length > 0) return subjects
  return ['语文', '数学', '英语']
})

const backendAnalysis = computed(() => analysisDashboard.value)

const numericScoreRows = computed(() => {
  return scoreRows.value
    .map((row) => ({
      ...row,
      totalValue: parseScoreValue(row.total)
    }))
    .filter((row) => row.totalValue > 0)
})

const totalScores = computed(() => numericScoreRows.value.map((row) => row.totalValue).sort((a, b) => b - a))

const classAverage = computed(() => average(totalScores.value))

const classMedian = computed(() => {
  const values = [...totalScores.value].sort((a, b) => a - b)
  if (values.length === 0) return 0
  const middle = Math.floor(values.length / 2)
  return values.length % 2 === 0 ? (values[middle - 1] + values[middle]) / 2 : values[middle]
})

const classHighest = computed(() => totalScores.value[0] ?? 0)

const classLowest = computed(() => totalScores.value[totalScores.value.length - 1] ?? 0)

const scoreSpread = computed(() => Math.max(0, classHighest.value - classLowest.value))

const excellenceLine = computed(() => classAverage.value + 30)

const riskLine = computed(() => Math.max(0, classAverage.value - 30))

const analysisMetrics = computed(() => {
  if (backendAnalysis.value?.examMetrics?.length) return backendAnalysis.value.examMetrics
  const rows = numericScoreRows.value
  const highCount = rows.filter((row) => row.totalValue >= excellenceLine.value).length
  const riskCount = rows.filter((row) => row.totalValue < riskLine.value).length
  return [
    {
      label: '班级均分',
      value: formatNumber(classAverage.value),
      note: `${selectedExam.value?.name ?? '当前考试'} / ${rows.length} 人`
    },
    {
      label: '中位分',
      value: formatNumber(classMedian.value),
      note: `最高 ${formatNumber(classHighest.value)} / 最低 ${formatNumber(classLowest.value)}`
    },
    {
      label: '高分层',
      value: `${highCount} 人`,
      note: `高于均分 30 分以上`
    },
    {
      label: '预警层',
      value: `${riskCount} 人`,
      note: `低于均分 30 分以上`
    }
  ]
})

const subjectDiagnostics = computed(() => {
  if (backendAnalysis.value?.subjectDiagnostics?.length) return backendAnalysis.value.subjectDiagnostics.map((item) => ({
    subject: item.subject,
    average: item.average,
    max: item.highest,
    min: item.lowest,
    weakCount: item.lowCount,
    excellentCount: item.excellentCount,
    stability: 0,
    barWidth: item.barWidth,
    riskLabel: item.riskLabel
  }))
  return selectedExamSubjects.value.map((subject) => {
    const values = numericScoreRows.value
      .map((row) => parseScoreValue(row.subjectScores?.[subject]))
      .filter((value) => value > 0)
    const avg = average(values)
    const max = Math.max(...values, 0)
    const min = values.length > 0 ? Math.min(...values) : 0
    const weakCount = values.filter((value) => value < 60).length
    const excellentCount = values.filter((value) => value >= 90).length
    const stability = Math.max(0, 100 - (max - min))
    return {
      subject,
      average: avg,
      max,
      min,
      weakCount,
      excellentCount,
      stability,
      barWidth: clamp((avg / Math.max(max, 150)) * 100, 8, 100),
      riskLabel: weakCount > 0 ? `${weakCount} 人低于 60` : '暂无低分预警'
    }
  })
})

const subjectDiagnosticBars = computed<BarChartRow[]>(() => {
  const maxScore = Math.max(...subjectDiagnostics.value.map((item) => item.max), 100)
  return subjectDiagnostics.value.map((item) => ({
    label: item.subject,
    value: item.average,
    max: maxScore,
    meta: `均分 ${formatNumber(item.average)}`,
    note: `最高 ${formatNumber(item.max)} / 最低 ${formatNumber(item.min)} · ${item.riskLabel} · 优秀 ${item.excellentCount} 人`,
    tone: item.weakCount > 0 ? 'risk' : 'steady'
  }))
})

const scoreBandsView = computed(() => {
  if (backendAnalysis.value?.scoreBands?.length) {
    return backendAnalysis.value.scoreBands.map((band) => ({
      label: band.label,
      range: band.range,
      count: band.count,
      percent: band.percent,
      width: band.width,
      tone: band.tone
    }))
  }
  const bands = [
    { label: '高分突破', range: `≥ ${formatNumber(excellenceLine.value)}`, count: 0, tone: 'strong' },
    { label: '稳定中坚', range: `${formatNumber(riskLine.value)} - ${formatNumber(excellenceLine.value)}`, count: 0, tone: 'steady' },
    { label: '重点帮扶', range: `< ${formatNumber(riskLine.value)}`, count: 0, tone: 'risk' }
  ]
  for (const row of numericScoreRows.value) {
    if (row.totalValue >= excellenceLine.value) bands[0].count += 1
    else if (row.totalValue < riskLine.value) bands[2].count += 1
    else bands[1].count += 1
  }
  const total = Math.max(numericScoreRows.value.length, 1)
  return bands.map((band) => ({
    ...band,
    percent: Math.round((band.count / total) * 100),
    width: Math.max(8, Math.round((band.count / total) * 100))
  }))
})

const distributionRows = computed<DistributionRow[]>(() => scoreBandsView.value.map((band) => ({
  label: band.label,
  range: band.range,
  count: band.count,
  percent: band.percent,
  width: band.width,
  tone: band.tone
})))

const layeredStudentGroups = computed(() => {
  if (backendAnalysis.value?.layerGroups?.length) return backendAnalysis.value.layerGroups
  const sorted = [...numericScoreRows.value].sort((a, b) => b.totalValue - a.totalValue)
  const high = sorted.filter((row) => row.totalValue >= excellenceLine.value)
  const middle = sorted.filter((row) => row.totalValue < excellenceLine.value && row.totalValue >= riskLine.value)
  const risk = sorted.filter((row) => row.totalValue < riskLine.value)
  return [
    {
      title: '高分突破组',
      count: `${high.length} 人`,
      students: namesOf(high),
      goal: '保持优势科，安排压轴题和高阶迁移训练。'
    },
    {
      title: '临界提升组',
      count: `${middle.length} 人`,
      students: namesOf(middle),
      goal: '找出最短板学科，优先做 2 周小目标提升。'
    },
    {
      title: '重点帮扶组',
      count: `${risk.length} 人`,
      students: namesOf(risk),
      goal: '班主任、单科老师、家长同步跟进，先止跌再补弱。'
    }
  ]
})

const riskStudentRows = computed(() => {
  if (backendAnalysis.value?.riskStudents?.length) return backendAnalysis.value.riskStudents
  return [...numericScoreRows.value]
    .map((row) => {
      const weakSubjects = selectedExamSubjects.value
        .map((subject) => ({ subject, value: parseScoreValue(row.subjectScores?.[subject]) }))
        .filter((item) => item.value > 0 && item.value < 60)
      const gap = classAverage.value - row.totalValue
      let level = '观察'
      if (gap >= 45 || weakSubjects.length >= 2) level = '高'
      else if (gap >= 20 || weakSubjects.length === 1) level = '中'
      return {
        name: row.studentName,
        total: row.total,
        gap,
        weakSubjects: weakSubjects.map((item) => item.subject).join('、') || '暂无明显低分科',
        level,
        reason: gap > 0 ? `低于均分 ${formatNumber(gap)} 分` : `高于均分 ${formatNumber(Math.abs(gap))} 分`
      }
    })
    .filter((row) => row.level !== '观察')
    .sort((a, b) => b.gap - a.gap)
})

const teachingActions = computed(() => {
  if (backendAnalysis.value?.teachingActions?.length) return backendAnalysis.value.teachingActions
  const weakestSubject = [...subjectDiagnostics.value].sort((a, b) => a.average - b.average)[0]
  const mostRisk = riskStudentRows.value[0]
  return [
    {
      title: '先抓班级最弱学科',
      detail: weakestSubject
        ? `${weakestSubject.subject} 均分 ${formatNumber(weakestSubject.average)}，${weakestSubject.riskLabel}，建议先做错因归类和小测回收。`
        : '暂无学科成绩数据。',
      tag: weakestSubject ? weakestSubject.subject : '待导入'
    },
    {
      title: '建立动态小组',
      detail: `按高分突破、临界提升、重点帮扶三组推进，当前重点帮扶 ${scoreBandsView.value[2]?.count ?? 0} 人。`,
      tag: '分层教学'
    },
    {
      title: '优先跟进高风险学生',
      detail: mostRisk
        ? `${mostRisk.name} ${mostRisk.reason}，薄弱点：${mostRisk.weakSubjects}。`
        : '当前没有明显高风险学生。',
      tag: mostRisk?.level ?? '稳定'
    },
    {
      title: '下次考试验证目标',
      detail: `建议把班级均分提升 ${Math.max(3, Math.round(scoreSpread.value * 0.04))} 分、重点帮扶组减少 1 人作为下一次阶段目标。`,
      tag: '闭环验证'
    }
  ]
})

const classTrendPoints = computed(() => backendAnalysis.value?.classTrend ?? [])

const classTrendDelta = computed(() => {
  const points = classTrendPoints.value
  if (points.length < 2) return 0
  const latest = points[points.length - 1]?.value ?? 0
  const previous = points[points.length - 2]?.value ?? latest
  return latest - previous
})

const analysisReportContext = computed(() => [
  {
    label: '考试',
    value: selectedExam.value?.name || latestInsightExamName.value || '暂无考试'
  },
  {
    label: '范围',
    value: isOverallScope.value ? `${activeScopeLabel.value} · ${insightScopeClassIds.value.length} 个班` : activeScopeLabel.value
  },
  {
    label: '学生',
    value: `${numericScoreRows.value.length} 人`
  },
  {
    label: '学科',
    value: selectedExamSubjects.value.join('、')
  }
])

const analysisReportSummary = computed(() => {
  const weakestSubject = [...subjectDiagnostics.value].sort((a, b) => a.average - b.average)[0]
  const strongestBand = [...scoreBandsView.value].sort((a, b) => b.count - a.count)[0]
  const delta = classTrendDelta.value
  const trendText = classTrendPoints.value.length < 2
    ? '历史考试不足，暂不判断趋势'
    : `较上次考试${delta >= 0 ? '提升' : '下降'} ${formatNumber(Math.abs(delta))} 分`

  return [
    `${analysisReportContext.value[0].value}覆盖 ${numericScoreRows.value.length} 名学生，当前均分 ${formatNumber(classAverage.value)}，中位分 ${formatNumber(classMedian.value)}。`,
    `本次最高分 ${formatNumber(classHighest.value)}，最低分 ${formatNumber(classLowest.value)}，分差 ${formatNumber(scoreSpread.value)}，${trendText}。`,
    weakestSubject
      ? `${weakestSubject.subject} 是当前最需要关注的学科，均分 ${formatNumber(weakestSubject.average)}，${weakestSubject.riskLabel}。`
      : '当前暂无可用学科诊断数据。',
    strongestBand
      ? `分层结构中「${strongestBand.label}」人数最多，为 ${strongestBand.count} 人，占比 ${strongestBand.percent}%。`
      : '当前暂无可用分层分布数据。',
    riskStudentRows.value.length > 0
      ? `重点预警 ${riskStudentRows.value.length} 人，建议优先从薄弱科目和连续波动学生开始跟进。`
      : '当前没有明显预警学生，可重点关注临界提升与高分突破。'
  ]
})

const subjectTrendSeries = computed(() => backendAnalysis.value?.subjectTrends ?? [])

const studentAnalysisRows = computed(() => backendAnalysis.value?.studentAnalyses ?? [])

const classTrendChartPoints = computed<ChartPoint[]>(() => classTrendPoints.value.map((point) => ({
  label: point.examName.replace('2026 年', '').replace('考试', ''),
  value: point.value,
  note: point.date
})))

const subjectTrendBars = computed<BarChartRow[]>(() => subjectTrendSeries.value.slice(0, 6).map((series) => {
  const latest = series.points[series.points.length - 1]?.value ?? 0
  const first = series.points[0]?.value ?? latest
  const delta = latest - first
  return {
    label: series.subject,
    value: latest,
    max: Math.max(...series.points.map((point) => point.value), 100),
    meta: `${formatNumber(first)} → ${formatNumber(latest)}`,
    note: `较首场${delta >= 0 ? '提升' : '下降'} ${formatNumber(Math.abs(delta))} 分`,
    tone: delta >= 0 ? 'strong' : 'risk'
  }
}))

const selectedAnalysisStudent = computed(() => {
  return studentAnalysisRows.value.find((item) => item.studentId === selectedAnalysisStudentId.value)
    ?? studentAnalysisRows.value[0]
})

const selectedAnalysisStudentSubjectTrends = computed(() => selectedAnalysisStudent.value?.subjectTrends ?? [])

const selectedStudentTotalTrendPoints = computed<ChartPoint[]>(() => (selectedAnalysisStudent.value?.totalTrend ?? []).map((point) => ({
  label: point.examName.replace('2026 年', '').replace('考试', ''),
  value: point.value,
  note: point.date
})))

const selectedStudentRankTrendPoints = computed<ChartPoint[]>(() => (selectedAnalysisStudent.value?.rankTrend ?? []).map((point) => ({
  label: point.examName.replace('2026 年', '').replace('考试', ''),
  value: point.value,
  note: point.date
})))

const classComparisonRows = computed(() => backendAnalysis.value?.classComparisons ?? [])

const classComparisonBars = computed<BarChartRow[]>(() => classComparisonRows.value.map((item) => ({
  label: item.className,
  value: item.average,
  max: Math.max(...classComparisonRows.value.map((row) => row.average), 1),
  meta: `均分 ${formatNumber(item.average)}`,
  note: `${item.examName} · ${item.studentCount} 人 · 最高 ${formatNumber(item.highest)} / 最低 ${formatNumber(item.lowest)} · 预警 ${item.riskCount} 人`,
  tone: item.riskCount > 0 ? 'risk' : 'strong'
})))

const selectedStudentSubjectBars = computed<BarChartRow[]>(() => selectedAnalysisStudentSubjectTrends.value.map((series) => {
  const latest = series.points[series.points.length - 1]?.value ?? 0
  const first = series.points[0]?.value ?? latest
  const delta = latest - first
  return {
    label: series.subject,
    value: latest,
    max: Math.max(...series.points.map((point) => point.value), 100),
    meta: `${formatNumber(first)} → ${formatNumber(latest)}`,
    note: `${series.subject}最近趋势：${series.points.map((point) => formatNumber(point.value)).join(' → ')}`,
    tone: delta >= 0 ? 'strong' : 'risk'
  }
}))

const rankingRows = computed(() => {
  return [...numericScoreRows.value]
    .sort((a, b) => b.totalValue - a.totalValue)
    .map((row, index) => {
      const weakSubjects = selectedExamSubjects.value
        .filter((subject) => parseScoreValue(row.subjectScores?.[subject]) > 0 && parseScoreValue(row.subjectScores?.[subject]) < 60)
      return {
        id: row.id,
        rank: index + 1,
        studentName: row.studentName,
        total: row.total,
        totalValue: row.totalValue,
        subjectScores: row.subjectScores ?? {},
        gapToAverage: row.totalValue - classAverage.value,
        weakSubjects: weakSubjects.join('、') || '无',
        tag: row.totalValue >= excellenceLine.value ? '高分突破' : row.totalValue < riskLine.value ? '重点帮扶' : '稳定中坚'
      }
    })
})

const transcriptGridStyle = computed(() => ({
  gridTemplateColumns: `0.55fr 1fr repeat(${selectedExamSubjects.value.length}, minmax(72px, 0.78fr)) 0.85fr 0.75fr 0.95fr 1.15fr`
}))

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

const homeReadinessCards = computed(() => {
  const totalStudents = students.value.length
  const parentReady = students.value.filter((item) => item.parentStatus === '已绑定').length
  const selectionReady = students.value.filter((item) => item.selectionStatus === '已登记').length
  const teacherReady = teachers.value.filter((item) => item.accountStatus === 'bound' && item.permissionStatus === 'synced').length
  const latestExam = exams.value[0]

  return [
    {
      label: '最近考试',
      value: latestExam?.name ?? '暂无考试',
      note: latestExam ? `${latestExam.type} / ${latestExam.importStatus}` : '先导入一次考试成绩'
    },
    {
      label: '学生档案',
      value: `${parentReady}/${totalStudents}`,
      note: totalStudents === parentReady ? '家长联系方式已齐全' : '仍有学生缺少家长绑定'
    },
    {
      label: '选科状态',
      value: activeSelectionScenario.value.label,
      note: `${selectionReady}/${totalStudents} 已完成登记`
    },
    {
      label: '任课授权',
      value: `${teacherReady}/${teachers.value.length}`,
      note: teacherReady === teachers.value.length ? '任课老师权限已同步' : '仍有老师待绑定或同步'
    }
  ]
})

const homeTodoItems = computed(() => {
  const missingParents = students.value.filter((item) => item.parentStatus !== '已绑定')
  const missingSelection = students.value.filter((item) => item.selectionStatus !== '已登记')
  const pendingTeachers = teachers.value.filter((item) => item.accountStatus !== 'bound' || item.permissionStatus !== 'synced')
  const openIssues = issues.value.filter((item) => item.status !== '已修复')
  const blockers = publishBlockers.value.length > 0 ? publishBlockers.value : []

  const items: Array<{
    title: string
    detail: string
    status: string
    action: string
    target: 'classes' | 'scores' | 'sync' | 'analysis'
  }> = []

  if (missingParents.length > 0) {
    items.push({
      title: '补齐家长联系方式',
      detail: `${missingParents.map((item) => item.name).slice(0, 3).join('、')} 等学生会影响家长端同步。`,
      status: `${missingParents.length} 人待处理`,
      action: '去学生档案',
      target: 'classes'
    })
  }

  if (missingSelection.length > 0) {
    items.push({
      title: '确认学生选科状态',
      detail: `${missingSelection.map((item) => item.name).slice(0, 3).join('、')} 的选科信息会影响选考科成绩解析。`,
      status: `${missingSelection.length} 人待确认`,
      action: '去班级维护',
      target: 'classes'
    })
  }

  if (pendingTeachers.length > 0) {
    items.push({
      title: '完成任课老师账号授权',
      detail: `${pendingTeachers.map((item) => `${item.subject}${item.teacher}`).slice(0, 3).join('、')} 还不能查看授权成绩。`,
      status: `${pendingTeachers.length} 位待处理`,
      action: '去任课维护',
      target: 'classes'
    })
  }

  if (openIssues.length > 0) {
    items.push({
      title: '处理成绩导入异常',
      detail: '仍有导入问题未标记修复，建议先处理再发布同步。',
      status: `${openIssues.length} 条异常`,
      action: '去成绩详情',
      target: 'scores'
    })
  }

  if (blockers.length > 0) {
    items.push({
      title: '解除发布阻塞',
      detail: blockers.slice(0, 2).join('；'),
      status: `${blockers.length} 项阻塞`,
      action: '去同步中心',
      target: 'sync'
    })
  }

  if (items.length === 0) {
    items.push({
      title: '当前无阻塞事项',
      detail: '班级档案、成绩数据和同步条件已满足，可以继续查看分析或发布给家长。',
      status: '可发布',
      action: '查看分析',
      target: 'analysis'
    })
  }

  return items
})

const homeQuickActions = computed(() => [
  {
    title: '导入新考试成绩',
    detail: '填写考试信息、上传文件、校验后生成成绩记录。',
    action: '导入成绩',
    run: openScoreImport
  },
  {
    title: '维护学生档案',
    detail: '补家长手机号、性别、选科组合和档案状态。',
    action: '去维护',
    run: () => {
      activeNav.value = 'classes'
      classPanelMode.value = 'student'
      actionPanel.value = 'none'
    }
  },
  {
    title: '发布最近考试',
    detail: latestInsightExamName.value ? `当前目标：${latestInsightExamName.value}` : '先完成一次考试导入和分析。',
    action: canPublishLatestExam.value ? '去发布' : '看阻塞',
    run: () => {
      activeNav.value = 'sync'
      actionPanel.value = 'none'
    }
  }
])

const knownStudentIds = computed(() => students.value.map((item) => item.studentNo))
const knownStudentNames = computed(() => students.value.map((item) => item.name))

function parseScoreValue(value: string | number | undefined) {
  if (value === undefined || value === '') return 0
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function average(values: number[]) {
  if (values.length === 0) return 0
  return values.reduce((sum, value) => sum + value, 0) / values.length
}

function formatNumber(value: number) {
  if (!Number.isFinite(value)) return '0'
  return Number.isInteger(value) ? `${value}` : value.toFixed(1)
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function namesOf(rows: Array<{ studentName: string }>) {
  if (rows.length === 0) return '暂无学生'
  return rows.map((row) => row.studentName).slice(0, 6).join('、')
}

function workContextFromIdentity(identity: WorkIdentity) {
  return {
    id: identity.id,
    roleLabel: identity.roleLabel,
    primaryLabel: identity.primaryLabel,
    secondaryLabel: identity.secondaryLabel,
    description: `${identity.roleLabel} · ${identity.secondaryLabel || identity.primaryLabel}`
  }
}

function applyCurrentUser(me: CurrentUser) {
  currentUser.value = me
  if (me.workIdentities.length > 0) {
    availableWorkContexts.value = me.workIdentities.map(workContextFromIdentity)
    activeContextId.value = me.defaultRoleId || me.workIdentities[0].id
  }
}

async function sendLoginCode() {
  try {
    authSubmitting.value = true
    authError.value = ''
    const resp = await sendSMSCode(loginForm.value.mobile)
    loginForm.value.code = resp.devCode ?? ''
    authMessage.value = resp.devCode
      ? `开发环境验证码：${resp.devCode}`
      : '验证码已发送，请查收短信。'
  } catch (error) {
    authError.value = error instanceof Error ? error.message : '发送验证码失败'
  } finally {
    authSubmitting.value = false
  }
}

async function submitLogin() {
  try {
    authSubmitting.value = true
    authError.value = ''
    const resp = await loginWithSMS(loginForm.value.mobile, loginForm.value.code)
    authToken.value = resp.token
    localStorage.setItem(AUTH_TOKEN_KEY, resp.token)
    applyCurrentUser(resp.me)
    authMessage.value = ''
    await loadInitialData()
  } catch (error) {
    authError.value = error instanceof Error ? error.message : '登录失败'
  } finally {
    authSubmitting.value = false
  }
}

async function restoreSession() {
  if (!authToken.value) {
    authLoading.value = false
    return
  }
  try {
    applyCurrentUser(await fetchCurrentUser(authToken.value))
    await loadInitialData()
  } catch {
    localStorage.removeItem(AUTH_TOKEN_KEY)
    authToken.value = ''
    currentUser.value = null
  } finally {
    authLoading.value = false
  }
}

function logout() {
  localStorage.removeItem(AUTH_TOKEN_KEY)
  authToken.value = ''
  currentUser.value = null
  activeNav.value = 'overview'
  actionPanel.value = 'none'
}

function applyClassroomWorkspace(workspace: ClassroomWorkspace) {
  if (workspace.classOptions?.length) {
    managedClassOptions.value = workspace.classOptions
    availableSubjectClasses.value = workspace.classOptions.map((item) => ({
      id: item.id,
      label: item.className
    }))
    if (!availableSubjectClasses.value.find((item) => item.id === activeHomeroomClassId.value)) {
      activeHomeroomClassId.value = workspace.classOptions[0].id
    }
    if (!availableSubjectClasses.value.find((item) => item.id === activeSubjectClassId.value)) {
      activeSubjectClassId.value = workspace.classOptions[0].id
    }
  }
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
    applyClassroomWorkspace(await fetchClassroomWorkspace(currentClassId.value))
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '获取班级与师生数据失败'
  } finally {
    classroomLoading.value = false
  }
}

async function loadInitialData() {
  await loadClassroomWorkspace()
  await loadExamList()
  if (selectedExamId.value) {
    await loadExamDetail(selectedExamId.value)
  }
  await loadInsightDashboard()
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

function openClassCreator() {
  actionPanel.value = 'class-create'
}

async function saveClassForm() {
  if (!classForm.value.className.trim()) {
    classroomError.value = '请填写行政班名称，例如：高一（9）班'
    return
  }
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    const workspace = await createClass({
      schoolName: classForm.value.schoolName,
      gradeName: classForm.value.gradeName,
      className: classForm.value.className,
      homeroomTeacher: classForm.value.homeroomTeacher,
      academicYear: classForm.value.academicYear,
      stageId: classForm.value.stageId
    })
    applyClassroomWorkspace(workspace)
    activeHomeroomClassId.value = workspace.classId
    activeSubjectClassId.value = workspace.classId
    scopeMode.value = 'single'
    activeNav.value = 'classes'
    actionPanel.value = 'none'
    classroomActionMessage.value = `${workspace.className} 已创建，请继续导入学生档案和维护任课老师。`
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '创建班级失败'
  }
}

function openExamDetail(examId: string) {
  selectedExamId.value = examId
  activeNav.value = 'scores'
  actionPanel.value = 'none'
  void loadExamDetail(examId)
}

async function changeAnalysisExam() {
  if (!selectedExamId.value || isOverallScope.value) return
  await loadExamDetail(selectedExamId.value)
  await loadInsightDashboard()
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
      ? await updateStudent(currentClassId.value, studentForm.value.id, payload)
      : await createStudent(currentClassId.value, payload)
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

async function removeStudent(studentId: string) {
  const target = students.value.find((item) => item.id === studentId)
  if (!target) return
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    applyClassroomWorkspace(await deleteStudent(currentClassId.value, studentId))
    classroomActionMessage.value = `${target.name} 已从当前班级档案中删除，历史成绩记录保留。`
    actionPanel.value = 'none'
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '删除学生失败'
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
      ? await updateTeacher(currentClassId.value, teacherForm.value.id, payload)
      : await createTeacher(currentClassId.value, payload)
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

async function removeTeacher(teacherId: string) {
  const target = teachers.value.find((item) => item.id === teacherId)
  if (!target) return
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    applyClassroomWorkspace(await deleteTeacher(currentClassId.value, teacherId))
    classPanelMode.value = 'teacher'
    actionPanel.value = 'none'
    classroomActionMessage.value = `${target.teacher} 已从当前班级任课关系中移除，教师账号和其他班级授权不受影响。`
  } catch (error) {
    classroomError.value = error instanceof Error ? error.message : '移除任课老师失败'
  }
}

async function bindTeacher(teacherId: string) {
  try {
    classroomError.value = ''
    classroomActionMessage.value = ''
    applyClassroomWorkspace(await bindTeacherAccount(currentClassId.value, teacherId))
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
    applyClassroomWorkspace(await syncTeacherPermission(currentClassId.value, teacherId))
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
    const result = await importStudents(currentClassId.value, studentImportFile.value)
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
    const result = await importTeachers(currentClassId.value, teacherImportFile.value)
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
    applyClassroomWorkspace(await updatePolicy(currentClassId.value, policyForm.value.id, {
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
    classId: currentClassId.value,
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
  updateExamScore(currentClassId.value, selectedExamId.value, scoreEditForm.value.id, {
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
    exams.value = await fetchExams(currentClassId.value)
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
    const result = await fetchExamDetail(currentClassId.value, examID)
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
    const dashboard = await fetchInsightDashboard(insightScopePayload.value)
    summaryMetricRows.value = dashboard.summaryMetrics
    studentTrendRows.value = dashboard.studentTrends
    cohortInsightRows.value = dashboard.cohortInsights
    alertRows.value = dashboard.alertItems
    analysisDashboard.value = dashboard.analysis
    if (!selectedAnalysisStudentId.value && dashboard.analysis.studentAnalyses.length > 0) {
      selectedAnalysisStudentId.value = dashboard.analysis.studentAnalyses[0].studentId
    }
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
    const dashboard = await publishLatestExam(currentClassId.value)
    summaryMetricRows.value = dashboard.summaryMetrics
    studentTrendRows.value = dashboard.studentTrends
    cohortInsightRows.value = dashboard.cohortInsights
    alertRows.value = dashboard.alertItems
    analysisDashboard.value = dashboard.analysis
    if (!selectedAnalysisStudentId.value && dashboard.analysis.studentAnalyses.length > 0) {
      selectedAnalysisStudentId.value = dashboard.analysis.studentAnalyses[0].studentId
    }
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
  await restoreSession()
})

watch(currentClassId, async () => {
  if (!isAuthenticated.value) return
  actionPanel.value = 'none'
  classroomActionMessage.value = ''
  await loadClassroomWorkspace()
  await loadExamList()
  if (selectedExamId.value) {
    await loadExamDetail(selectedExamId.value)
  } else {
    scoreRows.value = []
    issues.value = []
  }
  await loadInsightDashboard()
})

watch(scopeMode, async () => {
  if (!isAuthenticated.value) return
  if (isOverallScope.value) {
    await loadInsightDashboard()
    return
  }
  await loadClassroomWorkspace()
  await loadExamList()
  if (selectedExamId.value) {
    await loadExamDetail(selectedExamId.value)
  }
  await loadInsightDashboard()
})
</script>

<template>
  <div v-if="authLoading" class="auth-shell">
    <section class="auth-card">
      <p class="brand-kicker">OPC</p>
      <h1>正在恢复登录状态</h1>
      <p>正在确认账号、身份和授权范围...</p>
    </section>
  </div>

  <div v-else-if="!isAuthenticated" class="auth-shell">
    <section class="auth-card">
      <div>
        <p class="brand-kicker">OPC EDU INSIGHT</p>
        <h1>手机号验证码登录</h1>
        <p>先确认稳定账号，再识别班主任、任课老师等工作身份。手机号只是登录凭证，后续可更换。</p>
      </div>

      <div class="auth-form">
        <label>
          <span>手机号</span>
          <input v-model="loginForm.mobile" type="tel" maxlength="11" placeholder="请输入手机号" />
        </label>
        <label>
          <span>验证码</span>
          <input v-model="loginForm.code" type="text" maxlength="6" placeholder="点击获取验证码" />
        </label>
        <div class="auth-actions">
          <button type="button" class="ghost-btn" :disabled="authSubmitting" @click="sendLoginCode">获取验证码</button>
          <button type="button" class="solid-btn" :disabled="authSubmitting" @click="submitLogin">登录教师端</button>
        </div>
      </div>

      <div class="auth-demo">
        <strong>本地体验账号</strong>
        <span>李老师：13800001003，可看到班主任 + 任课老师身份</span>
        <span>王老师：13800001002，可看到任课老师身份</span>
        <span>开发环境会直接返回验证码并自动填入。</span>
      </div>

      <p v-if="authMessage" class="success-note">{{ authMessage }}</p>
      <p v-if="authError" class="error-note">{{ authError }}</p>
    </section>
  </div>

  <div v-else class="teacher-shell">
    <aside class="sidebar">
      <div class="brand">
        <p class="brand-kicker">OPC</p>
        <h1>成绩洞察平台</h1>
        <p class="brand-copy">教师端可体验原型</p>
      </div>

      <nav class="nav-list">
        <div v-for="item in navItems" :key="item.id" class="nav-group">
          <button
            :class="['nav-item', { active: activeNav === item.id }]"
            type="button"
            @click="activeNav = item.id"
          >
            <span class="nav-icon">
              <AppIcon :name="navIconMap[item.id] ?? 'dashboard'" />
            </span>
            <span class="nav-copy">
              <strong>{{ item.label }}</strong>
              <span>{{ item.hint }}</span>
            </span>
          </button>

          <div v-if="item.id === 'analysis' && activeNav === 'analysis'" class="nav-sub-list">
            <button
              v-for="mode in analysisModeOptions"
              :key="mode.id"
              type="button"
              :class="['nav-sub-item', { active: analysisMode === mode.id }]"
              @click="setAnalysisMode(mode.id)"
            >
              <AppIcon :name="mode.icon" :size="15" />
              <span>{{ mode.label }}</span>
            </button>
          </div>
        </div>
      </nav>
    </aside>

    <main class="main-area">
      <header class="topbar">
        <div>
          <p class="topbar-label">当前身份：{{ activeContext.roleLabel }} · 数据范围：{{ activeScopeLabel }}</p>
          <h2>{{ activeTitle }}</h2>
          <p class="account-line">{{ currentUserName }} · {{ currentUserMobile }}</p>
        </div>

        <div class="topbar-actions">
          <label class="topbar-select">
            <span>工作身份</span>
            <select v-model="activeContextId">
              <option v-for="context in availableWorkContexts" :key="context.id" :value="context.id">
                {{ context.roleLabel }} / {{ context.primaryLabel }}
              </option>
            </select>
          </label>
          <label class="topbar-select">
            <span>数据范围</span>
            <select v-model="scopeMode">
              <option value="single">单班级</option>
              <option value="overall">{{ isSubjectTeacherView ? '任课班级整体' : '所带班级整体' }}</option>
            </select>
          </label>
          <label v-if="!isSubjectTeacherView && scopeMode === 'single'" class="topbar-select">
            <span>当前班级</span>
            <select v-model="activeHomeroomClassId">
              <option v-for="item in availableSubjectClasses" :key="item.id" :value="item.id">
                {{ item.label }}
              </option>
            </select>
          </label>
          <label v-if="isSubjectTeacherView && scopeMode === 'single'" class="topbar-select">
            <span>当前班级</span>
            <select v-model="activeSubjectClassId">
              <option v-for="item in availableSubjectClasses" :key="item.id" :value="item.id">
                {{ item.label }}
              </option>
            </select>
          </label>
          <button type="button" class="solid-btn topbar-primary" @click="openScoreImport">录入成绩</button>
          <button type="button" class="ghost-btn small" @click="logout">退出</button>
        </div>
      </header>

      <template v-if="activeNav === 'overview'">
        <section class="metrics-grid">
          <article v-for="item in homeReadinessCards" :key="item.label" class="metric-card">
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
            <span>{{ item.note }}</span>
          </article>
        </section>

        <section class="content-grid">
          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">今日待办</p>
                <h3>影响发布和查看的阻塞事项</h3>
              </div>
            </div>

            <div class="workbench-list">
              <button
                v-for="todo in homeTodoItems"
                :key="todo.title"
                type="button"
                class="workbench-item"
                @click="activeNav = todo.target"
              >
                <div>
                  <strong>{{ todo.title }}</strong>
                  <p>{{ todo.detail }}</p>
                </div>
                <span>{{ todo.status }} · {{ todo.action }}</span>
              </button>
            </div>
          </article>

          <article class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">成绩分析快照</p>
                <h3>最近多次考试变化与本次分层</h3>
              </div>
              <button type="button" class="ghost-btn small" @click="activeNav = 'analysis'">进入完整分析</button>
            </div>

            <div class="overview-chart-grid">
              <div class="line-panel">
                <div class="chart-title-line">
                  <AppIcon name="lineChart" />
                  <strong>班级均分趋势</strong>
                </div>
                <LineChart :points="classTrendChartPoints" :height="220" />
              </div>

              <div class="line-panel">
                <div class="chart-title-line">
                  <AppIcon name="distribution" />
                  <strong>本次分数分布</strong>
                </div>
                <DistributionBars :rows="distributionRows" />
              </div>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">快捷入口</p>
                <h3>高频操作</h3>
              </div>
            </div>

            <div class="task-list">
              <button
                v-for="item in homeQuickActions"
                :key="item.title"
                type="button"
                class="task-item button-task"
                @click="item.run()"
              >
                <strong>{{ item.title }}</strong>
                <p>{{ item.detail }}</p>
                <span>{{ item.action }}</span>
              </button>
            </div>
          </article>

          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="panel-label">最近考试</p>
                <h3>{{ exams[0]?.name ?? '暂无考试' }}</h3>
              </div>
            </div>

            <div class="task-list">
              <div class="task-item">
                <strong>导入状态</strong>
                <p>{{ exams[0]?.subjectCoverage ?? '暂无学科覆盖' }}</p>
                <span>{{ exams[0]?.importStatus ?? '待导入' }}</span>
              </div>
              <div class="task-item">
                <strong>同步准备度</strong>
                <p>{{ canPublishLatestExam ? '当前最近考试满足发布条件。' : '仍需处理阻塞后再发布。' }}</p>
                <span>{{ canPublishLatestExam ? '可发布' : `${publishBlockers.length} 项阻塞` }}</span>
              </div>
            </div>
          </article>
        </section>
      </template>

      <template v-else-if="activeNav === 'classes'">
        <section class="content-grid">
          <article v-if="isOverallScope" class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">整体范围</p>
                <h3>{{ activeScopeLabel }}不直接维护明细</h3>
              </div>
            </div>

            <div class="status-inline">
              当前是整体数据范围，适合查看工作台、考试汇总、分析与同步准备度。学生档案、家长信息、任课老师和展示策略需要落到单个班级维护。
            </div>

            <div class="insight-grid compact-insights">
              <div class="insight-card">
                <strong>覆盖班级</strong>
                <div class="insight-count">{{ availableSubjectClasses.length }}</div>
                <p>{{ availableSubjectClasses.map((item) => item.label).join('、') }}</p>
              </div>
              <div class="insight-card">
                <strong>当前身份</strong>
                <div class="insight-count">{{ activeContext.roleLabel }}</div>
                <p>{{ isSubjectTeacherView ? '按授权任课班级聚合查看' : '按所带班级聚合查看' }}</p>
              </div>
              <div class="insight-card">
                <strong>维护方式</strong>
                <div class="insight-count">单班</div>
                <p>切换为单班级后再维护学生、家长、任课老师和策略。</p>
              </div>
            </div>

            <div class="row-actions top-gap">
              <button type="button" class="solid-btn small" @click="scopeMode = 'single'">切换到单班维护</button>
            </div>
          </article>

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

          <article v-if="!isSubjectTeacherView" class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">班级管理</p>
                <h3>已维护班级与创建入口</h3>
              </div>
              <button type="button" class="solid-btn small" @click="openClassCreator">新建班级</button>
            </div>

            <div class="table-list">
              <div class="table-row table-header class-cols">
                <span>年级</span>
                <span>班级</span>
                <span>学年</span>
                <span>当前状态</span>
                <span>操作</span>
              </div>
              <button
                v-for="item in managedClassOptions"
                :key="item.id"
                type="button"
                :class="['table-row', 'class-cols', 'clickable-row', { selected: currentClassId === item.id && scopeMode === 'single' }]"
                @click="activeHomeroomClassId = item.id; scopeMode = 'single'"
              >
                <span>{{ item.gradeName }}</span>
                <span>{{ item.className }}</span>
                <span>{{ item.academicYear }}</span>
                <span>{{ currentClassId === item.id && scopeMode === 'single' ? '当前维护中' : '可切换' }}</span>
                <span class="row-actions">
                  <span class="inline-action">切换维护</span>
                </span>
              </button>
            </div>
          </article>

          <article v-if="canMaintainClassData" class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">班级基础信息</p>
                <h3>班级归属、学段和选科阶段</h3>
              </div>
              <div class="panel-actions">
                <button type="button" class="ghost-btn small" @click="classPanelMode = 'policy'; openPolicyEditor()">
                  编辑班级策略
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
          </article>

          <article v-if="canMaintainClassData" class="panel panel-wide">
            <div class="panel-head">
              <div>
                <p class="panel-label">主数据完整度</p>
                <h3>学生、家长、选科与任课授权</h3>
              </div>
            </div>

            <div class="insight-grid compact-insights">
              <div v-for="item in rosterInsightRows" :key="item.title" class="insight-card">
                <strong>{{ item.title }}</strong>
                <div class="insight-count">{{ item.count }}</div>
                <p>{{ item.detail }}</p>
              </div>
            </div>
          </article>

          <article v-if="canMaintainClassData" class="panel panel-wide">
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
                  <button type="button" class="ghost-btn tiny danger" @click="removeStudent(student.id)">删除</button>
                </span>
              </div>
            </div>
          </article>

          <article v-if="canMaintainClassData" class="panel panel-wide">
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
                  <button type="button" class="ghost-btn tiny danger" @click="removeTeacher(assignment.id)">移除</button>
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

          <article v-if="canMaintainClassData" class="panel">
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

          <article v-if="canMaintainClassData" class="panel">
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

        <section class="analysis-context-bar">
          <div class="analysis-context-main">
            <span class="context-mode">{{ analysisModeOptions.find((item) => item.id === analysisMode)?.label }}</span>
            <span v-for="item in analysisReportContext" :key="item.label" class="context-chip">
              <em>{{ item.label }}</em>
              <strong>{{ item.value }}</strong>
            </span>
          </div>
          <div class="analysis-context-actions">
            <label v-if="!isOverallScope" class="context-select">
              <span>考试</span>
              <select v-model="selectedExamId" @change="changeAnalysisExam">
                <option v-for="exam in exams" :key="exam.id" :value="exam.id">
                  {{ exam.name }} · {{ exam.date }}
                </option>
              </select>
            </label>
            <span v-else class="context-note">整体范围按各班最近一次考试聚合，班级对比页展示具体考试口径。</span>
          </div>
        </section>

        <template v-if="analysisMode === 'report'">
          <section class="report-summary-panel">
            <div class="panel-head compact-head">
              <div>
                <p class="panel-label">本次结论</p>
                <h3>先看判断，再看明细</h3>
              </div>
              <span class="inline-note">规则化摘要</span>
            </div>
            <ul class="report-summary-list">
              <li v-for="item in analysisReportSummary" :key="item">{{ item }}</li>
            </ul>
          </section>

          <section class="metrics-grid report-metrics-grid">
            <article v-for="metric in analysisMetrics" :key="metric.label" class="metric-card">
              <p>{{ metric.label }}</p>
              <strong>{{ metric.value }}</strong>
              <span>{{ metric.note }}</span>
            </article>
          </section>

          <section class="analysis-chart-grid">
            <article class="panel chart-panel-main">
              <div class="panel-head">
                <div>
                  <p class="panel-label">趋势图表</p>
                  <h3>最近多次考试班级均分</h3>
                </div>
                <span class="inline-note">{{ classTrendChartPoints.length }} 次考试</span>
              </div>
              <LineChart :points="classTrendChartPoints" :height="260" />
            </article>

            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="panel-label">学科柱状图</p>
                  <h3>本次考试学科均分</h3>
                </div>
              </div>
              <HorizontalBarChart :rows="subjectDiagnosticBars" />
            </article>

            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="panel-label">分布图表</p>
                  <h3>本次考试分层占比</h3>
                </div>
              </div>
              <DistributionBars :rows="distributionRows" />
            </article>
          </section>

          <section class="content-grid">
            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">学科诊断</p>
                  <h3>{{ activeScopeLabel }} 本次考试学科表现</h3>
                </div>
                <span class="inline-note">{{ selectedExam?.name }}</span>
              </div>
              <HorizontalBarChart :rows="subjectDiagnosticBars" />
            </article>
          </section>
        </template>

        <template v-else-if="analysisMode === 'transcript'">
          <section class="content-grid">
            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">成绩单</p>
                  <h3>学生成绩、排名、分层和薄弱科目</h3>
                </div>
                <span class="inline-note">{{ rankingRows.length }} 名学生 · {{ selectedExamSubjects.length }} 个学科</span>
              </div>

              <div class="transcript-table">
                <div class="table-row table-header transcript-cols" :style="transcriptGridStyle">
                  <span>名次</span>
                  <span>学生</span>
                  <span v-for="subject in selectedExamSubjects" :key="`head-${subject}`">{{ subject }}</span>
                  <span>总分</span>
                  <span>均分差</span>
                  <span>分层</span>
                  <span>薄弱科目</span>
                </div>
                <div v-for="row in rankingRows" :key="row.id" class="table-row transcript-cols" :style="transcriptGridStyle">
                  <strong>{{ row.rank }}</strong>
                  <span>{{ row.studentName }}</span>
                  <span v-for="subject in selectedExamSubjects" :key="`${row.id}-${subject}`">
                    {{ row.subjectScores?.[subject] || '-' }}
                  </span>
                  <strong>{{ row.total }}</strong>
                  <span :class="row.gapToAverage >= 0 ? 'positive-text' : 'negative-text'">
                    {{ row.gapToAverage >= 0 ? '+' : '' }}{{ formatNumber(row.gapToAverage) }}
                  </span>
                  <span>{{ row.tag }}</span>
                  <span>{{ row.weakSubjects }}</span>
                </div>
              </div>
            </article>
          </section>
        </template>

        <template v-else-if="analysisMode === 'distribution'">
          <section class="content-grid">
            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="panel-label">水平分布</p>
                  <h3>分数段与分层结构</h3>
                </div>
              </div>
              <DistributionBars :rows="distributionRows" />
            </article>

            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="panel-label">班级离散度</p>
                  <h3>最高、最低和均分位置</h3>
                </div>
              </div>
              <div class="spread-card">
                <strong>{{ formatNumber(scoreSpread) }}</strong>
                <p>最高分与最低分差距</p>
                <div class="spread-scale">
                  <span>最低 {{ formatNumber(classLowest) }}</span>
                  <span>均分 {{ formatNumber(classAverage) }}</span>
                  <span>最高 {{ formatNumber(classHighest) }}</span>
                </div>
              </div>
            </article>

            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">名次段</p>
                  <h3>按分层名单观察结构</h3>
                </div>
              </div>
              <div class="layer-grid">
                <div v-for="group in layeredStudentGroups" :key="group.title" class="layer-card">
                  <div class="layer-card-head">
                    <strong>{{ group.title }}</strong>
                    <span>{{ group.count }}</span>
                  </div>
                  <p>{{ group.students }}</p>
                  <em>{{ group.goal }}</em>
                </div>
              </div>
            </article>
          </section>
        </template>

        <template v-else-if="analysisMode === 'compare'">
          <section v-if="isOverallScope" class="metrics-grid">
            <article class="metric-card">
              <p>对比班级</p>
              <strong>{{ classComparisonRows.length }} 个</strong>
              <span>当前整体范围内参与对比的班级</span>
            </article>
            <article class="metric-card">
              <p>最高均分</p>
              <strong>{{ formatNumber(classComparisonRows[0]?.average ?? 0) }}</strong>
              <span>{{ classComparisonRows[0]?.className ?? '暂无班级' }}</span>
            </article>
            <article class="metric-card">
              <p>总人数</p>
              <strong>{{ classComparisonRows.reduce((sum, item) => sum + item.studentCount, 0) }} 人</strong>
              <span>按各班最近同名考试汇总</span>
            </article>
            <article class="metric-card">
              <p>预警合计</p>
              <strong>{{ classComparisonRows.reduce((sum, item) => sum + item.riskCount, 0) }} 人</strong>
              <span>低于本班预警线学生</span>
            </article>
          </section>

          <section class="content-grid">
            <article v-if="isOverallScope" class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">班级对比图</p>
                  <h3>各班最近同口径考试均分</h3>
                </div>
                <span class="inline-note">{{ activeScopeLabel }}</span>
              </div>
              <HorizontalBarChart :rows="classComparisonBars" />
            </article>

            <article v-if="isOverallScope" class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">对比明细</p>
                  <h3>班级均分、最高最低与预警人数</h3>
                </div>
              </div>

              <div class="risk-table">
                <div class="table-row table-header compare-cols">
                  <span>班级</span>
                  <span>考试</span>
                  <span>人数</span>
                  <span>均分</span>
                  <span>最高 / 最低</span>
                  <span>预警</span>
                </div>
                <div v-for="item in classComparisonRows" :key="item.classId" class="table-row compare-cols">
                  <strong>{{ item.className }}</strong>
                  <span>{{ item.examName }}</span>
                  <span>{{ item.studentCount }}</span>
                  <span>{{ formatNumber(item.average) }}</span>
                  <span>{{ formatNumber(item.highest) }} / {{ formatNumber(item.lowest) }}</span>
                  <span>{{ item.riskCount }} 人</span>
                </div>
              </div>
            </article>

            <article v-else class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">单班对比提示</p>
                  <h3>切换到整体范围查看多班级对比</h3>
                </div>
              </div>
              <p class="empty-tip">当前是单班级范围，适合查看考试报告、成绩单、水平分布和学生分析。若要对比多个班级，请在右上角数据范围切换为整体。</p>
            </article>
          </section>
        </template>

        <template v-else-if="analysisMode === 'trend'">
          <section class="content-grid">
            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">多次考试趋势</p>
                  <h3>班级均分和学科变化</h3>
                </div>
                <span class="inline-note">{{ classTrendPoints.length }} 次考试</span>
              </div>

              <div class="trend-analysis-grid">
                <div class="line-panel">
                  <strong>班级均分趋势</strong>
                  <LineChart :points="classTrendChartPoints" :height="210" />
                </div>
                <div class="subject-trend-list">
                  <HorizontalBarChart :rows="subjectTrendBars" />
                </div>
              </div>
            </article>

            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">趋势判断</p>
                  <h3>这段时间最需要关注的变化</h3>
                </div>
              </div>

              <div class="action-grid">
                <div class="action-card">
                  <span>班级趋势</span>
                  <strong>{{ classTrendPoints.length > 1 && classTrendPoints[classTrendPoints.length - 1].value >= classTrendPoints[0].value ? '整体上升' : '需要观察' }}</strong>
                  <p>班级均分从 {{ formatNumber(classTrendPoints[0]?.value ?? 0) }} 到 {{ formatNumber(classTrendPoints[classTrendPoints.length - 1]?.value ?? 0) }}。</p>
                </div>
                <div class="action-card" v-for="series in subjectTrendSeries.slice(0, 3)" :key="`trend-${series.subject}`">
                  <span>{{ series.subject }}</span>
                  <strong>{{ series.points.map((point) => formatNumber(point.value)).join(' → ') }}</strong>
                  <p>用于判断该学科是持续提升、短期波动还是长期薄弱。</p>
                </div>
              </div>
            </article>
          </section>
        </template>

        <template v-else-if="analysisMode === 'student'">
          <section class="content-grid">
            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">选择学生</p>
                  <h3>查看单个学生的多次考试变化</h3>
                </div>
                <label class="topbar-select">
                  <span>学生</span>
                  <select v-model="selectedAnalysisStudentId">
                    <option v-for="student in studentAnalysisRows" :key="student.studentId || student.studentName" :value="student.studentId">
                      {{ student.studentName }}
                    </option>
                  </select>
                </label>
              </div>
            </article>

            <article class="panel panel-wide" v-if="selectedAnalysisStudent">
              <div class="panel-head">
                <div>
                  <p class="panel-label">单学生视角</p>
                  <h3>{{ selectedAnalysisStudent.studentName }} 成绩画像</h3>
                </div>
                <span class="inline-note">最近班级第 {{ selectedAnalysisStudent.latestRank }} 名</span>
              </div>

              <div class="student-analysis-grid">
                <div class="student-analysis-card">
                  <div class="layer-card-head">
                    <strong>总分趋势</strong>
                    <span>{{ selectedAnalysisStudent.latestTotal }}</span>
                  </div>
                  <LineChart :points="selectedStudentTotalTrendPoints" :height="170" />
                  <p>观察总分是否持续上升、下降或出现明显波动。</p>
                </div>
                <div class="student-analysis-card">
                  <div class="layer-card-head">
                    <strong>排名趋势</strong>
                    <span>第 {{ selectedAnalysisStudent.latestRank }} 名</span>
                  </div>
                  <LineChart :points="selectedStudentRankTrendPoints" :height="170" invert value-suffix="名" />
                  <p>排名越小越好，后续图表需要反向坐标展示。</p>
                </div>
                <div class="student-analysis-card">
                  <strong>短板建议</strong>
                  <p>{{ selectedAnalysisStudent.recommendation }}</p>
                  <em>薄弱科目：{{ selectedAnalysisStudent.weakSubjects }}</em>
                </div>
              </div>
            </article>

            <article class="panel panel-wide" v-if="selectedAnalysisStudentSubjectTrends.length">
              <div class="panel-head">
                <div>
                  <p class="panel-label">学科趋势</p>
                  <h3>{{ selectedAnalysisStudent?.studentName }} 各科学习变化</h3>
                </div>
              </div>
              <HorizontalBarChart :rows="selectedStudentSubjectBars" />
            </article>
          </section>
        </template>

        <template v-else>
          <section class="content-grid">
            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">重点预警</p>
                  <h3>需要老师立刻看的人和科目</h3>
                </div>
                <span class="inline-note">{{ riskStudentRows.length }} 人待跟进</span>
              </div>

              <div class="risk-table">
                <div class="table-row table-header risk-cols">
                  <span>学生</span>
                  <span>总分</span>
                  <span>风险等级</span>
                  <span>原因</span>
                  <span>薄弱科目</span>
                </div>
                <div v-for="student in riskStudentRows" :key="student.name" class="table-row risk-cols">
                  <strong>{{ student.name }}</strong>
                  <span>{{ student.total }}</span>
                  <span :class="['risk-pill', student.level === '高' ? 'high' : 'mid']">{{ student.level }}</span>
                  <span>{{ student.reason }}</span>
                  <span>{{ student.weakSubjects }}</span>
                </div>
                <p v-if="riskStudentRows.length === 0" class="empty-tip">当前没有明显预警学生。</p>
              </div>
            </article>

            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-label">教学行动建议</p>
                  <h3>把分析转成下一步动作</h3>
                </div>
              </div>
              <div class="action-grid">
                <div v-for="item in teachingActions" :key="item.title" class="action-card">
                  <span>{{ item.tag }}</span>
                  <strong>{{ item.title }}</strong>
                  <p>{{ item.detail }}</p>
                </div>
              </div>
            </article>
          </section>
        </template>

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
              actionPanel === 'class-create' ? '新建班级' :
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

      <div v-if="actionPanel === 'class-create'" class="form-stack">
        <label>
          <span>学校</span>
          <input v-model="classForm.schoolName" type="text" />
        </label>
        <label>
          <span>年级</span>
          <select v-model="classForm.gradeName">
            <option value="初一">初一</option>
            <option value="初二">初二</option>
            <option value="初三">初三</option>
            <option value="高一">高一</option>
            <option value="高二">高二</option>
            <option value="高三">高三</option>
          </select>
        </label>
        <label>
          <span>行政班</span>
          <input v-model="classForm.className" type="text" placeholder="例如：高一（9）班" />
        </label>
        <label>
          <span>班主任</span>
          <input v-model="classForm.homeroomTeacher" type="text" />
        </label>
        <label>
          <span>学年</span>
          <input v-model="classForm.academicYear" type="text" placeholder="例如：2025-2026 学年" />
        </label>
        <label>
          <span>选科阶段</span>
          <select v-model="classForm.stageId">
            <option value="pre-selection">未选科阶段</option>
            <option value="post-selection">已选科阶段</option>
          </select>
        </label>
        <div class="task-item">
          <strong>创建后下一步</strong>
          <p>新班级会自动切为当前班级，随后导入学生档案并维护任课老师。</p>
          <span>成绩链路主数据</span>
        </div>
        <div class="row-actions">
          <button type="button" class="solid-btn small" @click="saveClassForm">创建班级</button>
          <button type="button" class="ghost-btn small" @click="actionPanel = 'none'">取消</button>
        </div>
      </div>

      <div v-else-if="actionPanel === 'student-add' || actionPanel === 'student-edit'" class="form-stack">
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

.auth-shell {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 32px;
  background:
    radial-gradient(circle at 12% 18%, rgba(154, 91, 38, 0.14), transparent 28%),
    radial-gradient(circle at 88% 14%, rgba(32, 25, 17, 0.1), transparent 24%),
    linear-gradient(135deg, #fbf6ee 0%, #efe3d2 100%);
}

.auth-card {
  width: min(520px, 100%);
  padding: 34px;
  border: 1px solid rgba(74, 52, 24, 0.14);
  border-radius: var(--radius-xl);
  background: rgba(255, 253, 249, 0.9);
  box-shadow: 0 28px 90px rgba(70, 49, 23, 0.18);
  backdrop-filter: blur(18px);
}

.auth-card h1 {
  margin: 12px 0 10px;
  font-size: 2.35rem;
  letter-spacing: -0.05em;
}

.auth-card p {
  margin: 0;
  color: var(--muted);
  line-height: 1.8;
}

.auth-form {
  display: grid;
  gap: 14px;
  margin-top: 26px;
}

.auth-form label {
  display: grid;
  gap: 8px;
  color: var(--muted);
}

.auth-form input {
  width: 100%;
  padding: 13px 14px;
  border: 1px solid var(--line);
  border-radius: 14px;
  background: var(--panel);
  color: var(--ink);
  outline: none;
}

.auth-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}

.auth-demo {
  display: grid;
  gap: 8px;
  margin-top: 22px;
  padding: 14px;
  border-radius: 16px;
  background: var(--panel-alt);
  color: var(--muted);
}

.auth-demo strong {
  color: var(--ink);
}

.account-line {
  margin: 6px 0 0;
  color: var(--muted);
  font-size: 0.92rem;
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

.nav-group {
  display: grid;
  gap: 8px;
}

.nav-item {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr);
  gap: 12px;
  align-items: center;
  width: 100%;
  padding: 14px;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  text-align: left;
}

.nav-icon {
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border-radius: 12px;
  color: var(--accent);
  background: rgba(154, 91, 38, 0.09);
}

.nav-copy strong,
.nav-copy span {
  display: block;
}

.nav-copy span {
  margin-top: 6px;
  color: var(--muted);
  font-size: 0.9rem;
}

.nav-item.active {
  background: var(--panel);
  border-color: var(--line);
  box-shadow: var(--shadow);
}

.nav-item.active .nav-icon {
  color: #fffaf2;
  background: var(--ink);
}

.nav-sub-list {
  display: grid;
  gap: 6px;
  margin: 0 0 4px 46px;
  padding-left: 12px;
  border-left: 1px solid rgba(74, 52, 24, 0.14);
}

.nav-sub-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 9px 10px;
  border: 1px solid transparent;
  border-radius: 12px;
  background: transparent;
  color: var(--muted);
  text-align: left;
  font-size: 0.9rem;
  cursor: pointer;
}

.nav-sub-item.active {
  border-color: rgba(154, 91, 38, 0.2);
  background: rgba(154, 91, 38, 0.09);
  color: var(--ink);
  font-weight: 800;
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

.topbar > div:first-child {
  min-width: 220px;
}

.topbar h2 {
  font-size: 2rem;
  letter-spacing: -0.04em;
}

.topbar-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: nowrap;
  gap: 8px;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.topbar-actions::-webkit-scrollbar {
  display: none;
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
  flex: 0 1 auto;
  gap: 6px;
  padding: 7px 10px;
  border-radius: 999px;
  background: var(--panel);
  border: 1px solid var(--line);
  white-space: nowrap;
}

.topbar-select span {
  color: var(--muted);
  font-size: 0.8rem;
}

.topbar-select select {
  border: 0;
  background: transparent;
  color: var(--ink);
  outline: none;
  min-width: 96px;
  max-width: 150px;
}

.topbar-primary {
  flex: 0 0 auto;
  padding-inline: 16px;
  white-space: nowrap;
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

.analysis-mode-tabs {
  display: inline-flex;
  gap: 8px;
  margin-top: 22px;
  padding: 6px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(255, 253, 249, 0.72);
  box-shadow: var(--shadow);
}

.analysis-context-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-top: 16px;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: 20px;
  background:
    linear-gradient(135deg, rgba(255, 253, 249, 0.92), rgba(247, 239, 228, 0.86));
  box-shadow: 0 12px 34px rgba(70, 49, 23, 0.08);
}

.analysis-context-main,
.analysis-context-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.analysis-context-main {
  flex: 1 1 auto;
  overflow-x: auto;
  scrollbar-width: none;
}

.analysis-context-main::-webkit-scrollbar {
  display: none;
}

.context-mode,
.context-chip,
.context-select {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  flex: 0 0 auto;
  min-height: 34px;
  padding: 7px 10px;
  border-radius: 999px;
  background: rgba(255, 253, 249, 0.76);
  border: 1px solid rgba(74, 52, 24, 0.09);
}

.context-mode {
  background: var(--ink);
  color: #fff8ed;
  font-size: 0.86rem;
  font-weight: 800;
}

.context-chip em,
.context-select span {
  color: var(--muted);
  font-size: 0.78rem;
  font-style: normal;
}

.context-chip strong {
  max-width: 240px;
  overflow: hidden;
  color: var(--ink);
  font-size: 0.88rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.context-select select {
  max-width: 260px;
  border: 0;
  background: transparent;
  color: var(--ink);
  outline: none;
}

.context-note {
  max-width: 360px;
  color: var(--muted);
  font-size: 0.84rem;
  line-height: 1.5;
}

.report-summary-panel {
  margin-top: 16px;
  padding: 18px;
  border: 1px solid var(--line);
  border-radius: var(--radius-lg);
  background: var(--panel);
  box-shadow: var(--shadow);
}

.compact-head {
  margin-bottom: 10px;
}

.report-summary-list {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.report-summary-list li {
  position: relative;
  padding: 12px 14px 12px 34px;
  border-radius: 14px;
  background: var(--panel-alt);
  color: var(--ink);
  line-height: 1.7;
}

.report-summary-list li::before {
  content: '';
  position: absolute;
  top: 20px;
  left: 16px;
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: var(--accent);
}

.report-metrics-grid {
  margin-top: 18px;
}

.analysis-sub-tabs {
  display: inline-flex;
  gap: 8px;
  margin-top: 14px;
  padding: 6px;
  border: 1px solid rgba(74, 52, 24, 0.1);
  border-radius: 999px;
  background: rgba(248, 242, 232, 0.76);
}

.analysis-mode-tabs .tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: transparent;
  color: var(--muted);
}

.analysis-sub-tabs .tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  background: transparent;
  color: var(--muted);
}

.analysis-mode-tabs .tab-btn.active,
.analysis-sub-tabs .tab-btn.active {
  background: var(--ink);
  color: #f9f4ea;
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

.layer-grid,
.action-grid {
  display: grid;
  gap: 12px;
}

.layer-card-head,
.spread-scale {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.spread-card p,
.layer-card p,
.layer-card em,
.action-card p {
  color: var(--muted);
}

.layer-card,
.action-card,
.spread-card {
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--panel-alt);
}

.layer-card p,
.action-card p {
  margin: 10px 0 0;
  line-height: 1.6;
}

.spread-card strong {
  display: block;
  font-size: 3rem;
  letter-spacing: -0.06em;
}

.spread-scale {
  margin-top: 16px;
  font-size: 0.86rem;
  color: var(--muted);
}

.layer-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.layer-card em {
  display: block;
  margin-top: 12px;
  font-style: normal;
  line-height: 1.6;
}

.risk-table {
  display: grid;
  gap: 10px;
}

.transcript-table {
  display: grid;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
}

.transcript-cols {
  min-width: 980px;
}

.positive-text {
  color: #2f6f4e;
  font-weight: 800;
}

.negative-text {
  color: #a13d2d;
  font-weight: 800;
}

.risk-cols {
  grid-template-columns: 0.9fr 0.7fr 0.7fr 1.2fr 1.2fr;
}

.ranking-cols {
  grid-template-columns: 0.5fr 1fr 0.8fr 1fr 1.2fr;
}

.compare-cols {
  grid-template-columns: 1.2fr 1.2fr 0.6fr 0.7fr 0.9fr 0.7fr;
}

.risk-pill {
  justify-self: start;
  padding: 5px 10px;
  border-radius: 999px;
  color: #fffaf2;
  font-size: 0.84rem;
  font-weight: 700;
}

.risk-pill.high {
  background: #a13d2d;
}

.risk-pill.mid {
  background: #9a5b26;
}

.action-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.action-card span {
  color: var(--accent);
  font-size: 0.82rem;
  font-weight: 700;
}

.action-card strong {
  display: block;
  margin-top: 10px;
}

.trend-analysis-grid {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 14px;
}

.analysis-chart-grid {
  display: grid;
  grid-template-columns: 1.35fr 0.95fr;
  gap: 16px;
  margin-top: 18px;
}

.overview-chart-grid {
  display: grid;
  grid-template-columns: 1.15fr 0.85fr;
  gap: 14px;
}

.chart-panel-main {
  grid-row: span 2;
}

.chart-title-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  color: var(--accent);
}

.chart-title-line strong {
  color: var(--ink);
}

.line-panel,
.student-analysis-card {
  padding: 14px;
  border-radius: var(--radius-md);
  background: var(--panel-alt);
}

.subject-trend-list,
.student-analysis-grid {
  display: grid;
  gap: 10px;
}

.student-analysis-card p,
.student-analysis-card em {
  color: var(--muted);
}

.student-analysis-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.student-analysis-card p,
.student-analysis-card em {
  display: block;
  margin: 10px 0 0;
  line-height: 1.6;
  font-style: normal;
}

.table-list,
.alert-list,
.issue-list,
.alert-grid,
.workbench-list {
  display: grid;
  gap: 10px;
}

.workbench-item,
.button-task {
  width: 100%;
  border: 0;
  text-align: left;
}

.workbench-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 16px;
  border-radius: var(--radius-md);
  background: var(--panel-alt);
}

.workbench-item p,
.button-task p {
  margin: 8px 0 0;
  color: var(--muted);
  line-height: 1.6;
}

.workbench-item span,
.button-task span {
  flex: 0 0 auto;
  color: var(--accent);
  font-size: 0.86rem;
  font-weight: 700;
}

.button-task {
  display: grid;
  gap: 8px;
}

.button-task span {
  margin-top: 4px;
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

.class-cols {
  grid-template-columns: 0.8fr 1.2fr 1.1fr 1fr 1fr;
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

.inline-action {
  color: var(--accent);
  font-size: 0.86rem;
  font-weight: 700;
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

.ghost-btn.danger {
  color: #9f2f24;
  border-color: rgba(159, 47, 36, 0.24);
  background: #fff4f1;
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
  .analysis-chart-grid,
  .overview-chart-grid,
  .content-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .chart-panel-main {
    grid-row: auto;
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
  .analysis-chart-grid,
  .overview-chart-grid,
  .compact-metrics,
  .content-grid,
  .table-row,
  .four-cols,
  .five-cols,
  .six-cols,
  .class-cols,
  .score-cols,
  .score-cols-pre,
  .score-cols-post,
  .preview-cols {
    grid-template-columns: 1fr;
  }

  .analysis-context-bar {
    align-items: stretch;
    flex-direction: column;
  }

  .analysis-context-actions {
    justify-content: space-between;
  }

  .context-select {
    width: 100%;
  }

  .context-select select {
    flex: 1;
    max-width: none;
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

  .topbar-actions {
    justify-content: flex-start;
    overflow-x: visible;
  }
}
</style>
