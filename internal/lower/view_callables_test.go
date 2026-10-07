package lower

import "testing"

func TestViewCallableSignature(t *testing.T) {
	t.Parallel()
	if err := checkViewCallableSignature("fixture.a:1", "run", true); err != nil {
		t.Fatal(err)
	}
	const want = "fixture.a:1: Adamic 0.1 refuses a checked view call to member run; its signature cannot be checked at runtime and no compatible implementation is proven; prove the callable implementation before calling through this view"
	err := checkViewCallableSignature("fixture.a:1", "run", false)
	if err == nil || err.Error() != want {
		t.Fatalf("want %q, got %v", want, err)
	}
}
