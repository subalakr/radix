package radix

import (
	"fmt"
	"testing"
)

func printit(r *Radix, level int) {
	for i := 0; i < level; i++ {
		fmt.Print("\t")
	}
	fmt.Printf("%p '%v'  value: '%v'    parent %p\n", r, r.key, r.Value, r.parent)
	for _, child := range r.children {
		printit(child, level+1)
	}
}

func radixtree() *Radix {
	r := New()
	r.Insert("test", "a")
	r.Insert("tester", "a")
	r.Insert("team", "a")
	r.Insert("te", "a")
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

func TestPrint(t *testing.T) {
	// TODO(mg): fix
}

func TestNext(t *testing.T) {
	r := radixtree()
	f := r.Find("team")
	for l := f.next(); l != nil; l = l.next() {
//		println(l.String())
//		println(l.Key())
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
	r.Insert("test", nil)
	r.Insert("slow", nil)
	r.Insert("water", nil)
	r.Insert("tester", nil)
	r.Insert("testering", nil)
	r.Insert("rewater", nil)
	r.Insert("waterrat", nil)
	if !validate(r) {
		t.Log("Tree does not validate")
		t.Fail()
	}
}

func TestRemove(t *testing.T) {
	r := New()
	r.Insert("test", "aa")
	r.Insert("slow", "bb")

	if k := r.Remove("slow").Value; k != "bb" {
		t.Log("should be bb", k)
		t.Fail()
	}

	if r.Remove("slow") != nil {
		t.Log("should be nil")
		t.Fail()
	}
	r.Insert("test", "aa")
	r.Insert("tester", "aa")
	r.Insert("testering", "aa")
	r.Find("tester").Remove("test")
}

func TestCommonPrefix(t *testing.T) {
	r := radixtree()
	f := r.Find("tester")
	t.Logf("%s %+v\n", f.key, f.Keys())
}

func ExampleFind() {
	r := New()
	r.Insert("tester", nil)
	r.Insert("testering", nil)
	r.Insert("te", nil)
	r.Insert("testeringandmore", nil)
	iter(r.Find("tester"))
	// Output:
	// prefix tester
	// prefix testering
	// prefix testeringandmore
}

func iter(r *Radix) {
	fmt.Printf("prefix %s\n", r.Key())
	for _, child := range r.children {
		iter(child)
	}
}

func BenchmarkFind(b *testing.B) {
	b.StopTimer()
	r := radixtree()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_ = r.Find("tester")
	}
}
