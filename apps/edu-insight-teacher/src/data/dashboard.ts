export type NavItem = {
  id: string
  label: string
  hint: string
}

export type WorkContext = {
  id: string
  roleLabel: string
  primaryLabel: string
  secondaryLabel: string
  description: string
}

export type SubjectScopeClass = {
  id: string
  label: string
}

export type SelectionScenario = {
  id: string
  label: string
  description: string
}

export type SummaryMetric = {
  label: string
  value: string
  note: string
}

export type ClassBaseField = {
  label: string
  value: string
}

export type ClassMember = {
  id: string
  studentNo: string
  name: string
  gender: string
  combination: string
  electiveSubjects: string[]
  parentMobile: string
  status: string
  parentStatus: string
  selectionStatus: string
  profileStatus: 'ready' | 'missing_parent' | 'missing_selection' | 'pending'
}

export type TeacherAssignment = {
  id: string
  subject: string
  teacher: string
  classes: string
  syncStatus: string
  accountStatus: 'bound' | 'pending'
  permissionStatus: 'synced' | 'pending'
}

export type PolicyItem = {
  id: string
  title: string
  value: string
  note: string
}

export type RosterInsight = {
  title: string
  count: string
  detail: string
}

export type ExamRecord = {
  id: string
  name: string
  type: string
  date: string
  subjectCoverage: string
  subjects?: string[]
  importStatus: string
}

export type ImportIssue = {
  id: string
  row: string
  student: string
  issue: string
  suggestion: string
  status: string
}

export type ImportBatchMetric = {
  label: string
  value: string
  note: string
}

export type ScoreFlowStep = {
  title: string
  description: string
  status: string
}

export type ScoreDecision = {
  title: string
  result: string
  detail: string
}

export type UploadValidationItem = {
  field: string
  result: string
  note: string
}

export type ScoreTableRow = {
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
}

export type StudentTrend = {
  name: string
  totalScore: string
  delta: string
  tag: string
}

export type CohortInsight = {
  title: string
  students: string
  insight: string
}

export type AlertItem = {
  student: string
  subject: string
  level: string
  detail: string
}

export type SyncAudienceCard = {
  audience: string
  readiness: string
  note: string
}

export type SyncRecord = {
  target: string
  channel: string
  status: string
  time: string
}

export const navItems: NavItem[] = [
  { id: 'overview', label: '工作台', hint: '总览与待办' },
  { id: 'classes', label: '班级与师生', hint: '学生档案与任课老师' },
  { id: 'scores', label: '考试与成绩', hint: '考试、导入、校验' },
  { id: 'analysis', label: '分析与预警', hint: '趋势、分层、风险' },
  { id: 'sync', label: '同步中心', hint: '家长与学生触达' }
]

export const workContexts: WorkContext[] = [
  {
    id: 'homeroom-g2-3',
    roleLabel: '班主任',
    primaryLabel: '高二（3）班',
    secondaryLabel: '英语任课',
    description: '查看全班、全科、预警与家校同步'
  },
  {
    id: 'subject-math',
    roleLabel: '任课老师',
    primaryLabel: '数学学科',
    secondaryLabel: '2 个任课班级',
    description: '只看当前班数学成绩、波动与薄弱学生'
  }
]

export const subjectScopeClasses: SubjectScopeClass[] = [
  { id: 'g2-3', label: '高二（3）班' },
  { id: 'g1-8', label: '高一（8）班' }
]

export const selectionScenarios: SelectionScenario[] = [
  {
    id: 'pre-selection',
    label: '未选科阶段',
    description: '适用于高一前期或尚未完成选科分流的班级，以行政班和通用科目分析为主。'
  },
  {
    id: 'post-selection',
    label: '已选科阶段',
    description: '适用于完成选科后的班级，支持行政班、教学班、选科组合和赋分分析。'
  }
]

export const currentClassStage: SelectionScenario = selectionScenarios[1]

