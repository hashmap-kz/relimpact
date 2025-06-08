package gitutils

import (
	"log"
	"os"
	"os/exec"
)

func CheckoutWorktree(ref string) string {
	tmpDir, err := os.MkdirTemp("", "apidiff-"+ref)
	if err != nil {
		log.Fatal(err)
	}
	run("git", "worktree", "add", "--detach", tmpDir, ref)
	return tmpDir
}

func CleanupWorktree(path string) {
	run("git", "worktree", "remove", "--force", path)
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
