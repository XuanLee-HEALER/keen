package fp_test

import (
	"math/rand"
	"strconv"
	"testing"

	"gitea.fcdm.top/lixuan/keen/fp"
)

func TestFuncMapTo(t *testing.T) {
	is := []int{1, 2, 3}
	res := fp.Map[int, string](is, func(e int) string { return strconv.Itoa(e) + " alpha" })
	strs := []string{"1 alpha", "2 alpha", "3 alpha"}
	for i, e := range res {
		if e != strs[i] {
			t.FailNow()
		}
	}
}

func TestFuncContains(t *testing.T) {
	is := []string{"good", "morning"}

	if fp.Contains[string](is, "afternoon") {
		t.FailNow()
	}

	if !fp.Contains[string](is, "good") {
		t.FailNow()
	}
}

func TestFuncContainsAny(t *testing.T) {
	type TA struct {
		AA int
		BB map[string]string
	}

	tas := []TA{
		{1, map[string]string{"good": "ta"}},
	}

	if !fp.ContainsAny[TA](tas, TA{AA: 1}, func(a1, a2 TA) bool { return true }) {
		t.FailNow()
	}

	if !fp.ContainsAny[TA](tas, TA{AA: 1}, func(a1, a2 TA) bool { return a1.AA == a2.AA }) {
		t.FailNow()
	}

	if fp.ContainsAny[TA](tas, TA{AA: 1}, func(a1, a2 TA) bool { return a1.BB["good"] == a2.BB["good"] }) {
		t.FailNow()
	}
}

func TestFuncFilter(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8}
	a = fp.Filter[int](a, func(e int) bool { return e >= 5 })
	res := []int{5, 6, 7, 8}
	for i, e := range a {
		if e != res[i] {
			t.FailNow()
		}
	}
}

func TestFuncReduce(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	res, ok := fp.Reduce[int](a, func(e1, e2 int) int { return e1 + e2 })
	if !ok || res != 55 {
		t.FailNow()
	}

	b := []int{}
	_, ok = fp.Reduce[int](b, func(e1, e2 int) int { return e1 + e2 })
	if ok {
		t.FailNow()
	}

	c := []int{1}
	res, ok = fp.Reduce[int](c, func(e1, e2 int) int { return e1 + e2 })
	if !ok || res != 1 {
		t.FailNow()
	}
}

func TestFuncForeach(t *testing.T) {
	a := []string{"good", "morning", "bob"}

	fp.Foreach[string](a, func(e string) { println(e) })
}

func TestFuncGroup(t *testing.T) {
	type student struct {
		name  string
		age   int
		class string
	}

	names := []string{"tom", "jerry", "bob", "curry", "james"}
	classes := []string{"class1", "class2", "class3"}

	stus := make([]student, 0, 10)
	for i := 0; i < 10; i++ {
		nidx := rand.Int31n(5)
		cidx := rand.Int31n(3)
		age := rand.Int31n(20) + 1

		stus = append(stus, student{names[nidx], int(age), classes[cidx]})
	}

	res := fp.Group[student](stus, func(e1, e2 student) bool { return e1.name == e2.name && e1.class == e2.class }, func(e1, e2 student) student { return student{e1.name, e1.age + e2.age, e1.class} })

	for _, r := range res {
		t.Log(r)
	}
}

func TestFuncAccumulate(t *testing.T) {
	type ex struct {
		A int
	}

	examples := []ex{{1}, {2}, {3}}
	r, ok := fp.Accumulate[ex, int](examples, func(t int, e ex) int { return t + e.A }, func(e ex) int { return e.A })

	if !ok || r != 6 {
		t.FailNow()
	}
}
