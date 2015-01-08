package sync

import "fmt"

type Statement struct {
	SQL   string
	Owner string
	Notes string
}

func Statementf(s string, p ...interface{}) *Statement {
	return &Statement{
		SQL: fmt.Sprintf(s, p...),
	}
}

func (s *Statement) Notef(str string, p ...interface{}) {
	s.Notes = fmt.Sprintf(str, p...)
}
