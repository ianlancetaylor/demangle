// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package demangle

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
)

// casesExpectedFailures is a list of exceptions from cases that we
// don't expect to pass.  This is to let us incrementally fix them
// over time.
var casesExpectedFailures = map[string]bool{
	"_ZSteqIcEN9__gnu_cxx11__enable_ifIXsr9__is_charIT_EE7__valueEbE6__typeERKSbIS2_St11char_traitsIS2_ESaIS2_EESA_": true,
	"_Z1fPU11objcproto1A11objc_object":             true,
	"_Z1fPKU11objcproto1A7NSArray":                 true,
	"_ZNK1AIJ1Z1Y1XEEcv1BIJDpPT_EEIJS2_S1_S0_EEEv": true,
	"_ZNK3Ncr6Silver7Utility6detail12CallOnThreadIZ53-[DeploymentSetupController handleManualServerEntry:]E3$_5EclIJEEEDTclclL_ZNS2_4getTIS4_EERT_vEEspclsr3stdE7forwardIT_Efp_EEEDpOSA_": true,
	"_Z1fIJicEEvDp7MuncherIAstT__S1_E":                 true,
	"_ZN5test31aINS_1XEMS1_PiEEvT_T0_DTdsfL0p_fL0p0_E": true,
	"_Z1fPU3AS1KiS0_":                                  true,
	"_Z1pILb1EEiM1SKDOT_EFivRE":                        true,
	"_Z1pIJicfEEiM1SVKDwDpT_EFivOE":                    true,
	"_ZZ18test_assign_throwsI20small_throws_on_copyLb0EEvvENKUlRNSt3__13anyEOT_E_clIRS0_EEDaS3_S5_": true,
	"_ZN1Scv7MuncherIJDpPT_EEIJFivEA_iEEEv":                                                         true,
	"_Z2f8IiJ8identityIiES0_IfEEEvRAsPiDpT0_T_DpNS3_4typeEE_i":                                      true,
	"_ZZ11inline_funcvENKUlTyTyT_T0_E_clIiiEEDaS_S0_":                                               true,
	"_ZZ11inline_funcvENKUlTyTyT_T1_T0_E_clIiiiEEDaS_S0_S1_":                                        true,

	"_Z1fIXfLpm1x1yEEvv": true,
	"_Z1fIXfLds1x1yEEvv": true,
}

// caseExceptions is a list of exceptions from the LLVM list that we
// do not handle the same as the LLVM demangler.  We keep a list of
// exceptions so that we can use an exact copy of the test cases.  We
// map to an empty string if we expect a demangling failure; this
// differs from caseExpectedFailures in that we've decided that we
// intentionally should not demangle this case.  Otherwise this maps
// to the expected demangling.
var casesExceptions = map[string]string{
	"_ZN1XIZ1fIiEvOT_EUlOT_DpT0_E_EclIJEEEvDpT_": "void X<void f<int>(int&&)::'lambda'(auto&&, auto)>::operator()<>()",
	"_ZN1XIZ1fIiEvOT_EUlS2_DpT0_E_EclIJEEEvDpT_": "void X<void f<int>(int&&)::'lambda'(auto&&, auto)>::operator()<>()",
	"_Z1h1XIJZ1fIiEDaOT_E1AZ1gIdEDaS2_E1BEE":     "h(X<auto f<int>(int&&)::A, auto g<double>(double&&)::B>)",
	"_Zcv1BIRT_EIS1_E":                           "",

	// For the next four test cases it seems to me that
	// <int, int>(int) is more correct than <int, int>(auto).
	// It depends on how we handle a template parameter in
	// and out of a lambda expression.
	"_ZZN5test71fIiEEvvENKUlTyQaa1CIT_E1CITL0__ET0_E_clIiiEEDaS3_Q1CIDtfp_EE":        "auto void test7::f<int>()::'lambda'<typename $T> requires C<T> && C<TL0_> (auto)::operator()<int, int>(int) const requires C<decltype(fp)>",
	"_ZZN5test71fIiEEvvENKUlTyQaa1CIT_E1CITL0__ET0_E0_clIiiEEDaS3_Qaa1CIDtfp_EELb1E": "auto void test7::f<int>()::'lambda0'<typename $T> requires C<T> && C<TL0_> (auto)::operator()<int, int>(int) const requires C<decltype(fp)> && true",
	"_ZZN5test71fIiEEvvENKUlTyQaa1CIT_E1CITL0__ET0_E1_clIiiEEDaS3_Q1CIDtfp_EE":       "auto void test7::f<int>()::'lambda1'<typename $T> requires C<T> && C<TL0_> (auto)::operator()<int, int>(int) const requires C<decltype(fp)>",
	"_ZZN5test71fIiEEvvENKUlTyT0_E_clIiiEEDaS1_":                                     "auto void test7::f<int>()::'lambda'<typename $T>(auto)::operator()<int, int>(int) const",
}

