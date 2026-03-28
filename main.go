package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type Tunnel struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Type                 string    `json:"type"` // "local" or "remote"
	LocalPort            string    `json:"local_port"`
	RemoteHost           string    `json:"remote_host"`
	RemotePort           string    `json:"remote_port"`
	SshHost              string    `json:"ssh_host"`
	SshPort              string    `json:"ssh_port"`
	SshUser              string    `json:"ssh_user"`
	SshKey               string    `json:"ssh_key"` // SSH key path
	SshPass              string    `json:"-"`       // SSH password, never serialized
	Status               string    `json:"status"`  // "running" or "stopped"
	Pid                  int       `json:"-"`       // process PID, internal use
	Process              *exec.Cmd `json:"-"`       // internal use, never serialized
	CreatedAt            int64     `json:"-"`       // creation timestamp, internal use
	AutoReconnect        bool      `json:"auto_reconnect"`
	ReconnectDelay       int       `json:"reconnect_delay"`
	LastReconnectTime    int64     `json:"-"`
	ReconnectAttempts    int       `json:"-"`
	MaxReconnectAttempts int       `json:"-"`
}

var (
	tunnels   = make(map[string]*Tunnel)
	tunnelMux sync.RWMutex
)

const (
	defaultReconnectDelay     = 5
	defaultMaxReconnectPerMin = 5
)

func main() {
	if err := loadConfig(); err != nil {
		log.Printf("[ERROR] Failed to load config: %v", err)
	}

	checkExistingProcesses()

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/api/tunnels", listTunnelsHandler).Methods("GET")
	r.HandleFunc("/api/tunnels", createTunnelHandler).Methods("POST")
	r.HandleFunc("/api/tunnels/{id}", getTunnelHandler).Methods("GET")
	r.HandleFunc("/api/tunnels/{id}", updateTunnelHandler).Methods("PUT")
	r.HandleFunc("/api/tunnels/{id}", deleteTunnelHandler).Methods("DELETE")
	r.HandleFunc("/api/tunnels/{id}/start", startTunnelHandler).Methods("POST")
	r.HandleFunc("/api/tunnels/{id}/stop", stopTunnelHandler).Methods("POST")
	r.HandleFunc("/api/tunnels/{id}/status", statusTunnelHandler).Methods("GET")
	r.HandleFunc("/api/tunnels/{id}/stats", statsTunnelHandler).Methods("GET")
	r.HandleFunc("/api/ping", pingHandler).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("[INFO] SSH Tunnel Manager starting on :11108")
	log.Println("[INFO] Config file: config.json")
	log.Println("[INFO] Opening browser...")
	go openBrowser("http://localhost:11108")
	log.Fatal(http.ListenAndServe(":11108", r))
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(HOME_HTML))
}

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))
	} else {
		cmd = exec.Command("ps", "-p", fmt.Sprintf("%d", pid))
	}

	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), fmt.Sprintf(" %d ", pid)) ||
		strings.Contains(string(output), fmt.Sprintf("\n%d ", pid)) ||
		strings.Contains(string(output), fmt.Sprintf(" %d\n", pid))
}

