package main

import (
	jenkins "github.com/ns-cweber/jenkins-cli"
	"github.com/ns-cweber/jenkins-cli/boolsearch"
)

type compiler struct {
	f func(b jenkins.Build) bool
}

func (c *compiler) VisitComparison(cmp boolsearch.Comparison) {
	if cmp.Op == boolsearch.CmpOpEq {
		c.f = func(b jenkins.Build) bool {
			return get(b, cmp.Left) == cmp.Right
		}
		return
	}
	if cmp.Op == boolsearch.CmpOpNe {
		c.f = func(b jenkins.Build) bool {
			return get(b, cmp.Left) != cmp.Right
		}
		return
	}
	panic("Invalid comparison operator: '" + string(cmp.Op) + "'")
}

func (c *compiler) VisitConjugation(conj boolsearch.Conjugation) {
	var left compiler
	var right compiler
	conj.Left.Visit(&left)
	conj.Right.Visit(&right)

	c.f = func(b jenkins.Build) bool { return left.f(b) == right.f(b) }
}

func (c *compiler) VisitEmpty(e boolsearch.Empty) {
	c.f = func(b jenkins.Build) bool { return true }
}
