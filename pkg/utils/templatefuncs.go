package utils

import (
	"html/template"
	"time"
)

func UnsafeHtml(s string) template.HTML {
	return template.HTML(s)

}

func StripSummaryTags(s string) string {
	return RemoveAllTags(s)
}

func DisplayDateString(s time.Time) string {
	return DisplayDateWithTime(s)
}

func DisplayDateV2(s int32) string {
	return DisplayDate(int64(s))
}

func TruncateBody(length int, s string) string {
	if len(s) < length {
		return s
	}
	return s[0:length] + "..."
}