func checkExistingProcesses() {
	tunnelMux.Lock()
	defer tunnelMux.Unlock()

	for _, t := range tunnels {
		if t.Pid > 0 && processExists(t.Pid) {
			t.Status = "running"
			log.Printf("[INFO] Restored tunnel state: %s (PID: %d)", t.Name, t.Pid)
		} else {
			t.Status = "stopped"
			t.Pid = 0
		}
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	tunnelMux.RLock()
	runningCount := 0
	totalCount := len(tunnels)
	for _, t := range tunnels {
		if t.Status == "running" {
			runningCount++
		}
	}
	tunnelMux.RUnlock()

	sendJSON(w, map[string]interface{}{
		"status":          "ok",
		"timestamp":       time.Now().Unix(),
		"tunnels_total":   totalCount,
		"tunnels_running": runningCount,
	})
}

func waitForPortOpen(port string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "localhost:"+port, 1*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func listTunnelsHandler(w http.ResponseWriter, r *http.Request) {
	tunnelMux.RLock()
	defer tunnelMux.RUnlock()

	tunnelList := make([]*Tunnel, 0, len(tunnels))
	for _, t := range tunnels {
		tunnelList = append(tunnelList, t)
	}

	sort.Slice(tunnelList, func(i, j int) bool {
		return tunnelList[i].CreatedAt > tunnelList[j].CreatedAt
	})

	sendJSON(w, tunnelList)
}

func createTunnelHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string `json:"name"`
		Type           string `json:"type"` // "local" or "remote"
		LocalPort      string `json:"local_port"`
		RemoteHost     string `json:"remote_host"`
		RemotePort     string `json:"remote_port"`
		SshHost        string `json:"ssh_host"`
		SshPort        string `json:"ssh_port"`
		SshUser        string `json:"ssh_user"`
		SshKey         string `json:"ssh_key"`
		SshPass        string `json:"ssh_pass"`
		AutoReconnect  bool   `json:"auto_reconnect"`
		ReconnectDelay int    `json:"reconnect_delay"`
	}

	if err := parseJSON(r, &req); err != nil {
		log.Printf("[ERROR] Create tunnel failed: invalid request from %s", r.RemoteAddr)
		sendError(w, "Invalid request", 400)
		return
	}

	if req.Name == "" || req.SshHost == "" || req.SshUser == "" {
		log.Printf("[WARN] Create tunnel failed: missing required fields")
		sendError(w, "Missing required fields", 400)
		return
	}

	if req.Type != "local" && req.Type != "remote" {
		log.Printf("[WARN] Create tunnel failed: invalid type '%s'", req.Type)
		sendError(w, "Type must be 'local' or 'remote'", 400)
		return
	}

	if req.SshKey == "" && req.SshPass == "" {
		log.Printf("[WARN] Create tunnel failed: no SSH credentials provided for '%s'", req.Name)
		sendError(w, "Must provide SSH key path or password", 400)
		return
	}

	tunnelMux.Lock()
	id := req.Name
	tunnel := &Tunnel{
		ID:             id,
		Name:           req.Name,
		Type:           req.Type,
		LocalPort:      req.LocalPort,
		RemoteHost:     req.RemoteHost,
		RemotePort:     req.RemotePort,
		SshHost:        req.SshHost,
		SshPort:        req.SshPort,
		SshUser:        req.SshUser,
		SshKey:         req.SshKey,
		SshPass:        req.SshPass,
		Status:         "stopped",
		CreatedAt:      time.Now().Unix(),
		AutoReconnect:  req.AutoReconnect,
		ReconnectDelay: req.ReconnectDelay,
	}
	tunnels[req.Name] = tunnel
	tunnelMux.Unlock()

	if err := saveConfig(); err != nil {
		log.Printf("[ERROR] Create tunnel '%s' failed to save config: %v", req.Name, err)
	} else {
		log.Printf("[INFO] Tunnel created: %s (%s:%s -> %s:%s, ssh: %s@%s:%s)",
			req.Name, req.Type, req.LocalPort, req.RemoteHost, req.RemotePort, req.SshUser, req.SshHost, req.SshPort)
	}
	sendJSON(w, tunnel)
}

func getTunnelHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	tunnelMux.RLock()
	tunnel, ok := tunnels[id]
	tunnelMux.RUnlock()

	if !ok {
		sendError(w, "Tunnel not found", 404)
		return
	}

	sendJSON(w, tunnel)
}

func updateTunnelHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req struct {
		Name           string `json:"name"`
		Type           string `json:"type"`
		LocalPort      string `json:"local_port"`
		RemoteHost     string `json:"remote_host"`
		RemotePort     string `json:"remote_port"`
		SshHost        string `json:"ssh_host"`
		SshPort        string `json:"ssh_port"`
		SshUser        string `json:"ssh_user"`
		SshKey         string `json:"ssh_key"`
		SshPass        string `json:"ssh_pass"`
		AutoReconnect  bool   `json:"auto_reconnect"`
		ReconnectDelay int    `json:"reconnect_delay"`
	}

	if err := parseJSON(r, &req); err != nil {
		sendError(w, "Invalid request", 400)
		return
	}

	if req.Name == "" || req.SshHost == "" || req.SshUser == "" {
		sendError(w, "Missing required fields", 400)
		return
	}

	if req.Type != "local" && req.Type != "remote" {
		sendError(w, "Type must be 'local' or 'remote'", 400)
		return
	}

	tunnelMux.Lock()
	tunnel, ok := tunnels[id]
	if !ok {
		tunnelMux.Unlock()
		sendError(w, "Tunnel not found", 404)
		return
	}

	if tunnel.Status == "running" {
		tunnelMux.Unlock()
		sendError(w, "Cannot edit a running tunnel, stop it first", 400)
		return
	}

	if id != req.Name {
		delete(tunnels, id)
	}

	tunnel.Name = req.Name
	tunnel.Type = req.Type
	tunnel.LocalPort = req.LocalPort
	tunnel.RemoteHost = req.RemoteHost
	tunnel.RemotePort = req.RemotePort
	tunnel.SshHost = req.SshHost
	tunnel.SshPort = req.SshPort
	tunnel.SshUser = req.SshUser
	tunnel.SshKey = req.SshKey
	tunnel.SshPass = req.SshPass
	tunnel.AutoReconnect = req.AutoReconnect
	tunnel.ReconnectDelay = req.ReconnectDelay

	tunnels[req.Name] = tunnel
	tunnelMux.Unlock()

	if err := saveConfig(); err != nil {
		log.Printf("[ERROR] Update tunnel '%s' failed to save config: %v", req.Name, err)
	} else {
		log.Printf("[INFO] Tunnel updated: %s", req.Name)
	}
	sendJSON(w, tunnel)
}

func deleteTunnelHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	tunnelMux.Lock()
	tunnel, ok := tunnels[id]
	if !ok {
		tunnelMux.Unlock()
		log.Printf("[WARN] Delete tunnel failed: not found '%s'", id)
		sendError(w, "Tunnel not found", 404)
		return
	}

	if tunnel.Status == "running" && tunnel.Process != nil {
		tunnel.Process.Process.Kill()
		log.Printf("[INFO] Stopped running tunnel before delete: %s", id)
	}

	delete(tunnels, id)
	tunnelMux.Unlock()

	if err := saveConfig(); err != nil {
		log.Printf("[ERROR] Delete tunnel '%s' failed to save config: %v", id, err)
	}
	log.Printf("[INFO] Tunnel deleted: %s", tunnel.Name)
	sendJSON(w, map[string]string{"message": "Tunnel deleted"})
}

func startTunnelHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	tunnelMux.Lock()
	tunnel, ok := tunnels[id]
	tunnelMux.Unlock()

	if !ok {
		log.Printf("[WARN] Start tunnel failed: not found '%s'", id)
		sendError(w, "Tunnel not found", 404)
		return
	}

	if tunnel.Status == "running" || tunnel.Status == "starting" {
		log.Printf("[WARN] Start tunnel failed: tunnel '%s' is already %s", id, tunnel.Status)
		sendError(w, fmt.Sprintf("Tunnel is already %s", tunnel.Status), 400)
		return
	}

	const maxRetries = 3
	const retryDelay = 5
	const readyTimeout = 30

	var args []string
	knownHostsFile := os.Getenv("USERPROFILE") + "/.ssh/known_hosts"
	if runtime.GOOS == "windows" {
		knownHostsFile = os.Getenv("USERPROFILE") + "\\.ssh\\known_hosts"
	}

	localPort := tunnel.LocalPort
	if tunnel.Type == "remote" {
		localPort = tunnel.RemotePort
	}

	if tunnel.Type == "local" {
		sshPort := "22"
		if tunnel.SshPort != "" {
			sshPort = tunnel.SshPort
		}
		args = []string{
			"-N", "-L", fmt.Sprintf("%s:%s:%s", tunnel.LocalPort, tunnel.RemoteHost, tunnel.RemotePort),
			"-p", sshPort,
			"-o", "ServerAliveInterval=30",
			"-o", "ServerAliveCountMax=3",
			"-o", "TCPKeepAlive=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "UserKnownHostsFile=" + knownHostsFile,
		}
	} else {
		sshPort := "22"
		if tunnel.SshPort != "" {
			sshPort = tunnel.SshPort
		}
		args = []string{
			"-N", "-R", fmt.Sprintf("%s:%s:%s", tunnel.RemotePort, tunnel.LocalPort, tunnel.LocalPort),
			"-p", sshPort,
			"-o", "ServerAliveInterval=30",
			"-o", "ServerAliveCountMax=3",
			"-o", "TCPKeepAlive=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "UserKnownHostsFile=" + knownHostsFile,
		}
	}

	if tunnel.SshKey != "" {
		args = append(args, "-i", tunnel.SshKey)
	}

	args = append(args, fmt.Sprintf("%s@%s", tunnel.SshUser, tunnel.SshHost))

	var cmd *exec.Cmd
	var startErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		cmd = exec.Command("ssh", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if tunnel.SshPass != "" {
			cmd.Env = append(os.Environ(), "SSH_ASKPASS="+tunnel.SshPass, "DISPLAY=:0")
		}

		startErr = cmd.Start()
		if startErr != nil {
			log.Printf("[WARN] Start tunnel '%s' failed (attempt %d/%d): %v", tunnel.Name, attempt, maxRetries, startErr)
			if attempt < maxRetries {
				log.Printf("[INFO] Retrying in %d seconds...", retryDelay)
				time.Sleep(time.Duration(retryDelay) * time.Second)
			}
			continue
		}

		tunnelMux.Lock()
		tunnel.Process = cmd
		tunnel.Pid = cmd.Process.Pid
		tunnel.Status = "starting"
		tunnelMux.Unlock()

		log.Printf("[INFO] Tunnel '%s' starting, waiting for port %s to be ready...", tunnel.Name, localPort)

		if waitForPortOpen(localPort, time.Duration(readyTimeout)*time.Second) {
			tunnelMux.Lock()
			tunnel.Status = "running"
			tunnelMux.Unlock()
			log.Printf("[INFO] Tunnel ready: %s (PID: %d, port: %s)", tunnel.Name, cmd.Process.Pid, localPort)
			break
		}

		log.Printf("[WARN] Tunnel '%s' port not ready after %d seconds, killing process...", tunnel.Name, readyTimeout)
		cmd.Process.Kill()
		cmd.Wait()

		tunnelMux.Lock()
		tunnel.Process = nil
		tunnel.Pid = 0
		tunnel.Status = "stopped"
		tunnelMux.Unlock()

		if attempt < maxRetries {
			log.Printf("[INFO] Retrying tunnel '%s' in %d seconds (attempt %d/%d)...", tunnel.Name, retryDelay, attempt+1, maxRetries)
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
	}

	if tunnel.Status != "running" {
		log.Printf("[ERROR] Start tunnel failed after %d attempts: %s", maxRetries, tunnel.Name)
		sendError(w, fmt.Sprintf("Failed to start tunnel '%s' after %d attempts", tunnel.Name, maxRetries), 500)
		return
	}

	go func() {
		err := cmd.Wait()
		tunnelMux.Lock()
		wasRunning := tunnel.Status == "running"
		tunnel.Status = "stopped"
		tunnel.Process = nil
		tunnel.Pid = 0
		tunnelMux.Unlock()
		if err != nil {
			log.Printf("[WARN] Tunnel exited with error: %s - %v", tunnel.Name, err)
		} else {
			log.Printf("[INFO] Tunnel stopped: %s", tunnel.Name)
		}

		if wasRunning && tunnel.AutoReconnect {
			delay := tunnel.ReconnectDelay
			if delay <= 0 {
				delay = defaultReconnectDelay
			}
			maxAttempts := tunnel.MaxReconnectAttempts
			if maxAttempts <= 0 {
				maxAttempts = defaultMaxReconnectPerMin
			}

			tunnelMux.Lock()
			tunnel.ReconnectAttempts++
			tunnel.LastReconnectTime = time.Now().Unix()
			shouldReconnect := tunnel.ReconnectAttempts <= maxAttempts
			tunnelMux.Unlock()

			if shouldReconnect {
				log.Printf("[INFO] Auto-reconnecting tunnel '%s' in %d seconds (attempt %d/%d)",
					tunnel.Name, delay, tunnel.ReconnectAttempts, maxAttempts)
				time.Sleep(time.Duration(delay) * time.Second)
				startTunnelAsync(tunnel)
			} else {
				log.Printf("[WARN] Tunnel '%s' exceeded max reconnection attempts (%d), stopping", tunnel.Name, maxAttempts)
			}
		}
	}()

	sendJSON(w, tunnel)
}

func stopTunnelHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	tunnelMux.Lock()
	tunnel, ok := tunnels[id]
	tunnelMux.Unlock()

	if !ok {
		log.Printf("[WARN] Stop tunnel failed: not found '%s'", id)
		sendError(w, "Tunnel not found", 404)
		return
	}

	if tunnel.Status != "running" || tunnel.Process == nil {
		log.Printf("[WARN] Stop tunnel failed: not running '%s'", id)
		sendError(w, "Tunnel not running", 400)
		return
	}

	if err := tunnel.Process.Process.Kill(); err != nil {
		log.Printf("[ERROR] Stop tunnel failed: %s - %v", tunnel.Name, err)
		sendError(w, fmt.Sprintf("Failed to stop tunnel: %v", err), 500)
		return
	}

	tunnelMux.Lock()
	tunnel.Status = "stopped"
	tunnel.Process = nil
	tunnel.Pid = 0
	tunnelMux.Unlock()

	log.Printf("[INFO] Tunnel stopped: %s", tunnel.Name)
	sendJSON(w, tunnel)
}

func statusTunnelHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	tunnelMux.RLock()
	tunnel, ok := tunnels[id]
	tunnelMux.RUnlock()

	if !ok {
		sendError(w, "Tunnel not found", 404)
		return
	}

	sendJSON(w, map[string]string{"status": tunnel.Status})
}

type tunnelStats struct {
	Status    string `json:"status"`
	BytesSent int64  `json:"bytes_sent"`
	BytesRecv int64  `json:"bytes_recv"`
}

func statsTunnelHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	tunnelMux.RLock()
	tunnel, ok := tunnels[id]
	tunnelMux.RUnlock()

	if !ok {
		sendError(w, "Tunnel not found", 404)
		return
	}

	stats := tunnelStats{
		Status:    tunnel.Status,
		BytesSent: 0,
		BytesRecv: 0,
	}

	if tunnel.Pid > 0 && tunnel.Status == "running" {
		stats.BytesSent, stats.BytesRecv = getProcessBytes(tunnel.Pid)
	}

	sendJSON(w, stats)
}

func getProcessBytes(pid int) (int64, int64) {
	if runtime.GOOS == "windows" {
		return getWindowsProcessBytes(pid)
	}
	return getLinuxProcessBytes(pid)
}

func getLinuxProcessBytes(pid int) (int64, int64) {
	cmd := exec.Command("cat", fmt.Sprintf("/proc/%d/net/dev", pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	var rx, tx int64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 10 {
			rxBytes, _ := strconv.ParseInt(fields[1], 10, 64)
			txBytes, _ := strconv.ParseInt(fields[9], 10, 64)
			rx += rxBytes
			tx += txBytes
		}
	}
	return rx, tx
}

func getWindowsProcessBytes(pid int) (int64, int64) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("(Get-Process -Id %d -ErrorAction SilentlyContinue).WorkingSet64", pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	bytes, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, 0
	}
	return bytes, bytes
}

