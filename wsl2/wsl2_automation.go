package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Colors for output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
)

// AutomationState represents the state passed between phases
type AutomationState struct {
	Phase1Completed bool   `json:"phase1_completed"`
	KernelBuilt     bool   `json:"kernel_built"`
	WindowsUser     string `json:"windows_user"`
	Timestamp       int64  `json:"timestamp"`
}

// WSL2Automator handles the two-phase automation
type WSL2Automator struct {
	IsWSL        bool
	WindowsUser  string
	TempDir      string
	StateFile    string
	KernelBranch string
	KernelDest   string
	KernelRepo   string
	KernelDir    string
	AutoClone    bool
}

// NewWSL2Automator creates a new automator instance
func NewWSL2Automator(kernelBranch, kernelDest, kernelRepo, kernelDir string, autoClone bool) *WSL2Automator {
	automator := &WSL2Automator{
		KernelBranch: kernelBranch,
		KernelDest:   kernelDest,
		KernelRepo:   kernelRepo,
		KernelDir:    kernelDir,
		AutoClone:    autoClone,
	}
	automator.IsWSL = automator.checkWSL()
	automator.WindowsUser = automator.getWindowsUser()

	if automator.IsWSL {
		automator.TempDir = "/tmp/wsl2_automation"
	} else {
		automator.TempDir = "C:/temp/wsl2_automation"
	}

	automator.StateFile = filepath.Join(automator.TempDir, "automation_state.json")
	return automator
}

// checkWSL determines if running in WSL environment
func (w *WSL2Automator) checkWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// getWindowsUser retrieves the Windows username
func (w *WSL2Automator) getWindowsUser() string {
	if w.IsWSL {
		cmd := exec.Command("cmd.exe", "/c", "echo %USERNAME%")
		output, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(output))
	}
	return os.Getenv("USERNAME")
}

