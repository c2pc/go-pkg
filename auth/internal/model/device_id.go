package model

const (
	IOSDeviceID     = 1
	AndroidDeviceID = 2
	DesktopDeviceID = 3
	WebDeviceID     = 4
	ConsoleDeviceID = 5
	LinuxDeviceID   = 6
	WindowsDeviceID = 7

	IOSDeviceStr     = "IOS"
	AndroidDeviceStr = "Android"
	DesktopDeviceStr = "Desktop"
	WebDeviceStr     = "Web"
	ConsoleDeviceStr = "Console"
	LinuxDeviceStr   = "Linux"
	WindowsDeviceStr = "Windows"
)

var DeviceIDs = []int{
	IOSDeviceID,
	AndroidDeviceID,
	DesktopDeviceID,
	WebDeviceID,
	ConsoleDeviceID,
	LinuxDeviceID,
	WindowsDeviceID,
}

var DeviceStrs = []string{
	IOSDeviceStr,
	AndroidDeviceStr,
	DesktopDeviceStr,
	WebDeviceStr,
	ConsoleDeviceStr,
	LinuxDeviceStr,
	WindowsDeviceStr,
}

var DeviceID2Name = map[int]string{
	IOSDeviceID:     IOSDeviceStr,
	AndroidDeviceID: AndroidDeviceStr,
	DesktopDeviceID: DesktopDeviceStr,
	WebDeviceID:     WebDeviceStr,
	ConsoleDeviceID: ConsoleDeviceStr,
	LinuxDeviceID:   LinuxDeviceStr,
	WindowsDeviceID: WindowsDeviceStr,
}

var DeviceName2ID = map[string]int{
	IOSDeviceStr:     IOSDeviceID,
	AndroidDeviceStr: AndroidDeviceID,
	DesktopDeviceStr: DesktopDeviceID,
	WebDeviceStr:     WebDeviceID,
	ConsoleDeviceStr: ConsoleDeviceID,
	LinuxDeviceStr:   LinuxDeviceID,
	WindowsDeviceStr: WindowsDeviceID,
}

func DeviceIDToName(num int) string {
	return DeviceID2Name[num]
}

func DeviceNameToID(name string) int {
	return DeviceName2ID[name]
}
