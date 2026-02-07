package formatting

import "time"

// RepeatString repeats a string n times
func RepeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Separator returns a line separator of given width
func Separator(width int) string {
	return RepeatString("=", width)
}

// ParseDate parses a date string in multiple formats
func ParseDate(dateStr string) time.Time {
	formats := []string{
		"2006-01-02", // YYYY-MM-DD (standard)
		"02/01/2006", // DD/MM/YYYY
		"02.01.2006", // DD.MM.YYYY
		"01-02-2006", // MM-DD-YYYY (US format)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}
