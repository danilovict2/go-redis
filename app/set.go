package main

import (
	"container/heap"
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

func (s *Set) FindByIndex(member string) int {
	for i, elem := range *s {
		if elem.Member == member {
			return i
		}
	}

	return -1
}

func (s *Set) FindByRank(member string) int {
	idx := -1
	elems := make([]SetMember, 0)

	for i := 0; len(*s) > 0; i++ {
		elem := heap.Pop(s).(SetMember)
		elems = append(elems, elem)

		if elem.Member == member {
			idx = i
			break
		}
	}

	for _, elem := range elems {
		heap.Push(s, elem)
	}

	return idx
}
