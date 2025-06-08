package gitutils

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func CheckoutWorktree(repoDir, ref string) string {
	tmpDir, err := os.MkdirTemp("", "apidiff-"+ref)
	if err != nil {
		log.Fatal(err)
	}
	runGitInDir(repoDir, "worktree", "add", "--detach", tmpDir, ref)
	return tmpDir
}

func CleanupWorktree(repoDir, path string) {
	runGitInDir(repoDir, "worktree", "remove", "--force", path)
}

// TODO: testutils
func runGitInDir(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		log.Fatalf("git %v failed: %v", args, err)
	}
}
