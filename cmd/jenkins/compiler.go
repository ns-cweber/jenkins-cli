package main

import (
	"github.com/ns-cweber/jenkins-cli/lib/boolsearch"
	"github.com/ns-cweber/jenkins-cli/lib/jenkins"
)

func get(b jenkins.Build, s string) string {
	switch s {
	case "number":
		return b.Number
	case "worker":
		return b.BuiltOn
	case "status":
		return string(b.Result)
	default:
		for _, action := range b.Actions {
			if action.Class == jenkins.ActionClassParameters {
				return action.Parameters.Get(s)
			}
		}
		return ""
	}
}

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

	if conj.Op == boolsearch.ConjOpAnd {
		c.f = func(b jenkins.Build) bool {
			return left.f(b) && right.f(b)
		}
		return
	}

	if conj.Op == boolsearch.ConjOpOr {
		c.f = func(b jenkins.Build) bool {
			return left.f(b) || right.f(b)
		}
		return
	}
	panic("Invalid conjugation operator: '" + string(conj.Op) + "'")
}

func (c *compiler) VisitEmpty(e boolsearch.Empty) {
	c.f = func(b jenkins.Build) bool { return true }
}
