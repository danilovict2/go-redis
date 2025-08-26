package main

import (
	"math"
)

type Set []SetMember

type SetMember struct {
	Member string
	Score  float64
}

func (s Set) Len() int {
	return len(s)
}

func (s Set) Less(i, j int) bool {
	const tolerance = 1e-9

	if math.Abs(s[i].Score-s[j].Score) <= tolerance {
		return s[i].Member < s[j].Member
	}

	return s[i].Score < s[j].Score
}

func (s Set) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s *Set) Push(x any) {
	item, ok := x.(SetMember)
	if !ok {
		return
	}

	*s = append(*s, item)
}

func (s *Set) Pop() any {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

func (s Set) Find(member string) int {
	for i, m := range s {
		if m.Member == member {
			return i
		}
	}

	return -1
}