func TestCases(t *testing.T) {
	t.Parallel()

	cases := readCases(t)

	expectedErrors := 0
	expectedDifferent := 0
	found := make(map[string]bool)
	for _, test := range cases {
		expectedFail := casesExpectedFailures[test[0]]
		exception, haveException := casesExceptions[test[0]]
		if expectedFail && haveException {
			t.Errorf("test case error: %s in both expectedFailures and exceptions", test[0])
		}
		want := test[1]
		if haveException && exception != "" {
			if want == exception {
				t.Errorf("test case error: %s exception is expected result", test[0])
			}
			want = exception
		}

		// We don't strip an extra underscore.
		testStr := test[0]
		if strings.HasPrefix(testStr, "__Z") || strings.HasPrefix(testStr, "____Z") {
			testStr = testStr[1:]
		}

		// We don't demangle plain types, so just skip them.
		if !strings.HasPrefix(testStr, "_") {
			continue
		}

		if got, err := ToString(testStr, LLVMStyle); err != nil {
			if expectedFail || (haveException && exception == "") {
				t.Logf("demangling %s: expected failure: error %v", test[0], err)
				if expectedFail {
					expectedErrors++
					found[test[0]] = true
				}
			} else {
				t.Errorf("demangling %s: unexpected error %v", test[0], err)
			}
		} else if got != want {
			if expectedFail {
				t.Logf("demangling %s: expected failure: got %s, want %s", test[0], got, want)
				expectedDifferent++
				found[test[0]] = true
			} else if haveException && exception == "" {
				t.Errorf("demangling %s: expected to fail, but succeeded with %s", test[0], got)
			} else {
				t.Errorf("demangling %s: got %s, want %s", test[0], got, want)
			}
		} else if expectedFail || (haveException && exception == "") {
			t.Errorf("demangling %s: expected to fail, but succeeded with %s", test[0], got)
			if expectedFail {
				found[test[0]] = true
			}
		}
	}
	if len(found) != len(casesExpectedFailures) {
		for expected := range casesExpectedFailures {
			if !found[expected] {
				t.Errorf("expected %s to fail but did not see it", expected)
			}
		}
		for f := range found {
			if !casesExpectedFailures[f] {
				t.Errorf("internal error: found failing but unexpected %s", f)
			}
		}
	}
	if expectedDifferent > 0 {
		t.Logf("%d different demanglings out of %d cases", expectedDifferent, len(cases))
	}
	if expectedErrors > 0 {
		t.Logf("%d expected failures out of %d cases", expectedErrors, len(cases))
	}
}

// readCases reads the LLVM test cases from DemangleTestCases.inc.
// That file is copied from
//
//	llvm-project/libcxxabi/test/DemangleTestCases.inc
//
// That file has no license, but LLVM in general has the license text:
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
// This file is only used for testing and does not form part of the
// demangle package when it is used by other code.
//
// The file does not contain any code, only test cases and comments.
func readCases(t testing.TB) [][2]string {
	const fn = "testdata/DemangleTestCases.inc"
	f, err := os.Open(fn)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var cases [][2]string
	r := bufio.NewReader(f)
	lineno := 1
	for {
		b, atEOF := readCasesUnquotedByte(t, r, fn, &lineno)
		if atEOF {
			break
		}
		if b != '{' {
			t.Fatalf("%s:%d: got %c, want {", fn, lineno, b)
		}

		s1 := readCasesString(t, r, fn, &lineno)

		for {
			b = readCasesUnquotedByteNoEOF(t, r, fn, &lineno)
			if b == ',' {
				break
			}
			if b != '"' {
				t.Fatalf("%s:%d: got %c, want ,", fn, lineno, b)
			}

			r.UnreadByte()
			s1 += readCasesString(t, r, fn, &lineno)
		}

		s2 := readCasesString(t, r, fn, &lineno)

		for {
			b = readCasesUnquotedByteNoEOF(t, r, fn, &lineno)
			if b == '}' {
				break
			}
			if b != '"' {
				t.Fatalf("%s:%d: got %c, want }", fn, lineno, b)
			}

			r.UnreadByte()
			s2 += readCasesString(t, r, fn, &lineno)
		}

		b = readCasesUnquotedByteNoEOF(t, r, fn, &lineno)
		if b != ',' {
			t.Fatalf("%s:%d: got %c, want ,", fn, lineno, b)
		}

		cases = append(cases, [2]string{s1, s2})
	}

	return cases
}

