package constant

// Label is label settings
type Label struct {
	Name        string
	Description string
	Color       string
}

var (
	// LabelOutdatedDiff expresses the PR has outdated differences
	LabelOutdatedDiff = Label{
		Name:        "cdkbot:outdated diff",
		Description: "Diffs are outdated. Run /diff again.",
		Color:       "e4e669",
	}
	// LabelNoDiff expresses the PR has no differences
	LabelNoDiff = Label{
		Name:        "cdkbot:no diffs",
		Description: "No diffs. Let's merge!",
		Color:       "008672",
	}
	// LabelDeploying expresses the PR is now being deployed
	LabelDeploying = Label{
		Name:        "cdkbot:deploying",
		Description: "Now deploying",
		Color:       "0075ca",
	}
	// NameToLabel is map of label's name to label
	NameToLabel = map[string]Label{
		LabelOutdatedDiff.Name: LabelOutdatedDiff,
		LabelNoDiff.Name:       LabelNoDiff,
		LabelDeploying.Name:    LabelDeploying,
	}
)
