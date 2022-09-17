package load_balancing

import (
	"math"
	"sort"
)

type Pair struct {
	Key   string
	Value HwStatusInfo
}

type PairList []Pair

func (p PairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PairList) Len() int {
	return len(p)
}

//sorted by RAM, then if equal - CPU
//RAM values are ceiled to 10th, so e.g. 16 becomes 20 etc.
//needed to average RAM load a bit and allow picking node with lower CPU between several nodes with close RAM load values
func (p PairList) Less(i, j int) bool {
	roundRam1 := ceilToTen(p[i].Value.LastRamUsagePerc)
	cpu1 := p[i].Value.AvgRecentCPUUsagePerc

	roundRam2 := ceilToTen(p[j].Value.LastRamUsagePerc)
	cpu2 := p[j].Value.AvgRecentCPUUsagePerc

	if roundRam1 != roundRam2 {
		return roundRam1 < roundRam2
	} else {
		return cpu1 < cpu2
	}
}

//rounds towards next 10th e.g. - 0 to 10, 3 to 10, 12 to 20, 18 to 20 etc.
func ceilToTen(val float64) int {
	if val == 0 {
		return 10
	}
	return int(math.Ceil(val/10)) * 10
}

func sortMapByValue(m *map[string]HwStatusInfo) PairList {
	p := make(PairList, len(*m))
	i := 0
	for k, v := range *m {
		p[i] = Pair{k, v}
		i += 1
	}

	sort.Sort(p)
	return p
}