func startTunnelAsync(tunnel *Tunnel) {
	const maxRetries = 3
	const retryDelay = 5
	const readyTimeout = 30

	var args []string
	knownHostsFile := os.Getenv("USERPROFILE") + "/.ssh/known_hosts"
	if runtime.GOOS == "windows" {
		knownHostsFile = os.Getenv("USERPROFILE") + "\\.ssh\\known_hosts"
	}

	localPort := tunnel.LocalPort
	if tunnel.Type == "remote" {
		localPort = tunnel.RemotePort
	}

	if tunnel.Type == "local" {
		sshPort := "22"
		if tunnel.SshPort != "" {
			sshPort = tunnel.SshPort
		}
		args = []string{
			"-N", "-L", fmt.Sprintf("%s:%s:%s", tunnel.LocalPort, tunnel.RemoteHost, tunnel.RemotePort),
			"-p", sshPort,
			"-o", "ServerAliveInterval=30",
			"-o", "ServerAliveCountMax=3",
			"-o", "TCPKeepAlive=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "UserKnownHostsFile=" + knownHostsFile,
		}
	} else {
		sshPort := "22"
		if tunnel.SshPort != "" {
			sshPort = tunnel.SshPort
		}
		args = []string{
			"-N", "-R", fmt.Sprintf("%s:%s:%s", tunnel.RemotePort, tunnel.LocalPort, tunnel.LocalPort),
			"-p", sshPort,
			"-o", "ServerAliveInterval=30",
			"-o", "ServerAliveCountMax=3",
			"-o", "TCPKeepAlive=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", "UserKnownHostsFile=" + knownHostsFile,
		}
	}

	if tunnel.SshKey != "" {
		args = append(args, "-i", tunnel.SshKey)
	}

	args = append(args, fmt.Sprintf("%s@%s", tunnel.SshUser, tunnel.SshHost))

	var cmd *exec.Cmd
	var startErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		cmd = exec.Command("ssh", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if tunnel.SshPass != "" {
			cmd.Env = append(os.Environ(), "SSH_ASKPASS="+tunnel.SshPass, "DISPLAY=:0")
		}

		startErr = cmd.Start()
		if startErr != nil {
			log.Printf("[WARN] Start tunnel '%s' failed (attempt %d/%d): %v", tunnel.Name, attempt, maxRetries, startErr)
			if attempt < maxRetries {
				log.Printf("[INFO] Retrying in %d seconds...", retryDelay)
				time.Sleep(time.Duration(retryDelay) * time.Second)
			}
			continue
		}

		tunnelMux.Lock()
		tunnel.Process = cmd
		tunnel.Pid = cmd.Process.Pid
		tunnel.Status = "starting"
		tunnelMux.Unlock()

		log.Printf("[INFO] Tunnel '%s' starting, waiting for port %s to be ready...", tunnel.Name, localPort)

		if waitForPortOpen(localPort, time.Duration(readyTimeout)*time.Second) {
			tunnelMux.Lock()
			tunnel.Status = "running"
			tunnel.ReconnectAttempts = 0
			tunnelMux.Unlock()
			log.Printf("[INFO] Tunnel ready: %s (PID: %d, port: %s)", tunnel.Name, cmd.Process.Pid, localPort)
			break
		}

		log.Printf("[WARN] Tunnel '%s' port not ready after %d seconds, killing process...", tunnel.Name, readyTimeout)
		cmd.Process.Kill()
		cmd.Wait()

		tunnelMux.Lock()
		tunnel.Process = nil
		tunnel.Pid = 0
		tunnel.Status = "stopped"
		tunnelMux.Unlock()

		if attempt < maxRetries {
			log.Printf("[INFO] Retrying tunnel '%s' in %d seconds (attempt %d/%d)...", tunnel.Name, retryDelay, attempt+1, maxRetries)
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}
	}

	if tunnel.Status != "running" {
		log.Printf("[ERROR] Start tunnel failed after %d attempts: %s", maxRetries, tunnel.Name)
		return
	}

	go func() {
		err := cmd.Wait()
		tunnelMux.Lock()
		wasRunning := tunnel.Status == "running"
		tunnel.Status = "stopped"
		tunnel.Process = nil
		tunnel.Pid = 0
		tunnelMux.Unlock()
		if err != nil {
			log.Printf("[WARN] Tunnel exited with error: %s - %v", tunnel.Name, err)
		} else {
			log.Printf("[INFO] Tunnel stopped: %s", tunnel.Name)
		}

		if wasRunning && tunnel.AutoReconnect {
			delay := tunnel.ReconnectDelay
			if delay <= 0 {
				delay = defaultReconnectDelay
			}
			maxAttempts := tunnel.MaxReconnectAttempts
			if maxAttempts <= 0 {
				maxAttempts = defaultMaxReconnectPerMin
			}

			tunnelMux.Lock()
			tunnel.ReconnectAttempts++
			tunnel.LastReconnectTime = time.Now().Unix()
			shouldReconnect := tunnel.ReconnectAttempts <= maxAttempts
			tunnelMux.Unlock()

			if shouldReconnect {
				log.Printf("[INFO] Auto-reconnecting tunnel '%s' in %d seconds (attempt %d/%d)",
					tunnel.Name, delay, tunnel.ReconnectAttempts, maxAttempts)
				time.Sleep(time.Duration(delay) * time.Second)
				startTunnelAsync(tunnel)
			} else {
				log.Printf("[WARN] Tunnel '%s' exceeded max reconnection attempts (%d), stopping", tunnel.Name, maxAttempts)
			}
		}
	}()
}
