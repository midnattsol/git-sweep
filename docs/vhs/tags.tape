# git sweep tags - orphan tags cleanup demo
# Run: vhs tags.tape

Output tags.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "# Find orphan tags (local without remote)"
Enter
Sleep 500ms

Type "git sweep tags"
Enter
Sleep 3s

Type "# Clean up orphan tags"
Enter
Sleep 500ms

Type "git sweep tags --clean"
Enter
Sleep 2s

# Confirm selection
Enter
Sleep 2s
