// Package replace puts a finished file in place of another, as one rename, on every platform cohere ships.
package replace

import (
	"os"
	"time"
)

// File renames from over to, replacing what is there.
//
// On Unix a rename replaces a file whoever holds it open. Windows refuses while another process holds the
// destination open without sharing deletion, which Go's own os.Open does not share. So a run reading the
// cache table, or an antivirus scan of a source file, fails the rename beside it for as long as the read
// takes. There the rename is retried for a bounded while, then fails as it always did.
func File(from string, to string) error {
	return replacing(os.Rename, holdsTheDestination, time.Sleep, from, to)
}

// retryBound is how long a refused rename is retried. A read of the largest cache table takes well under
// it; what still holds the file after it is holding it for its own reasons, and waiting longer would only
// make a loud failure a slow one.
const retryBound = 2 * time.Second

// replacing is File with what it calls handed in, so a test can refuse a rename the way Windows does on
// a platform that never would.
func replacing(rename func(string, string) error, retryable func(error) bool, sleep func(time.Duration), from string, to string) error {
	pause := 5 * time.Millisecond
	waited := time.Duration(0)
	for {
		err := rename(from, to)
		if err == nil || !retryable(err) || waited >= retryBound {
			return err
		}
		sleep(pause)
		waited += pause
		pause = min(pause*2, 200*time.Millisecond)
	}
}
