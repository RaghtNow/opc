package insight

type SummaryMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type StudentTrend struct {
	Name       string `json:"name"`
	TotalScore string `json:"totalScore"`
	Delta      string `json:"delta"`
	Tag        string `json:"tag"`
}

type CohortInsight struct {
	Title    string `json:"title"`
	Students string `json:"students"`
	Insight  string `json:"insight"`
}

type AlertItem struct {
	Student string `json:"student"`
	Subject string `json:"subject"`
	Level   string `json:"level"`
	Detail  string `json:"detail"`
}

type SyncAudienceCard struct {
	Audience  string `json:"audience"`
	Readiness string `json:"readiness"`
	Note      string `json:"note"`
}

type SyncRecord struct {
	Target  string `json:"target"`
	Channel string `json:"channel"`
	Status  string `json:"status"`
	Time    string `json:"time"`
}

type ExamMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type SubjectDiagnostic struct {
	Subject        string  `json:"subject"`
	Average        float64 `json:"average"`
	Highest        float64 `json:"highest"`
	Lowest         float64 `json:"lowest"`
	ExcellentCount int     `json:"excellentCount"`
	PassCount      int     `json:"passCount"`
	LowCount       int     `json:"lowCount"`
	BarWidth       float64 `json:"barWidth"`
	RiskLabel      string  `json:"riskLabel"`
}

type ScoreBand struct {
	Label   string `json:"label"`
	Range   string `json:"range"`
	Count   int    `json:"count"`
	Percent int    `json:"percent"`
	Width   int    `json:"width"`
	Tone    string `json:"tone"`
}

type LayerGroup struct {
	Title    string `json:"title"`
	Count    string `json:"count"`
	Students string `json:"students"`
	Goal     string `json:"goal"`
}

type RiskStudent struct {
	Name         string  `json:"name"`
	Total        string  `json:"total"`
	Gap          float64 `json:"gap"`
	WeakSubjects string  `json:"weakSubjects"`
	Level        string  `json:"level"`
	Reason       string  `json:"reason"`
}

type TeachingAction struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Tag    string `json:"tag"`
}

type TrendPoint struct {
	ExamID   string  `json:"examId"`
	ExamName string  `json:"examName"`
	Date     string  `json:"date"`
	Value    float64 `json:"value"`
}

type SubjectTrend struct {
	Subject string       `json:"subject"`
	Points  []TrendPoint `json:"points"`
}

type StudentAnalysis struct {
	StudentID      string         `json:"studentId"`
	StudentName    string         `json:"studentName"`
	TotalTrend     []TrendPoint   `json:"totalTrend"`
	RankTrend      []TrendPoint   `json:"rankTrend"`
	SubjectTrends  []SubjectTrend `json:"subjectTrends"`
	LatestRank     int            `json:"latestRank"`
	LatestTotal    string         `json:"latestTotal"`
	WeakSubjects   string         `json:"weakSubjects"`
	Recommendation string         `json:"recommendation"`
}

type ClassComparison struct {
	ClassID      string  `json:"classId"`
	ClassName    string  `json:"className"`
	StudentCount int     `json:"studentCount"`
	Average      float64 `json:"average"`
	Highest      float64 `json:"highest"`
	Lowest       float64 `json:"lowest"`
	RiskCount    int     `json:"riskCount"`
	ExamName     string  `json:"examName"`
}

type AnalysisDashboard struct {
	ExamMetrics        []ExamMetric        `json:"examMetrics"`
	SubjectDiagnostics []SubjectDiagnostic `json:"subjectDiagnostics"`
	ScoreBands         []ScoreBand         `json:"scoreBands"`
	LayerGroups        []LayerGroup        `json:"layerGroups"`
	RiskStudents       []RiskStudent       `json:"riskStudents"`
	TeachingActions    []TeachingAction    `json:"teachingActions"`
	ClassTrend         []TrendPoint        `json:"classTrend"`
	SubjectTrends      []SubjectTrend      `json:"subjectTrends"`
	StudentAnalyses    []StudentAnalysis   `json:"studentAnalyses"`
	ClassComparisons   []ClassComparison   `json:"classComparisons"`
}

type Dashboard struct {
	Scope             string             `json:"scope"`
	ScopeLabel        string             `json:"scopeLabel"`
	SourceClassIDs    []string           `json:"sourceClassIds"`
	SummaryMetrics    []SummaryMetric    `json:"summaryMetrics"`
	StudentTrends     []StudentTrend     `json:"studentTrends"`
	CohortInsights    []CohortInsight    `json:"cohortInsights"`
	AlertItems        []AlertItem        `json:"alertItems"`
	Analysis          AnalysisDashboard  `json:"analysis"`
	SyncAudienceCards []SyncAudienceCard `json:"syncAudienceCards"`
	SyncRecords       []SyncRecord       `json:"syncRecords"`
	LatestExamName    string             `json:"latestExamName"`
	CanPublish        bool               `json:"canPublish"`
	PublishBlockers   []string           `json:"publishBlockers"`
}
