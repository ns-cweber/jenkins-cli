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

func (e Empty) String() string {
	return "()"
}

type Expression interface {
	Visit(Visitor)
	String() string
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

func (c Comparison) String() string {
	return c.Left + " " + string(c.Op) + " " + c.Right
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

func (c Conjugation) String() string {
	return c.Left.String() + " " + string(c.Op) + " " + c.Right.String()
}
