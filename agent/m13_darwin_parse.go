package agent

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	protocol "github.com/wwjd4u/MeshAlot/protocol/v1"
)

var m13DarwinIdleRE = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)%\s+idle`)
var m13DarwinPageRE = regexp.MustCompile(`page size of ([0-9]+) bytes`)

// Mac's top reports an idle percentage. The final CPU usage line represents
// the second sample instead of a since-boot first sample.
func m13DarwinCPUUsage(raw []byte) (float64, error) {
	matches := m13DarwinIdleRE.FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		return 0, errors.New("top CPU idle sample missing")
	}
	idle, err := strconv.ParseFloat(matches[len(matches)-1][1], 64)
	if err != nil || math.IsNaN(idle) || math.IsInf(idle, 0) || idle < 0 || idle > 100 {
		return 0, errors.New("invalid top CPU idle percentage")
	}
	return 100 - idle, nil
}

// vm_stat's reclaimable count is an approximate snapshot: free + inactive +
// speculative + purgeable pages. This is not an assertion of physical VRAM.
func m13DarwinAvailableRAM(raw []byte) (uint64, error) {
	txt := string(raw)
	match := m13DarwinPageRE.FindStringSubmatch(txt)
	if len(match) != 2 {
		return 0, errors.New("vm_stat page size missing")
	}
	size, err := strconv.ParseUint(match[1], 10, 64)
	if err != nil || size == 0 || size > (1<<20) {
		return 0, errors.New("invalid vm_stat page size")
	}
	want := map[string]bool{"Pages free": true, "Pages inactive": true,
		"Pages speculative": true, "Pages purgeable": true}
	var pageSum uint64
	found := 0
	for _, line := range strings.Split(txt, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if !want[key] {
			continue
		}
		value = strings.TrimSuffix(strings.TrimSpace(value), ".")
		count, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid vm_stat count for %s", key)
		}
		if count > protocol.M13MaxResourceBytes/size-pageSum {
			return 0, errors.New("vm_stat available RAM exceeds sane limit")
		}
		pageSum += count
		if key == "Pages free" || key == "Pages inactive" {
			found++
		}
	}
	if found < 2 {
		return 0, errors.New("required vm_stat counters missing")
	}
	return pageSum * size, nil
}
