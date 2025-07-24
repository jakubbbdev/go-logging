package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// ContainerInfo holds information about the current container environment
type ContainerInfo struct {
	ID           string            `json:"container_id"`
	Name         string            `json:"container_name"`
	Image        string            `json:"image"`
	ImageTag     string            `json:"image_tag"`
	Hostname     string            `json:"hostname"`
	PodName      string            `json:"pod_name,omitempty"`      // Kubernetes
	PodNamespace string            `json:"pod_namespace,omitempty"` // Kubernetes
	ServiceName  string            `json:"service_name,omitempty"`  // Docker Compose
	NetworkMode  string            `json:"network_mode,omitempty"`
	RestartCount int               `json:"restart_count,omitempty"`
	CreatedAt    time.Time         `json:"created_at,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Environment  string            `json:"environment,omitempty"` // prod, staging, dev
}

// DockerFormatter formats logs for Docker/Kubernetes consumption
type DockerFormatter struct {
	includeContainer bool
	containerInfo    *ContainerInfo
	baseFormatter    Formatter
}

// NewDockerFormatter creates a Docker-optimized formatter
func NewDockerFormatter(baseFormatter Formatter) *DockerFormatter {
	return &DockerFormatter{
		includeContainer: true,
		containerInfo:    DetectContainerEnvironment(),
		baseFormatter:    baseFormatter,
	}
}

// Format formats log entries with container information
func (df *DockerFormatter) Format(entry *Entry) ([]byte, error) {
	if df.includeContainer && df.containerInfo != nil {
		// Add container info to fields
		if entry.Fields == nil {
			entry.Fields = make(Fields)
		}

		entry.Fields["container_id"] = df.containerInfo.ID
		entry.Fields["container_name"] = df.containerInfo.Name
		entry.Fields["image"] = df.containerInfo.Image
		entry.Fields["hostname"] = df.containerInfo.Hostname

		if df.containerInfo.PodName != "" {
			entry.Fields["pod_name"] = df.containerInfo.PodName
			entry.Fields["pod_namespace"] = df.containerInfo.PodNamespace
		}

		if df.containerInfo.ServiceName != "" {
			entry.Fields["service_name"] = df.containerInfo.ServiceName
		}

		if df.containerInfo.Environment != "" {
			entry.Fields["environment"] = df.containerInfo.Environment
		}
	}

	return df.baseFormatter.Format(entry)
}

// DetectContainerEnvironment detects if running in a container and extracts info
func DetectContainerEnvironment() *ContainerInfo {
	info := &ContainerInfo{
		Labels: make(map[string]string),
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// Check if running in Docker
	if isDocker() {
		populateDockerInfo(info)
	}

	// Check if running in Kubernetes
	if isKubernetes() {
		populateKubernetesInfo(info)
	}

	// Get environment from ENV vars
	info.Environment = getEnv("ENVIRONMENT", getEnv("ENV", "unknown"))

	return info
}

// isDocker checks if running inside a Docker container
func isDocker() bool {
	// Check for Docker-specific files
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for docker
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		return strings.Contains(string(data), "docker") ||
			strings.Contains(string(data), "containerd")
	}

	return false
}

// isKubernetes checks if running inside a Kubernetes pod
func isKubernetes() bool {
	// Check for Kubernetes service account
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		return true
	}

	// Check for Kubernetes environment variables
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}

// populateDockerInfo populates Docker-specific information
func populateDockerInfo(info *ContainerInfo) {
	// Get container ID from cgroup or hostname
	if data, err := os.ReadFile("/proc/self/cgroup"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.Contains(line, "docker") {
				parts := strings.Split(line, "/")
				if len(parts) > 0 {
					id := parts[len(parts)-1]
					if len(id) >= 12 {
						info.ID = id[:12] // Short Docker ID
					}
				}
				break
			}
		}
	}

	// Fallback to hostname for container ID
	if info.ID == "" {
		info.ID = info.Hostname
	}

	// Get container name from Docker labels or hostname
	info.Name = getEnv("CONTAINER_NAME", info.Hostname)

	// Get image information
	info.Image = getEnv("IMAGE_NAME", "unknown")
	info.ImageTag = getEnv("IMAGE_TAG", "latest")

	// Get service name (Docker Compose)
	info.ServiceName = getEnv("COMPOSE_SERVICE", "")

	// Parse Docker labels from environment
	parseDockerLabels(info)
}

// populateKubernetesInfo populates Kubernetes-specific information
func populateKubernetesInfo(info *ContainerInfo) {
	// Kubernetes environment variables
	info.PodName = os.Getenv("POD_NAME")
	info.PodNamespace = os.Getenv("POD_NAMESPACE")

	// Fallback to hostname for pod name
	if info.PodName == "" {
		info.PodName = info.Hostname
	}

	// Default namespace
	if info.PodNamespace == "" {
		info.PodNamespace = "default"
	}

	// Get service account name
	if saName := os.Getenv("SERVICE_ACCOUNT_NAME"); saName != "" {
		info.Labels["service_account"] = saName
	}

	// Get node name
	if nodeName := os.Getenv("NODE_NAME"); nodeName != "" {
		info.Labels["node_name"] = nodeName
	}

	// Read service account info
	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		info.PodNamespace = strings.TrimSpace(string(data))
	}
}

// parseDockerLabels parses Docker labels from environment variables
func parseDockerLabels(info *ContainerInfo) {
	// Common Docker label environment variables
	labelEnvs := map[string]string{
		"DOCKER_LABEL_VERSION":    "version",
		"DOCKER_LABEL_MAINTAINER": "maintainer",
		"DOCKER_LABEL_BUILD_DATE": "build_date",
		"DOCKER_LABEL_VCS_REF":    "vcs_ref",
	}

	for envVar, labelKey := range labelEnvs {
		if value := os.Getenv(envVar); value != "" {
			info.Labels[labelKey] = value
		}
	}
}

// ContainerHandler wraps another handler and adds container-specific features
type ContainerHandler struct {
	handler       Handler
	containerInfo *ContainerInfo
	logPath       string
	enableStdout  bool
}

// NewContainerHandler creates a container-optimized handler
func NewContainerHandler(handler Handler) *ContainerHandler {
	return &ContainerHandler{
		handler:       handler,
		containerInfo: DetectContainerEnvironment(),
		logPath:       getEnv("LOG_PATH", "/var/log/app"),
		enableStdout:  getEnvBool("ENABLE_STDOUT_LOGS", true),
	}
}

// Handle implements the Handler interface with container optimizations
func (ch *ContainerHandler) Handle(entry *Entry) error {
	// Add container fields if not already present
	if ch.containerInfo != nil {
		if entry.Fields == nil {
			entry.Fields = make(Fields)
		}

		// Only add if not already set
		if _, exists := entry.Fields["container_id"]; !exists {
			entry.Fields["container_id"] = ch.containerInfo.ID
		}
		if _, exists := entry.Fields["pod_name"]; !exists && ch.containerInfo.PodName != "" {
			entry.Fields["pod_name"] = ch.containerInfo.PodName
		}
	}

	return ch.handler.Handle(entry)
}

// CloudNativeLogger provides cloud-native logging features
type CloudNativeLogger struct {
	baseLogger    Logger
	containerInfo *ContainerInfo
	structured    bool
	jsonOutput    bool
}

// NewCloudNativeLogger creates a cloud-native optimized logger
func NewCloudNativeLogger(baseLogger Logger) *CloudNativeLogger {
	return &CloudNativeLogger{
		baseLogger:    baseLogger,
		containerInfo: DetectContainerEnvironment(),
		structured:    true,
		jsonOutput:    getEnvBool("JSON_LOGS", true),
	}
}

// Log implements cloud-native logging with structured output
func (cnl *CloudNativeLogger) LogWithFields(level Level, msg string, fields Fields) {
	// Merge container info with provided fields
	logFields := make(Fields)

	// Add container context
	if cnl.containerInfo != nil {
		logFields["container_id"] = cnl.containerInfo.ID
		logFields["image"] = cnl.containerInfo.Image
		logFields["hostname"] = cnl.containerInfo.Hostname

		if cnl.containerInfo.PodName != "" {
			logFields["pod_name"] = cnl.containerInfo.PodName
			logFields["pod_namespace"] = cnl.containerInfo.PodNamespace
		}
	}

	// Add runtime info
	logFields["go_version"] = runtime.Version()
	logFields["arch"] = runtime.GOARCH
	logFields["os"] = runtime.GOOS

	// Merge user fields
	for k, v := range fields {
		logFields[k] = v
	}

	// Log with enriched fields
	cnl.baseLogger.WithFields(logFields).Log(level, msg)
}

// Log implements the Logger interface
func (cnl *CloudNativeLogger) Log(level Level, args ...interface{}) {
	cnl.LogWithFields(level, fmt.Sprint(args...), nil)
}

// Implement Logger interface
func (cnl *CloudNativeLogger) Debug(args ...interface{}) {
	cnl.LogWithFields(DebugLevel, fmt.Sprint(args...), nil)
}

func (cnl *CloudNativeLogger) Info(args ...interface{}) {
	cnl.LogWithFields(InfoLevel, fmt.Sprint(args...), nil)
}

func (cnl *CloudNativeLogger) Warn(args ...interface{}) {
	cnl.LogWithFields(WarnLevel, fmt.Sprint(args...), nil)
}

func (cnl *CloudNativeLogger) Error(args ...interface{}) {
	cnl.LogWithFields(ErrorLevel, fmt.Sprint(args...), nil)
}

func (cnl *CloudNativeLogger) Fatal(args ...interface{}) {
	cnl.LogWithFields(FatalLevel, fmt.Sprint(args...), nil)
	os.Exit(1)
}

func (cnl *CloudNativeLogger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	cnl.LogWithFields(PanicLevel, msg, nil)
	panic(msg)
}

// Simplified implementations for interface compliance
func (cnl *CloudNativeLogger) Debugf(format string, args ...interface{}) {
	cnl.Debug(fmt.Sprintf(format, args...))
}
func (cnl *CloudNativeLogger) Infof(format string, args ...interface{}) {
	cnl.Info(fmt.Sprintf(format, args...))
}
func (cnl *CloudNativeLogger) Warnf(format string, args ...interface{}) {
	cnl.Warn(fmt.Sprintf(format, args...))
}
func (cnl *CloudNativeLogger) Errorf(format string, args ...interface{}) {
	cnl.Error(fmt.Sprintf(format, args...))
}
func (cnl *CloudNativeLogger) Fatalf(format string, args ...interface{}) {
	cnl.Fatal(fmt.Sprintf(format, args...))
}
func (cnl *CloudNativeLogger) Panicf(format string, args ...interface{}) {
	cnl.Panic(fmt.Sprintf(format, args...))
}
func (cnl *CloudNativeLogger) DebugFast(msg string)            { cnl.LogWithFields(DebugLevel, msg, nil) }
func (cnl *CloudNativeLogger) InfoFast(msg string)             { cnl.LogWithFields(InfoLevel, msg, nil) }
func (cnl *CloudNativeLogger) WarnFast(msg string)             { cnl.LogWithFields(WarnLevel, msg, nil) }
func (cnl *CloudNativeLogger) ErrorFast(msg string)            { cnl.LogWithFields(ErrorLevel, msg, nil) }
func (cnl *CloudNativeLogger) LogFast(level Level, msg string) { cnl.LogWithFields(level, msg, nil) }
func (cnl *CloudNativeLogger) Logf(level Level, format string, args ...interface{}) {
	cnl.LogWithFields(level, fmt.Sprintf(format, args...), nil)
}
func (cnl *CloudNativeLogger) WithFields(fields Fields) Logger {
	return &fieldLogger{cnl, fields}
}
func (cnl *CloudNativeLogger) WithContext(ctx Context) Logger   { return cnl }
func (cnl *CloudNativeLogger) WithTrace(ctx Context) Logger     { return cnl }
func (cnl *CloudNativeLogger) SetLevel(level Level)             {}
func (cnl *CloudNativeLogger) SetHandler(handler Handler)       {}
func (cnl *CloudNativeLogger) SetFormatter(formatter Formatter) {}
func (cnl *CloudNativeLogger) AddHook(hook Hook)                {}

// fieldLogger wraps CloudNativeLogger with additional fields
type fieldLogger struct {
	logger *CloudNativeLogger
	fields Fields
}

func (fl *fieldLogger) Log(level Level, args ...interface{}) {
	fl.logger.LogWithFields(level, fmt.Sprint(args...), fl.fields)
}

// Implement remaining Logger interface methods for fieldLogger
func (fl *fieldLogger) Debug(args ...interface{}) { fl.Log(DebugLevel, args...) }
func (fl *fieldLogger) Info(args ...interface{})  { fl.Log(InfoLevel, args...) }
func (fl *fieldLogger) Warn(args ...interface{})  { fl.Log(WarnLevel, args...) }
func (fl *fieldLogger) Error(args ...interface{}) { fl.Log(ErrorLevel, args...) }
func (fl *fieldLogger) Fatal(args ...interface{}) { fl.Log(FatalLevel, args...) }
func (fl *fieldLogger) Panic(args ...interface{}) { fl.Log(PanicLevel, args...) }
func (fl *fieldLogger) Debugf(format string, args ...interface{}) {
	fl.Debug(fmt.Sprintf(format, args...))
}
func (fl *fieldLogger) Infof(format string, args ...interface{}) {
	fl.Info(fmt.Sprintf(format, args...))
}
func (fl *fieldLogger) Warnf(format string, args ...interface{}) {
	fl.Warn(fmt.Sprintf(format, args...))
}
func (fl *fieldLogger) Errorf(format string, args ...interface{}) {
	fl.Error(fmt.Sprintf(format, args...))
}
func (fl *fieldLogger) Fatalf(format string, args ...interface{}) {
	fl.Fatal(fmt.Sprintf(format, args...))
}
func (fl *fieldLogger) Panicf(format string, args ...interface{}) {
	fl.Panic(fmt.Sprintf(format, args...))
}
func (fl *fieldLogger) DebugFast(msg string)            { fl.Log(DebugLevel, msg) }
func (fl *fieldLogger) InfoFast(msg string)             { fl.Log(InfoLevel, msg) }
func (fl *fieldLogger) WarnFast(msg string)             { fl.Log(WarnLevel, msg) }
func (fl *fieldLogger) ErrorFast(msg string)            { fl.Log(ErrorLevel, msg) }
func (fl *fieldLogger) LogFast(level Level, msg string) { fl.Log(level, msg) }
func (fl *fieldLogger) Logf(level Level, format string, args ...interface{}) {
	fl.Log(level, fmt.Sprintf(format, args...))
}
func (fl *fieldLogger) WithFields(fields Fields) Logger {
	newFields := make(Fields)
	for k, v := range fl.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &fieldLogger{fl.logger, newFields}
}
func (fl *fieldLogger) WithContext(ctx Context) Logger   { return fl }
func (fl *fieldLogger) WithTrace(ctx Context) Logger     { return fl }
func (fl *fieldLogger) SetLevel(level Level)             {}
func (fl *fieldLogger) SetHandler(handler Handler)       {}
func (fl *fieldLogger) SetFormatter(formatter Formatter) {}
func (fl *fieldLogger) AddHook(hook Hook)                {}

// Helper to create container-optimized configuration
func NewContainerConfig() *Config {
	return &Config{
		Level:           getEnv("LOG_LEVEL", "info"),
		Format:          getEnv("LOG_FORMAT", "json"),
		Output:          getEnv("LOG_OUTPUT", "console"),
		IncludeCaller:   getEnvBool("LOG_INCLUDE_CALLER", false),
		IncludeStack:    getEnvBool("LOG_INCLUDE_STACK", false),
		TimestampFormat: getEnv("LOG_TIMESTAMP_FORMAT", time.RFC3339),
		UseColors:       getEnvBool("LOG_USE_COLORS", false), // Disable colors in containers
		DefaultFields: map[string]string{
			"service": getEnv("SERVICE_NAME", "app"),
			"version": getEnv("VERSION", "unknown"),
		},
	}
}

// CreateContainerLogger creates a fully configured container logger
func CreateContainerLogger() Logger {
	config := NewContainerConfig()

	// Create base logger from config
	baseLogger, err := config.ToLogger()
	if err != nil {
		// Fallback to simple logger
		baseLogger = NewLogger(
			WithLevel(InfoLevel),
			WithFormatter(NewJSONFormatter()),
			WithHandler(NewConsoleHandler()),
		)
	}

	// Wrap with Docker formatter
	dockerFormatter := NewDockerFormatter(NewJSONFormatter())
	baseLogger.SetFormatter(dockerFormatter)

	// Wrap with container handler
	containerHandler := NewContainerHandler(NewConsoleHandler())
	baseLogger.SetHandler(containerHandler)

	// Create cloud-native logger
	return NewCloudNativeLogger(baseLogger)
}

// ContainerHealthCheck creates a health check for container environments
func ContainerHealthCheck() HealthCheck {
	return func(ctx Context) HealthCheckResult {
		containerInfo := DetectContainerEnvironment()

		if containerInfo == nil {
			return HealthCheckResult{
				Status:  HealthStatusDegraded,
				Message: "Container info not available",
			}
		}

		// Check if we're in the expected container environment
		if !isDocker() && !isKubernetes() {
			return HealthCheckResult{
				Status:  HealthStatusDegraded,
				Message: "Not running in container environment",
			}
		}

		return HealthCheckResult{
			Status:  HealthStatusHealthy,
			Message: fmt.Sprintf("Container %s healthy", containerInfo.ID),
		}
	}
}

// Helper to export container info as JSON for debugging
func (ci *ContainerInfo) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ci, "", "  ")
}

// String representation of container info
func (ci *ContainerInfo) String() string {
	if ci.PodName != "" {
		return fmt.Sprintf("Pod: %s/%s (Image: %s)", ci.PodNamespace, ci.PodName, ci.Image)
	}
	return fmt.Sprintf("Container: %s (Image: %s)", ci.Name, ci.Image)
}
