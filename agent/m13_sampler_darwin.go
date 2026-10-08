//go:build darwin

package agent

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

// SampleM13Heartbeat uses macOS's built-in top and vm_stat interfaces.
// GPU/VRAM reporting is intentionally nil until a reliable Metal/Intel/AMD
// device metric source is verified; it never reports unknown as idle/zero.
func SampleM13Heartbeat(ctx context.Context,opts M13SamplerOptions)(protocol.M13Heartbeat,error){
	var empty protocol.M13Heartbeat
	sampleCtx,cancel:=context.WithTimeout(ctx,5*time.Second)
	defer cancel()
	top,err:=exec.CommandContext(sampleCtx,"top","-l","2","-n","0","-s","1").Output()
	if err!=nil {return empty,fmt.Errorf("sample macOS CPU: %w",err)}
	cpu,err:=m13DarwinCPUUsage(top)
	if err!=nil {return empty,err}
	vm,err:=exec.CommandContext(sampleCtx,"vm_stat").Output()
	if err!=nil {return empty,fmt.Errorf("sample macOS RAM: %w",err)}
	ram,err:=m13DarwinAvailableRAM(vm)
	if err!=nil {return empty,err}
	latency:=opts.RecentLatencyMS
	h:=protocol.M13Heartbeat{
		Type:"heartbeat",Online:true,
		AvailableRAMBytes:&ram,CPULoadPercent:&cpu,
		ActiveJobState:opts.ActiveJobState,RecentLatencyMS:&latency,
		AvailabilityMode:opts.AvailabilityMode,
		ManualPause:opts.ManualPause,ObservedAt:time.Now().UTC(),
	}
	if err=h.Validate(time.Now().UTC());err!=nil {return empty,err}
	return h,nil
}
