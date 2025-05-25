# WSL2 Two-Phase Automation Solution

This repository contains solutions for automating tasks that require changes in WSL2, then Windows, followed by a WSL2 restart - solving the challenge where the WSL2 restart kills the automation script.

## The Problem

When automating WSL2 kernel builds or similar tasks, you often need to:
1. Build/modify something inside WSL2
2. Copy files to Windows and modify Windows configuration
3. Restart WSL2 to apply changes

The challenge is that step 3 kills the WSL2 VM running your automation script!

## The Solution: Two-Phase Automation

Our solution uses a **two-phase approach**:

- **Phase 1** (runs in WSL2): Performs WSL2-side tasks, prepares Windows-side changes, and triggers Phase 2
- **Phase 2** (runs on Windows): Applies Windows configuration and restarts WSL2

## 🚀 Quick Start (TL;DR)

### The "I Just Want It To Work" Section 😅

Look, we know there are a million options below. Here's what you probably want:

```bash
# Python version
python3 wsl2_automation.py

# Go version
go run wsl2_automation.go
```

That's it! 🎉 It clones the kernel, builds it, and sets everything up. Go grab a ☕ (or three) while it builds.

## 🎛️ All The Options (For The Control Freaks)

### Configuration Options

Both implementations support extensive customization through command-line arguments:

#### Python Version

```bash
# Basic usage
python3 wsl2_automation.py --phase 1  # Run in WSL2

# Use a different kernel branch
python3 wsl2_automation.py --kernel-branch linux-msft-wsl-6.1.y

# Use a different destination for the kernel
python3 wsl2_automation.py --kernel-dest "C:/my-kernel/bzImage"

# Use a different git repository
python3 wsl2_automation.py --kernel-repo "https://github.com/myuser/custom-kernel.git"

# Use existing cloned directory (don't auto-clone)
python3 wsl2_automation.py --kernel-dir "my-kernel-dir" --no-clone

# Combine ALL the options! 🤯
python3 wsl2_automation.py \
    --kernel-branch main \
    --kernel-dest "D:/kernels/wsl2.img" \
    --kernel-repo "https://github.com/torvalds/linux.git" \
    --kernel-dir "linux-mainline" \
    --no-clone
```

#### Go Version

```bash
# Basic usage
./wsl2_automation -phase=1  # Run in WSL2

# Use a different kernel branch
./wsl2_automation -kernel-branch=linux-msft-wsl-6.1.y

# Use a different destination for the kernel
./wsl2_automation -kernel-dest="C:/my-kernel/bzImage"

# Use a different git repository
./wsl2_automation -kernel-repo="https://github.com/myuser/custom-kernel.git"

# Use existing cloned directory (don't auto-clone)
./wsl2_automation -kernel-dir="my-kernel-dir" -no-clone

# The kitchen sink approach! 🚰
./wsl2_automation \
    -kernel-branch=main \
    -kernel-dest="D:/kernels/wsl2.img" \
    -kernel-repo="https://github.com/torvalds/linux.git" \
    -kernel-dir="linux-mainline" \
    -no-clone
```

### Available Options Explained

| Option | Default | Description |
|--------|---------|-------------|
| `--kernel-repo` | `https://github.com/microsoft/WSL2-Linux-Kernel.git` | Git repository URL |
| `--kernel-branch` | `linux-msft-wsl-6.6.y` | Branch to build |
| `--kernel-dest` | `C:/bzImage` | Where to copy the built kernel |
| `--kernel-dir` | `WSL2-Linux-Kernel` | Local directory name |
| `--no-clone` | `false` | Skip cloning (use existing source) |
| `--phase` | `1` | Which phase to run (1 or 2) |
| `--cleanup` | `false` | Remove temporary files |

## 😵 Option Overload? 

### The "Analysis Paralysis" Prevention Guide

Feeling overwhelmed by all these options? Getting flashbacks to ordering at that fancy coffee shop? 😰

Here's a handy flowchart:

```
Do you just want a custom WSL2 kernel?
    │
    ├─ YES → Just run: python3 wsl2_automation.py
    │
    └─ NO → Are you SURE you need all those options?
            │
            ├─ NO → Just run: python3 wsl2_automation.py
            │
            └─ YES → OK, you do you... scroll up for options 📜
```

### Real Talk 💬

99% of users just need:

```bash
# This is all you need, seriously
python3 wsl2_automation.py
```

The other options are for:
- 🤓 Kernel hackers with custom forks
- 🏢 Enterprise users with specific requirements  
- 🎮 People who like to press all the buttons
- 🧪 Those building experimental kernels
- 📁 Folks with "unique" directory structures

## Available Implementations

### Python Version (`wsl2_automation.py`)

**Features:**
- Object-oriented design with comprehensive error handling
- State persistence between phases using JSON
- Colored logging output
- Automatic PowerShell script generation for Phase 2
- Command-line interface with options

