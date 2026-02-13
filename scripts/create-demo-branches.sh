#!/bin/bash
# Creates demo branches for VHS recordings
# These branches will appear as "eligible" in git-sweep (merged PR + upstream gone)

set -e

BRANCHES=(
  "feature/user-auth"
  "fix/login-bug"
  "refactor/cleanup"
)

BASE_BRANCH="develop"

echo "Creating demo branches for git-sweep VHS recording..."
echo ""

# Make sure we're on base branch and up to date
git checkout $BASE_BRANCH
git pull

for branch in "${BRANCHES[@]}"; do
  echo "→ Creating $branch..."
  
  # Create branch with a dummy commit
  git checkout -b "$branch"
  echo "Demo commit for $branch" > ".demo-${branch//\//-}.txt"
  git add .
  git commit -m "demo: add test file for $branch"
  
  # Push to remote
  git push -u origin "$branch"
  
  # Create and merge PR
  echo "  Creating PR..."
  gh pr create --title "Demo: $branch" --body "Demo branch for VHS recording" --base $BASE_BRANCH
  
  echo "  Merging PR..."
  gh pr merge --merge --delete-branch
  
  # Back to base
  git checkout $BASE_BRANCH
  git pull
  
  echo "  ✓ Done"
  echo ""
done

# Prune to detect gone upstreams
git fetch --prune

echo ""
echo "✓ All demo branches created!"
echo ""
echo "Now you should see these as 'eligible' when running:"
echo "  git sweep --dry-run"
echo ""
echo "To clean up after recording:"
echo "  git sweep --force"
