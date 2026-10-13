package fuzz

import (
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_Adamic(t *testing.T) {
	t.Parallel()
	buildcache.Adamic(t)
}
