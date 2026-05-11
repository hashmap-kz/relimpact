package gitutils

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
)

func CheckoutWorktree(ctx context.Context, repoDir, ref string) string {
	tmpDir, err := os.MkdirTemp("", "apidiff-"+ref)
	if err != nil {
		log.Fatal(err)
	}
	runGitInDir(ctx, repoDir, "worktree", "add", "--detach", tmpDir, ref)
	return tmpDir
}

// CleanupWorktree removes the worktree unconditionally; it uses a fresh
// context so that an expired deadline does not prevent cleanup.
func CleanupWorktree(repoDir, path string) {
	runGitInDir(context.Background(), repoDir, "worktree", "remove", "--force", path)
}

func runGitInDir(ctx context.Context, dir string, args ...string) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		log.Fatalf("git %v failed: %v", args, err)
	}
}