// readCasesString reads a quoted string from the cases file.
func readCasesString(t testing.TB, r *bufio.Reader, fn string, lineno *int) string {
	b, atEOF := readCasesUnquotedByte(t, r, fn, lineno)
	if atEOF {
		t.Fatalf(`%s:%d: got EOF, want "`, fn, *lineno)
	}
	if b != '"' {
		t.Fatalf(`%s:%d: got %c, want "`, fn, *lineno, b)
	}

	var sb strings.Builder
	for {
		b = readCasesByteNoEOF(t, r, fn, *lineno)
		if b == '"' {
			break
		}
		if b != '\\' {
			sb.WriteByte(b)
			continue
		}

		b = readCasesByteNoEOF(t, r, fn, *lineno)
		switch b {
		case '"', '\'', '?', '\\':
			sb.WriteByte(b)
			continue
		case 't':
			sb.WriteByte('\t')
			continue

		case '0', '1', '2', '3', '4', '5', '6', '7':
			val := b - '0'
		octalLoop:
			for {
				b = readCasesByteNoEOF(t, r, fn, *lineno)
				switch b {
				case '0', '1', '2', '3', '4', '5', '6', '7':
					val <<= 3
					val += b - '0'
				default:
					r.UnreadByte()
					break octalLoop
				}
			}
			sb.WriteByte(val)

		case 'x':
			val := byte(0)
			seen := false
		hexLoop:
			for {
				b = readCasesByteNoEOF(t, r, fn, *lineno)
				bval := byte(0)
				switch b {
				case '0', '1', '2', '3', '4', '5', '6', '7':
					bval = b - '0'
				case 'A', 'B', 'C', 'D', 'E', 'F':
					bval = b - 'A'
				case 'a', 'b', 'c', 'd', 'e', 'f':
					bval = b - 'a'
				default:
					r.UnreadByte()
					break hexLoop
				}
				val <<= 4
				val += bval
				seen = true
			}
			if !seen {
				t.Fatalf(`%s:%d: no hex digits after \x`, fn, *lineno)
			}
			sb.WriteByte(val)

		default:
			t.Fatalf(`%s:%d: unexpected escape sequence \%c in string`, fn, *lineno, b)
		}
	}

	return sb.String()
}

// readCasesUnquotedByte reads a byte from the cases file,
// where the byte is not in a quoted string.
// This skips comments and whitespace.
// The bool result reports whether we are at EOF.
func readCasesUnquotedByte(t testing.TB, r *bufio.Reader, fn string, lineno *int) (byte, bool) {
	for {
		b, atEOF := readCasesByte(t, r, fn, *lineno)
		if atEOF {
			return 0, true
		}
		if b == ' ' || b == '\t' {
			continue
		}
		if b == '\n' {
			*lineno++
			continue
		}
		if b == '/' {
			b, atEOF = readCasesByte(t, r, fn, *lineno)
			if atEOF || b != '/' {
				t.Fatalf("%s:%d: unexpected single /", fn, *lineno)
			}
			for {
				_, err := r.ReadSlice('\n')
				if err == nil {
					*lineno++
					break
				}
				switch err {
				case bufio.ErrBufferFull:
				case io.EOF:
					return 0, true
				default:
					t.Fatalf("%s:%d: %v", fn, *lineno, err)
				}
			}
			continue
		}

		return b, false
	}
}

// readCasesUnquotedByteNoEOF is like readCasesUnquotedByte,
// but fails on EOF.
func readCasesUnquotedByteNoEOF(t testing.TB, r *bufio.Reader, fn string, lineno *int) byte {
	b, atEOF := readCasesUnquotedByte(t, r, fn, lineno)
	if atEOF {
		t.Helper()
		t.Fatalf("%s:%d: unexpected EOF", fn, *lineno)
	}
	return b
}

// readCasesByte reads a byte from the cases file.
// The bool result reports whether we are at EOF.
func readCasesByte(t testing.TB, r *bufio.Reader, fn string, lineno int) (byte, bool) {
	b, err := r.ReadByte()
	if err != nil {
		if err == io.EOF {
			return 0, true
		}
		t.Helper()
		t.Fatalf("%s:%d: %v", fn, lineno, err)
	}
	return b, false
}

// readCasesByteNoEOF is like readCasesByte, but fails on EOF.
func readCasesByteNoEOF(t testing.TB, r *bufio.Reader, fn string, lineno int) byte {
	b, atEOF := readCasesByte(t, r, fn, lineno)
	if atEOF {
		t.Helper()
		t.Fatalf("%s:%d: unexpected EOF", fn, lineno)
	}
	return b
}
