package clipboard

import "testing"

func TestSetProviderOverridesCopy(t *testing.T) {
	var captured string
	SetProvider(ProviderFunc(func(text string) error {
		captured = text
		return nil
	}))
	defer SetProvider(nil)

	expected := "hello world"
	if err := Copy(expected); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}

	if captured != expected {
		t.Fatalf("expected %q copied, got %q", expected, captured)
	}
}
