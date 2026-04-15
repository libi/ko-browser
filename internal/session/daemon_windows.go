//go:build windows

package session

import (
	"os/exec"
	"syscall"
)

// setDaemonSysProcAttr detaches the daemon on Windows by using
// CREATE_NEW_PROCESS_GROUP so it is not affected by the parent's
// console signals (Ctrl+C, etc.).
func setDaemonSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