export const summaryMetrics: SummaryMetric[] = [
  { label: '当前班级', value: '高二（3）班', note: '行政班 / 48 人' },
  { label: '最近考试', value: '6 月月考', note: '已完成分析' },
  { label: '重点预警', value: '7 人', note: '连续下滑或单科异常' },
  { label: '待同步家长', value: '18 条', note: '包含 4 条失败重试' }
]

export const classBaseFields: ClassBaseField[] = [
  { label: '学校', value: '星河高级中学' },
  { label: '年级', value: '高二' },
  { label: '行政班', value: '高二（3）班' },
  { label: '班主任', value: '李老师' },
  { label: '学年', value: '2025-2026 学年' },
  { label: '班级状态', value: '正常使用中' }
]

export const classMembers: ClassMember[] = [
  {
    id: 'student-g230301',
    studentNo: 'G230301',
    name: '林书言',
    gender: '女',
    combination: '物化生',
    electiveSubjects: ['物理', '化学', '生物'],
    parentMobile: '138****3201',
    status: '已完整',
    parentStatus: '已绑定',
    selectionStatus: '已登记',
    profileStatus: 'ready'
  },
  {
    id: 'student-g230302',
    studentNo: 'G230302',
    name: '许一诺',
    gender: '男',
    combination: '物化生',
    electiveSubjects: ['物理', '化学', '生物'],
    parentMobile: '139****1882',
    status: '已完整',
    parentStatus: '已绑定',
    selectionStatus: '已登记',
    profileStatus: 'ready'
  },
  {
    id: 'student-g230317',
    studentNo: 'G230317',
    name: '陈可心',
    gender: '女',
    combination: '史地政',
    electiveSubjects: ['历史', '地理', '政治'],
    parentMobile: '137****5210',
    status: '待补家长',
    parentStatus: '待补充',
    selectionStatus: '已登记',
    profileStatus: 'missing_parent'
  },
  {
    id: 'student-g230329',
    studentNo: 'G230329',
    name: '赵博文',
    gender: '男',
    combination: '物化生',
    electiveSubjects: ['物理', '化学', '生物'],
    parentMobile: '136****9913',
    status: '选科待确认',
    parentStatus: '已绑定',
    selectionStatus: '待确认',
    profileStatus: 'missing_selection'
  }
]

export const teacherAssignments: TeacherAssignment[] = [
  {
    id: 'teacher-chinese-zhang',
    subject: '语文',
    teacher: '张老师',
    classes: '高二（3）班、高二（5）班',
    syncStatus: '已同步',
    accountStatus: 'bound',
    permissionStatus: 'synced'
  },
  {
    id: 'teacher-math-wang',
    subject: '数学',
    teacher: '王老师',
    classes: '高二（3）班、高一（8）班',
    syncStatus: '已同步',
    accountStatus: 'bound',
    permissionStatus: 'synced'
  },
  {
    id: 'teacher-english-li',
    subject: '英语',
    teacher: '李老师',
    classes: '高二（3）班',
    syncStatus: '班主任本人',
    accountStatus: 'bound',
    permissionStatus: 'synced'
  },
  {
    id: 'teacher-chemistry-zhao',
    subject: '化学',
    teacher: '赵老师',
    classes: '高二（3）班教学班',
    syncStatus: '待补账号绑定',
    accountStatus: 'pending',
    permissionStatus: 'pending'
  }
]

export const displayPolicies: PolicyItem[] = [
  {
    id: 'policy-parent-class-rank',
    title: '家长端班级位置',
    value: '班主任可配置',
    note: '当前班级允许展示班级位置，不允许展示年级排名。'
  },
  {
    id: 'policy-student-band',
    title: '学生端分数段',
    value: '已开启',
    note: '班主任已开启班级分数段和班级均分对比。'
  },
  {
    id: 'policy-sync-trigger',
    title: '同步策略',
    value: '考试发布后触发',
    note: '成绩分析完成后，再统一同步家长、学生与任课老师。'
  }
]

