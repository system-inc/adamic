//go:build darwin

package oracle

import "testing"

func TestMovesAgreesWithNode_malloc(t *testing.T) {
	t.Parallel()

	moveNativeVariant(t, "malloc")
}

func TestProduct_MovesNative_malloc(t *testing.T) {
	t.Parallel()
	moveBinary(t, "malloc")
}
