//go:build mage

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	packageName = "github.com/dunamismax/go-discord-bots"
	buildDir    = "bin"
	tmpDir      = "tmp"
)

// Default target to run when none is specified
var Default = Build

// loadEnvFile loads environment variables from .env file if it exists
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
	defer file.Close()

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
				os.Setenv(key, value)
			}
		}
	}

	return scanner.Err()
}

// Build builds all Discord bots
func Build() error {
	fmt.Println("Building all Discord bots...")
	
	bots, err := getBotDirectories()
	if err != nil {
		return err
	}
	
	if len(bots) == 0 {
		fmt.Println("No bots found to build")
		return nil
	}
	
	for _, bot := range bots {
		if err := buildBot(bot); err != nil {
			return fmt.Errorf("failed to build %s: %w", bot, err)
		}
	}
	
	fmt.Printf("Successfully built %d bot(s)!\n", len(bots))
	return showBuildInfo()
}

func buildBot(bot string) error {
	fmt.Printf("  Building %s...\n", bot)
	
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	
	ldflags := "-s -w -X main.version=1.0.0 -X main.buildTime=" + getCurrentTime()
	binaryPath := filepath.Join(buildDir, bot)
	
	// Add .exe extension on Windows
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}
	
	return sh.Run("go", "build", "-ldflags="+ldflags, "-o", binaryPath, fmt.Sprintf("./bots/%s/main.go", bot))
}

func getCurrentTime() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// getGoBinaryPath finds the path to a Go binary, checking GOBIN, GOPATH/bin, and PATH
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

// Run runs a specific Discord bot
func Run(bot string) error {
	if bot == "" {
		bots, _ := getBotDirectories()
		if len(bots) > 0 {
			return fmt.Errorf("bot name is required. Available bots: %s", strings.Join(bots, ", "))
		}
		return fmt.Errorf("bot name is required. No bots found in bots/ directory")
	}
	
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	
	botDir := filepath.Join("bots", bot)
	if _, err := os.Stat(botDir); os.IsNotExist(err) {
		return fmt.Errorf("bot %s does not exist", bot)
	}
	
	fmt.Printf("Starting %s Discord bot...\n", bot)
	return sh.RunWith(map[string]string{"BOT_NAME": bot}, "go", "run", fmt.Sprintf("./bots/%s/main.go", bot))
}

// RunAll runs all Discord bots concurrently
func RunAll() error {
	fmt.Println("Starting all Discord bots...")
	
	// Load environment variables from .env file
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	
	bots, err := getBotDirectories()
	if err != nil {
		return err
	}
	
	if len(bots) == 0 {
		return fmt.Errorf("no bots found to run")
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	for _, bot := range bots {
		go func(botName string) {
			cmd := exec.CommandContext(ctx, "go", "run", fmt.Sprintf("./bots/%s/main.go", botName))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(), fmt.Sprintf("BOT_NAME=%s", botName))
			
			fmt.Printf("[%s] Starting bot...\n", botName)
			if err := cmd.Run(); err != nil {
				fmt.Printf("[%s] Bot exited with error: %v\n", botName, err)
			}
		}(bot)
	}
	
	// Keep running until interrupted
	fmt.Printf("All %d bot(s) are running. Press Ctrl+C to stop.\n", len(bots))
	select {}
}

