//go:build !windows

package session

import (
	"os/exec"
	"syscall"
)

// setDaemonSysProcAttr detaches the daemon into a new session/process group
// so it survives after the parent CLI process exits.
func setDaemonSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
