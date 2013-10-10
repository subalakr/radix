package radix

import (
	"fmt"
	"strings"
	"testing"
)

func printit(r *Radix, level int) {
	for i := 0; i < level; i++ {
		fmt.Print("\t")
	}

	fmt.Printf("%p '%v' value: '%v' parent %p\n", r, r.key, r.Value, r.parent)
	for _, child := range r.children {
		printit(child, level+1)
	}
}

func radixtree() *Radix {
	r := New()
	r.Insert(strings.NewReader("test"), "a")
	r.Insert(strings.NewReader("tester"), "a")
	r.Insert(strings.NewReader("team"), "a")
	r.Insert(strings.NewReader("te"), "a")
	return r
}

// None, of the childeren must have a prefix incommon with r.key
func validate(r *Radix) bool {
	return true
	for _, child := range r.children {
		_, i := longestCommonPrefix(r.key, child.key)
		if i != 0 {
			return false
		}
		validate(child)
	}
	return true
}
func TestSuccessor(t *testing.T) {
	a := make(map[byte]*Radix)
	// fake fill it, this is randomized by Go
	a['a'] = nil
	a['b'] = nil
	a['c'] = nil
	a['d'] = nil
	a['e'] = nil
	a['f'] = nil
	s, f := smallestSuccessor(a, 'f')
	if f {
		t.Logf("Should be false")
		t.Fail()
	}
	s, f = smallestSuccessor(a, 'b')
	if s != 'c' {
		t.Logf("Should be c (%s)!", string(s))
		t.Fail()
	}
}

func TestInsert(t *testing.T) {
	r := New()
	if !validate(r) {
		t.Log("Tree does not validate")
		t.Fail()
	}
	if r.Len() != 0 {
		t.Log("Len should be 0", r.Len())
	}
	r.Insert(strings.NewReader("test"), nil)
	r.Insert(strings.NewReader("slow"), nil)
	r.Insert(strings.NewReader("water"), nil)
	r.Insert(strings.NewReader("tester"), nil)
	r.Insert(strings.NewReader("testering"), nil)
	r.Insert(strings.NewReader("rewater"), nil)
	r.Insert(strings.NewReader("waterrat"), nil)
	if !validate(r) {
		t.Log("Tree does not validate")
		t.Fail()
	}
}

func TestRemove(t *testing.T) {
	r := New()
	r.Insert(strings.NewReader("test"), "aa")
	r.Insert(strings.NewReader("slow"), "bb")

	if k := r.Remove("slow").Value; k != "bb" {
		t.Log("should be bb", k)
		t.Fail()
	}

	if r.Remove("slow") != nil {
		t.Log("should be nil")
		t.Fail()
	}
	r.Insert(strings.NewReader("test"), "aa")
	r.Insert(strings.NewReader("tester"), "aa")
	r.Insert(strings.NewReader("testering"), "aa")
	//	r.Find("tester").Remove("test")
}

func TestNext(t *testing.T) {
	r := New()
	r.Insert(strings.NewReader("nl.miek"), "xx")
	r.Insert(strings.NewReader("nl.miek.a"), "xx")
	r.Insert(strings.NewReader("nl.miek.c"), "xx")
	r.Insert(strings.NewReader("nl.miek.d"), "xx")
	r.Insert(strings.NewReader("nl.miek.c.a"), "xx")
	r.Insert(strings.NewReader("nl.miek.c.c"), "xx")
	next := map[string]string{
		"nl.miek":     "nl.miek.a",
		"nl.miek.a":   "nl.miek.c",
		"nl.miek.c.a": "nl.miek.c.c",
		"nl.miek.c.c": "nl.miek.d",
		"nl.miek.d":   "nl.miek",
	}
	for x, nxt := range next {
		r1, _ := r.Find(strings.NewReader(x))
		if n := r1.Next(); n.Key() != nxt {
			fmt.Println(len(n.Key()))
			fmt.Println(len(nxt))
			t.Logf("Next of %s must be %s, is %s\n", x, nxt, n.Key())
			t.Fail()
		}
	}
}

func TestPrev(t *testing.T) {
	r := New()
	r.Insert(strings.NewReader("nl.miek"), "xx")
	r.Insert(strings.NewReader("nl.miek.a"), "xx")
	r.Insert(strings.NewReader("nl.miek.c"), "xx")
	r.Insert(strings.NewReader("nl.miek.d"), "xx")
	r.Insert(strings.NewReader("nl.miek.c.a"), "xx")
	r.Insert(strings.NewReader("nl.miek.c.c"), "xx")
	prev := map[string]string{
		"nl.miek.a":   "nl.miek",
		"nl.miek.c":   "nl.miek.a",
		"nl.miek.c.c": "nl.miek.c.a",
		"nl.miek.d":   "nl.miek.c.c",
		"nl.miek":     "nl.miek.d",
	}
	for x, prv := range prev {
		r1, _ := r.Find(strings.NewReader(x))
		if n := r1.Prev(); n.Key() != prv {
			t.Logf("Prev of %s must be %s, is %s\n", x, prv, n.Key())
			t.Fail()
		}
	}
}

//xn--fake-rr4221.miek.nl.	345600	IN	NS	ns2.example.com.
//xn--fake-rr4223.miek.nl.	345600	IN	NS	ns1.example.com

func TestNextPrev(t *testing.T) {
	r := New()
	r.Insert(strings.NewReader("nl.miek"), "xx")

	r1, _ := r.Find(strings.NewReader("nl.miek"))
	if r1.Next().Key() != "nl.miek" {
		t.Logf("I'm not my own next")
		t.Fail()
	}
	if r1.Prev().Key() != "nl.miek" {
		t.Logf("I'm not my own prev")
		t.Fail()
	}
	// Add another element, making it two (with non-nil values)
	r.Insert(strings.NewReader("nl.miek.a"), "xx")
	// r1 hasn't changed
	r2, _ := r.Find(strings.NewReader("nl.miek.a"))
	if r1.Next().Key() != "nl.miek.a" {
		t.Logf("r1 next should be nl.miek.a")
		t.Fail()
	}
	if r2.Next().Key() != "nl.miek" {
		t.Logf("r2 next should be nl.miek")
		t.Fail()
	}
	if r1.Prev().Key() != "nl.miek.a" {
		t.Logf("r1 prev should be nl.miek.a")
		t.Fail()
	}
	if r2.Prev().Key() != "nl.miek" {
		t.Logf("r2 prev should be nl.miek")
		t.Fail()
	}
	if r1.Next().Prev().Key() != r1.Key() {
		t.Logf("Next()->Prev() should map to me")
		t.Fail()
	}
}
func TestNextPrevEmpty(t *testing.T) {
	r := New()
	nxt := r.Next()
	prev := r.Prev()
	_ = nxt
	_ = prev
}

func ExampleFind() {
	r := New()

	r.Insert(strings.NewReader("tester"), nil)
	r.Insert(strings.NewReader("testering"), nil)
	r.Insert(strings.NewReader("te"), nil)
	r.Insert(strings.NewReader("testeringandmore"), nil)
	iter(r)
	// Output:
	// prefix te
	// prefix tester
	// prefix testering
	// prefix testeringandmore
}

func iter(r *Radix) {
	if r.Key() != "" {
		fmt.Printf("prefix %s\n", r.Key())
	}
	//	fmt.Println(len(r.children))
	for _, child := range r.children {
		iter(child)
	}
}

func BenchmarkFind(b *testing.B) {
	b.StopTimer()
	r := radixtree()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Find(strings.NewReader("tester"))
	}
}
