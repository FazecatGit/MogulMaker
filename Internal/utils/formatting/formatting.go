package formatting

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
