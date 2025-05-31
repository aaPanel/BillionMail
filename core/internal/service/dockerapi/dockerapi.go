package docker

import (
	v1 "billionmail-core/api/dockerapi/v1"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/gogf/gf/v2/util/gconv"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// DockerAPI struct
type DockerAPI struct {
	client     *client.Client
	containers map[string]bool
}

// NewDockerAPI creates a new DockerAPI instance
func NewDockerAPI() (*DockerAPI, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	return &DockerAPI{
		client:     cli,
		containers: make(map[string]bool),
	}, nil
}

// StartContainer starts a container
func (d *DockerAPI) StartContainer(ctx context.Context, containerID string) error {
	return d.client.ContainerStart(ctx, containerID, container.StartOptions{})
}

// StopContainer stops a container
func (d *DockerAPI) StopContainer(ctx context.Context, containerID string) error {
	timeout := gconv.Int(10 * time.Second)
	return d.client.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
}

// RestartContainer restarts a container
func (d *DockerAPI) RestartContainer(ctx context.Context, containerID string) error {
	timeout := gconv.Int(10 * time.Second)
	return d.client.ContainerRestart(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
}

// RestartContainerByName restarts a container by its name
func (d *DockerAPI) RestartContainerByName(ctx context.Context, name string) error {
	ctnr, err := d.GetContainerByName(ctx, name)
	if err != nil {
		return err
	}

	return d.RestartContainer(ctx, ctnr.ID)
}

// GetContainerStats gets container status
func (d *DockerAPI) GetContainerStats(ctx context.Context, containerID string) (*v1.ContainerStats, error) {
	stats, err := d.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var statsJSON v1.DockerStatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return nil, err
	}

	// Calculate CPU usage
	cpuPercent := calculateCPUPercentUnix(statsJSON)

	// Calculate memory usage
	memoryPercent := float64(statsJSON.MemoryStats.Usage) / float64(statsJSON.MemoryStats.Limit) * 100.0
	memoryMB := float64(statsJSON.MemoryStats.Usage) / 1024 / 1024

	// Calculate network IO
	var netIn, netOut float64
	for _, v := range statsJSON.Networks {
		netIn += float64(v.RxBytes)
		netOut += float64(v.TxBytes)
	}

	// Calculate disk IO
	var blockIn, blockOut float64
	for _, bioEntry := range statsJSON.BlkioStats.IoServiceBytesRecursive {
		if bioEntry.Op == "Read" {
			blockIn += float64(bioEntry.Value)
		}
		if bioEntry.Op == "Write" {
			blockOut += float64(bioEntry.Value)
		}
	}

	return &v1.ContainerStats{
		CPU:      cpuPercent,
		Memory:   memoryPercent,
		MemoryMB: memoryMB,
		NetIO: struct {
			In  float64 `json:"in"`
			Out float64 `json:"out"`
		}{
			In:  netIn,
			Out: netOut,
		},
		BlockIO: struct {
			In  float64 `json:"in"`
			Out float64 `json:"out"`
		}{
			In:  blockIn,
			Out: blockOut,
		},
	}, nil
}

// Calculate CPU usage percentage
func calculateCPUPercentUnix(stats v1.DockerStatsJSON) float64 {
	cpuPercent := 0.0
	// Calculate CPU usage
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

// ExecCommand executes a command in a container
func (d *DockerAPI) ExecCommand(ctx context.Context, containerID string, cmd []string, user string) (*v1.ExecResult, error) {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		User:         user,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}

	execID, err := d.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// Use stdcopy to copy output
	var outBuf, errBuf strings.Builder
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
	if err != nil {
		return nil, err
	}

	// Get exit code
	inspect, err := d.client.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return nil, err
	}

	output := outBuf.String()
	if errBuf.String() != "" {
		if output != "" {
			output += "\n"
		}
		output += errBuf.String()
	}

	return &v1.ExecResult{
		ExitCode: inspect.ExitCode,
		Output:   output,
	}, nil
}

