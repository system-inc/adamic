package checkplanted

import "testing"

func TestWorks(t *testing.T) {
	t.Parallel()
	if Misformatted() != 2 {
		t.Fatal("planted: a failing test")
	}
}
