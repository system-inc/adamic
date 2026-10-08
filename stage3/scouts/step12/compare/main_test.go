package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestCompare(t *testing.T) {
	t.Parallel()
	cases := []struct {
		left, right string
		first       *Difference
	}{
		{"", "", nil}, {"a\nb\n", "a\nb\n", nil},
		{"a\nb\n", "a\nc\n", &Difference{2, 2, 1, 'b', 'c'}},
		{"a", "ab", &Difference{1, 1, 2, -1, 'b'}},
		{"ab", "a", &Difference{1, 1, 2, 'b', -1}},
		{"\x00\xff", "\x00\xfe", &Difference{1, 1, 2, 255, 254}},
	}
	for _, c := range cases {
		r, err := Compare(strings.NewReader(c.left), strings.NewReader(c.right))
		if err != nil {
			t.Fatal(err)
		}
		if r.Equal != (c.first == nil) {
			t.Fatalf("equal: %+v", r)
		}
		if c.first != nil && (r.First == nil || *r.First != *c.first) {
			t.Fatalf("difference: %+v", r.First)
		}
		if r.Left.Bytes != int64(len(c.left)) || r.Right.Bytes != int64(len(c.right)) {
			t.Fatal(r)
		}
		if r.Left.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(c.left))) || r.Right.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(c.right))) {
			t.Fatal(r)
		}
	}
}

// This is a synthetic scale/control measurement, not native scanner evidence.
func Test509014RowsAndMutants(t *testing.T) {
	t.Parallel()
	original := bytes.Repeat([]byte("Identifier\t0\t0\t1\t0\t\"x\"\n"), 509014)
	control, err := Compare(bytes.NewReader(original), bytes.NewReader(original))
	if err != nil || !control.Equal || control.Left.Lines != 509014 {
		t.Fatalf("%+v %v", control, err)
	}
	for _, index := range []int{0, 65535, 65536, len(original) - 2} {
		mutant := bytes.Clone(original)
		mutant[index] ^= 1
		r, err := Compare(bytes.NewReader(original), bytes.NewReader(mutant))
		if err != nil || r.Equal || r.First == nil || r.First.Offset != int64(index) {
			t.Fatalf("byte mutant %d survived: %+v %v", index, r, err)
		}
		t.Logf("byte mutant %d caught", index)
	}
	for _, mutant := range [][]byte{original[:len(original)-1], append(bytes.Clone(original), 'x')} {
		r, err := Compare(bytes.NewReader(original), bytes.NewReader(mutant))
		if err != nil || r.Equal || r.First == nil {
			t.Fatalf("EOF mutant survived: %+v %v", r, err)
		}
		t.Logf("EOF mutant caught at %d", r.First.Offset)
	}
}

type failedReader struct{}

func (failedReader) Read([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestReadError(t *testing.T) {
	t.Parallel()
	for _, pair := range [][2]io.Reader{{failedReader{}, strings.NewReader("")}, {strings.NewReader(""), failedReader{}}} {
		if _, err := Compare(pair[0], pair[1]); err == nil {
			t.Fatal("read error reported as equality")
		}
	}
}