// log prints colored log messages
func (w *WSL2Automator) log(message, level string) {
	colors := map[string]string{
		"INFO":    ColorBlue,
		"SUCCESS": ColorGreen,
		"WARNING": ColorYellow,
		"ERROR":   ColorRed,
	}

	color := colors[level]
	if color == "" {
		color = ColorBlue
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s%s [%s]%s %s\n", color, timestamp, level, ColorReset, message)
}

// saveState saves automation state to file
func (w *WSL2Automator) saveState(state AutomationState) error {
	err := os.MkdirAll(w.TempDir, 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(w.StateFile, data, 0644)
}

// loadState loads automation state from file
func (w *WSL2Automator) loadState() (AutomationState, error) {
	var state AutomationState

	data, err := os.ReadFile(w.StateFile)
	if err != nil {
		return state, err
	}

	err = json.Unmarshal(data, &state)
	return state, err
}

// runCommand executes a command with logging
func (w *WSL2Automator) runCommand(name string, args ...string) error {
	w.log(fmt.Sprintf("Running: %s %s", name, strings.Join(args, " ")), "INFO")
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runCommandOutput executes a command and returns output
func (w *WSL2Automator) runCommandOutput(name string, args ...string) (string, error) {
	w.log(fmt.Sprintf("Running: %s %s", name, strings.Join(args, " ")), "INFO")
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// phase1WSLTasks performs tasks inside WSL2
func (w *WSL2Automator) phase1WSLTasks() error {
	if !w.IsWSL {
		w.log("Phase 1 must run inside WSL2", "ERROR")
		return fmt.Errorf("not running in WSL2")
	}

	w.log("Starting Phase 1: WSL2 tasks", "INFO")

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	// Ensure we return to original directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			w.log(fmt.Sprintf("Failed to return to original directory: %v", err), "WARNING")
		}
	}()

	// Check if kernel directory exists
	if _, err := os.Stat(w.KernelDir); os.IsNotExist(err) {
		if w.AutoClone {
			w.log(fmt.Sprintf("Cloning kernel repository from %s...", w.KernelRepo), "INFO")
			err := w.runCommand("git", "clone",
				w.KernelRepo, w.KernelDir,
				"--depth=1", "-b", w.KernelBranch)
			if err != nil {
				return fmt.Errorf("failed to clone kernel: %v", err)
			}
		} else {
			w.log(fmt.Sprintf("Kernel directory '%s' not found!", w.KernelDir), "ERROR")
			w.log("Please clone the repository manually or set --auto-clone", "ERROR")
			return fmt.Errorf("kernel directory not found and auto-clone disabled")
		}
	} else {
		w.log(fmt.Sprintf("Using existing kernel directory: %s", w.KernelDir), "INFO")
	}

	// Change to kernel directory
	err = os.Chdir(w.KernelDir)
	if err != nil {
		return fmt.Errorf("failed to change directory: %v", err)
	}

	// Install dependencies
	w.log("Installing build dependencies...", "INFO")
	err = w.runCommand("sudo", "apt", "update")
	if err != nil {
		return fmt.Errorf("failed to update packages: %v", err)
	}

	err = w.runCommand("sudo", "apt", "install", "-y",
		"build-essential", "flex", "bison", "libssl-dev",
		"libelf-dev", "bc", "python3", "pahole", "cpio")
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	// Build kernel
	w.log("Building kernel...", "INFO")
	nproc, err := w.runCommandOutput("nproc")
	if err != nil {
		nproc = "4" // fallback
	}

	err = w.runCommand("make", fmt.Sprintf("-j%s", nproc), "KCONFIG_CONFIG=Microsoft/config-wsl")
	if err != nil {
		return fmt.Errorf("failed to build kernel: %v", err)
	}

	// Install modules
	w.log("Installing kernel modules...", "INFO")
	err = w.runCommand("sudo", "make", "modules_install", "headers_install")
	if err != nil {
		return fmt.Errorf("failed to install modules: %v", err)
	}

	// Copy kernel to Windows
	kernelPath := "arch/x86/boot/bzImage"
	if _, err := os.Stat(kernelPath); err == nil {
		err = w.runCommand("cp", kernelPath, w.KernelDest)
		if err != nil {
			return fmt.Errorf("failed to copy kernel: %v", err)
		}
		w.log(fmt.Sprintf("Kernel copied to %s", w.KernelDest), "SUCCESS")
	} else {
		return fmt.Errorf("kernel image not found at %s - build may have failed", kernelPath)
	}

	// Save state for Phase 2
	state := AutomationState{
		Phase1Completed: true,
		KernelBuilt:     true,
		WindowsUser:     w.WindowsUser,
		Timestamp:       time.Now().Unix(),
	}

	err = w.saveState(state)
	if err != nil {
		return fmt.Errorf("failed to save state: %v", err)
	}

	// Create Phase 2 script
	err = w.createPhase2Script()
	if err != nil {
		return fmt.Errorf("failed to create phase 2 script: %v", err)
	}

	w.log("Phase 1 completed successfully", "SUCCESS")
	w.log("Starting Phase 2 on Windows...", "INFO")

	// Trigger Phase 2
	w.triggerPhase2()

	return nil
}

// createPhase2Script creates the PowerShell script for Phase 2
func (w *WSL2Automator) createPhase2Script() error {
	if w.WindowsUser == "" {
		w.log("Cannot determine Windows user", "WARNING")
		return nil
	}

	psScript := fmt.Sprintf(`# Phase 2: Windows tasks
$ErrorActionPreference = "Stop"

Write-Host "Starting Phase 2: Windows tasks" -ForegroundColor Blue

try {
    # Create .wslconfig
    $wslConfigPath = "C:\Users\%s\.wslconfig"
    $wslConfig = @"
[wsl2]
kernel=%s
"@
    
    Write-Host "Creating WSL config at $wslConfigPath" -ForegroundColor Green
    $wslConfig | Out-File -FilePath $wslConfigPath -Encoding UTF8
    
    # Wait a moment
    Start-Sleep -Seconds 2
    
    # Shutdown WSL
    Write-Host "Shutting down WSL..." -ForegroundColor Yellow
    wsl --shutdown
    
    # Wait for shutdown
    Start-Sleep -Seconds 5
    
    # Success message
    Write-Host "WSL automation completed successfully!" -ForegroundColor Green
    Write-Host "WSL will use the new kernel on next startup." -ForegroundColor Green
    
} catch {
    Write-Host "Phase 2 failed: $_" -ForegroundColor Red
    exit 1
}

# Cleanup
Remove-Item "C:\temp\wsl2_automation\phase2.ps1" -ErrorAction SilentlyContinue
`, w.WindowsUser, w.KernelDest)

	// Create directory on Windows
	psScriptDir := "/mnt/c/temp/wsl2_automation"
	err := os.MkdirAll(psScriptDir, 0755)
	if err != nil {
		return err
	}

	// Write PowerShell script
	psScriptPath := filepath.Join(psScriptDir, "phase2.ps1")
	err = os.WriteFile(psScriptPath, []byte(psScript), 0644)
	if err != nil {
		return err
	}

	w.log("Phase 2 script created on Windows", "SUCCESS")
	return nil
}

// triggerPhase2 executes the Phase 2 script on Windows
func (w *WSL2Automator) triggerPhase2() {
	// Execute PowerShell script on Windows
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass",
		"-File", "C:\\temp\\wsl2_automation\\phase2.ps1")

	// This is expected to fail as WSL will be shut down
	cmd.Run()
	w.log("Phase 2 triggered (WSL shutdown expected)", "SUCCESS")
}

// phase2WindowsTasks performs tasks on Windows
func (w *WSL2Automator) phase2WindowsTasks() error {
	if w.IsWSL {
		w.log("Phase 2 should run on Windows, not WSL", "ERROR")
		return fmt.Errorf("running in WSL2")
	}

	w.log("Starting Phase 2: Windows tasks", "INFO")

	// Load state
	state, err := w.loadState()
	if err != nil || !state.Phase1Completed {
		w.log("Phase 1 not completed", "ERROR")
		return fmt.Errorf("phase 1 not completed")
	}

	// Create .wslconfig
	wslConfigPath := fmt.Sprintf("C:/Users/%s/.wslconfig", w.WindowsUser)
	wslConfigContent := fmt.Sprintf("[wsl2]\nkernel=%s\n", w.KernelDest)

	err = os.WriteFile(wslConfigPath, []byte(wslConfigContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create .wslconfig: %v", err)
	}

	w.log(fmt.Sprintf("WSL config created at %s", wslConfigPath), "SUCCESS")

	// Shutdown WSL
	w.log("Shutting down WSL...", "INFO")
	err = w.runCommand("wsl", "--shutdown")
	if err != nil {
		return fmt.Errorf("failed to shutdown WSL: %v", err)
	}

	w.log("Phase 2 completed successfully", "SUCCESS")
	return nil
}

// cleanup removes temporary files
func (w *WSL2Automator) cleanup() error {
	err := os.RemoveAll(w.TempDir)
	if err != nil {
		w.log(fmt.Sprintf("Cleanup failed: %v", err), "WARNING")
		return err
	}
	w.log("Cleanup completed", "SUCCESS")
	return nil
}

func main() {
	var phase = flag.String("phase", "1", "Which phase to run (1 or 2)")
	var cleanup = flag.Bool("cleanup", false, "Clean up temporary files")
	var kernelBranch = flag.String("kernel-branch", "linux-msft-wsl-6.6.y", "Kernel branch to build")
	var kernelDest = flag.String("kernel-dest", "C:/bzImage", "Destination for kernel image")
	var kernelRepo = flag.String("kernel-repo", "https://github.com/microsoft/WSL2-Linux-Kernel.git", "Kernel repository URL")
	var kernelDir = flag.String("kernel-dir", "WSL2-Linux-Kernel", "Local kernel directory name")
	var noClone = flag.Bool("no-clone", false, "Don't auto-clone the repository")
	flag.Parse()

	automator := NewWSL2Automator(*kernelBranch, *kernelDest, *kernelRepo, *kernelDir, !*noClone)

	if *cleanup {
		automator.cleanup()
		return
	}

	var err error
	switch *phase {
	case "1":
		err = automator.phase1WSLTasks()
	case "2":
		err = automator.phase2WindowsTasks()
	default:
		fmt.Printf("Invalid phase: %s\n", *phase)
		os.Exit(1)
	}

	if err != nil {
		automator.log(fmt.Sprintf("Automation failed: %v", err), "ERROR")
		os.Exit(1)
	}
}
