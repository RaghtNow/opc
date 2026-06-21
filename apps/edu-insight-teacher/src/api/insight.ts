import type { AlertItem, CohortInsight, StudentTrend, SummaryMetric, SyncAudienceCard, SyncRecord } from '../data/dashboard'

const API_BASE = import.meta.env.VITE_EDU_INSIGHT_API_BASE ?? 'http://127.0.0.1:8088/api'

export type InsightDashboard = {
  scope: string
  scopeLabel: string
  sourceClassIds: string[]
  summaryMetrics: SummaryMetric[]
  studentTrends: StudentTrend[]
  cohortInsights: CohortInsight[]
  alertItems: AlertItem[]
  analysis: AnalysisDashboard
  syncAudienceCards: SyncAudienceCard[]
  syncRecords: SyncRecord[]
  latestExamName: string
  canPublish: boolean
  publishBlockers: string[]
}

export type AnalysisDashboard = {
  examMetrics: Array<{ label: string; value: string; note: string }>
  subjectDiagnostics: Array<{
    subject: string
    average: number
    highest: number
    lowest: number
    excellentCount: number
    passCount: number
    lowCount: number
    barWidth: number
    riskLabel: string
  }>
  scoreBands: Array<{ label: string; range: string; count: number; percent: number; width: number; tone: string }>
  layerGroups: Array<{ title: string; count: string; students: string; goal: string }>
  riskStudents: Array<{ name: string; total: string; gap: number; weakSubjects: string; level: string; reason: string }>
  teachingActions: Array<{ title: string; detail: string; tag: string }>
  classTrend: Array<{ examId: string; examName: string; date: string; value: number }>
  subjectTrends: Array<{ subject: string; points: Array<{ examId: string; examName: string; date: string; value: number }> }>
  studentAnalyses: Array<{
    studentId: string
    studentName: string
    totalTrend: Array<{ examId: string; examName: string; date: string; value: number }>
    rankTrend: Array<{ examId: string; examName: string; date: string; value: number }>
    subjectTrends: Array<{ subject: string; points: Array<{ examId: string; examName: string; date: string; value: number }> }>
    latestRank: number
    latestTotal: string
    weakSubjects: string
    recommendation: string
  }>
  classComparisons: Array<{
    classId: string
    className: string
    studentCount: number
    average: number
    highest: number
    lowest: number
    riskCount: number
    examName: string
  }>
}

export type InsightScope = {
  mode: 'single' | 'overall'
  classId: string
  classIds?: string[]
  examId?: string
}

export async function fetchInsightDashboard(scope: InsightScope): Promise<InsightDashboard> {
  const response = await fetch(withInsightScope('/insights/dashboard', scope))
  if (!response.ok) throw new Error('获取分析与同步数据失败')
  return response.json()
}

export async function publishLatestExam(classId: string): Promise<InsightDashboard> {
  const response = await fetch(withClassId('/sync/publish-latest', classId), {
    method: 'POST'
  })
  const data = await response.json()
  if (!response.ok) throw new Error(data.error ?? data.message ?? '发布失败')
  return data
}

function withClassId(path: string, classId: string) {
  const url = new URL(`${API_BASE}${path}`)
  if (classId) url.searchParams.set('classId', classId)
  return url.toString()
}

function withInsightScope(path: string, scope: InsightScope) {
  const url = new URL(`${API_BASE}${path}`)
  if (scope.classId) url.searchParams.set('classId', scope.classId)
  url.searchParams.set('scope', scope.mode)
  if (scope.examId) url.searchParams.set('examId', scope.examId)
  if (scope.mode === 'overall' && scope.classIds?.length) {
    url.searchParams.set('classIds', scope.classIds.join(','))
  }
  return url.toString()
}
