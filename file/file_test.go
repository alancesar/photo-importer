package file

import (
	"fmt"
	"testing"
	"time"
)

func TestDateToPath(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2020-02-13")
	path := DateToPath(date)
	if path != fmt.Sprintf("2020%s2020-02-13", separator) {
		t.Error("invalid path")
	}
}
