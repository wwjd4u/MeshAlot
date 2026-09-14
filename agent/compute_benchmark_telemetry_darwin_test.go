//go:build darwin

package agent

import "testing"

func TestParseDarwinProcessRSS(
	t *testing.T,
) {
	got, err := parseDarwinProcessRSS(
		[]byte("  97980\n"),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := uint64(97980 * 1024)

	if got != want {
		t.Fatalf(
			"RSS = %d, want %d",
			got,
			want,
		)
	}
}

func TestParseDarwinProcessRSSRejectsBadValue(
	t *testing.T,
) {
	for _, raw := range []string{
		"",
		"0",
		"not-a-number",
		"123 456",
	} {
		if _, err := parseDarwinProcessRSS(
			[]byte(raw),
		); err == nil {
			t.Fatalf(
				"invalid RSS %q accepted",
				raw,
			)
		}
	}
}

func TestParseDarwinThermalStateNormal(
	t *testing.T,
) {
	raw := `
Note: No thermal warning level has been recorded
Note: No performance warning level has been recorded
CPU Power notify
    CPU_Scheduler_Limit = 100
    CPU_Available_CPUs = 16
    CPU_Speed_Limit = 100
`

	throttled, err :=
		parseDarwinThermalState(
			[]byte(raw),
			16,
		)

	if err != nil {
		t.Fatal(err)
	}

	if throttled {
		t.Fatal(
			"normal thermal state marked throttled",
		)
	}
}

func TestParseDarwinThermalStateDetectsSpeedThrottle(
	t *testing.T,
) {
	raw := `
CPU_Scheduler_Limit = 100
CPU_Available_CPUs = 16
CPU_Speed_Limit = 75
`

	throttled, err :=
		parseDarwinThermalState(
			[]byte(raw),
			16,
		)

	if err != nil {
		t.Fatal(err)
	}

	if !throttled {
		t.Fatal(
			"CPU speed throttle was not detected",
		)
	}
}

func TestParseDarwinThermalStateDetectsCPURestriction(
	t *testing.T,
) {
	raw := `
CPU_Scheduler_Limit = 100
CPU_Available_CPUs = 8
CPU_Speed_Limit = 100
`

	throttled, err :=
		parseDarwinThermalState(
			[]byte(raw),
			16,
		)

	if err != nil {
		t.Fatal(err)
	}

	if !throttled {
		t.Fatal(
			"CPU availability restriction was not detected",
		)
	}
}

func TestParseDarwinThermalStateDetectsSchedulerThrottle(
	t *testing.T,
) {
	raw := `
CPU_Scheduler_Limit = 80
CPU_Available_CPUs = 16
CPU_Speed_Limit = 100
`

	throttled, err :=
		parseDarwinThermalState(
			[]byte(raw),
			16,
		)

	if err != nil {
		t.Fatal(err)
	}

	if !throttled {
		t.Fatal(
			"scheduler throttle was not detected",
		)
	}
}

func TestParseDarwinThermalStateRejectsIncompleteData(
	t *testing.T,
) {
	raw := `
CPU_Scheduler_Limit = 100
CPU_Speed_Limit = 100
`

	if _, err := parseDarwinThermalState(
		[]byte(raw),
		16,
	); err == nil {
		t.Fatal(
			"incomplete thermal state was accepted",
		)
	}
}

func TestParseDarwinTimePeakRSS(
	t *testing.T,
) {
	raw := `
        0.40 real         0.06 user         0.08 sys
            73805824  maximum resident set size
                   0  average shared memory size
                   0  swaps
            70766592  peak memory footprint
`

	got, err := parseDarwinTimePeakRSS(
		[]byte(raw),
	)
	if err != nil {
		t.Fatal(err)
	}

	const want uint64 = 73805824

	if got != want {
		t.Fatalf(
			"peak RSS = %d, want %d",
			got,
			want,
		)
	}
}

func TestParseDarwinTimePeakRSSDoesNotUseMemoryFootprint(
	t *testing.T,
) {
	raw := `
            73596928  maximum resident set size
            99999999  peak memory footprint
`

	got, err := parseDarwinTimePeakRSS(
		[]byte(raw),
	)
	if err != nil {
		t.Fatal(err)
	}

	if got != 73596928 {
		t.Fatalf(
			"peak RSS = %d, want 73596928",
			got,
		)
	}
}

func TestParseDarwinTimePeakRSSRejectsMissingValue(
	t *testing.T,
) {
	raw := `
        0.00 real
        123456 peak memory footprint
`

	if _, err := parseDarwinTimePeakRSS(
		[]byte(raw),
	); err == nil {
		t.Fatal(
			"missing maximum RSS was accepted",
		)
	}
}

func TestParseDarwinTimePeakRSSRejectsZero(
	t *testing.T,
) {
	raw := `
        0 maximum resident set size
`

	if _, err := parseDarwinTimePeakRSS(
		[]byte(raw),
	); err == nil {
		t.Fatal(
			"zero maximum RSS was accepted",
		)
	}
}
