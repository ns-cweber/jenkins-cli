package boolsearch

type Visitor interface {
	VisitComparison(cmp Comparison)
	VisitConjugation(conj Conjugation)
	VisitEmpty(e Empty)
}

type Empty struct{}

func (e Empty) Visit(visitor Visitor) {
	visitor.VisitEmpty(e)
}

type Expression interface {
	Visit(Visitor)
}

type CmpOp string

const (
	CmpOpEq CmpOp = "="
	CmpOpNe CmpOp = "!="
)

type Comparison struct {
	Left, Right string
	Op          CmpOp
}

func (c Comparison) Visit(visitor Visitor) {
	visitor.VisitComparison(c)
}

type ConjOp string

const (
	ConjOpAnd ConjOp = "&"
	ConjOpOr  ConjOp = "|"
)

type Conjugation struct {
	Op    ConjOp
	Left  Expression
	Right Expression
}

func (c Conjugation) Visit(visitor Visitor) {
	visitor.VisitConjugation(c)
}
