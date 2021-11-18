package yandex

type MDBResourcePreset struct {
	Cores        int64
	Memory       int64
	CpuPlatform  string
	CoreFraction int8
}

type MDBResourceItem struct {
	CpuPlatform  string
	Cores        int64
	CoreFraction int8
	Memory       int64
	NetworkSSD   int64
	NetworkHDD   int64
}

var cpuPlatforms = []string{"Intel Broadwell", "Intel Cascade Lake", "Intel Ice Lake"}
