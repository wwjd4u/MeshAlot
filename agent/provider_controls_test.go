package agent

import "testing"

func TestValidateProviderResourceControls(t *testing.T) {
	valid := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8,
		MaxRAMBytes:         16,
		MaxCPUPercent:       50,
		AllowedStartHour:    8,
		AllowedEndHour:      22,
		Mode:                ProviderModeNormal,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}
	if err := ValidateProviderResourceControls(valid); err != nil {
		t.Fatalf("valid controls rejected: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*ProviderResourceControls)
	}{
		{name: "negative GPU percent", mutate: func(c *ProviderResourceControls) { c.MaxGPUPercent = -1 }},
		{name: "GPU percent above 100", mutate: func(c *ProviderResourceControls) { c.MaxGPUPercent = 101 }},
		{name: "negative CPU percent", mutate: func(c *ProviderResourceControls) { c.MaxCPUPercent = -1 }},
		{name: "CPU percent above 100", mutate: func(c *ProviderResourceControls) { c.MaxCPUPercent = 101 }},
		{name: "negative start hour", mutate: func(c *ProviderResourceControls) { c.AllowedStartHour = -1 }},
		{name: "start hour above 23", mutate: func(c *ProviderResourceControls) { c.AllowedStartHour = 24 }},
		{name: "negative end hour", mutate: func(c *ProviderResourceControls) { c.AllowedEndHour = -1 }},
		{name: "end hour above 23", mutate: func(c *ProviderResourceControls) { c.AllowedEndHour = 24 }},
		{name: "unknown provider mode", mutate: func(c *ProviderResourceControls) { c.Mode = ProviderMode("unknown") }},
		{name: "unknown local return behavior", mutate: func(c *ProviderResourceControls) { c.LocalReturnBehavior = LocalReturnBehavior("unknown") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := valid
			test.mutate(&current)
			if err := ValidateProviderResourceControls(current); err == nil {
				t.Fatal("invalid controls accepted")
			}
		})
	}
}

func TestEvaluateProviderResourceRequest(t *testing.T) {
	controls := ProviderResourceControls{
		MaxGPUPercent: 50,
		MaxVRAMBytes:  8,
		MaxRAMBytes:   16,
		MaxCPUPercent: 50,
	}
	atLimit := ProviderWorkloadResources{
		GPUPercent: 50,
		VRAMBytes:  8,
		RAMBytes:   16,
		CPUPercent: 50,
	}
	if result := EvaluateProviderResourceRequest(controls, atLimit); !result.Allowed {
		t.Fatalf("request at provider limits rejected: %v", result.Limitations)
	}

	tooLarge := atLimit
	tooLarge.GPUPercent = 51
	result := EvaluateProviderResourceRequest(controls, tooLarge)
	if result.Allowed || len(result.Limitations) == 0 {
		t.Fatalf("over-limit request was not rejected with explanation: %+v", result)
	}
}

func TestProviderManualPauseBlocksNewWork(t *testing.T) {
	controls := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8,
		MaxRAMBytes:         16,
		MaxCPUPercent:       50,
		AllowedStartHour:    0,
		AllowedEndHour:      0,
		Mode:                ProviderModeNormal,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}
	request := ProviderWorkloadResources{
		GPUPercent: 25,
		VRAMBytes:  4,
		RAMBytes:   8,
		CPUPercent: 25,
	}

	if result := EvaluateProviderNewWork(controls, request); !result.Allowed {
		t.Fatalf("compatible work rejected before manual pause: %v", result.Limitations)
	}
	controls.ManualPause = true
	result := EvaluateProviderNewWork(controls, request)
	if result.Allowed {
		t.Fatal("provider accepted new work while manually paused")
	}
}

func TestProviderAllowedHours(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)
	controls := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8 * gib,
		MaxRAMBytes:         16 * gib,
		MaxCPUPercent:       50,
		AllowedStartHour:    8,
		AllowedEndHour:      22,
		Mode:                ProviderModeNormal,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}
	request := ProviderWorkloadResources{
		GPUPercent: 25,
		VRAMBytes:  4 * gib,
		RAMBytes:   8 * gib,
		CPUPercent: 25,
	}
	tests := []struct {
		name    string
		start   int
		end     int
		hour    int
		allowed bool
	}{
		{name: "inside daytime window", start: 8, end: 22, hour: 12, allowed: true},
		{name: "before daytime window", start: 8, end: 22, hour: 7, allowed: false},
		{name: "end hour is excluded", start: 8, end: 22, hour: 22, allowed: false},
		{name: "inside overnight window before midnight", start: 22, end: 6, hour: 23, allowed: true},
		{name: "inside overnight window after midnight", start: 22, end: 6, hour: 3, allowed: true},
		{name: "outside overnight window", start: 22, end: 6, hour: 12, allowed: false},
		{name: "equal hours mean all day", start: 0, end: 0, hour: 15, allowed: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := controls
			current.AllowedStartHour = test.start
			current.AllowedEndHour = test.end
			result := EvaluateProviderNewWorkAtHour(current, request, test.hour)
			if result.Allowed != test.allowed {
				t.Fatalf("Allowed=%v, want %v; limitations=%v", result.Allowed, test.allowed, result.Limitations)
			}
		})
	}
}

