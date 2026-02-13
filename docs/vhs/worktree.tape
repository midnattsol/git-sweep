# git sweep worktree - broken worktree cleanup demo
# Run: vhs worktree.tape

Output worktree.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "# Check worktree status"
Enter
Sleep 500ms

Type "git sweep worktree"
Enter
Sleep 3s

Type "# Clean up broken worktrees"
Enter
Sleep 500ms

Type "git sweep worktree --clean"
Enter
Sleep 2s

Enter
Sleep 2s