// ExecCommandByName executes a command in a container by its name
func (d *DockerAPI) ExecCommandByName(ctx context.Context, name string, cmd []string, user string) (*v1.ExecResult, error) {
	ctnr, err := d.GetContainerByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return d.ExecCommand(ctx, ctnr.ID, cmd, user)
}

// ListContainers lists all containers
func (d *DockerAPI) ListContainers(ctx context.Context) ([]container.Summary, error) {
	return d.client.ContainerList(ctx, container.ListOptions{All: true})
}

// GetContainerByName gets a container by its name
func (d *DockerAPI) GetContainerByName(ctx context.Context, name string) (*container.Summary, error) {
	args := filters.NewArgs()
	args.Add("name", name)

	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: args,
	})

	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("container with name %s not found", name)
	}

	return &containers[0], nil
}

// GetContainer gets container details
func (d *DockerAPI) GetContainer(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return d.client.ContainerInspect(ctx, containerID)
}

// Close closes the Docker client connection
func (d *DockerAPI) Close() error {
	return d.client.Close()
}

// HostCommandResult stores the result of a command executed on the host
type HostCommandResult struct {
	ExitCode int    `json:"exit_code"` // Command exit status code
	Output   string `json:"output"`    // Command output
	Error    string `json:"error"`     // Execution error
	Duration int64  `json:"duration"`  // Execution time (milliseconds)
}

// ExecHostCommand executes a command on the host through Docker API
// Principle: Creates a temporary privileged container mounting the host's root directory,
// executing commands inside the container that actually operate on the host's file system
func (d *DockerAPI) ExecHostCommand(ctx context.Context, command []string) (*HostCommandResult, error) {
	startTime := time.Now()
	result := &HostCommandResult{
		ExitCode: -1,
	}

	// Check if docker.sock is mounted
	if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
		return nil, fmt.Errorf("docker.sock not mounted, cannot access Docker API")
	}

	// Specify the base image to use
	baseImage := "alpine:latest"

	// Check if the image exists, pull it if it doesn't
	_, err := d.client.ImageInspect(ctx, baseImage)
	if err != nil {
		// Image doesn't exist, try to pull it
		reader, err := d.client.ImagePull(ctx, baseImage, image.PullOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to pull image: %w", err)
		}
		defer reader.Close()

		// Wait for the image pull to complete
		_, _ = io.Copy(io.Discard, reader)
	}

	// Create temporary container configuration
	config := &container.Config{
		Image:      baseImage,
		Cmd:        command,
		WorkingDir: "/host_root", // Set working directory to mount point
		Tty:        false,
		Entrypoint: []string{},
	}

	// Use privileged mode and mount the host's root directory
	hostConfig := &container.HostConfig{
		Privileged: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/",          // Host root directory
				Target: "/host_root", // Mount point in the container
			},
		},
		NetworkMode: "host", // No network needed
	}

	// Create temporary container
	resp, err := d.client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		nil,
		nil,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary container: %w", err)
	}
	containerID := resp.ID
	defer d.cleanupContainer(ctx, containerID) // Ensure container is cleaned up

	// Start container
	if err := d.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start temporary container: %w", err)
	}

	// Wait for container execution to complete
	statusCh, errCh := d.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("failed waiting for container execution: %w", err)
		}
	case status := <-statusCh:
		result.ExitCode = int(status.StatusCode)
	}

	// Get container logs
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	logs, err := d.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get command output: %v", err)
	} else {
		defer logs.Close()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, logs)
		if err != nil {
			result.Error = fmt.Sprintf("failed to read command output: %v", err)
		} else {
			result.Output = strings.TrimSpace(buf.String())
		}
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// ExecHostShellCommand executes a shell command (supports pipes, redirects, etc.)
func (d *DockerAPI) ExecHostShellCommand(ctx context.Context, shellCommand string) (*HostCommandResult, error) {
	return d.ExecHostCommand(ctx, []string{"/bin/sh", "-c", "chroot /host_root " + shellCommand})
}

// cleanupContainer cleans up the temporary container
func (d *DockerAPI) cleanupContainer(ctx context.Context, containerID string) {
	// Stop container
	timeout := 1 // 1 second timeout
	_ = d.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})

	// Remove container
	_ = d.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}
