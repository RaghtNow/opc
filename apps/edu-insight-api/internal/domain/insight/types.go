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

type Dashboard struct {
	SummaryMetrics    []SummaryMetric    `json:"summaryMetrics"`
	StudentTrends     []StudentTrend     `json:"studentTrends"`
	CohortInsights    []CohortInsight    `json:"cohortInsights"`
	AlertItems        []AlertItem        `json:"alertItems"`
	SyncAudienceCards []SyncAudienceCard `json:"syncAudienceCards"`
	SyncRecords       []SyncRecord       `json:"syncRecords"`
	LatestExamName    string             `json:"latestExamName"`
	CanPublish        bool               `json:"canPublish"`
	PublishBlockers   []string           `json:"publishBlockers"`
}
