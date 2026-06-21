export type ParsedCsvRow = Record<string, string>

export type CsvValidationResult = {
  ok: boolean
  rows: ParsedCsvRow[]
  headers: string[]
  issues: Array<{
    id: string
    row: string
    student: string
    issue: string
    suggestion: string
    status: string
  }>
  metrics: {
    total: number
    matched: number
    issues: number
  }
  validationSummary: Array<{
    field: string
    result: string
    note: string
  }>
  scoreRows: Array<{
    id: string
    studentId: string
    studentName: string
    chinese?: string
    math?: string
    english?: string
    electiveLabel?: string
    electiveScore?: string
    subjectScores: Record<string, string>
    total: string
  }>
  error?: string
}

const API_BASE = import.meta.env.VITE_EDU_INSIGHT_API_BASE ?? 'http://127.0.0.1:8088/api'

export async function parseAndValidateCsv(params: {
  file: File
  selectedSubjects: string[]
  stage: 'pre-selection' | 'post-selection'
  knownStudentIds: string[]
  knownStudentNames: string[]
}): Promise<CsvValidationResult> {
  try {
    const formData = new FormData()
    formData.append('file', params.file)
    formData.append('subjects', params.selectedSubjects.join(','))

    const response = await fetch(`${API_BASE}/exams/import/validate`, {
      method: 'POST',
      body: formData
    })

    if (!response.ok) throw new Error('后端校验失败')
    return response.json()
  } catch (error) {
    return {
      ok: false,
      rows: [],
      headers: [],
      issues: [],
      metrics: { total: 0, matched: 0, issues: 0 },
      validationSummary: [],
      scoreRows: [],
      error: error instanceof Error ? error.message : 'CSV 解析失败'
    }
  }
}
