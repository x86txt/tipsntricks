#!/usr/bin/env python3
"""
WSL2 Two-Phase Automation Script
Handles the challenge of making changes in WSL2, then Windows, then restarting WSL2
"""

import os
import sys
import subprocess
import json
import time
import argparse
import shutil
from pathlib import Path
from typing import Dict, Any

class WSL2Automator:
    def __init__(self, 
                 kernel_branch="linux-msft-wsl-6.6.y", 
                 kernel_dest="C:/bzImage",
                 kernel_repo="https://github.com/microsoft/WSL2-Linux-Kernel.git",
                 kernel_dir="WSL2-Linux-Kernel",
                 auto_clone=True):
        self.is_wsl = self._check_wsl()
        self.windows_user = self._get_windows_user()
        self.temp_dir = Path("/tmp/wsl2_automation") if self.is_wsl else Path("C:/temp/wsl2_automation")
        self.state_file = self.temp_dir / "automation_state.json"
        self.kernel_branch = kernel_branch
        self.kernel_dest = kernel_dest
        self.kernel_repo = kernel_repo
        self.kernel_dir = kernel_dir
        self.auto_clone = auto_clone
        
    def _check_wsl(self) -> bool:
        """Check if running in WSL environment"""
        try:
            with open('/proc/version', 'r') as f:
                return 'microsoft' in f.read().lower()
        except:
            return False
    
    def _get_windows_user(self) -> str:
        """Get Windows username"""
        if self.is_wsl:
            try:
                result = subprocess.run(['cmd.exe', '/c', 'echo %USERNAME%'], 
                                      capture_output=True, text=True)
                return result.stdout.strip()
            except:
                return ""
        else:
            return os.environ.get('USERNAME', '')
    
    def _log(self, message: str, level: str = "INFO"):
        """Logging function with colors"""
        colors = {
            "INFO": "\033[0;34m",
            "SUCCESS": "\033[0;32m", 
            "WARNING": "\033[1;33m",
            "ERROR": "\033[0;31m"
        }
        reset = "\033[0m"
        timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        print(f"{colors.get(level, '')}{timestamp} [{level}]{reset} {message}")
    
    def _save_state(self, state: Dict[str, Any]):
        """Save automation state to file"""
        self.temp_dir.mkdir(exist_ok=True)
        with open(self.state_file, 'w') as f:
            json.dump(state, f, indent=2)
    
    def _load_state(self) -> Dict[str, Any]:
        """Load automation state from file"""
        if self.state_file.exists():
            with open(self.state_file, 'r') as f:
                return json.load(f)
        return {}
    
    def _run_command(self, cmd: list, check: bool = True, capture_output: bool = False) -> subprocess.CompletedProcess:
        """Run command with logging"""
        self._log(f"Running: {' '.join(cmd)}")
        if capture_output:
            return subprocess.run(cmd, check=check, capture_output=True, text=True)
        return subprocess.run(cmd, check=check, text=True)
    
    def phase1_wsl_tasks(self):
        """Phase 1: Tasks to run inside WSL2"""
        if not self.is_wsl:
            self._log("Phase 1 must run inside WSL2", "ERROR")
            return False
            
        self._log("Starting Phase 1: WSL2 tasks")
        
        # Save current directory
        original_dir = os.getcwd()
        
        try:
            # Example WSL2 tasks (customize as needed)
            self._log("Building kernel...")
            
            # Clone and build kernel (simplified version of your script)
            if not Path(self.kernel_dir).exists():
                if self.auto_clone:
                    self._log(f"Cloning kernel repository from {self.kernel_repo}...")
                    self._run_command([
                        "git", "clone", self.kernel_repo, self.kernel_dir,
                        "--depth=1", "-b", self.kernel_branch
                    ])
                else:
                    self._log(f"Kernel directory '{self.kernel_dir}' not found!", "ERROR")
                    self._log(f"Please clone the repository manually or set auto_clone=True", "ERROR")
                    return False
            else:
                self._log(f"Using existing kernel directory: {self.kernel_dir}")
            
            os.chdir(self.kernel_dir)
            
            # Install dependencies
            self._run_command([
                "sudo", "apt", "update"
            ])
            self._run_command([
                "sudo", "apt", "install", "-y",
                "build-essential", "flex", "bison", "libssl-dev", 
                "libelf-dev", "bc", "python3", "pahole", "cpio"
            ])
            
            # Build kernel
            try:
                nproc_result = self._run_command(["nproc"], capture_output=True)
                nproc = nproc_result.stdout.strip() if nproc_result.stdout else "4"
            except:
                nproc = "4"  # fallback
            self._run_command([
                "make", f"-j{nproc}", "KCONFIG_CONFIG=Microsoft/config-wsl"
            ])
            
            # Install modules
            self._run_command(["sudo", "make", "modules_install", "headers_install"])
            
            # Copy kernel to Windows
            kernel_path = Path("arch/x86/boot/bzImage")
            if kernel_path.exists():
                self._run_command(["cp", str(kernel_path), self.kernel_dest])
                self._log(f"Kernel copied to {self.kernel_dest}", "SUCCESS")
            else:
                self._log("Kernel image not found! Build may have failed.", "ERROR")
                return False
            
            os.chdir("..")
            
            # Save state for Phase 2
            state = {
                "phase1_completed": True,
                "kernel_built": True,
                "windows_user": self.windows_user,
                "timestamp": time.time()
            }
            self._save_state(state)
            
            # Create Phase 2 script on Windows
            self._create_phase2_script()
            
            self._log("Phase 1 completed successfully", "SUCCESS")
            self._log("Starting Phase 2 on Windows...")
            
            # Trigger Phase 2 on Windows
            self._trigger_phase2()
            
            return True
            
        except subprocess.CalledProcessError as e:
            self._log(f"Command failed: {e}", "ERROR")
            return False
        except Exception as e:
            self._log(f"Phase 1 failed: {e}", "ERROR")
            return False
        finally:
            # Always return to original directory
            os.chdir(original_dir)
    
    def _create_phase2_script(self):
        """Create the Phase 2 PowerShell script on Windows"""
        if not self.windows_user:
            self._log("Cannot determine Windows user", "ERROR")
            self._log("Please set WINDOWS_USER environment variable or run from WSL2", "ERROR")
            raise RuntimeError("Windows user not detected")
            
        ps_script = f"""
# Phase 2: Windows tasks
$ErrorActionPreference = "Stop"

Write-Host "Starting Phase 2: Windows tasks" -ForegroundColor Blue

try {{
    # Create .wslconfig
    $wslConfigPath = "C:\\Users\\{self.windows_user}\\.wslconfig"
    $wslConfig = @"
[wsl2]
kernel={self.kernel_dest}
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
    
    # Restart WSL (optional - will happen automatically on next use)
    Write-Host "WSL automation completed successfully!" -ForegroundColor Green
    Write-Host "WSL will use the new kernel on next startup." -ForegroundColor Green
    
}} catch {{
    Write-Host "Phase 2 failed: $_" -ForegroundColor Red
    exit 1
}}

# Cleanup
Remove-Item "C:\\temp\\wsl2_automation\\phase2.ps1" -ErrorAction SilentlyContinue
"""
        
        # Write PowerShell script to Windows
        ps_script_path = Path("/mnt/c/temp/wsl2_automation")
        ps_script_path.mkdir(parents=True, exist_ok=True)
        
        with open(ps_script_path / "phase2.ps1", 'w') as f:
            f.write(ps_script)
        
        self._log("Phase 2 script created on Windows", "SUCCESS")
    
    def _trigger_phase2(self):
        """Trigger Phase 2 execution on Windows"""
        try:
            # Execute PowerShell script on Windows
            self._run_command([
                "powershell.exe", "-ExecutionPolicy", "Bypass",
                "-File", "C:\\temp\\wsl2_automation\\phase2.ps1"
            ])
        except subprocess.CalledProcessError:
            # This is expected as WSL will be shut down
            self._log("Phase 2 triggered (WSL shutdown expected)", "SUCCESS")
    
    def phase2_windows_tasks(self):
        """Phase 2: Tasks to run on Windows (fallback if called directly)"""
        if self.is_wsl:
            self._log("Phase 2 should run on Windows, not WSL", "ERROR")
            return False
            
        self._log("Starting Phase 2: Windows tasks")
        
        try:
            state = self._load_state()
            if not state.get("phase1_completed"):
                self._log("Phase 1 not completed", "ERROR")
                return False
            
            # Create .wslconfig
            wsl_config_path = Path(f"C:/Users/{self.windows_user}/.wslconfig")
            wsl_config_content = f"[wsl2]\nkernel={self.kernel_dest}\n"
            
            with open(wsl_config_path, 'w') as f:
                f.write(wsl_config_content)
            
            self._log(f"WSL config created at {wsl_config_path}", "SUCCESS")
            
            # Shutdown WSL
            self._log("Shutting down WSL...")
            self._run_command(["wsl", "--shutdown"])
            
            self._log("Phase 2 completed successfully", "SUCCESS")
            return True
            
        except Exception as e:
            self._log(f"Phase 2 failed: {e}", "ERROR")
            return False
    
    def cleanup(self):
        """Clean up temporary files"""
        try:
            if self.temp_dir.exists():
                shutil.rmtree(self.temp_dir)
            self._log("Cleanup completed", "SUCCESS")
        except Exception as e:
            self._log(f"Cleanup failed: {e}", "WARNING")

