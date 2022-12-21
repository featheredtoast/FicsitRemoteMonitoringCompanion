package exporter
import (
	"regexp"
	"strings"
	"time"
)

var timeRegex = regexp.MustCompile(`\d\d:\d\d:\d\d`)
func parseTimeSeconds(timeStr string) (bool, float64) {
	match := timeRegex.FindStringSubmatch(timeStr)
	if len(match) < 1 {
		return false, 0
	}
	parts := strings.Split(match[0], ":")
	duration := parts[0] + "h" + parts[1] + "m" + parts[2] + "s"
	t, _ := time.ParseDuration(duration)
	return true, t.Seconds()
}
