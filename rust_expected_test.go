// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package demangle

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

const rustFilename = "testdata/rust-demangle-expected"

// TestRustExpected is like TestExpected, but for Rust demangling.
// We ignore all test inputs that do not start with _R.
func TestRustExpected(t *testing.T) {
	t.Parallel()
	f, err := os.Open(rustFilename)
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	lineno := 1
	for {
		format, got := getOptLine(t, scanner, &lineno)
		if !got {
			break
		}
		report := lineno
		input := getLine(t, scanner, &lineno)
		expect := getLine(t, scanner, &lineno)

		skip := false
		if len(format) > 0 && format[0] == '-' {
			for _, arg := range strings.Fields(format) {
				switch arg {
				case "--format=gnu-v3":
					skip = true
				case "--format=auto":
				case "--format=rust":
				default:
					t.Errorf("%s:%d: unrecognized argument %s", rustFilename, report, arg)
				}
			}
		}

		if skip {
			continue
		}

		oneRustTest(t, report, input, expect)
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
}

// oneRustTest tests one entry from rust-demangle-expected.
func oneRustTest(t *testing.T, report int, input, expect string) {
	if *verbose {
		fmt.Println(input)
	}

	s, err := ToString(input)
	if err != nil {
		if err != ErrNotMangledName {
			if input == expect {
				return
			}
			t.Errorf("%s:%d: %v", rustFilename, report, err)
			return
		}
		s = input
	}

	if s != expect {
		t.Errorf("%s:%d: got %q, want %q", rustFilename, report, s, expect)
	}
}

const rustCheckFilename = "testdata/rust.test"

func TestRustCheck(t *testing.T) {
	t.Parallel()
	f, err := os.Open(rustCheckFilename)
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	lineno := 1
	for scanner.Scan() {
		report := lineno
		line := strings.TrimSpace(scanner.Text())
		lineno++
		if !strings.HasPrefix(line, "CHECK: ") {
			continue
		}
		want := strings.TrimPrefix(line, "CHECK: ")
		if !scanner.Scan() {
			t.Fatalf("%s:%d: unexpected EOF", rustCheckFilename, report)
		}
		lineno++
		input := strings.TrimSpace(scanner.Text())

		got, err := ToString(input, LLVMStyle)
		if err != nil {
			if want != input {
				t.Errorf("%s:%d: %v", rustCheckFilename, report, err)
			}
		} else if got != want {
			t.Errorf("%s:%d: got %q, want %q", rustCheckFilename, report, got, want)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
}
