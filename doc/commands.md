
# Git Helper Commands

A collection of helpful Git commands to streamline your workflow.

## Table of Contents
- [Sync](#sync)
- [Sync Fork](#sync-fork)
- [Cherry Pick](#cherry-pick)
- [Prune](#prune)
- [Blame](#blame)
- [Rescue](#rescue)
- [Refresh](#refresh)
- [Squash](#squash)
- [Switch](#switch)
- [Worktree](#worktree)

## Sync

Safely synchronize local and remote changes when your push is rejected.

```bash
# Sync current branch
githelper sync

# Sync specific branch
githelper sync main

# Skip stashing changes
githelper sync --no-stash
```

**Use when:**
- Push is rejected due to remote changes
- You want to update local branch without merge commits
- You have local changes you don't want to lose

## Sync Fork

Keep your fork in sync with the upstream repository.

```bash
# Auto-detect and sync with upstream
githelper sync-fork

# Sync with specific upstream
githelper sync-fork --upstream original/repo

# Sync with different main branch
githelper sync-fork --branch develop
```

**Use when:**
- Maintaining a fork of another repository
- Need to get latest changes from upstream
- Want to keep your fork in sync

## Cherry Pick

Interactively cherry-pick commits from a pull request.

```bash
# Cherry-pick from PR #123
githelper cherry-pick 123

# Interactive selection with fzf:
# - Use TAB to select multiple commits
# - ENTER to confirm
# - ESC to cancel
```

**Use when:**
- You want specific changes from a PR
- Need to apply fixes to multiple branches
- Want to test specific commits

## Prune

Clean up local branches that have been merged.

```bash
# Interactive branch cleanup
githelper prune

# Delete without confirmation
githelper prune --force

# Use different main branch
githelper prune --main develop
```

**Use when:**
- You have many stale branches
- Want to clean up after merging PRs
- Need to remove old feature branches

## Blame

Track changes to a specific line across all commits.

```bash
# Show history of line 42 in main.go
githelper blame main.go 42

# Shows:
# - Who modified the line
# - When it was modified
# - Commit message for each change
```

**Use when:**
- Investigating code history
- Finding out who wrote specific code
- Understanding why code changed

## Rescue

Create a new branch from detached HEAD state.

```bash
# Interactive branch creation
githelper rescue

# Create specific branch name
githelper rescue new-feature
```

**Use when:**
- You checked out a specific commit without -b
- You're in "detached HEAD" state
- You need to save your work before switching branches

## Refresh

Fix Git index and line ending issues.

```bash
# Refresh all files
githelper refresh

# Refresh specific file
githelper refresh file.txt

# Fix line ending issues
githelper refresh --crlf

# Clean up untracked files
githelper refresh --clean
```

**Use when:**
- Git shows files as modified but you haven't changed them
- Line ending (CRLF/LF) issues causing false modifications
- Need to clean up and start fresh

## Squash

Quickly squash your recent commits into a single commit.

```bash
# Squash last 3 commits
githelper squash 3

# Squash with custom message
githelper squash 5 -m "New feature"

# Generate message with AI
githelper squash 3 --ai
```

**Use when:**
- Your commit history is too granular
- You want to clean up WIP commits
- You need a clean history before merging

## Switch

Interactively switch between Git branches.

```bash
# Interactive branch selection
githelper switch

# Show all branches including remote
githelper switch --all

# Sort by name instead of date
githelper switch --sort=name
```

**Use when:**
- Working across multiple branches
- Need to find a specific branch quickly
- Want to see branch details before switching

## Worktree

Manage multiple working trees for your repository.

```bash
# List existing worktrees
githelper worktree list

# Create new worktree
githelper worktree add feature-branch

# Remove worktree
githelper worktree remove path/to/worktree
```

**Use when:**
- Working on multiple features simultaneously
- Need to test changes in isolation
- Want to work on different branches without stashing

## Tips

1. Most commands support interactive mode with `fzf` when available
2. Use `--help` with any command for detailed usage information
3. Commands with destructive operations will ask for confirmation
4. Many commands support both simple and advanced usage patterns

## Installation

```bash
# Install from source
go install github.com/yourusername/githelper@latest

# Required dependencies
brew install fzf  # Optional but recommended for better interaction
```

## Configuration

Some commands support configuration through environment variables or config files:

```bash
# Configure OpenAI API key for AI features
export OPENAI_API_KEY=your_key_here

# Or use config file (~/.githelper.yaml):
openai_api_key: your_key_here
```
```
