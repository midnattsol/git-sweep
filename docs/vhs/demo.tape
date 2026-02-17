# git-sweep full demo - all commands overview
# Run: vhs demo.tape

Output demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 20

# Title
Type "# git-sweep: all-in-one git cleanup"
Enter
Sleep 1s

# Branch cleanup
Type "git sweep --yes"
Enter
Sleep 4s

# Stash
Type "git sweep stash"
Enter
Sleep 3s

# Tags
Type "git sweep tags"
Enter
Sleep 3s

# Worktree
Type "git sweep worktree"
Enter
Sleep 3s

# Done
Type "# Done! Run with --clean to clean up"
Enter
Sleep 2s