func TestMaximumEarningsModeRespectsResourceLimits(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)
	controls := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8 * gib,
		MaxRAMBytes:         16 * gib,
		MaxCPUPercent:       50,
		AllowedStartHour:    0,
		AllowedEndHour:      0,
		ManualPause:         false,
		Mode:                ProviderModeMaximumEarnings,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}
	compatible := ProviderWorkloadResources{GPUPercent: 50, VRAMBytes: 8 * gib, RAMBytes: 16 * gib, CPUPercent: 50}
	result := EvaluateProviderNewWorkAtHour(controls, compatible, 12)
	if !result.Allowed {
		t.Fatalf("maximum-earnings mode rejected compatible work: %v", result.Limitations)
	}
	overLimit := ProviderWorkloadResources{GPUPercent: 51, VRAMBytes: 8 * gib, RAMBytes: 16 * gib, CPUPercent: 50}
	result = EvaluateProviderNewWorkAtHour(controls, overLimit, 12)
	if result.Allowed {
		t.Fatal("maximum-earnings mode bypassed provider GPU limit")
	}
	controls.ManualPause = true
	result = EvaluateProviderNewWorkAtHour(controls, compatible, 12)
	if result.Allowed {
		t.Fatal("maximum-earnings mode bypassed manual pause")
	}
}

func TestProviderLocalUserReturnBehavior(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)
	controls := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8 * gib,
		MaxRAMBytes:         16 * gib,
		MaxCPUPercent:       50,
		AllowedStartHour:    0,
		AllowedEndHour:      0,
		ManualPause:         false,
		Mode:                ProviderModeAway,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}
	request := ProviderWorkloadResources{GPUPercent: 25, VRAMBytes: 4 * gib, RAMBytes: 8 * gib, CPUPercent: 25}
	result := EvaluateProviderNewWorkWithActivity(controls, request, 12, false)
	if !result.Allowed {
		t.Fatalf("away provider rejected compatible work while local user inactive: %v", result.Limitations)
	}
	result = EvaluateProviderNewWorkWithActivity(controls, request, 12, true)
	if result.Allowed {
		t.Fatal("provider accepted new work after local user returned")
	}
	controls.LocalReturnBehavior = LocalReturnContinue
	result = EvaluateProviderNewWorkWithActivity(controls, request, 12, true)
	if !result.Allowed {
		t.Fatalf("continue policy rejected compatible work after local user returned: %v", result.Limitations)
	}
}

func TestEvaluateActiveProviderWork(t *testing.T) {
	tests := []struct {
		name                         string
		safetyTerminationRequired    bool
		terminationExplicitlyAllowed bool
		want                         ActiveProviderWorkAction
	}{
		{name: "normal sharing change continues running work", want: ActiveProviderWorkContinue},
		{name: "safety requirement permits termination", safetyTerminationRequired: true, want: ActiveProviderWorkTerminate},
		{name: "explicit authorization permits termination", terminationExplicitlyAllowed: true, want: ActiveProviderWorkTerminate},
		{name: "safety and explicit authorization permit termination", safetyTerminationRequired: true, terminationExplicitlyAllowed: true, want: ActiveProviderWorkTerminate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := EvaluateActiveProviderWork(test.safetyTerminationRequired, test.terminationExplicitlyAllowed)
			if got != test.want {
				t.Fatalf("EvaluateActiveProviderWork()=%q, want %q", got, test.want)
			}
		})
	}
}


