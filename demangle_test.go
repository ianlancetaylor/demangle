// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package demangle

import (
	"strconv"
	"strings"
	"testing"
)

// Check test cases discovered after the code passed the tests in
// demangle-expected (which are tested by TestExpected).  Some of this
// are cases where we differ from the standard demangler, some we are
// the same but we weren't initially.
func TestDemangler(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{
			"_ZNSaIcEC1ERKS_",
			"std::allocator<char>::allocator(std::allocator<char> const&)",
		},
		{
			"_ZN9__gnu_cxx13stdio_filebufIcSt11char_traitsIcEEC1EP8_IO_FILESt13_Ios_Openmodem",
			"__gnu_cxx::stdio_filebuf<char, std::char_traits<char> >::stdio_filebuf(_IO_FILE*, std::_Ios_Openmode, unsigned long)",
		},
		{
			"_ZN1n1CcvNS_1DIT_EEI1EEEv",
			"n::C::operator n::D<E><E>()",
		},
		{
			"_Z1CIvPN1D1E1FIdJEEEdEPN1GILb0ET_T0_T1_E1HEPFS6_S7_S8_EN1H1I1JIS7_E1KENSG_IS8_E1KE",
			"G<false, void, D::E::F<double>*, double>::H* C<void, D::E::F<double>*, double>(void (*)(D::E::F<double>*, double), H::I::J<D::E::F<double>*>::K, H::I::J<double>::K)",
		},
		{
			"_ZZNK1CI1DIcSt1EIcESaIcEEJEE1FEvE1F",
			"C<D<char, std::E<char>, std::allocator<char> > >::F() const::F",
		},
		{
			"_ZN1CI1DSt1EIK1FN1G1HEEE1I1JIJRKS6_EEEvDpOT_",
			"void C<D, std::E<F const, G::H> >::I::J<std::E<F const, G::H> const&>(std::E<F const, G::H> const&)",
		},
		{
			"_ZN1C1D1E1FIJEEEvi1GDpT_",
			"void C::D::E::F<>(int, G)",
		},
		{
			"_ZN1CILj50ELb1EE1DEv",
			"C<50u, true>::D()",
		},
		{
			"_ZN1CUt_C2Ev",
			"C::{unnamed type#1}::{unnamed type#1}()",
		},
		{
			"_ZN1C12_GLOBAL__N_11DINS_1EEEEN1F1GIDTadcldtcvT__E1HEEEERKS5_NS_1I1JE",
			"F::G<decltype (&((((C::E)()).H)()))> C::(anonymous namespace)::D<C::E>(C::E const&, C::I::J)",
		},
		{
			"_ZN1CI1DE1EIJiRiRPKcRA1_S4_S8_bS6_S3_RjRPKN1F1GERPKN1H1IEEEEvDpOT_",
			"void C<D>::E<int, int&, char const*&, char const (&) [1], char const (&) [1], bool, char const*&, int&, unsigned int&, F::G const*&, H::I const*&>(int&&, int&, char const*&, char const (&) [1], char const (&) [1], bool&&, char const*&, int&, unsigned int&, F::G const*&, H::I const*&)",
		},
		{
			"_ZN1C12_GLOBAL__N_11DIFbPKNS_1EEEEEvPNS_1FERKT_",
			"void C::(anonymous namespace)::D<bool (C::E const*)>(C::F*, bool (&)(C::E const*) const)",
		},
		{
			"_ZN1C1D1EIJRFviSt1FIFvRKN1G1H1IEEERKSt6vectorINS_1JESaISB_EEERiS9_EvEENS0_1K1LIJDpNSt1MIT_E1NEEEEDpOSM_",
			"C::D::K::L<std::M<void (&)(int, std::F<void (G::H::I const&)>, std::vector<C::J, std::allocator<C::J> > const&)>::N, std::M<int&>::N, std::M<std::F<void (G::H::I const&)> >::N> C::D::E<void (&)(int, std::F<void (G::H::I const&)>, std::vector<C::J, std::allocator<C::J> > const&), int&, std::F<void (G::H::I const&)>, void>(void (&)(int, std::F<void (G::H::I const&)>, std::vector<C::J, std::allocator<C::J> > const&), int&, std::F<void (G::H::I const&)>&&)",
		},
		{
			"_ZN1C1D1E1FcvNS_1GIT_EEI1HEEv",
			"C::D::E::F::operator C::G<H><H>()",
		},
		{
			"_ZN9__gnu_cxx17__normal_iteratorIPK1EIN1F1G1HEESt6vectorIS5_SaIS5_EEEC2IPS5_EERKNS0_IT_NS_11__enable_ifIXsr3std10__are_sameISE_SD_EE7__valueESA_E1IEEE",
			"__gnu_cxx::__normal_iterator<E<F::G::H> const*, std::vector<E<F::G::H>, std::allocator<E<F::G::H> > > >::__normal_iterator<E<F::G::H>*>(__gnu_cxx::__normal_iterator<E<F::G::H>*, __gnu_cxx::__enable_if<std::__are_same<E<F::G::H>*, E<F::G::H>*>::__value, std::vector<E<F::G::H>, std::allocator<E<F::G::H> > > >::I> const&)",
		},
		{
			"_ZNKSt1CIM1DKFjvEEclIJEvEEjPKS0_DpOT_",
			"unsigned int std::C<unsigned int (D::*)() const>::operator()<void>(D const*) const",
		},
		{
			"_ZNSt10_HashtableI12basic_stringIcSt11char_traitsIcESaIcEESt4pairIKS4_N1C1D1EEESaISA_ENSt8__detail10_Select1stESt8equal_toIS4_ESt4hashIS4_ENSC_18_Mod_range_hashingENSC_20_Default_ranged_hashENSC_20_Prime_rehash_policyENSC_17_Hashtable_traitsILb1ELb0ELb1EEEE9_M_assignIZNSN_C1ERKSN_EUlPKNSC_10_Hash_nodeISA_Lb1EEEE_EEvSQ_RKT_",
			"void std::_Hashtable<basic_string<char, std::char_traits<char>, std::allocator<char> >, std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, std::allocator<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E> >, std::__detail::_Select1st, std::equal_to<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::hash<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::__detail::_Mod_range_hashing, std::__detail::_Default_ranged_hash, std::__detail::_Prime_rehash_policy, std::__detail::_Hashtable_traits<true, false, true> >::_M_assign<std::_Hashtable<basic_string<char, std::char_traits<char>, std::allocator<char> >, std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, std::allocator<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E> >, std::__detail::_Select1st, std::equal_to<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::hash<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::__detail::_Mod_range_hashing, std::__detail::_Default_ranged_hash, std::__detail::_Prime_rehash_policy, std::__detail::_Hashtable_traits<true, false, true> >::_Hashtable(std::_Hashtable<basic_string<char, std::char_traits<char>, std::allocator<char> >, std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, std::allocator<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E> >, std::__detail::_Select1st, std::equal_to<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::hash<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::__detail::_Mod_range_hashing, std::__detail::_Default_ranged_hash, std::__detail::_Prime_rehash_policy, std::__detail::_Hashtable_traits<true, false, true> > const&)::{lambda(std::__detail::_Hash_node<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, true> const*)#1}>(std::_Hashtable<basic_string<char, std::char_traits<char>, std::allocator<char> >, std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, std::allocator<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E> >, std::__detail::_Select1st, std::equal_to<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::hash<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::__detail::_Mod_range_hashing, std::__detail::_Default_ranged_hash, std::__detail::_Prime_rehash_policy, std::__detail::_Hashtable_traits<true, false, true> > const&, std::_Hashtable<basic_string<char, std::char_traits<char>, std::allocator<char> >, std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, std::allocator<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E> >, std::__detail::_Select1st, std::equal_to<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::hash<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::__detail::_Mod_range_hashing, std::__detail::_Default_ranged_hash, std::__detail::_Prime_rehash_policy, std::__detail::_Hashtable_traits<true, false, true> >::_Hashtable(std::_Hashtable<basic_string<char, std::char_traits<char>, std::allocator<char> >, std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, std::allocator<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E> >, std::__detail::_Select1st, std::equal_to<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::hash<basic_string<char, std::char_traits<char>, std::allocator<char> > >, std::__detail::_Mod_range_hashing, std::__detail::_Default_ranged_hash, std::__detail::_Prime_rehash_policy, std::__detail::_Hashtable_traits<true, false, true> > const&)::{lambda(std::__detail::_Hash_node<std::pair<basic_string<char, std::char_traits<char>, std::allocator<char> > const, C::D::E>, true> const*)#1} const&)",
		},
		{
			"_ZSt3maxIVdERKT_S3_S3_",
			"double const volatile& std::max<double volatile>(double const volatile&, double const volatile&)",
		},
		{
			"_ZZN1C1D1E1F1G1HEvENUlvE_C2EOS4_",
			"C::D::E::F::G::H()::{lambda()#1}::{lambda()#1}({lambda()#1}&&)",
		},
		{
			"_ZThn8_NK1C1D1EEv",
			"non-virtual thunk to C::D::E() const",
		},
		{
			"_ZTv0_n96_NK1C1D1E1FEv",
			"virtual thunk to C::D::E::F() const",
		},
		{
			"_ZTCSt9strstream16_So",
			"construction vtable for std::ostream-in-std::strstream",
		},
		{
			"_ZGVZZN1C1D1EEvENK3$_0clEvE1F",
			"guard variable for C::D::E()::$_0::operator()() const::F",
		},
		{
			"_Z1fICiEvT_",
			"void f<int _Complex>(int _Complex)",
		},
		{
			"_GLOBAL__D__Z2fnv",
			"global destructors keyed to fn()",
		},
		{
			"_Z1fIXadL_Z1hvEEEvv",
			"void f<&h>()",
		},
		{
			"_Z1CIP1DEiRK1EPT_N1F1GIS5_Xaasr1HIS5_E1IntsrSA_1JEE1KE",
			"int C<D*>(E const&, D**, F::G<D*, H<D*>::I&&(!H<D*>::J)>::K)",
		},
	}

	for _, test := range tests {
		got, err := ToString(test.input)
		if err != nil {
			t.Errorf("demangling %s: unexpected error %v", test.input, err)
		} else if got != test.want {
			t.Errorf("demangling %s: got %s, want %s", test.input, got, test.want)
		}

		// Test Filter also.
		if got = Filter(test.input); got != test.want {
			t.Errorf("Filter(%s) == %s, want %s", test.input, got, test.want)
		}
	}
}

// Test for some failure cases.
func TestFailure(t *testing.T) {
	var tests = []struct {
		input string
		error string
		off   int
	}{
		{
			"_Z1FE",
			"unparsed characters at end of mangled name",
			4,
		},
		{
			"_Z1FQ",
			"unrecognized type code",
			4,
		},
	}

	for _, test := range tests {
		got, err := ToString(test.input)
		if err == nil {
			t.Errorf("unexpected success for %s: %s", test.input, got)
		} else if !strings.Contains(err.Error(), test.error) {
			t.Errorf("unexpected error for %s: %v", test.input, err)
		} else {
			s := err.Error()
			i := strings.LastIndex(s, " at ")
			if i < 0 {
				t.Errorf("missing offset in error for %s: %v", test.input, err)
			} else {
				off, oerr := strconv.Atoi(s[i+4:])
				if oerr != nil {
					t.Errorf("can't parse offset (%s) for %s: %v", s[i+4:], test.input, err)
				} else if off != test.off {
					t.Errorf("unexpected offset for %s: got %d, want %d", test.input, off, test.off)
				}
			}
		}

		if got := Filter(test.input); got != test.input {
			t.Errorf("Filter(%s) == %s, want %s", test.input, got, test.input)
		}
	}
}
