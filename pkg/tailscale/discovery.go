package tailscale

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Machine represents a device on the Tailscale network
type Machine struct {
	Name        string `json:"name"`
	IP          string `json:"ip"`
	Online      bool   `json:"online"`
	OS          string `json:"os"`
	HostName    string `json:"hostname"`
	DNSName     string `json:"dns_name"`
	TailscaleIP string `json:"tailscale_ip"`
}

// TailscaleStatus represents the output from `tailscale status --json`
type TailscaleStatus struct {
	Self  PeerInfo            `json:"Self"`
	Peer  map[string]PeerInfo `json:"Peer"`
	User  map[string]UserInfo `json:"User"`
}

// PeerInfo represents a peer in the Tailscale network
type PeerInfo struct {
	ID            string   `json:"ID"`
	PublicKey     string   `json:"PublicKey"`
	HostName      string   `json:"HostName"`
	DNSName       string   `json:"DNSName"`
	OS            string   `json:"OS"`
	UserID        int      `json:"UserID"`
	TailscaleIPs  []string `json:"TailscaleIPs"`
	Online        bool     `json:"Online"`
	Active        bool     `json:"Active"`
	ExitNode      bool     `json:"ExitNode"`
	ExitNodeOption bool    `json:"ExitNodeOption"`
}

// UserInfo represents a user in the Tailscale network
type UserInfo struct {
	ID          int    `json:"ID"`
	LoginName   string `json:"LoginName"`
	DisplayName string `json:"DisplayName"`
}

// DiscoverMachines discovers all machines on the Tailscale network
// Returns a list of machines with their connection information
func DiscoverMachines() ([]Machine, error) {
	// Check if tailscale is installed
	if !isTailscaleInstalled() {
		return nil, fmt.Errorf("tailscale is not installed or not in PATH")
	}

	// Run tailscale status --json
	cmd := exec.Command("tailscale", "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run tailscale status: %w", err)
	}

	// Parse JSON output
	var status TailscaleStatus
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse tailscale status: %w", err)
	}

	// Convert peers to Machine list
	machines := make([]Machine, 0, len(status.Peer))

	// Add self (local machine)
	if len(status.Self.TailscaleIPs) > 0 {
		machines = append(machines, Machine{
			Name:        status.Self.HostName,
			IP:          status.Self.TailscaleIPs[0],
			Online:      true, // Self is always online
			OS:          status.Self.OS,
			HostName:    status.Self.HostName,
			DNSName:     status.Self.DNSName,
			TailscaleIP: status.Self.TailscaleIPs[0],
		})
	}

	// Add peers
	for _, peer := range status.Peer {
		if len(peer.TailscaleIPs) == 0 {
			continue
		}

		machine := Machine{
			Name:        peer.HostName,
			IP:          peer.TailscaleIPs[0],
			Online:      peer.Online,
			OS:          peer.OS,
			HostName:    peer.HostName,
			DNSName:     peer.DNSName,
			TailscaleIP: peer.TailscaleIPs[0],
		}

		machines = append(machines, machine)
	}

	return machines, nil
}

// ProbeForIOSAgent checks if a machine is running ios-agent server
// This is a simple TCP connection check to port 4723 (default WebDriverAgent port)
// Returns true if the port is accessible, false otherwise
func ProbeForIOSAgent(ip string) bool {
	// For MVP, we skip the actual probe and return false
	// In a full implementation, this would:
	// 1. Try to connect to port 4723 (WebDriverAgent)
	// 2. Or try SSH and check for ios-agent process
	// 3. Or try a custom discovery protocol

	// TODO: Implement actual probe in post-MVP
	return false
}

// isTailscaleInstalled checks if tailscale CLI is available
func isTailscaleInstalled() bool {
	cmd := exec.Command("which", "tailscale")
	err := cmd.Run()
	return err == nil
}

// GetMachineByName finds a machine by hostname
func GetMachineByName(name string) (*Machine, error) {
	machines, err := DiscoverMachines()
	if err != nil {
		return nil, err
	}

	// Normalize name for comparison
	normalizedName := strings.ToLower(strings.TrimSpace(name))

	for _, machine := range machines {
		if strings.ToLower(machine.HostName) == normalizedName ||
		   strings.ToLower(machine.Name) == normalizedName {
			return &machine, nil
		}
	}

	return nil, fmt.Errorf("machine not found: %s", name)
}

// GetMachineByIP finds a machine by its Tailscale IP
func GetMachineByIP(ip string) (*Machine, error) {
	machines, err := DiscoverMachines()
	if err != nil {
		return nil, err
	}

	for _, machine := range machines {
		if machine.TailscaleIP == ip || machine.IP == ip {
			return &machine, nil
		}
	}

	return nil, fmt.Errorf("machine not found with IP: %s", ip)
}
