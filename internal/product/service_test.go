package product

import "testing"

func TestCents(t *testing.T) {
	got, err := Cents(79.99)
	if err != nil || got != 7999 {
		t.Fatalf("got %d, %v", got, err)
	}
}
func TestCentsRejectsInvalidPrices(t *testing.T) {
	for _, v := range []float64{0, -1, 1.001} {
		if _, err := Cents(v); err == nil {
			t.Fatalf("Cents(%v) accepted", v)
		}
	}
}
