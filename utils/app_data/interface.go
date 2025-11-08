package app_data

import "net"

type IPInfo struct {
	IP      net.IP
	Family  string
	Netmask net.IP
}

type InterfaceInfo struct {
	Name      string
	MAC       string
	Addresses []IPInfo
	Active    bool
}

func Interfaces() ([]InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var interfaces []InterfaceInfo
	for _, iface := range ifaces {
		// Пропускаем loopback
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ipInfos []IPInfo
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip := v.IP
				if ip.IsLoopback() {
					continue
				}
				var family string
				var netmask net.IP

				if ip4 := ip.To4(); ip4 != nil {
					ip = ip4
					netmask = net.IP(v.Mask).To4()
					family = "inet"
				} else {
					// Для IPv6
					netmask = net.IP(v.Mask)
					family = "inet6"
				}

				ipInfos = append(ipInfos, IPInfo{
					IP:      ip,
					Family:  family,
					Netmask: netmask,
				})
			}
		}

		if len(ipInfos) > 0 {
			interfaces = append(interfaces, InterfaceInfo{
				Name:      iface.Name,
				MAC:       iface.HardwareAddr.String(),
				Addresses: ipInfos,
				Active:    (iface.Flags & net.FlagUp) != 0,
			})
		}
	}

	return interfaces, nil
}
