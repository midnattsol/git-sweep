package sweep

import (
	"time"

	"github.com/midnattsol/git-sweep/internal/config"
	"github.com/midnattsol/git-sweep/internal/git"
)

// Category represents the branch classification used in interactive cleanup.
type Category string

const (
	CategorySuggested Category = "suggested" // Stale, no upstream, no PR - safe to delete
	CategoryGone      Category = "gone"      // Upstream gone but no merged PR found
	CategoryOrphan    Category = "orphan"    // No upstream but recent commits
	CategoryActive    Category = "active"    // Has upstream that exists
	CategoryProtected Category = "protected" // Protected or current branch
)

// BranchResult represents the analysis result for a branch
type BranchResult struct {
	Branch    git.Branch
	Skip      git.SkipReason
	Category  Category
	Eligible  bool // Has merged PR - safe to delete
	Suggested bool
}

// Stats holds sweep statistics
type Stats struct {
	Total     int
	Eligible  int
	Deleted   int
	Skipped   int
	Suggested int
}

// Result contains the full sweep result
type Result struct {
	Branches []BranchResult
	Stats    Stats
}

// AnalyzeBranches returns all branches categorized for interactive cleanup.
// If mergedPRs is provided, branches with merged PRs are marked as Suggested.
func AnalyzeBranches(cfg *config.Config, mergedPRs map[string]bool) (*Result, error) {
	branches, err := git.ListBranches(cfg.Remote)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Branches: make([]BranchResult, 0, len(branches)),
	}

	staleThreshold := time.Now().AddDate(0, 0, -cfg.StaleDays)

	for _, b := range branches {
		br := BranchResult{Branch: b}
		result.Stats.Total++

		isStale := !b.LastCommit.IsZero() && b.LastCommit.Before(staleThreshold)
		hasMergedPR := mergedPRs != nil && mergedPRs[b.Name]

		switch {
		case b.IsCurrent:
			br.Skip = git.SkipCurrent
			br.Category = CategoryProtected
		case cfg.IsProtected(b.Name):
			br.Skip = git.SkipProtected
			br.Category = CategoryProtected
		case hasMergedPR:
			// Branch has a merged PR - always suggest deletion
			br.Eligible = true
			br.Suggested = true
			br.Category = CategorySuggested
			result.Stats.Eligible++
			result.Stats.Suggested++
		case b.UpstreamGone || (b.Upstream != "" && !git.UpstreamExists(b.Upstream)):
			// Upstream is gone (no merged PR found) - suggest deletion
			br.Eligible = true
			br.Suggested = true
			br.Category = CategoryGone
			result.Stats.Eligible++
			result.Stats.Suggested++
		case b.Upstream == "" && isStale:
			// No upstream and stale - suggest deletion
			br.Eligible = true
			br.Suggested = true
			br.Category = CategorySuggested
			result.Stats.Eligible++
			result.Stats.Suggested++
		case b.Upstream == "":
			// No upstream but recent - orphan
			br.Eligible = true
			br.Category = CategoryOrphan
			result.Stats.Eligible++
		default:
			// Has active upstream
			br.Eligible = true
			br.Category = CategoryActive
			result.Stats.Eligible++
		}

		if br.Skip != git.SkipNone {
			result.Stats.Skipped++
		}

		result.Branches = append(result.Branches, br)
	}

	return result, nil
}

// DeleteBranches deletes specific branches by name
func DeleteBranches(names []string) (deleted int, errors []error) {
	for _, name := range names {
		if err := git.DeleteBranch(name); err != nil {
			errors = append(errors, err)
		} else {
			deleted++
		}
	}
	return
}
