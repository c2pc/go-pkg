package tls

import "crypto/tls"

func GetCipherSuiteIDs() []uint16 {
	suites := tls.CipherSuites()
	ids := make([]uint16, len(suites))
	for i, suite := range suites {
		ids[i] = suite.ID
	}
	return ids
}
