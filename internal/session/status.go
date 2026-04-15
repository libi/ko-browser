package session

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

// PidInfo is persisted to disk by the daemon so that status can be
// queried independently of whether the daemon is responsive.
type PidInfo struct {
	Pid       int       `json:"pid"`
	StartTime time.Time `json:"startTime"`
	Headed    bool      `json:"headed"`
	Session   string    `json:"session"`
}

// StatusInfo holds the status of a daemon and its associated browser process.
type StatusInfo struct {
	Session string `json:"session"`

	// Daemon status
	DaemonRunning  bool   `json:"daemonRunning"`
	DaemonPID      int    `json:"daemonPid,omitempty"`
	DaemonUptime   string `json:"daemonUptime,omitempty"`
	SocketPath     string `json:"socketPath"`
	SocketExists   bool   `json:"socketExists"`
	DaemonPingOK   bool   `json:"daemonPingOk"`
	DaemonPingTime string `json:"daemonPingTime,omitempty"`

	// Browser status — inferred from daemon lifecycle.
	// The daemon owns the browser process; if the daemon is alive the browser is alive.
	BrowserRunning bool `json:"browserRunning"`
	BrowserHeaded  bool `json:"browserHeaded"`
}

// ----- pidfile helpers -----

func pidfilePath(name string) string {
	return socketPath(name) + ".pid" // e.g. /tmp/ko-browser-default.sock.pid
}

func writePidfile(name string, headed bool) error {
	info := PidInfo{
		Pid:       os.Getpid(),
		StartTime: time.Now(),
		Headed:    headed,
		Session:   name,
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(pidfilePath(name), data, 0600)
}

func readPidfile(name string) (*PidInfo, error) {
	data, err := os.ReadFile(pidfilePath(name))
	if err != nil {
		return nil, err
	}
	var info PidInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func removePidfile(name string) {
	_ = os.Remove(pidfilePath(name))
}

// isProcessAlive checks whether a process with the given PID exists.
// Works on macOS, Linux and Windows.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 does not actually send a signal — it only checks if the
	// process exists.  This works on all Unix systems.  On Windows the
	// syscall package maps Signal(0) to a process-existence check as well
	// (since Go 1.20+).
	return p.Signal(syscall.Signal(0)) == nil
}

// ----- public API -----

// GetStatus probes the daemon pidfile, socket and process to build a StatusInfo.
// This works independently of whether the daemon is running or responsive.
func GetStatus(sessionName string) StatusInfo {
	opts := Options{Name: sessionName}.normalized()
	sockPath := socketPath(opts.Name)

	info := StatusInfo{
		Session:    opts.Name,
		SocketPath: sockPath,
	}

	// 1. Check socket file existence.
	if _, err := os.Stat(sockPath); err == nil {
		info.SocketExists = true
	}

	// 2. Read pidfile (written by daemon on startup).
	pidInfo, _ := readPidfile(opts.Name)
	if pidInfo != nil {
		info.DaemonPID = pidInfo.Pid
		info.BrowserHeaded = pidInfo.Headed
	}

	// 3. Try pinging the daemon via its socket.
	if info.SocketExists {
		start := time.Now()
		conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
		if err == nil {
			_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
			_ = json.NewEncoder(conn).Encode(&Request{Command: "session.info"})
			var resp Response
			if decErr := json.NewDecoder(conn).Decode(&resp); decErr == nil && resp.OK {
				info.DaemonPingOK = true
				info.DaemonPingTime = time.Since(start).Round(time.Millisecond).String()
			}
			conn.Close()
		}
	}

	// 4. Determine daemon status by combining ping result and PID liveness.
	if info.DaemonPingOK {
		info.DaemonRunning = true
	} else if info.DaemonPID > 0 && isProcessAlive(info.DaemonPID) {
		// Process is alive but not answering pings — probably still starting
		// up, or currently blocked on a long operation.
		info.DaemonRunning = true
	}

	if info.DaemonRunning && pidInfo != nil {
		info.DaemonUptime = time.Since(pidInfo.StartTime).Round(time.Second).String()
	}

	// 5. Browser status — the daemon manages the browser lifecycle.
	//    If the daemon is alive the browser is alive.
	info.BrowserRunning = info.DaemonRunning

	return info
}

// FormatStatus formats StatusInfo as human-readable text.
func FormatStatus(s StatusInfo) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Session:  %s\n", s.Session))
	b.WriteString(fmt.Sprintf("Socket:   %s\n", s.SocketPath))

	// Daemon
	switch {
	case s.DaemonRunning && s.DaemonPingOK:
		detail := fmt.Sprintf("running (pid: %d", s.DaemonPID)
		if s.DaemonUptime != "" {
			detail += fmt.Sprintf(", uptime: %s", s.DaemonUptime)
		}
		detail += fmt.Sprintf(", ping: %s)", s.DaemonPingTime)
		b.WriteString(fmt.Sprintf("Daemon:   %s\n", detail))
	case s.DaemonRunning && !s.DaemonPingOK:
		b.WriteString(fmt.Sprintf("Daemon:   busy / not responding (pid: %d, uptime: %s)\n", s.DaemonPID, s.DaemonUptime))
	case s.SocketExists:
		b.WriteString("Daemon:   not running (stale socket)\n")
	default:
		b.WriteString("Daemon:   not running\n")
	}

	// Browser
	if s.BrowserRunning {
		mode := "headless"
		if s.BrowserHeaded {
			mode = "headed"
		}
		b.WriteString(fmt.Sprintf("Browser:  running (%s)\n", mode))
	} else {
		b.WriteString("Browser:  not running\n")
	}

	return b.String()
}
