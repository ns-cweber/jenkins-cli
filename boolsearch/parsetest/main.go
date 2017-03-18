package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ns-cweber/jenkins-cli/boolsearch"
)

type MapVisitor map[string]string

func (mv MapVisitor) VisitComparison(cmp boolsearch.Comparison) bool {
	if cmp.Op == boolsearch.CmpOpEq {
		value, found := mv[cmp.Left]
		if !found {
			panic("Variable not found: " + cmp.Left)
		}
		return value == cmp.Right
	} else if cmp.Op == boolsearch.CmpOpNe {
		value, found := mv[cmp.Left]
		if !found {
			panic("Variable not found: " + cmp.Left)
		}
		return value != cmp.Right
	}
	panic("Invalid comparison operator: " + string(cmp.Op))
}

func main() {
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(" > ")
		if !s.Scan() {
			fmt.Println("BYE!")
			break
		}

		toks, err := boolsearch.Tokenize(bufio.NewReader(bytes.NewReader(s.Bytes())))
		if err != nil && err != io.EOF {
			log.Fatal("TOKENIZE:", err)
		}

		expr, err := boolsearch.ParseTokens(toks)
		if err != nil {
			log.Fatal("PARSE:", err)
		}

		fmt.Println(expr.Visit(MapVisitor{"ab": "1234", "cd": "4567"}))

		// data, err := json.MarshalIndent(expr, "", "    ")
		// if err != nil {
		// 	log.Fatal("MARSHAL:", err)
		// }
		// fmt.Printf("%s\n", data)
	}
}
