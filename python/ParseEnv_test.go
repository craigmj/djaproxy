package python

import (
	`fmt`
	"strings"
	"testing"
)

type envtest struct {
	In string
	Expect string
}
func (et *envtest) Check(i int, T *testing.T) {
	pairs, err := ParseEnv(et.In)
	if nil!=err {
		T.Fatalf("Test %d: %s", i, err.Error())
	}
	pt := make([]string, len(pairs))
	for j, p := range pairs {
		pt[j] = fmt.Sprintf("%s=%s", p[0],p[1])
	}
	got := strings.Join(pt, `,`)
	expect := et.Expect
	if ``==expect {
		expect = et.In
	}
	if got!=expect {
		T.Errorf(`Failed test %d: expected '%s', got '%s'`, i, expect, got)
	}
}

func TestEnvParsing(T *testing.T) {
	for i, test := range []*envtest{
		{ "a=1,b=2,c=3", ""},
		{ "a,  b=2,c=3", "a=,b=2,c=3"},
		{ ` "a"="a test", b  =   'another test'; c='three'`, "a=a test,b=another test,c=three"},
		{ `"fancy key"  =  "fancy key with \n delimiter"`, "fancy key=fancy key with \n delimiter"},
	} {
		test.Check(i, T)
	}
}