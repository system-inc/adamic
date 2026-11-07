package fuzz

import (
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_Adamic(test *testing.T) {
	test.Parallel()
	buildcache.AdamicBuiltElsewhere(test)
}
