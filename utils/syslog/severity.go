package syslog

type Severity int

const (
	SeverityUnknown  Severity = 0
	SeverityLow      Severity = 1
	SeverityMedium   Severity = 3
	SeverityHigh     Severity = 6
	SeverityCritical Severity = 9
)
