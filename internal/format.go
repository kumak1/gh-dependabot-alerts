package internal

import (
	"github.com/fatih/color"
	"time"
)

func formatIndex(index string) string {
	return color.GreenString("#" + index)
}

func formatDate(t time.Time) string {
	return color.WhiteString(t.Format("2006-01-02"))
}

func formatSeverity(severity string) string {
	switch severity {
	case "low":
		return color.WhiteString(severity)
	case "medium":
		return color.YellowString(severity)
	case "high":
		return color.RedString(severity)
	case "critical":
		return color.HiRedString(severity)
	default:
		return severity
	}
}
