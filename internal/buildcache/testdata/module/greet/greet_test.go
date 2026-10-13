package greet

import "testing"

func TestGreeting(t *testing.T) {
	if Greeting() != "hello" {
		t.Fatal(Greeting())
	}
}
