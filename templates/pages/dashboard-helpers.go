package pages

import (
	"math"

	"etop/services"
)

type segmentData struct {
	Color        string
	Length       float64
	Offset       float64
	Circumference float64
}

func computeSegments(types []services.TypeCount, total int64) []segmentData {
	colors := []string{"#3b82f6", "#22c55e", "#f59e0b", "#ef4444", "#8b5cf6", "#ec4899", "#14b8a6"}
	c := 2 * math.Pi * 13
	var segments []segmentData
	offset := 0.0
	for i, t := range types {
		pct := float64(t.Count) / float64(total)
		length := c * pct
		segments = append(segments, segmentData{
			Color:         colors[i%len(colors)],
			Length:        length,
			Offset:        offset,
			Circumference: c,
		})
		offset += length
	}
	return segments
}

func circumference(value float64) float64 {
	c := 2 * math.Pi * 15.5
	return c * clampPercentage(value) / 100
}

func clampPercentage(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func progressBarColor(v float64) string {
	if v >= 80 {
		return "bg-green-500"
	}
	if v >= 50 {
		return "bg-amber-500"
	}
	return "bg-red-500"
}

func maxMonthlyCount(ms []services.MonthlyCount) int64 {
	var m int64
	for _, v := range ms {
		if v.Count > m {
			m = v.Count
		}
	}
	if m == 0 {
		m = 1
	}
	return m
}

func maxDistCount(dist []services.StatusCount) int64 {
	var m int64
	for _, s := range dist {
		if s.Count > m {
			m = s.Count
		}
	}
	if m == 0 {
		m = 1
	}
	return m
}

func statusColor(color string) string {
	if color == "" {
		return "bg-primary"
	}
	return color
}

func sumTypeCount(types []services.TypeCount) int64 {
	var total int64
	for _, t := range types {
		total += t.Count
	}
	if total == 0 {
		total = 1
	}
	return total
}

func typeColor(t string) string {
	switch t {
	case "PROJECT":
		return "bg-blue-500"
	case "DAILY":
		return "bg-green-500"
	case "TICKET":
		return "bg-amber-500"
	default:
		return "bg-gray-500"
	}
}

func monthLabel(m int) string {
	names := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	if m >= 1 && m <= 12 {
		return names[m]
	}
	return ""
}
