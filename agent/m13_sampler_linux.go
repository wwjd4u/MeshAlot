//go:build linux

package agent

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

const m13CPUSampleWindow = 150 * time.Millisecond

// SampleM13Heartbeat reads passive host counters and one bounded nvidia-smi
// query. It does NOT benchmark the GPU, change drivers, or probe user files.
// Unavailable GPU telemetry is represented by two nil fields, never made-up 0.
func SampleM13Heartbeat(ctx context.Context, opts M13SamplerOptions) (protocol.M13Heartbeat, error) {
	return sampleM13Linux(ctx,opts,os.ReadFile,osCommandRunner{},m13CPUSampleWindow)
}

func sampleM13Linux(ctx context.Context, opts M13SamplerOptions,
	read func(string)([]byte,error), runner commandRunner, delay time.Duration) (protocol.M13Heartbeat,error) {
	var empty protocol.M13Heartbeat
	before,err:=read("/proc/stat")
	if err!=nil {return empty,fmt.Errorf("read CPU counters: %w",err)}
	totalBefore,idleBefore,err:=m13CPUStat(before)
	if err!=nil {return empty,err}
	wait:=time.NewTimer(delay)
	select {
	case <-ctx.Done():
		wait.Stop()
		return empty,ctx.Err()
	case <-wait.C:
	}
	after,err:=read("/proc/stat")
	if err!=nil {return empty,fmt.Errorf("reread CPU counters: %w",err)}
	totalAfter,idleAfter,err:=m13CPUStat(after)
	if err!=nil {return empty,err}
	if totalAfter<=totalBefore || idleAfter<idleBefore {
		return empty,errors.New("invalid CPU counter progression")
	}
	totalDelta:=totalAfter-totalBefore
	idleDelta:=idleAfter-idleBefore
	if idleDelta>totalDelta {return empty,errors.New("invalid idle counter progression")}
	cpu:=100*float64(totalDelta-idleDelta)/float64(totalDelta)
	if math.IsNaN(cpu)||math.IsInf(cpu,0) {return empty,errors.New("invalid CPU load")}
	mem,err:=read("/proc/meminfo")
	if err!=nil {return empty,fmt.Errorf("read memory counters: %w",err)}
	ram,err:=m13MemAvailable(mem)
	if err!=nil {return empty,err}
	latency:=opts.RecentLatencyMS
	h:=protocol.M13Heartbeat{
		Type:"heartbeat",Online:true,
		AvailableRAMBytes:&ram,CPULoadPercent:&cpu,
		ActiveJobState:opts.ActiveJobState,RecentLatencyMS:&latency,
		AvailabilityMode:opts.AvailabilityMode,
		ManualPause:opts.ManualPause,ObservedAt:time.Now().UTC(),
	}
	// Query only basic read-only GPU metrics. A missing/unsupported nvidia-smi
	// is not a zero-utilization GPU: leave both GPU fields absent.
	if runner!=nil {
		queryCtx,cancel:=context.WithTimeout(ctx,2*time.Second)
		output,queryErr:=runner.Run(queryCtx,"nvidia-smi",
			"--query-gpu=utilization.gpu,memory.free",
			"--format=csv,noheader,nounits")
		cancel()
		if queryErr==nil {
			if utilization,freeBytes,ok:=m13NvidiaSnapshot(output);ok {
				h.GPUUtilizationPercent=&utilization
				h.AvailableVRAMBytes=&freeBytes
			}
		}
	}
	if err:=h.Validate(time.Now().UTC());err!=nil {
		return empty,fmt.Errorf("invalid host snapshot: %w",err)
	}
	return h,nil
}

func m13CPUStat(raw []byte)(total uint64,idle uint64,err error){
	lines:=strings.SplitN(string(raw),"\n",2)
	fields:=strings.Fields(lines[0])
	if len(fields)<9 || fields[0]!="cpu" {return 0,0,errors.New("invalid /proc/stat CPU line")}
	var values [8]uint64
	for i:=0;i<8;i++ {
		values[i],err=strconv.ParseUint(fields[i+1],10,64)
		if err!=nil {return 0,0,errors.New("invalid CPU counter")}
		total+=values[i]
	}
	idle=values[3]+values[4] // idle + iowait
	if idle>total {return 0,0,errors.New("invalid idle counters")}
	return total,idle,nil
}

func m13MemAvailable(raw []byte)(uint64,error) {
	scanner:=bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		f:=strings.Fields(scanner.Text())
		if len(f)>=3 && f[0]=="MemAvailable:" && f[2]=="kB" {
			kb,err:=strconv.ParseUint(f[1],10,64)
			if err!=nil || kb>protocol.M13MaxResourceBytes/1024 {
				return 0,errors.New("invalid MemAvailable")
			}
			return kb*1024,nil
		}
	}
	if scanner.Err()!=nil{return 0,scanner.Err()}
	return 0,errors.New("MemAvailable not reported")
}

// For the initial single-GPU POC, use the first reported NVIDIA GPU. Future
// multi-GPU node metrics require an explicit per-device schema.
func m13NvidiaSnapshot(raw []byte)(float64,uint64,bool){
	line:=strings.TrimSpace(strings.SplitN(string(raw),"\n",2)[0])
	parts:=strings.Split(line,",")
	if len(parts)<2 {return 0,0,false}
	percentage,err:=strconv.ParseFloat(strings.TrimSpace(parts[0]),64)
	if err!=nil || math.IsNaN(percentage) || math.IsInf(percentage,0) ||
		percentage<0 || percentage>100 {return 0,0,false}
	mib,err:=strconv.ParseUint(strings.TrimSpace(parts[1]),10,64)
	if err!=nil || mib>protocol.M13MaxResourceBytes/(1<<20) {return 0,0,false}
	return percentage,mib*(1<<20),true
}
