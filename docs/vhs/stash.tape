# git sweep stash - stash cleanup demo
# Run: vhs stash.tape

Output stash.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20

Type "# List stashes with context"
Enter
Sleep 500ms

Type "git sweep stash"
Enter
Sleep 3s

Type "# Clean up old stashes interactively"
Enter
Sleep 500ms

Type "git sweep stash --clean"
Enter
Sleep 2s

# Navigate and select
Down
Sleep 300ms
Type " "
Sleep 300ms
Down
Sleep 300ms
Type " "
Sleep 500ms

Enter
Sleep 2s
