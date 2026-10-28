package client

import (
	"context"
	"net/http"
)

type Process struct {
	ID                    string           `json:"id"`
	AppID                 *string          `json:"app_id"`
	Key                   string           `json:"key"`
	Type                  *string          `json:"type"`
	DisplayName           *string          `json:"display_name"`
	Entrypoint            string           `json:"entrypoint"`
	Port                  *int64           `json:"port"`
	IsIngressEnabled      bool             `json:"is_ingress_enabled"`
	IngressProtocol       *string          `json:"ingress_protocol"`
	ScalingStrategy       *ScalingStrategy `json:"scaling_strategy"`
	ResourceTypeID        *string          `json:"resource_type_id"`
	ResourceTypeName      *string          `json:"resource_type_name"`
	CPULimit              *int64           `json:"cpu_limit"`
	MemoryLimit           *int64           `json:"memory_limit"`
	Schedule              *string          `json:"schedule"`
	TimeZone              *string          `json:"time_zone"`
	ConcurrencyPolicy     *string          `json:"concurrency_policy"`
	ActiveDeadlineSeconds *int64           `json:"active_deadline_seconds"`
	JobStartPolicy        *string          `json:"job_start_policy"`
	LivenessProbe         *Probe           `json:"liveness_probe"`
	ReadinessProbe        *Probe           `json:"readiness_probe"`
	InternalHostname      *string          `json:"internal_hostname"`
	CreatedAt             string           `json:"created_at"`
	UpdatedAt             string           `json:"updated_at"`
}

type CreateProcessRequest struct {
	DisplayName           string           `json:"display_name"`
	Type                  string           `json:"type"`
	Entrypoint            *string          `json:"entrypoint,omitempty"`
	Port                  *int64           `json:"port,omitempty"`
	ResourceTypeID        string           `json:"resource_type_id"`
	ScalingStrategy       *ScalingStrategy `json:"scaling_strategy"`
	Schedule              *string          `json:"schedule,omitempty"`
	TimeZone              *string          `json:"time_zone,omitempty"`
	ConcurrencyPolicy     *string          `json:"concurrency_policy,omitempty"`
	ActiveDeadlineSeconds *int64           `json:"active_deadline_seconds,omitempty"`
	JobStartPolicy        *string          `json:"job_start_policy,omitempty"`
	LivenessProbe         *Probe           `json:"liveness_probe,omitempty"`
	ReadinessProbe        *Probe           `json:"readiness_probe,omitempty"`
}

type UpdateProcessRequest struct {
	DisplayName           *string          `json:"display_name,omitempty"`
	Entrypoint            *string          `json:"entrypoint,omitempty"`
	Port                  *int64           `json:"port,omitempty"`
	ResourceTypeID        *string          `json:"resource_type_id,omitempty"`
	ScalingStrategy       *ScalingStrategy `json:"scaling_strategy,omitempty"`
	Schedule              *string          `json:"schedule,omitempty"`
	TimeZone              *string          `json:"time_zone,omitempty"`
	ConcurrencyPolicy     *string          `json:"concurrency_policy,omitempty"`
	ActiveDeadlineSeconds *int64           `json:"active_deadline_seconds,omitempty"`
	JobStartPolicy        *string          `json:"job_start_policy,omitempty"`
	IsIngressEnabled      *bool            `json:"is_ingress_enabled,omitempty"`
	IngressProtocol       *string          `json:"ingress_protocol,omitempty"`
	LivenessProbe         *Probe           `json:"liveness_probe,omitempty"`
	ReadinessProbe        *Probe           `json:"readiness_probe,omitempty"`
}

// ScalingStrategy is either {type: manual, config: {instanceCount}} or
// {type: horizontal, config: {minInstanceCount, maxInstanceCount, ...}}.
type ScalingStrategy struct {
	Type   string        `json:"type"`
	Config ScalingConfig `json:"config"`
}

type ScalingConfig struct {
	InstanceCount            *int64 `json:"instanceCount,omitempty"`
	MinInstanceCount         *int64 `json:"minInstanceCount,omitempty"`
	MaxInstanceCount         *int64 `json:"maxInstanceCount,omitempty"`
	TargetCPUPercent         *int64 `json:"targetCpuPercent,omitempty"`
	TargetMemoryPercent      *int64 `json:"targetMemoryPercent,omitempty"`
	ScaleUpIntervalSeconds   *int64 `json:"scaleUpIntervalSeconds,omitempty"`
	ScaleUpIncrement         *int64 `json:"scaleUpIncrement,omitempty"`
	ScaleDownIntervalSeconds *int64 `json:"scaleDownIntervalSeconds,omitempty"`
	ScaleDownIncrement       *int64 `json:"scaleDownIncrement,omitempty"`
}

// Probe is a Kubernetes probe. The API only accepts httpGet probes on write;
// exec and tcpSocket are read-only.
type Probe struct {
	Exec                *ProbeExec      `json:"exec,omitempty"`
	HTTPGet             *ProbeHTTPGet   `json:"httpGet,omitempty"`
	TCPSocket           *ProbeTCPSocket `json:"tcpSocket,omitempty"`
	InitialDelaySeconds *int64          `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       *int64          `json:"periodSeconds,omitempty"`
	TimeoutSeconds      *int64          `json:"timeoutSeconds,omitempty"`
	SuccessThreshold    *int64          `json:"successThreshold,omitempty"`
	FailureThreshold    *int64          `json:"failureThreshold,omitempty"`
}

type ProbeExec struct {
	Command []string `json:"command"`
}

type ProbeHTTPGet struct {
	Path        *string      `json:"path,omitempty"`
	Port        *int64       `json:"port,omitempty"`
	Host        *string      `json:"host,omitempty"`
	Scheme      *string      `json:"scheme,omitempty"`
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`
}

type HTTPHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ProbeTCPSocket struct {
	Host *string `json:"host,omitempty"`
	Port int64   `json:"port"`
}

func (c *SevallaClient) GetProcess(ctx context.Context, appID, processID string) (*Process, error) {
	var out Process
	if err := c.do(ctx, http.MethodGet, path("/applications/%s/processes/%s", appID, processID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) CreateProcess(ctx context.Context, appID string, input *CreateProcessRequest) (*Process, error) {
	var out Process
	if err := c.do(ctx, http.MethodPost, path("/applications/%s/processes", appID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) UpdateProcess(ctx context.Context, appID, processID string, input *UpdateProcessRequest) (*Process, error) {
	var out Process
	if err := c.do(ctx, http.MethodPatch, path("/applications/%s/processes/%s", appID, processID), input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SevallaClient) DeleteProcess(ctx context.Context, appID, processID string) error {
	return c.do(ctx, http.MethodDelete, path("/applications/%s/processes/%s", appID, processID), nil, nil)
}
