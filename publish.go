package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	// Define build targets
	targets := []struct {
		OS   string
		Arch string
	}{
		{"windows", "amd64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"linux", "amd64"},
	}

	// Build for each target
	for _, target := range targets {
		log.Printf("Building for %s/%s...\n", target.OS, target.Arch)
		cmd := exec.Command("wails", "build", "-platform", target.OS+"/"+target.Arch)
		cmd.Env = append(os.Environ(), "GOOS="+target.OS, "GOARCH="+target.Arch)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = "./app"
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to build for %s/%s: %v", target.OS, target.Arch, err)
		}
	}

	// Authenticate with GitHub CLI
	log.Println("Authenticating with GitHub CLI...")
	if err := exec.Command("gh", "auth", "status").Run(); err != nil {
		log.Fatalf("GitHub CLI authentication failed: %v", err)
	}

	// Create a new release
	releaseTag := "v1.0.0" // Replace with your version
	log.Printf("Creating GitHub release %s...\n", releaseTag)
	if err := exec.Command("gh", "release", "create", releaseTag, "--title", releaseTag, "--notes", "Release notes here").Run(); err != nil {
		log.Fatalf("Failed to create GitHub release: %v", err)
	}

	// Upload artifacts to the release
	buildDir := "./app/build/bin" // Replace with your build output directory
	files, err := os.ReadDir(buildDir)
	if err != nil {
		log.Fatalf("Failed to read build directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := buildDir + "/" + file.Name()
			log.Printf("Uploading %s to release...\n", filePath)
			if err := exec.Command("gh", "release", "upload", releaseTag, filePath).Run(); err != nil {
				log.Fatalf("Failed to upload %s: %v", filePath, err)
			}
		}
	}

	log.Println("Build and release process completed successfully!")
}
