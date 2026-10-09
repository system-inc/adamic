package lint

import "testing"

// Declare the exact recipes fetched by these independently selected shards.
// C-emission performance is tracked separately; no shard budget exemption.
func TestProduct_EmittedMismatchLowered(t *testing.T) { t.Parallel(); emittedMismatchLowered(t) }
func TestProduct_EmittedMismatchNative(t *testing.T)  { t.Parallel(); emittedMismatchNative(t) }
func TestProduct_RulesAgreeLowered(t *testing.T)      { t.Parallel(); rulesAgreeLowered(t) }
func TestProduct_RulesAgreeNative(t *testing.T)       { t.Parallel(); rulesAgreeNative(t) }
func TestProduct_RulesAgreeCapture(t *testing.T)      { t.Parallel(); rulesAgreeCapture(t) }