// Dev runs a Discord bot in development mode with auto-restart and environment loading
func Dev(bot string) error {
	if bot == "" {
		bots, _ := getBotDirectories()
		if len(bots) > 0 {
			return fmt.Errorf("bot name is required. Available bots: %s", strings.Join(bots, ", "))
		}
		return fmt.Errorf("bot name is required. No bots found in bots/ directory")
	}
	
	// Check if bot exists
	botDir := filepath.Join("bots", bot)
	if _, err := os.Stat(botDir); os.IsNotExist(err) {
		return fmt.Errorf("bot %s does not exist", bot)
	}
	
	fmt.Printf("Starting %s in development mode with auto-restart...\n", bot)
	fmt.Println("Press Ctrl+C to stop.")
	
	restartCount := 0
	for {
		// Load environment variables fresh each restart
		if err := loadEnvFile(); err != nil {
			fmt.Printf("Warning: failed to load .env file: %v\n", err)
		}
		
		cmd := exec.Command("go", "run", fmt.Sprintf("./bots/%s/main.go", bot))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), fmt.Sprintf("BOT_NAME=%s", bot))
		
		if restartCount > 0 {
			fmt.Printf("[Restart #%d] Starting %s...\n", restartCount, bot)
		}
		
		if err := cmd.Run(); err != nil {
			restartCount++
			fmt.Printf("Bot crashed: %v. Restarting in 3 seconds... (restart #%d)\n", err, restartCount)
			time.Sleep(3 * time.Second)
		} else {
			fmt.Printf("Bot %s exited cleanly.\n", bot)
			break
		}
		
		// Prevent infinite restart loop
		if restartCount > 10 {
			return fmt.Errorf("bot has crashed too many times (>10), stopping auto-restart")
		}
	}
	
	return nil
}

// DevAll runs all bots in development mode
func DevAll() error {
	fmt.Println("Starting all bots in development mode...")
	
	bots, err := getBotDirectories()
	if err != nil {
		return err
	}
	
	if len(bots) == 0 {
		return fmt.Errorf("no bots found to run")
	}
	
	// Load environment variables
	if err := loadEnvFile(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	for _, bot := range bots {
		go func(botName string) {
			restartCount := 0
			for {
				cmd := exec.CommandContext(ctx, "go", "run", fmt.Sprintf("./bots/%s/main.go", botName))
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Env = append(os.Environ(), fmt.Sprintf("BOT_NAME=%s", botName))
				
				if restartCount == 0 {
					fmt.Printf("[%s] Starting bot...\n", botName)
				} else {
					fmt.Printf("[%s] Restarting bot... (restart #%d)\n", botName, restartCount)
				}
				
				if err := cmd.Run(); err != nil {
					restartCount++
					if restartCount > 5 {
						fmt.Printf("[%s] Bot has crashed too many times, stopping auto-restart\n", botName)
						return
					}
					fmt.Printf("[%s] Bot crashed: %v. Restarting in 3 seconds...\n", botName, err)
					time.Sleep(3 * time.Second)
				} else {
					fmt.Printf("[%s] Bot exited cleanly\n", botName)
					return
				}
			}
		}(bot)
	}
	
	fmt.Printf("All %d bot(s) running in development mode. Press Ctrl+C to stop.\n", len(bots))
	select {}
}

// Test runs tests for all packages
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "-v", "./...")
}

// TestCoverage runs tests with coverage report
func TestCoverage() error {
	fmt.Println("Running tests with coverage...")
	
	coverageFile := "coverage.out"
	if err := sh.RunV("go", "test", "-coverprofile="+coverageFile, "./..."); err != nil {
		return fmt.Errorf("failed to run tests with coverage: %w", err)
	}
	
	fmt.Println("Generating coverage report...")
	if err := sh.RunV("go", "tool", "cover", "-html="+coverageFile, "-o", "coverage.html"); err != nil {
		return fmt.Errorf("failed to generate coverage report: %w", err)
	}
	
	fmt.Println("Coverage report generated: coverage.html")
	return nil
}

// Fmt formats and tidies code using goimports and standard tooling
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

// Vet analyzes code for common errors
func Vet() error {
	fmt.Println("Running go vet...")
	return sh.RunV("go", "vet", "./...")
}

// VulnCheck scans for known vulnerabilities
func VulnCheck() error {
	fmt.Println("Running vulnerability check...")
	govulncheckPath, err := getGoBinaryPath("govulncheck")
	if err != nil {
		return fmt.Errorf("govulncheck not found: %w", err)
	}
	return sh.RunV(govulncheckPath, "./...")
}

// Lint runs golangci-lint with comprehensive linting rules
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

	return sh.RunV(lintPath, "run", "./...")
}

// Clean removes built binaries and generated files
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

	// Remove coverage files
	coverageFiles := []string{"coverage.out", "coverage.html"}
	for _, file := range coverageFiles {
		if err := sh.Rm(file); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove %s: %v\n", file, err)
		}
	}

	fmt.Println("Clean complete!")
	return nil
}

