//go:build mage

// Package main provides Mage build targets for the MTG Card Bot project.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	packageName = "github.com/dunamismax/MTG-Card-Bot"
	botName     = "mtg-card-bot"
	buildDir    = "bin"
	tmpDir      = "tmp"
)

// Default target to run when none is specified.
var Default = Build

// loadEnvFile loads environment variables from .env file if it exists.
func loadEnvFile() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// .env file doesn't exist, that's okay
		return nil
	}

	file, err := os.Open(envFile)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Warning: failed to close .env file: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
				(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
				value = value[1 : len(value)-1]
			}

			// Only set if not already set by system environment
			if os.Getenv(key) == "" {
				if err := os.Setenv(key, value); err != nil {
					fmt.Printf("Warning: failed to set environment variable %s: %v\n", key, err)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan .env file: %w", err)
	}

	return nil
}

// Build builds the MTG Card Bot.
func Build() error {
	fmt.Println("Building MTG Card Bot...")

	if err := buildBot(botName); err != nil {
		return fmt.Errorf("failed to build %s: %w", botName, err)
	}

	fmt.Println("Successfully built MTG Card Bot!")

	return showBuildInfo()
}

func buildBot(bot string) error {
	fmt.Printf("  Building %s...\n", bot)

	if err := os.MkdirAll(buildDir, 0750); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	ldflags := "-s -w -X main.version=1.0.0 -X main.buildTime=" + getCurrentTime()
	binaryPath := filepath.Join(buildDir, bot)

	// Add .exe extension on Windows
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	if err := sh.Run("go", "build", "-ldflags="+ldflags, "-o", binaryPath, "main.go"); err != nil {
		return fmt.Errorf("failed to build %s: %w", bot, err)
	}

	return nil
}

func getCurrentTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// getGoBinaryPath finds the path to a Go binary, checking GOBIN, GOPATH/bin, and PATH.
func getGoBinaryPath(binaryName string) (string, error) {
	// First check if it's in PATH
	if err := sh.Run("which", binaryName); err == nil {
		return binaryName, nil
	}

	// Check GOBIN first
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		binaryPath := filepath.Join(gobin, binaryName)
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, nil
		}
	}

	// Check GOPATH/bin
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		if home := os.Getenv("HOME"); home != "" {
			gopath = filepath.Join(home, "go")
		}
	}

	if gopath != "" {
		binaryPath := filepath.Join(gopath, "bin", binaryName)
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, nil
		}
	}

	return "", fmt.Errorf("%s not found in PATH, GOBIN, or GOPATH/bin", binaryName)
}

// Run runs the MTG Card Bot.
func Run() error {
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		return fmt.Errorf("main.go does not exist")
	}

	fmt.Printf("Starting %s Discord bot...\n", botName)

	if err := sh.RunWith(map[string]string{"BOT_NAME": botName}, "go", "run", "main.go"); err != nil {
		return fmt.Errorf("failed to run bot: %w", err)
	}

	return nil
}

// Dev runs the MTG Card Bot in development mode with auto-restart.
func Dev() error {
	// Check if main.go exists
	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		return fmt.Errorf("main.go does not exist")
	}

	fmt.Printf("Starting %s in development mode with auto-restart...\n", botName)
	fmt.Println("Press Ctrl+C to stop.")

	// Setup signal handling for the dev mode itself
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	restartCount := 0

	for {
		// Load environment variables fresh each restart
		if err := loadEnvFile(); err != nil {
			fmt.Printf("Warning: failed to load .env file: %v\n", err)
		}

		cmd := exec.Command("go", "run", "main.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), fmt.Sprintf("BOT_NAME=%s", botName))

		if restartCount > 0 {
			fmt.Printf("[Restart #%d] Starting %s...\n", restartCount, botName)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start bot: %w", err)
		}

		// Wait for either the process to finish or a signal
		done := make(chan error, 1)

		go func() {
			done <- cmd.Wait()
		}()

		select {
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal: %v. Stopping development mode...\n", sig)
			// Forward signal to the bot process
			if err := cmd.Process.Signal(sig); err != nil {
				fmt.Printf("Warning: failed to send signal to bot process: %v\n", err)
			}
			// Wait for the process to finish gracefully
			<-done
			fmt.Println("Development mode stopped.")

			return nil

		case err := <-done:
			if err != nil {
				// Check if it was interrupted (graceful shutdown)
				var exitError *exec.ExitError
				if errors.As(err, &exitError) {
					if exitError.ExitCode() == 1 {
						// Exit code 1 could be graceful shutdown, check if it was a signal
						fmt.Printf("Bot %s exited with code 1 (likely graceful shutdown).\n", botName)
						return nil
					}
				}

				restartCount++
				fmt.Printf("Bot crashed: %v. Restarting in 3 seconds... (restart #%d)\n", err, restartCount)
				time.Sleep(3 * time.Second)
			} else {
				fmt.Printf("Bot %s exited cleanly.\n", botName)
				return nil
			}
		}

		// Prevent infinite restart loop
		if restartCount > 10 {
			return fmt.Errorf("bot has crashed too many times (>10), stopping auto-restart")
		}
	}
}

