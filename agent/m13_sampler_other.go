//go:build !linux && !darwin

package agent

import (
	"context"
	"errors"
	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// Unsupported hosts fail closed rather than submitting invented metrics.
func SampleM13Heartbeat(_ context.Context,_ M13SamplerOptions)(protocol.M13Heartbeat,error){
	return protocol.M13Heartbeat{},errors.New("M13 host telemetry is not supported on this operating system")
}
