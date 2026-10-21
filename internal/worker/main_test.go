package worker

import (
	"os"
	"testing"

	"github.com/system-inc/adamic/internal/oracletest"
)

func TestMain(main *testing.M) {
	os.Exit(oracletest.Run(main.Run))
}