export const rosterInsights: RosterInsight[] = [
  {
    title: '已完成家长绑定',
    count: '46 / 48',
    detail: '2 位学生缺少有效家长手机号，影响成绩同步。'
  },
  {
    title: '已完成选科登记',
    count: '48 / 48',
    detail: '当前班级所有学生均已完成选科组合登记。'
  },
  {
    title: '任课老师已绑定账号',
    count: '3 / 4',
    detail: '化学老师尚未绑定账号，暂无法自动同步单科分析。'
  }
]

export const examRecords: ExamRecord[] = [
  {
    id: 'exam-june-monthly',
    name: '2026 年 6 月月考',
    type: '月考',
    date: '2026-06-08',
    subjectCoverage: '语数英 + 选考科',
    importStatus: '已完成'
  },
  {
    id: 'exam-midterm',
    name: '2026 年期中考试',
    type: '期中',
    date: '2026-05-12',
    subjectCoverage: '全科',
    importStatus: '已完成'
  },
  {
    id: 'exam-weekly-07',
    name: '2026 年 4 月周测 07',
    type: '周测',
    date: '2026-04-18',
    subjectCoverage: '语数英',
    importStatus: '已归档'
  }
]

export const importIssues: ImportIssue[] = [
  {
    id: 'issue-12',
    row: '12',
    student: '陈可心 / 化学',
    issue: '学生选科中未登记化学',
    suggestion: '检查选科档案或确认科目列是否误填',
    status: '待处理'
  },
  {
    id: 'issue-24',
    row: '24',
    student: '周子昂 / 生物',
    issue: '赋分缺失',
    suggestion: '若本次为选考统计，请补充赋分字段',
    status: '待处理'
  },
  {
    id: 'issue-31',
    row: '31',
    student: '高泽宇 / 数学',
    issue: '实际分超过上限',
    suggestion: '核对原始表格分值或满分配置',
    status: '待处理'
  }
]

export const importBatchMetrics: ImportBatchMetric[] = [
  {
    label: '导入记录数',
    value: '286',
    note: '覆盖语数英和已开放选考科'
  },
  {
    label: '自动匹配成功',
    value: '283',
    note: '学号与姓名双匹配成功'
  },
  {
    label: '异常条目',
    value: '3',
    note: '待人工修复后重新校验'
  }
]

export const scoreFlowSteps: ScoreFlowStep[] = [
  {
    title: '1. 填写考试信息',
    description: '在导入成绩时一并填写考试名称、类型、时间和学科覆盖。',
    status: '不再单独维护创建考试动作'
  },
  {
    title: '2. 上传成绩',
    description: '支持学号、姓名、科目、实际分、赋分、缺考。',
    status: '已支持在抽屉内录入'
  },
  {
    title: '3. 校验修复',
    description: '处理选科冲突、赋分缺失、异常分值等问题。',
    status: '支持异常修复与状态更新'
  },
  {
    title: '4. 查看与修改',
    description: '按考试查看全班成绩表，并支持逐行修改成绩。',
    status: '当前原型已支持'
  }
]

export const scoreDecisions: ScoreDecision[] = [
  {
    title: '总分统计口径',
    result: '已采用赋分口径',
    detail: '当前班级处于已选科阶段，选考科纳入赋分统计。'
  },
  {
    title: '异常记录处理',
    result: '3 条待人工确认',
    detail: '系统已定位到选科冲突、赋分缺失和异常分值问题。'
  },
  {
    title: '同步触发条件',
    result: '分析完成后触发',
    detail: '家长与学生同步基于最终分析结果，不在导入成功时立即发送。'
  }
]

export const uploadValidationItems: UploadValidationItem[] = [
  {
    field: '学号匹配',
    result: '46 / 48 成功',
    note: '2 条记录需班主任确认学生学号或姓名。'
  },
  {
    field: '学科匹配',
    result: '全部通过',
    note: '当前勾选学科与导入表头一致。'
  },
  {
    field: '赋分字段',
    result: '1 条缺失',
    note: '已选科阶段下，政治科目缺少赋分。'
  }
]

