package platform

const (
	IOSPlatformID = iota + 1
	AndroidPlatformID
	HuaweiPlatformID
	BrowserPlatformID
	ConsolePlatformID
	LinuxPlatformID
	WindowsPlatformID

	IOSPlatformStr     = "IOS"
	AndroidPlatformStr = "Android"
	HuaweiPlatformStr  = "Huawei"
	BrowserPlatformStr = "Browser"
	ConsolePlatformStr = "Console"
	LinuxPlatformStr   = "Linux"
	WindowsPlatformStr = "Windows"
)

var PlatformID = []int{
	IOSPlatformID,
	AndroidPlatformID,
	HuaweiPlatformID,
	BrowserPlatformID,
	ConsolePlatformID,
	LinuxPlatformID,
	WindowsPlatformID,
}

var PlatformStr = []string{
	IOSPlatformStr,
	AndroidPlatformStr,
	HuaweiPlatformStr,
	BrowserPlatformStr,
	ConsolePlatformStr,
	LinuxPlatformStr,
	WindowsPlatformStr,
}

var PlatformID2Name = map[int]string{
	IOSPlatformID:     IOSPlatformStr,
	AndroidPlatformID: AndroidPlatformStr,
	HuaweiPlatformID:  HuaweiPlatformStr,
	BrowserPlatformID: BrowserPlatformStr,
	ConsolePlatformID: ConsolePlatformStr,
	LinuxPlatformID:   LinuxPlatformStr,
	WindowsPlatformID: WindowsPlatformStr,
}

var PlatformName2ID = map[string]int{
	IOSPlatformStr:     IOSPlatformID,
	AndroidPlatformStr: AndroidPlatformID,
	HuaweiPlatformStr:  HuaweiPlatformID,
	BrowserPlatformStr: BrowserPlatformID,
	ConsolePlatformStr: ConsolePlatformID,
	LinuxPlatformStr:   LinuxPlatformID,
	WindowsPlatformStr: WindowsPlatformID,
}

func PlatformIDToName(num int) string {
	return PlatformID2Name[num]
}

func PlatformNameToID(name string) int {
	return PlatformName2ID[name]
}
