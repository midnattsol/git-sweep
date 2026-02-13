package sweep

import (
	"time"

	"github.com/midnattsol/git-sweep/internal/config"
	"github.com/midnattsol/git-sweep/internal/git"
	"github.com/midnattsol/git-sweep/internal/github"
)

// Category represents the category of a branch for nuke mode
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
	Eligible  bool
	Suggested bool
	Deleted   bool
	DeleteErr error
}

// Stats holds sweep statistics
type Stats struct {
	Total      int
	Candidates int
	Eligible   int
	Deleted    int
	Skipped    int
	Suggested  int
}

// Result contains the full sweep result
type Result struct {
	Branches []BranchResult
	Stats    Stats
}

// Analyze examines all branches and determines which are eligible for deletion
func Analyze(cfg *config.Config, mergedPRs map[string]bool) (*Result, error) {
	branches, err := git.ListBranches(cfg.Remote)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Branches: make([]BranchResult, 0, len(branches)),
	}

	for _, b := range branches {
		br := BranchResult{Branch: b}
		result.Stats.Total++

		// Check skip conditions
		switch {
		case b.IsCurrent:
			br.Skip = git.SkipCurrent
		case cfg.IsProtected(b.Name):
			br.Skip = git.SkipProtected
		case b.Upstream == "":
			br.Skip = git.SkipNoUpstream
		case !git.IsOnRemote(b.Upstream, cfg.Remote):
			br.Skip = git.SkipWrongRemote
		case !b.UpstreamGone && git.UpstreamExists(b.Upstream):
			br.Skip = git.SkipUpstreamExists
		default:
			// Upstream is gone, check for merged PR
			result.Stats.Candidates++
			if !github.HasMergedPR(b.Name, mergedPRs) {
				br.Skip = git.SkipNoPR
			} else {
				br.Eligible = true
				result.Stats.Eligible++
			}
		}

		if br.Skip != git.SkipNone {
			result.Stats.Skipped++
		}

		result.Branches = append(result.Branches, br)
	}

	return result, nil
}

// Execute deletes all eligible branches
func Execute(result *Result) {
	for i, br := range result.Branches {
		if !br.Eligible {
			continue
		}

		if err := git.DeleteBranch(br.Branch.Name); err != nil {
			result.Branches[i].DeleteErr = err
		} else {
			result.Branches[i].Deleted = true
			result.Stats.Deleted++
		}
	}
}

// AnalyzeForNuke returns all branches categorized for nuke mode
func AnalyzeForNuke(cfg *config.Config) (*Result, error) {
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

		switch {
		case b.IsCurrent:
			br.Skip = git.SkipCurrent
			br.Category = CategoryProtected
		case cfg.IsProtected(b.Name):
			br.Skip = git.SkipProtected
			br.Category = CategoryProtected
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
		case b.UpstreamGone || !git.UpstreamExists(b.Upstream):
			// Upstream is gone - suggest deletion
			br.Eligible = true
			br.Suggested = true
			br.Category = CategoryGone
			result.Stats.Eligible++
			result.Stats.Suggested++
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
