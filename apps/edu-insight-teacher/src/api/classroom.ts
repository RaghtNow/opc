import type { ClassBaseField, ClassMember, PolicyItem, RosterInsight, SelectionScenario, TeacherAssignment } from '../data/dashboard'

const API_BASE = import.meta.env.VITE_EDU_INSIGHT_API_BASE ?? 'http://127.0.0.1:8088/api'

export type ClassroomWorkspace = {
  classId: string
  className: string
  classOptions: ClassOption[]
  stage: SelectionScenario
  baseFields: ClassBaseField[]
  rosterInsights: RosterInsight[]
  students: ClassMember[]
  teachers: TeacherAssignment[]
  policies: PolicyItem[]
}

export type ClassOption = {
  id: string
  className: string
  gradeName: string
  academicYear: string
  roleLabel: string
  primaryLabel: string
  secondaryLabel: string
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

export type CreateClassPayload = {
  schoolName: string
  gradeName: string
  className: string
  homeroomTeacher: string
  academicYear: string
  stageId: string
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

export async function fetchClassroomWorkspace(classId: string): Promise<ClassroomWorkspace> {
  const response = await fetch(withClassId('/classes/current', classId))
  if (!response.ok) throw new Error('获取班级与师生数据失败')
  return response.json()
}

export async function createClass(payload: CreateClassPayload): Promise<ClassroomWorkspace> {
  const response = await fetch(`${API_BASE}/classes`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(payload)
  })

  if (!response.ok) throw new Error('创建班级失败')
  return response.json()
}

export async function createStudent(classId: string, payload: SaveStudentPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace('/classes/current/students', classId, 'POST', payload)
}

export async function updateStudent(classId: string, studentID: string, payload: SaveStudentPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/students/${studentID}`, classId, 'PATCH', payload)
}

export async function deleteStudent(classId: string, studentID: string): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/students/${studentID}`, classId, 'DELETE')
}

export async function createTeacher(classId: string, payload: SaveTeacherPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace('/classes/current/teachers', classId, 'POST', payload)
}

export async function updateTeacher(classId: string, teacherID: string, payload: SaveTeacherPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/teachers/${teacherID}`, classId, 'PATCH', payload)
}

export async function deleteTeacher(classId: string, teacherID: string): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/teachers/${teacherID}`, classId, 'DELETE')
}

export async function bindTeacherAccount(classId: string, teacherID: string): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/teachers/${teacherID}/bind-account`, classId, 'POST')
}

export async function syncTeacherPermission(classId: string, teacherID: string): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/teachers/${teacherID}/sync-permission`, classId, 'POST')
}

export async function importStudents(classId: string, file: File): Promise<{ workspace: ClassroomWorkspace; summary: ImportSummary }> {
  return importFile('/classes/current/students/import', classId, file)
}

export async function importTeachers(classId: string, file: File): Promise<{ workspace: ClassroomWorkspace; summary: ImportSummary }> {
  return importFile('/classes/current/teachers/import', classId, file)
}

export async function updatePolicy(classId: string, policyID: string, payload: SavePolicyPayload): Promise<ClassroomWorkspace> {
  return requestWorkspace(`/classes/current/policies/${policyID}`, classId, 'PATCH', payload)
}

async function requestWorkspace(path: string, classId: string, method: 'POST' | 'PATCH' | 'DELETE', payload?: unknown): Promise<ClassroomWorkspace> {
  const response = await fetch(withClassId(path, classId), {
    method,
    headers: {
      'Content-Type': 'application/json'
    },
    body: payload === undefined ? undefined : JSON.stringify(payload)
  })

  if (!response.ok) throw new Error('保存班级与师生数据失败')
  return response.json()
}

async function importFile(path: string, classId: string, file: File): Promise<{ workspace: ClassroomWorkspace; summary: ImportSummary }> {
  const formData = new FormData()
  formData.append('file', file)
  const response = await fetch(withClassId(path, classId), {
    method: 'POST',
    body: formData
  })

  if (!response.ok) {
    const data = await response.json().catch(() => ({}))
    throw new Error(data.error ?? data.message ?? '导入失败')
  }
  return response.json()
}

function withClassId(path: string, classId: string) {
  const url = new URL(`${API_BASE}${path}`)
  if (classId) url.searchParams.set('classId', classId)
  return url.toString()
}