def main():
    parser = argparse.ArgumentParser(description="WSL2 Two-Phase Automation")
    parser.add_argument("--phase", choices=["1", "2"], default="1",
                       help="Which phase to run (default: 1)")
    parser.add_argument("--cleanup", action="store_true",
                       help="Clean up temporary files")
    parser.add_argument("--kernel-branch", default="linux-msft-wsl-6.6.y",
                       help="Kernel branch to build (default: linux-msft-wsl-6.6.y)")
    parser.add_argument("--kernel-dest", default="C:/bzImage",
                       help="Destination for kernel image (default: C:/bzImage)")
    parser.add_argument("--kernel-repo", 
                       default="https://github.com/microsoft/WSL2-Linux-Kernel.git",
                       help="Kernel repository URL")
    parser.add_argument("--kernel-dir", default="WSL2-Linux-Kernel",
                       help="Local kernel directory name (default: WSL2-Linux-Kernel)")
    parser.add_argument("--no-clone", action="store_true",
                       help="Don't auto-clone the repository")
    
    args = parser.parse_args()
    
    automator = WSL2Automator(
        kernel_branch=args.kernel_branch,
        kernel_dest=args.kernel_dest,
        kernel_repo=args.kernel_repo,
        kernel_dir=args.kernel_dir,
        auto_clone=not args.no_clone
    )
    
    if args.cleanup:
        automator.cleanup()
        return
    
    if args.phase == "1":
        success = automator.phase1_wsl_tasks()
    else:
        success = automator.phase2_windows_tasks()
    
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main() 