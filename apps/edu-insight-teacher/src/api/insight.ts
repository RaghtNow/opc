import type { AlertItem, CohortInsight, StudentTrend, SummaryMetric, SyncAudienceCard, SyncRecord } from '../data/dashboard'

const API_BASE = import.meta.env.VITE_EDU_INSIGHT_API_BASE ?? 'http://127.0.0.1:8088/api'

export type InsightDashboard = {
  summaryMetrics: SummaryMetric[]
  studentTrends: StudentTrend[]
  cohortInsights: CohortInsight[]
  alertItems: AlertItem[]
  syncAudienceCards: SyncAudienceCard[]
  syncRecords: SyncRecord[]
  latestExamName: string
  canPublish: boolean
  publishBlockers: string[]
}

export async function fetchInsightDashboard(): Promise<InsightDashboard> {
  const response = await fetch(`${API_BASE}/insights/dashboard`)
  if (!response.ok) throw new Error('获取分析与同步数据失败')
  return response.json()
}

export async function publishLatestExam(): Promise<InsightDashboard> {
  const response = await fetch(`${API_BASE}/sync/publish-latest`, {
    method: 'POST'
  })
  const data = await response.json()
  if (!response.ok) throw new Error(data.error ?? data.message ?? '发布失败')
  return data
}
