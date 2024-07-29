package rest

type TaskIdStuct struct {
	Success bool
	TaskId  string
	ErrStr  string
}

type VideoTaskPayload struct {
	FileName string
	UserName string
	UserUuid string
	Widths   []string
	Matter   *Matter `json:"matter"`
}

type VideoTaskResult struct {
	Success   bool
	TaskId    string
	ResultStr string
	ErrStr    string
}

type VideoTaskQueryResult struct {
	VideoTaskPayload VideoTaskPayload `json:"videoTaskPayload"`
	VideoTaskResult  VideoTaskResult  `json:"videoTaskResult"`
}