// Reset completely resets the repository to a fresh state
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

// Setup installs required development tools
func Setup() error {
	fmt.Println("Setting up Discord bot development environment...")

	tools := map[string]string{
		"govulncheck":    "golang.org/x/vuln/cmd/govulncheck@latest",
		"goimports":      "golang.org/x/tools/cmd/goimports@latest",
		"golangci-lint":  "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest",
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

// CI runs the complete CI pipeline
func CI() error {
	fmt.Println("Running complete CI pipeline...")
	mg.SerialDeps(Fmt, Vet, Lint, Build, Test, showBuildInfo)
	return nil
}

// Quality runs all quality checks
func Quality() error {
	fmt.Println("Running all quality checks...")
	mg.Deps(Vet, Lint, VulnCheck)
	return nil
}

// ListBots lists all available Discord bots
func ListBots() error {
	fmt.Println("Available Discord bots:")
	
	bots, err := getBotDirectories()
	if err != nil {
		return err
	}
	
	if len(bots) == 0 {
		fmt.Println("  No bots found in bots/ directory")
		return nil
	}
	
	for i, bot := range bots {
		fmt.Printf("  %d. %s\n", i+1, bot)
	}
	
	fmt.Printf("\nTotal: %d bot(s)\n", len(bots))
	return nil
}

// Status shows the current status of the development environment
func Status() error {
	fmt.Println("Discord Bot Development Environment Status")
	fmt.Println("=========================================")
	
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
	
	// List available bots
	bots, err := getBotDirectories()
	if err != nil {
		fmt.Printf("Bots: Error reading bots directory (%v)\n", err)
	} else if len(bots) == 0 {
		fmt.Println("Bots: No bots found")
	} else {
		fmt.Printf("Bots: %d found (%s)\n", len(bots), strings.Join(bots, ", "))
	}
	
	// Check if binaries exist
	if _, err := os.Stat(buildDir); err == nil {
		entries, _ := os.ReadDir(buildDir)
		fmt.Printf("Built binaries: %d found in %s/\n", len(entries), buildDir)
	} else {
		fmt.Println("Built binaries: None found")
	}
	
	return nil
}

// Help prints a help message with available commands
func Help() {
	fmt.Println(`
Go Discord Bots Magefile

Available commands:

Development:
  mage setup (s)        Install all development tools and dependencies
  mage dev <bot>        Run a specific bot in development mode with auto-restart
  mage devAll           Run all bots in development mode with auto-restart  
  mage run <bot>        Build and run a specific bot
  mage runAll           Run all bots concurrently
  mage build (b)        Build all Discord bot binaries
  mage listBots         List all available Discord bots
  mage status           Show development environment status

Testing:
  mage test (t)         Run all tests
  mage testCoverage     Run tests with coverage report

Quality:
  mage fmt (f)          Format code with goimports and tidy modules
  mage vet (v)          Run go vet static analysis
  mage lint (l)         Run golangci-lint comprehensive linting
  mage vulncheck (vc)   Check for security vulnerabilities
  mage quality (q)      Run all quality checks (vet + lint + vulncheck)

Production:
  mage ci               Complete CI pipeline (fmt + quality + build + test)
  mage clean (c)        Clean build artifacts and temporary files
  mage reset            Reset repository to fresh state (clean + tidy + download)

Other:
  mage help (h)         Show this help message

Examples:
  mage dev mtg-card-bot    # Run MTG bot in dev mode
  mage run mtg-card-bot    # Run MTG bot once  
  mage build               # Build all bots
  mage runAll              # Run all bots at once
    `)
}

// showBuildInfo displays information about the built binaries
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

func getBotDirectories() ([]string, error) {
	botsDir := "bots"
	if _, err := os.Stat(botsDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	entries, err := os.ReadDir(botsDir)
	if err != nil {
		return nil, err
	}
	
	var bots []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if main.go exists in the bot directory
			mainFile := filepath.Join(botsDir, entry.Name(), "main.go")
			if _, err := os.Stat(mainFile); err == nil {
				bots = append(bots, entry.Name())
			}
		}
	}
	
	return bots, nil
}

// Aliases for common commands
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
	"t":  Test,
	"h":  Help,
}