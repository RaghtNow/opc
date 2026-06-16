export type ExamItem = {
  id: string
  name: string
  type: string
  date: string
  subjectCoverage: string
  subjects?: string[]
  importStatus: string
}

export type ScoreRow = {
  id: string
  studentId: string
  studentName: string
  chinese: string
  math: string
  english: string
  electiveLabel: string
  electiveScore: string
  total: string
}

export type ImportIssue = {
  id: string
  row: string
  student: string
  issue: string
  suggestion: string
  status: string
}

const API_BASE = 'http://127.0.0.1:8088/api'

export async function fetchExams(): Promise<ExamItem[]> {
  const response = await fetch(`${API_BASE}/exams`)
  if (!response.ok) throw new Error('获取考试列表失败')
  const data = await response.json()
  return data.items ?? []
}

export async function fetchExamDetail(examID: string): Promise<{
  exam: ExamItem
  scores: ScoreRow[]
  issues: ImportIssue[]
}> {
  const response = await fetch(`${API_BASE}/exams/${examID}`)
  if (!response.ok) throw new Error('获取考试详情失败')
  return response.json()
}

export async function importExam(payload: {
  name: string
  type: string
  date: string
  subjects: string[]
  subjectCoverage: string
  fileName: string
  scores: ScoreRow[]
  issues: ImportIssue[]
}) {
  const response = await fetch(`${API_BASE}/exams/import`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  if (!response.ok) throw new Error('导入考试失败')
  return response.json()
}

export async function updateExamScore(
  examID: string,
  scoreID: string,
  payload: {
    chinese: string
    math: string
    english: string
    electiveLabel: string
    electiveScore: string
    total: string
  }
) {
  const response = await fetch(`${API_BASE}/exams/${examID}/scores/${scoreID}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  if (!response.ok) throw new Error('修改成绩失败')
  return response.json()
}
