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
	// LabelNoDiff expresses the PR has no differences
	LabelNoDiff = Label{
		Name:        fmt.Sprintf("%sno diffs", labelPrefix),
		Description: "No diffs. Let's merge!",
		Color:       "008672",
	}
	// LabelRunning expresses operation is running on the PR
	LabelRunning = Label{
		Name:        fmt.Sprintf("%srunning", labelPrefix),
		Description: "Now running",
		Color:       "0075ca",
	}
	// NameToLabel is map of label's name to label
	NameToLabel = map[string]Label{
		LabelOutdatedDiff.Name: LabelOutdatedDiff,
		LabelNoDiff.Name:       LabelNoDiff,
		LabelRunning.Name:    LabelRunning,
	}
)
