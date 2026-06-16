import type { ClassBaseField, ClassMember, PolicyItem, RosterInsight, SelectionScenario, TeacherAssignment } from '../data/dashboard'

const API_BASE = import.meta.env.VITE_EDU_INSIGHT_API_BASE ?? 'http://127.0.0.1:8088/api'

export type ClassroomWorkspace = {
  classId: string
  className: string
  stage: SelectionScenario
  baseFields: ClassBaseField[]
  rosterInsights: RosterInsight[]
  students: ClassMember[]
  teachers: TeacherAssignment[]
  policies: PolicyItem[]
}

export type SaveStudentPayload = {
  studentNo: string
  name: string
  gender: string
  combination: string
  parentMobile: string
  parentStatus: string
  selectionStatus: string
}

export type SaveTeacherPayload = {
  subject: string
  teacher: string
  mobile: string
  classes: string
}

export type SavePolicyPayload = {
  value: string
  note: string
}

export type ImportSummary = {
  created: number
  updated: number
  skipped: number
  errors: string[]
}

export async function fetchClassroomWorkspace(): Promise<ClassroomWorkspace> {
  const response = await fetch(`${API_BASE}/classes/current`)
  if (!response.ok) throw new Error('获取班级与师生数据失败')
  return response.json()
}

export async function createStudent(payload: SaveStudentPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace('/classes/current/students', 'POST', payload)
}

export async function updateStudent(studentID: string, payload: SaveStudentPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/students/${studentID}`, 'PATCH', payload)
}

export async function createTeacher(payload: SaveTeacherPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace('/classes/current/teachers', 'POST', payload)
}

export async function updateTeacher(teacherID: string, payload: SaveTeacherPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/teachers/${teacherID}`, 'PATCH', payload)
}

export async function bindTeacherAccount(teacherID: string): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/teachers/${teacherID}/bind-account`, 'POST')
}

export async function syncTeacherPermission(teacherID: string): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/teachers/${teacherID}/sync-permission`, 'POST')
}

export async function importStudents(file: File): Promise<{ workspace: ClassroomWorkspace; summary: ImportSummary }> {
  return importFile('/classes/current/students/import', file)
}

export async function importTeachers(file: File): Promise<{ workspace: ClassroomWorkspace; summary: ImportSummary }> {
  return importFile('/classes/current/teachers/import', file)
}

export async function updatePolicy(policyID: string, payload: SavePolicyPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/policies/${policyID}`, 'PATCH', payload)
}

async function requestWorkspace(path: string, method: 'POST' | 'PATCH', payload?: unknown): Promise<ClassroomWorkspace> {
  const response = await fetch(`${API_BASE}${path}`, {
    method,
    headers: {
      'Content-Type': 'application/json'
    },
    body: payload === undefined ? undefined : JSON.stringify(payload)
  })

  if (!response.ok) throw new Error('保存班级与师生数据失败')
  return response.json()
}

async function importFile(path: string, file: File): Promise<{ workspace: ClassroomWorkspace; summary: ImportSummary }> {
  const formData = new FormData()
  formData.append('file', file)
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    body: formData
  })

  if (!response.ok) {
    const data = await response.json().catch(() => ({}))
    throw new Error(data.error ?? data.message ?? '导入失败')
  }
  return response.json()
}
