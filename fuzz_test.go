// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package demangle

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func FuzzDemangle(f *testing.F) {
	for _, test := range demanglerTests {
		f.Add(test.input)
	}
	for _, test := range failureTests {
		f.Add(test.input)
	}

	// Add all the entries from demangle-expected.

	file, err := os.Open(filename)
	if err != nil {
		f.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	lineno := 1
	for {
		format, got := getOptLine(f, scanner, &lineno)
		if !got {
			break
		}
		report := lineno
		input := getLine(f, scanner, &lineno)
		getLine(f, scanner, &lineno)

		skip, testNoParams := skipExpectedTest(f, format, input, report)
		if testNoParams {
			getLine(f, scanner, &lineno)
		}

		if skip {
			continue
		}

		f.Add(input)
	}

	if err := scanner.Err(); err != nil {
		f.Error(err)
	}

	file.Close()

	// Add all the entries from the LLVM cases.

	cases := readCases(f)
	for _, test := range cases {
		testStr := test[0]
		if strings.HasPrefix(testStr, "__Z") || strings.HasPrefix(testStr, "____Z") {
			testStr = testStr[1:]
		}
		if !strings.HasPrefix(testStr, "_") {
			continue
		}
		f.Add(testStr)
	}

	f.Fuzz(func(t *testing.T, in string) {
		_, err := ToString(in)
		if err != nil {
			t.Skip()
		}
	})
}