**Usage:**
```bash
# Run Phase 1 (inside WSL2)
python3 wsl2_automation.py --phase 1

# Run Phase 2 (on Windows - usually automatic)
python wsl2_automation.py --phase 2

# Cleanup temporary files
python3 wsl2_automation.py --cleanup
```

### Go Version (`wsl2_automation.go`)

**Features:**
- Compiled binary for better performance
- No external dependencies (uses only Go standard library)
- Cross-platform compatibility
- Structured logging with colors
- Type-safe state management

**Usage:**
```bash
# Build the Go binary
go build -o wsl2_automation wsl2_automation.go

# Run Phase 1 (inside WSL2)
./wsl2_automation -phase=1

# Run Phase 2 (on Windows - usually automatic)
./wsl2_automation.exe -phase=2

# Cleanup temporary files
./wsl2_automation -cleanup
```

## How It Works

### Phase 1 (WSL2)
1. **Environment Detection**: Checks if running in WSL2
2. **Kernel Build**: Clones and builds the WSL2 kernel
3. **File Preparation**: Copies kernel image to Windows filesystem
4. **State Persistence**: Saves automation state to shared location
5. **Script Generation**: Creates PowerShell script for Phase 2
6. **Phase 2 Trigger**: Executes PowerShell script on Windows

### Phase 2 (Windows)
1. **State Validation**: Verifies Phase 1 completed successfully
2. **Configuration**: Creates/updates `.wslconfig` file
3. **WSL Restart**: Shuts down WSL2 (applies new kernel)
4. **Cleanup**: Removes temporary files

## File Structure

```
samples/wsl2/
├── wsl2_automation.py      # Python implementation
├── wsl2_automation.go      # Go implementation
├── go.mod                  # Go module file
├── wsl2_kernel_build.sh    # Original bash script (reference)
└── README.md               # This file
```

## Temporary Files

Both implementations use temporary directories for state management:
- **WSL2**: `/tmp/wsl2_automation/`
- **Windows**: `C:/temp/wsl2_automation/`

Files created:
- `automation_state.json`: State data passed between phases
- `phase2.ps1`: PowerShell script for Windows tasks

## Customization

### Adding Your Own Tasks

Both implementations are designed to be easily customizable. Key areas to modify:

**Phase 1 (WSL2 tasks):**
- Modify `phase1_wsl_tasks()` (Python) or `phase1WSLTasks()` (Go)
- Add your custom build steps, file operations, etc.

**Phase 2 (Windows tasks):**
- Modify `_create_phase2_script()` (Python) or `createPhase2Script()` (Go)
- Add Windows registry changes, service restarts, etc.

### Example Customizations

```python
# Python: Add custom WSL2 task
def custom_wsl_task(self):
    self._log("Running custom task...")
    self._run_command(["your-custom-command", "arg1", "arg2"])
```

```go
// Go: Add custom Windows task
func (w *WSL2Automator) customWindowsTask() error {
    w.log("Running custom Windows task...", "INFO")
    return w.runCommand("your-command.exe", "arg1", "arg2")
}
```

## Error Handling

Both implementations include comprehensive error handling:
- **Command failures**: Logged with full error details
- **State corruption**: Validation between phases
- **Environment issues**: WSL2 detection and user identification
- **File operations**: Proper error propagation and cleanup

## Security Considerations

- PowerShell execution policy is temporarily bypassed for Phase 2
- Temporary files are cleaned up automatically
- No sensitive data is stored in state files
- Scripts run with current user privileges

## Troubleshooting

### Common Issues

1. **"Phase 1 must run inside WSL2"**
   - Ensure you're running the script from within WSL2
   - Check `/proc/version` contains "microsoft"

2. **"Cannot determine Windows user"**
   - Verify `cmd.exe` is accessible from WSL2
   - Check Windows PATH is properly configured

3. **"Phase 1 not completed"**
   - Check if state file exists in temp directory
   - Verify Phase 1 ran successfully without errors

4. **PowerShell execution errors**
   - Ensure PowerShell is available on Windows
   - Check Windows execution policy settings

### Debug Mode

Add verbose logging by modifying the log level in either implementation:

```python
# Python: Add debug logging
self._log(f"Debug: Current directory is {os.getcwd()}", "INFO")
```

```go
// Go: Add debug logging
w.log(fmt.Sprintf("Debug: Current directory is %s", pwd), "INFO")
```

## Performance Notes

- **Go version**: Faster startup, smaller memory footprint
- **Python version**: More readable, easier to modify
- **Build time**: Kernel compilation is the bottleneck (both versions similar)
- **State persistence**: JSON files are small and fast to read/write

## Contributing

When adding features:
1. Maintain compatibility between Python and Go versions
2. Add appropriate error handling and logging
3. Update this README with new functionality
4. Test both phases thoroughly

## License

This code is provided as-is for educational and automation purposes. Modify as needed for your specific use case. 