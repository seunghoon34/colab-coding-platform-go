package handlers

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"
)

type ExecuteRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type ExecuteResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func testDocker() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}

	_, err = cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping Docker daemon: %v", err)
	}

	return nil
}

func ExecuteCode(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Executing code. Language: %s, Code length: %d", req.Language, len(req.Code))

	output, err := executeInDocker(req.Code, req.Language)
	if err != nil {
		log.Printf("Error executing code: %v", err)
		c.JSON(http.StatusInternalServerError, ExecuteResponse{Error: err.Error()})
		return
	}

	log.Printf("Code executed successfully. Output length: %d", len(output))

	// Broadcast the output to all clients in the room
	roomCode := c.Query("roomCode")
	if room, exists := rooms[roomCode]; exists {
		message := ExecuteResponse{Output: output}
		jsonMessage, _ := json.Marshal(message)
		room.BroadcastMessage(jsonMessage)
	}

	c.JSON(http.StatusOK, ExecuteResponse{Output: output})
}
func executeInDocker(code, language string) (string, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %v", err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %v", err)
	}
	log.Printf("Current working directory: %s", cwd)

	// Construct the path to the dockerfiles directory
	dockerfilesDir := filepath.Join(filepath.Dir(filepath.Dir(cwd)), "dockerfiles")
	log.Printf("Dockerfiles directory: %s", dockerfilesDir)

	// Create a temporary directory to store the script
	tmpDir, err := ioutil.TempDir("", "docker-exec")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write the code to a file
	var scriptName string
	switch language {
	case "javascript":
		scriptName = "script.js"
	case "python":
		scriptName = "script.py"
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
	scriptPath := filepath.Join(tmpDir, scriptName)
	if err := ioutil.WriteFile(scriptPath, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write script file: %v", err)
	}

	// Convert Windows path to Docker-friendly path
	dockerfilePath := filepath.Join("dockerfiles", fmt.Sprintf("Dockerfile.%s", language))
	dockerfilePath = filepath.ToSlash(dockerfilePath)
	log.Printf("Using Dockerfile at: %s", dockerfilePath)

	// Create a tar archive containing both the script and the Dockerfile
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add script to tar
	if err := addFileToTar(tw, scriptPath, scriptName); err != nil {
		return "", fmt.Errorf("failed to add script to tar: %v", err)
	}

	// Add Dockerfile to tar
	dockerfileFullPath := filepath.Join(dockerfilesDir, fmt.Sprintf("Dockerfile.%s", language))
	if err := addFileToTar(tw, dockerfileFullPath, dockerfilePath); err != nil {
		return "", fmt.Errorf("failed to add Dockerfile to tar: %v", err)
	}

	if err := tw.Close(); err != nil {
		return "", fmt.Errorf("failed to close tar writer: %v", err)
	}

	// Build the Docker image
	buildResponse, err := cli.ImageBuild(ctx, &buf, types.ImageBuildOptions{
		Dockerfile: dockerfilePath,
		Tags:       []string{fmt.Sprintf("code-exec-%s", language)},
		Remove:     true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to build Docker image: %v", err)
	}
	defer buildResponse.Body.Close()

	// Read the response
	response, err := ioutil.ReadAll(buildResponse.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read build response: %v", err)
	}
	log.Printf("Build response: %s", string(response))

	// Create a container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: fmt.Sprintf("code-exec-%s", language),
	}, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	// Wait for the container to finish
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("error waiting for container: %v", err)
		}
	case <-statusCh:
	}

	// Get the logs from the container
	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	stdcopy.StdCopy(&outBuf, &errBuf, out)

	return outBuf.String() + errBuf.String(), nil
}

// Helper function to add a file to a tar archive
func addFileToTar(tw *tar.Writer, filePath, name string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats for %s: %v", filePath, err)
	}

	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: stat.Size(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("failed to write header for %s: %v", filePath, err)
	}

	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("failed to copy %s contents to tar: %v", filePath, err)
	}

	return nil
}
