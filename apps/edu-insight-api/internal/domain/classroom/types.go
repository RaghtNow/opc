package classroom

type SelectionStage struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type BaseField struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Student struct {
	ID               string   `json:"id"`
	StudentNo        string   `json:"studentNo"`
	Name             string   `json:"name"`
	Gender           string   `json:"gender"`
	Combination      string   `json:"combination"`
	ElectiveSubjects []string `json:"electiveSubjects"`
	ParentMobile     string   `json:"parentMobile"`
	Status           string   `json:"status"`
	ParentStatus     string   `json:"parentStatus"`
	SelectionStatus  string   `json:"selectionStatus"`
	ProfileStatus    string   `json:"profileStatus"`
}

type TeacherAssignment struct {
	ID                 string `json:"id"`
	Subject            string `json:"subject"`
	Teacher            string `json:"teacher"`
	Mobile             string `json:"mobile"`
	Classes            string `json:"classes"`
	SyncStatus         string `json:"syncStatus"`
	AccountStatus      string `json:"accountStatus"`
	AccountID          string `json:"accountId"`
	AccountBoundAt     string `json:"accountBoundAt"`
	PermissionStatus   string `json:"permissionStatus"`
	PermissionSyncedAt string `json:"permissionSyncedAt"`
}

type Policy struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type RosterInsight struct {
	Title  string `json:"title"`
	Count  string `json:"count"`
	Detail string `json:"detail"`
}

type Workspace struct {
	ClassID        string              `json:"classId"`
	ClassName      string              `json:"className"`
	Stage          SelectionStage      `json:"stage"`
	BaseFields     []BaseField         `json:"baseFields"`
	RosterInsights []RosterInsight     `json:"rosterInsights"`
	Students       []Student           `json:"students"`
	Teachers       []TeacherAssignment `json:"teachers"`
	Policies       []Policy            `json:"policies"`
}

type SaveStudentRequest struct {
	StudentNo       string `json:"studentNo"`
	Name            string `json:"name"`
	Gender          string `json:"gender"`
	Combination     string `json:"combination"`
	ParentMobile    string `json:"parentMobile"`
	ParentStatus    string `json:"parentStatus"`
	SelectionStatus string `json:"selectionStatus"`
}

type SaveTeacherRequest struct {
	Subject          string `json:"subject"`
	Teacher          string `json:"teacher"`
	Mobile           string `json:"mobile"`
	Classes          string `json:"classes"`
	AccountStatus    string `json:"accountStatus"`
	PermissionStatus string `json:"permissionStatus"`
}

type SavePolicyRequest struct {
	Value string `json:"value"`
	Note  string `json:"note"`
}

type ImportSummary struct {
	Created int      `json:"created"`
	Updated int      `json:"updated"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors"`
}