func TestProviderFiftyPercentSharing(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)

	controls := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8 * gib,
		MaxRAMBytes:         16 * gib,
		MaxCPUPercent:       50,
		AllowedStartHour:    0,
		AllowedEndHour:      0,
		ManualPause:         false,
		Mode:                ProviderModeNormal,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}

	atLimit := ProviderWorkloadResources{
		GPUPercent: 50,
		VRAMBytes:  8 * gib,
		RAMBytes:   16 * gib,
		CPUPercent: 50,
	}

	result := EvaluateProviderNewWorkWithActivity(
		controls,
		atLimit,
		12,
		false,
	)
	if !result.Allowed {
		t.Fatalf(
			"workload exactly at 50-percent sharing limit was rejected: %v",
			result.Limitations,
		)
	}

	tests := []struct {
		name    string
		request ProviderWorkloadResources
	}{
		{
			name: "GPU exceeds shared amount",
			request: ProviderWorkloadResources{
				GPUPercent: 51,
				VRAMBytes:  8 * gib,
				RAMBytes:   16 * gib,
				CPUPercent: 50,
			},
		},
		{
			name: "VRAM exceeds shared amount",
			request: ProviderWorkloadResources{
				GPUPercent: 50,
				VRAMBytes:  8*gib + 1,
				RAMBytes:   16 * gib,
				CPUPercent: 50,
			},
		},
		{
			name: "RAM exceeds shared amount",
			request: ProviderWorkloadResources{
				GPUPercent: 50,
				VRAMBytes:  8 * gib,
				RAMBytes:   16*gib + 1,
				CPUPercent: 50,
			},
		},
		{
			name: "CPU exceeds shared amount",
			request: ProviderWorkloadResources{
				GPUPercent: 50,
				VRAMBytes:  8 * gib,
				RAMBytes:   16 * gib,
				CPUPercent: 51,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := EvaluateProviderNewWorkWithActivity(
				controls,
				test.request,
				12,
				false,
			)
			if result.Allowed {
				t.Fatal(
					"agent accepted workload exceeding owner's 50-percent sharing limit",
				)
			}
			if len(result.Limitations) == 0 {
				t.Fatal(
					"rejected workload did not explain the exceeded provider limit",
				)
			}
		})
	}
}

func TestProviderIncreaseToMaximumEarningsMode(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)

	controls := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8 * gib,
		MaxRAMBytes:         16 * gib,
		MaxCPUPercent:       50,
		AllowedStartHour:    0,
		AllowedEndHour:      0,
		ManualPause:         false,
		Mode:                ProviderModeNormal,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}

	atLimit := ProviderWorkloadResources{
		GPUPercent: 50,
		VRAMBytes:  8 * gib,
		RAMBytes:   16 * gib,
		CPUPercent: 50,
	}

	if result := EvaluateProviderNewWorkWithActivity(
		controls,
		atLimit,
		12,
		false,
	); !result.Allowed {
		t.Fatalf(
			"normal mode rejected workload at configured limits: %v",
			result.Limitations,
		)
	}

	controls.Mode = ProviderModeMaximumEarnings

	if result := EvaluateProviderNewWorkWithActivity(
		controls,
		atLimit,
		12,
		false,
	); !result.Allowed {
		t.Fatalf(
			"maximum-earnings mode rejected workload at configured limits: %v",
			result.Limitations,
		)
	}

	overLimit := atLimit
	overLimit.GPUPercent = 51

	if result := EvaluateProviderNewWorkWithActivity(
		controls,
		overLimit,
		12,
		false,
	); result.Allowed {
		t.Fatal(
			"maximum-earnings mode accepted work above the owner's configured GPU limit",
		)
	}
}

func TestMaximumEarningsStopsNewWorkWhenLocalUserReturns(t *testing.T) {
	const gib = uint64(1024 * 1024 * 1024)

	controls := ProviderResourceControls{
		MaxGPUPercent:       50,
		MaxVRAMBytes:        8 * gib,
		MaxRAMBytes:         16 * gib,
		MaxCPUPercent:       50,
		AllowedStartHour:    0,
		AllowedEndHour:      0,
		ManualPause:         false,
		Mode:                ProviderModeMaximumEarnings,
		LocalReturnBehavior: LocalReturnStopNewWork,
	}

	request := ProviderWorkloadResources{
		GPUPercent: 25,
		VRAMBytes:  4 * gib,
		RAMBytes:   8 * gib,
		CPUPercent: 25,
	}

	result := EvaluateProviderNewWorkWithActivity(
		controls,
		request,
		12,
		false,
	)
	if !result.Allowed {
		t.Fatalf(
			"maximum-earnings mode rejected compatible work while local user inactive: %v",
			result.Limitations,
		)
	}

	result = EvaluateProviderNewWorkWithActivity(
		controls,
		request,
		12,
		true,
	)
	if result.Allowed {
		t.Fatal(
			"maximum-earnings mode accepted new work after the local user returned",
		)
	}

	if action := EvaluateActiveProviderWork(false, false); action != ActiveProviderWorkContinue {
		t.Fatalf(
			"local user return unexpectedly terminated active work: got %q, want %q",
			action,
			ActiveProviderWorkContinue,
		)
	}
}
