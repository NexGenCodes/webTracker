package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Find the backend root (where sqlc.yaml lives)
	root := findRoot()
	if root == "" {
		log.Fatal("Could not find sqlc.yaml. Run from the backend directory.")
	}

	fmt.Printf("Running sqlc generate in: %s\n", root)

	cmd := exec.Command("sqlc", "generate")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Fallback: try with full path
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		sqlcBin := filepath.Join(gopath, "bin", "sqlc")

		cmd2 := exec.Command(sqlcBin, "generate")
		cmd2.Dir = root
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		if err2 := cmd2.Run(); err2 != nil {
			log.Fatalf("sqlc generate failed: %v (also tried %s: %v)", err, sqlcBin, err2)
		}
	}

	fmt.Println("SQLC code generation complete.")
}

func findRoot() string {
	// Try CWD first
	if _, err := os.Stat("sqlc.yaml"); err == nil {
		abs, _ := filepath.Abs(".")
		return abs
	}
	// Try parent dirs
	dir, _ := os.Getwd()
	for dir != "" && dir != filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "sqlc.yaml")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}
