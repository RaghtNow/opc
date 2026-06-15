import * as XLSX from 'xlsx'

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
    chinese: string
    math: string
    english: string
    electiveLabel: string
    electiveScore: string
    total: string
  }>
  error?: string
}

const REQUIRED_BASE_HEADERS = ['学号', '姓名']

function parseCsv(text: string): ParsedCsvRow[] {
  const lines = text
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)

  if (lines.length < 2) return []

  const headers = lines[0].split(',').map((item) => item.trim())

  return lines.slice(1).map((line) => {
    const cols = line.split(',').map((item) => item.trim())
    const row: ParsedCsvRow = {}
    headers.forEach((header, index) => {
      row[header] = cols[index] ?? ''
    })
    return row
  })
}

function parseWorkbook(file: File): Promise<ParsedCsvRow[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()

    reader.onload = (event) => {
      try {
        const data = event.target?.result
        if (!data) throw new Error('文件读取失败')

        const workbook = XLSX.read(data, { type: 'array' })
        const firstSheetName = workbook.SheetNames[0]
        const worksheet = workbook.Sheets[firstSheetName]
        const rows = XLSX.utils.sheet_to_json<Record<string, string>>(worksheet, {
          defval: ''
        })

        resolve(
          rows.map((row) => {
            const normalized: ParsedCsvRow = {}
            Object.entries(row).forEach(([key, value]) => {
              normalized[String(key).trim()] = String(value).trim()
            })
            return normalized
          })
        )
      } catch (error) {
        reject(error)
      }
    }

    reader.onerror = () => reject(new Error('文件读取失败'))
    reader.readAsArrayBuffer(file)
  })
}

export async function parseAndValidateCsv(params: {
  file: File
  selectedSubjects: string[]
  stage: 'pre-selection' | 'post-selection'
  knownStudentIds: string[]
  knownStudentNames: string[]
}): Promise<CsvValidationResult> {
  try {
    const fileName = params.file.name.toLowerCase()
    const rows = fileName.endsWith('.csv')
      ? parseCsv(await params.file.text())
      : fileName.endsWith('.xlsx') || fileName.endsWith('.xls')
        ? await parseWorkbook(params.file)
        : []

    if (!rows.length) {
      return {
        ok: false,
        rows: [],
        headers: [],
        issues: [],
        metrics: { total: 0, matched: 0, issues: 0 },
        validationSummary: [],
        scoreRows: [],
        error: '文件内容为空、格式无效，或当前格式暂不支持'
      }
    }

    const headers = Object.keys(rows[0])
    const missingBase = REQUIRED_BASE_HEADERS.filter((header) => !headers.includes(header))
    const missingSubjects = params.selectedSubjects.filter((subject) => !headers.includes(subject))

    const issues: CsvValidationResult['issues'] = []
    const scoreRows: CsvValidationResult['scoreRows'] = []

    rows.forEach((row, index) => {
      const studentNo = row['学号'] || ''
      const studentName = row['姓名'] || `第${index + 2}行学生`

      if (!params.knownStudentIds.includes(studentNo) && !params.knownStudentNames.includes(studentName)) {
        issues.push({
          id: `issue-student-${index}`,
          row: `${index + 2}`,
          student: `${studentName}`,
          issue: '学生未匹配到班级档案',
          suggestion: '检查学号或姓名是否与班级档案一致',
          status: '待处理'
        })
      }

      const chinese = row['语文'] || ''
      const math = row['数学'] || ''
      const english = row['英语'] || ''
      const electiveSubjects = ['物理', '化学', '生物', '历史', '地理', '政治'].filter((subject) =>
        params.selectedSubjects.includes(subject)
      )

      const electiveTotal = electiveSubjects.reduce((sum, subject) => {
        const value = Number(row[subject] || '0')
        return sum + (Number.isNaN(value) ? 0 : value)
      }, 0)

      const baseTotal = [chinese, math, english].reduce((sum, value) => {
        const num = Number(value || '0')
        return sum + (Number.isNaN(num) ? 0 : num)
      }, 0)

      scoreRows.push({
        id: `score-${studentNo || index}`,
        studentId: studentNo || `unknown-${index}`,
        studentName,
        chinese,
        math,
        english,
        electiveLabel: electiveSubjects.length ? electiveSubjects.join('') : '-',
        electiveScore: electiveSubjects.length ? `${electiveTotal}` : '-',
        total: `${baseTotal + electiveTotal}`
      })
    })

    const validationSummary: CsvValidationResult['validationSummary'] = [
      {
        field: '基础字段',
        result: missingBase.length ? '缺失' : '通过',
        note: missingBase.length ? `缺少：${missingBase.join('、')}` : '学号、姓名字段齐全'
      },
      {
        field: '学科字段',
        result: missingSubjects.length ? '缺失' : '通过',
        note: missingSubjects.length ? `缺少：${missingSubjects.join('、')}` : '已勾选学科全部存在'
      },
      {
        field: '学生匹配',
        result: `${rows.length - issues.length} / ${rows.length} 通过`,
        note: issues.length ? `${issues.length} 条学生档案未匹配` : '全部学生已成功匹配'
      }
    ]

    return {
      ok: missingBase.length === 0 && missingSubjects.length === 0,
      rows,
      headers,
      issues,
      metrics: {
        total: rows.length,
        matched: rows.length - issues.length,
        issues: issues.length
      },
      validationSummary,
      scoreRows
    }
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
