package notes

import (
	"fmt"
	"time"
)

func RenderDailyTemplate(date time.Time) string {
	return fmt.Sprintf(`---
date: %s
---
# %s

## Tasks

## Notes
`, date.Format("2006-01-02"), date.Format("Monday, January 2, 2006"))
}
