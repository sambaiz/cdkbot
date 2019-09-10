package constant

// State is operation state
type State int

const (
	// StateRunning expresses operation is running
	StateRunning State = iota + 1
	// StateMergeReady expresses the PR is ready to merge
	StateMergeReady
	// StateNotMergeReady expresses the PR is not ready to merge
	StateNotMergeReady
	// StateError expresses error occurred
	StateError
)