// Fmt formats and tidies code using goimports and standard tooling.
func Fmt() error {
	fmt.Println("Formatting and tidying...")

	// Tidy go modules
	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy modules: %w", err)
	}

	// Use goimports for better import management and formatting
	fmt.Println("  Running goimports...")

	goimportsPath, err := getGoBinaryPath("goimports")
	if err != nil {
		fmt.Printf("Warning: goimports not found, falling back to go fmt: %v\n", err)

		if err := sh.RunV("go", "fmt", "./..."); err != nil {
			return fmt.Errorf("failed to format code: %w", err)
		}
	} else {
		if err := sh.RunV(goimportsPath, "-w", "."); err != nil {
			fmt.Printf("Warning: goimports failed, falling back to go fmt: %v\n", err)

			if err := sh.RunV("go", "fmt", "./..."); err != nil {
				return fmt.Errorf("failed to format code: %w", err)
			}
		}
	}

	return nil
}

// Vet analyzes code for common errors.
func Vet() error {
	fmt.Println("Running go vet...")

	if err := sh.RunV("go", "vet", "./..."); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}

	return nil
}

// VulnCheck scans for known vulnerabilities.
func VulnCheck() error {
	fmt.Println("Running vulnerability check...")

	govulncheckPath, err := getGoBinaryPath("govulncheck")
	if err != nil {
		return fmt.Errorf("govulncheck not found: %w", err)
	}

	if err := sh.RunV(govulncheckPath, "./..."); err != nil {
		return fmt.Errorf("govulncheck failed: %w", err)
	}

	return nil
}

// Lint runs golangci-lint with comprehensive linting rules.
func Lint() error {
	fmt.Println("Running golangci-lint...")

	// Ensure the correct version of golangci-lint v2 is installed
	fmt.Println("  Ensuring golangci-lint v2 is installed...")

	if err := sh.RunV("go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"); err != nil {
		return fmt.Errorf("failed to install golangci-lint v2: %w", err)
	}

	// Find golangci-lint binary
	lintPath, err := getGoBinaryPath("golangci-lint")
	if err != nil {
		return fmt.Errorf("golangci-lint not found after installation: %w", err)
	}

	if err := sh.RunV(lintPath, "run", "./..."); err != nil {
		return fmt.Errorf("golangci-lint failed: %w", err)
	}

	return nil
}

// Clean removes built binaries and generated files.
func Clean() error {
	fmt.Println("Cleaning up...")

	// Remove build directory
	if err := sh.Rm(buildDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove build directory: %w", err)
	}

	// Remove tmp directory
	if err := sh.Rm(tmpDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove tmp directory: %w", err)
	}

	fmt.Println("Clean complete!")

	return nil
}

// Reset completely resets the repository to a fresh state.
func Reset() error {
	fmt.Println("Resetting repository to clean state...")

	// First run clean to remove built artifacts
	if err := Clean(); err != nil {
		return fmt.Errorf("failed to clean build artifacts: %w", err)
	}

	// Tidy modules
	fmt.Println("Tidying Go modules...")

	if err := sh.RunV("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy modules: %w", err)
	}

	// Download dependencies
	fmt.Println("Downloading fresh dependencies...")

	if err := sh.RunV("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	fmt.Println("Reset complete! Repository is now in fresh state.")

	return nil
}

