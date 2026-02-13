package sweep

import (
	"github.com/midnattsol/git-sweep/internal/config"
	"github.com/midnattsol/git-sweep/internal/git"
	"github.com/midnattsol/git-sweep/internal/github"
)

// BranchResult represents the analysis result for a branch
type BranchResult struct {
	Branch    git.Branch
	Skip      git.SkipReason
	Eligible  bool
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

// AnalyzeForNuke returns all branches that can potentially be deleted (for nuke mode)
func AnalyzeForNuke(cfg *config.Config) (*Result, error) {
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

		switch {
		case b.IsCurrent:
			br.Skip = git.SkipCurrent
		case cfg.IsProtected(b.Name):
			br.Skip = git.SkipProtected
		default:
			br.Eligible = true
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