export const scoreEntries: ScoreTableRow[] = [
  {
    id: 'score-row-1',
    studentId: 'student-g230301',
    studentName: '林书言',
    chinese: '121',
    math: '128',
    english: '136',
    electiveLabel: '物化生',
    electiveScore: '278',
    subjectScores: { 语文: '121', 数学: '128', 英语: '136', 物理: '94', 化学: '91', 生物: '93' },
    total: '663'
  },
  {
    id: 'score-row-2',
    studentId: 'student-g230302',
    studentName: '许一诺',
    chinese: '116',
    math: '122',
    english: '129',
    electiveLabel: '物化生',
    electiveScore: '266',
    subjectScores: { 语文: '116', 数学: '122', 英语: '129', 物理: '89', 化学: '88', 生物: '89' },
    total: '633'
  },
  {
    id: 'score-row-3',
    studentId: 'student-g230317',
    studentName: '陈可心',
    chinese: '113',
    math: '104',
    english: '125',
    electiveLabel: '史地政',
    electiveScore: '249',
    subjectScores: { 语文: '113', 数学: '104', 英语: '125', 历史: '84', 地理: '82', 政治: '83' },
    total: '591'
  },
  {
    id: 'score-row-4',
    studentId: 'student-g230329',
    studentName: '赵博文',
    chinese: '108',
    math: '96',
    english: '118',
    electiveLabel: '物化生',
    electiveScore: '241',
    subjectScores: { 语文: '108', 数学: '96', 英语: '118', 物理: '82', 化学: '80', 生物: '79' },
    total: '563'
  }
]

export const cohortInsights: CohortInsight[] = [
  {
    title: '尖子生群体',
    students: '8 人',
    insight: '总分稳定，高分段集中在物化生组合，适合持续跟踪拔高。'
  },
  {
    title: '临界生群体',
    students: '11 人',
    insight: '数学和英语波动最明显，适合班主任和单科老师联合关注。'
  },
  {
    title: '待帮扶群体',
    students: '7 人',
    insight: '连续下滑与偏科叠加，建议优先同步家长并建立后续跟踪。'
  }
]

export const studentTrends: StudentTrend[] = [
  { name: '林书言', totalScore: '612', delta: '+19', tag: '持续进步' },
  { name: '许一诺', totalScore: '584', delta: '+7', tag: '稳定提升' },
  { name: '赵博文', totalScore: '468', delta: '-16', tag: '重点关注' },
  { name: '孙嘉禾', totalScore: '521', delta: '-8', tag: '临界波动' }
]

export const alertItems: AlertItem[] = [
  {
    student: '刘星宇',
    subject: '数学',
    level: '高',
    detail: '连续两次下滑，较上次下降 18 分'
  },
  {
    student: '陈可心',
    subject: '英语',
    level: '中',
    detail: '班级排名下滑 9 位，作文部分异常低分'
  },
  {
    student: '周子昂',
    subject: '化学',
    level: '高',
    detail: '选考赋分波动明显，需联合单科老师跟进'
  },
  {
    student: '赵博文',
    subject: '总分',
    level: '中',
    detail: '总分连续两次低于班级中位线'
  }
]

export const syncAudienceCards: SyncAudienceCard[] = [
  {
    audience: '家长',
    readiness: '46 / 48 已可触达',
    note: '2 位家长手机号待补充，当前不影响其余家长同步。'
  },
  {
    audience: '学生',
    readiness: '48 / 48 已可触达',
    note: '学生侧已全部绑定，可按考试结果统一发起同步。'
  },
  {
    audience: '任课老师',
    readiness: '3 / 4 已可查看',
    note: '化学老师尚未完成账号绑定，当前只能由班主任代查看。'
  }
]

export const syncRecords: SyncRecord[] = [
  {
    target: '家长（高二 3 班）',
    channel: '小程序订阅消息',
    status: '已完成 44 / 48',
    time: '2026-06-14 09:20'
  },
  {
    target: '学生（高二 3 班）',
    channel: '小程序订阅消息',
    status: '已完成 46 / 48',
    time: '2026-06-14 09:23'
  },
  {
    target: '单科老师（数学、英语）',
    channel: '站内消息',
    status: '已完成',
    time: '2026-06-14 09:26'
  }
]
