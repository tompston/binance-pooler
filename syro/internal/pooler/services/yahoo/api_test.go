package yahoo

import (
	"fmt"
	"testing"
	"time"
)

func TestApi(t *testing.T) {

	from := time.Now().AddDate(0, 0, -2)
	to := time.Now()

	fmt.Printf("From: %v To: %v\n", from.UTC(), to.UTC())

	if err := GetStockData("ASM.AS", from, to, Interval5m); err != nil {
		t.Errorf("Error: %v", err)
	}
}
