package user

import "testing"

func TestValidate(t *testing.T) {
	n, e, err := Validate("  Jane Doe ", " JANE@example.com ")
	if err != nil || n != "Jane Doe" || e != "jane@example.com" {
		t.Fatalf("got %q %q %v", n, e, err)
	}
}
func TestValidateRejectsBlankAndInvalidEmail(t *testing.T) {
	for _, tc := range [][2]string{{"   ", "a@example.com"}, {"Jane", "bad"}} {
		if _, _, err := Validate(tc[0], tc[1]); err == nil {
			t.Fatalf("Validate(%q,%q) accepted", tc[0], tc[1])
		}
	}
}
