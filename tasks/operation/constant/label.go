package constant

import "fmt"

// Label is label settings
type Label struct {
	Name        string
	Description string
	Color       string
}

const labelPrefix = "cdkbot:"

var (
	// LabelOutdatedDiff expresses the PR has outdated differences
	LabelOutdatedDiff = Label{
		Name:        fmt.Sprintf("%soutdated diffs", labelPrefix),
		Description: "Diffs are outdated. Run /diff again.",
		Color:       "e4e669",
	}
	// LabelRunning expresses command is running on the PR
	LabelRunning = Label{
		Name:        fmt.Sprintf("%srunning", labelPrefix),
		Description: "Now running",
		Color:       "0075ca",
	}
	// LabelDeployed expresses some stacks are deployed on the PR
	LabelDeployed = Label{
		Name:        fmt.Sprintf("%sdeployed", labelPrefix),
		Description: "Some stacks are deployed. Complete /deploy and merge, or /rollback them.",
		Color:       "a2eeef",
	}
	// NameToLabel is map of label's name to label
	NameToLabel = map[string]Label{
		LabelOutdatedDiff.Name: LabelOutdatedDiff,
		LabelRunning.Name:      LabelRunning,
		LabelDeployed.Name:     LabelDeployed,
	}
)