// Setup installs required development tools.
func Setup() error {
	fmt.Println("Setting up Discord bot development environment...")

	tools := map[string]string{
		"govulncheck":   "golang.org/x/vuln/cmd/govulncheck@latest",
		"goimports":     "golang.org/x/tools/cmd/goimports@latest",
		"golangci-lint": "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest",
	}

	for tool, pkg := range tools {
		fmt.Printf("  Installing %s...\n", tool)

		if err := sh.RunV("go", "install", pkg); err != nil {
			return fmt.Errorf("failed to install %s: %w", tool, err)
		}
	}

	// Download module dependencies
	fmt.Println("Downloading dependencies...")

	if err := sh.RunV("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	fmt.Println("Setup complete!")
	fmt.Println("Next steps:")
	fmt.Println("   • Copy .env.example to .env and add your Discord bot token")
	fmt.Println("   • Run 'mage dev <bot-name>' to start development with auto-restart")
	fmt.Println("   • Run 'mage build' to create production binaries")
	fmt.Println("   • Run 'mage help' to see all available commands")

	return nil
}

// CI runs the complete CI pipeline.
func CI() {
	fmt.Println("Running complete CI pipeline...")
	mg.SerialDeps(Fmt, Vet, Lint, Build, showBuildInfo)
}

// Quality runs all quality checks.
func Quality() error {
	fmt.Println("Running all quality checks...")
	mg.Deps(Vet, Lint, VulnCheck)

	return nil
}

// Info shows information about the MTG Card Bot.
func Info() {
	fmt.Println("MTG Card Discord Bot")
	fmt.Println("===================")

	if _, err := os.Stat("main.go"); os.IsNotExist(err) {
		fmt.Printf("Main file not found: main.go\n")
		return
	}

	fmt.Printf("Bot name: %s\n", botName)
	fmt.Printf("Project root: %s\n", ".")
	fmt.Printf("Main file: %s\n", "main.go")
}

// Status shows the current status of the development environment.
func Status() {
	fmt.Println("MTG Card Bot Development Environment Status")
	fmt.Println("==========================================")

	// Check Go version
	if version, err := sh.Output("go", "version"); err == nil {
		fmt.Printf("Go: %s\n", version)
	} else {
		fmt.Printf("Go: Not found or error (%v)\n", err)
	}

	// Check if .env file exists
	if _, err := os.Stat(".env"); err == nil {
		fmt.Println("Environment: .env file found ✓")
	} else {
		fmt.Println("Environment: .env file missing ✗")
		fmt.Println("  Run: cp .env.example .env")
	}

	// Check main file
	if _, err := os.Stat("main.go"); err == nil {
		fmt.Printf("Bot: %s main.go found ✓\n", botName)
	} else {
		fmt.Printf("Bot: %s main.go missing ✗\n", botName)
	}

	// Check if binaries exist
	if _, err := os.Stat(buildDir); err == nil {
		entries, _ := os.ReadDir(buildDir)
		fmt.Printf("Built binaries: %d found in %s/\n", len(entries), buildDir)
	} else {
		fmt.Println("Built binaries: None found")
	}
}

// Help prints a help message with available commands.
func Help() {
	fmt.Println(`
MTG Card Bot Magefile

Available commands:

Development:
  mage setup (s)        Install all development tools and dependencies
  mage dev              Run the bot in development mode with auto-restart
  mage run              Build and run the bot
  mage build (b)        Build the bot binary
  mage info             Show bot information
  mage status           Show development environment status

Quality:
  mage fmt (f)          Format code with goimports and tidy modules
  mage vet (v)          Run go vet static analysis
  mage lint (l)         Run golangci-lint comprehensive linting
  mage vulncheck (vc)   Check for security vulnerabilities
  mage quality (q)      Run all quality checks (vet + lint + vulncheck)

Production:
  mage ci               Complete CI pipeline (fmt + quality + build)
  mage clean (c)        Clean build artifacts and temporary files
  mage reset            Reset repository to fresh state (clean + tidy + download)

Other:
  mage help (h)         Show this help message

Examples:
  mage dev              # Run MTG bot in dev mode
  mage run              # Run MTG bot once  
  mage build            # Build the bot
  mage info             # Show bot information
    `)
}

// showBuildInfo displays information about the built binaries.
func showBuildInfo() error {
	fmt.Println("\nBuild Information:")

	// Show Go version
	if version, err := sh.Output("go", "version"); err == nil {
		fmt.Printf("   Go version: %s\n", version)
	}

	// Show built binaries info
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		fmt.Println("   No binaries found")
		return nil
	}

	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return fmt.Errorf("failed to read build directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("   No binaries found")
		return nil
	}

	fmt.Printf("   Built binaries (%d):\n", len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			if info, err := entry.Info(); err == nil {
				size := info.Size()
				fmt.Printf("     %s: %.2f MB\n", entry.Name(), float64(size)/1024/1024)
			} else {
				fmt.Printf("     %s\n", entry.Name())
			}
		}
	}

	return nil
}

// Aliases for common commands.
var Aliases = map[string]interface{}{
	"b":  Build,
	"f":  Fmt,
	"v":  Vet,
	"l":  Lint,
	"vc": VulnCheck,
	"d":  Dev,
	"c":  Clean,
	"s":  Setup,
	"q":  Quality,
	"h":  Help,
}
