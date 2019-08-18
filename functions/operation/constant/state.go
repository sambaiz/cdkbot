package constant

// State is operation state
type State int

const (
	// StateRunning expresses operation is running
	StateRunning State = iota + 1
	// StateMergeReady expresses the PR is ready to merge
	StateMergeReady
	// StateNeedDeploy expresses the PR needs to deploy
	StateNeedDeploy
	// StateError expresses error occurred
	StateError
)
