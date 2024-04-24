package query

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

type SQLQueryBuilder string

func NewExp(key string, assignment string, value interface{}) SQLExpression {
	return SQLExpression{Key: key, Exp: assignment, Value: value}
}

func And(component ...SQLQueryBuilder) SQLQueryBuilder {
	return constructQuery("AND", component...)
}

func Or(component ...SQLQueryBuilder) SQLQueryBuilder {
	return constructQuery("OR", component...)
}

func NotEq(component ...interface{}) SQLQueryBuilder {
	return getOperationExpression("OR", component...)
}

func Eq(component ...interface{}) SQLQueryBuilder {
	return getOperationExpression("AND", component...)
}

func constructQuery(operation string, queries ...SQLQueryBuilder) SQLQueryBuilder {
	if len(queries) == 0 {
		return ""
	}

	str := make([]string, len(queries))
	for i, query := range queries {
		str[i] = string(query)
	}

	return SQLQueryBuilder(strings.Join(str, " "+operation+" "))
}

func getOperationExpression(operation string, expressions ...interface{}) SQLQueryBuilder {
	if len(expressions) == 0 {
		return ""
	}

	if len(expressions) == 1 {
		return expressionToString(expressions[0])
	} else {
		clauses := make([]string, 0)
		for _, v := range expressions {
			value := expressionToString(v)
			if value != "" {
				clauses = append(clauses, ""+string(value)+"")
			}
		}

		if len(clauses) > 0 {
			return SQLQueryBuilder("( " + strings.Join(clauses, " "+operation+" ") + " )")
		}
	}

	return ""
}

func (qb SQLQueryBuilder) String() string {
	return string(qb)
}

type SQLExpression struct {
	Key   string
	Exp   string
	Value interface{}
}

func (e SQLExpression) String() string {

	switch e.Value.(type) {
	case int, int16, int32, int64:
		val := strconv.Itoa(e.Value.(int))
		clause := e.Key + e.Exp + e.getReplaceExp()
		return fmt.Sprintf(clause, val)

	default:
		if strings.TrimSpace(e.Value.(string)) == "" {
			return ""
		} else {
			e.Value = template.HTMLEscapeString(e.Value.(string))
			clause := e.Key + e.Exp + e.getReplaceExp()
			val := fmt.Sprintf(clause, e.Value)
			return val
		}
	}
}

func (e SQLExpression) getReplaceExp() string {
	switch e.Value.(type) {
	case int, int64, int32, int16:
		return "%s"
	default:
		return "'%s'"
	}
}

func expressionToString(c interface{}) SQLQueryBuilder {
	switch v := c.(type) {
	case SQLQueryBuilder:
		return v
	case SQLExpression:
		return SQLQueryBuilder(v.String())
	case string, *string:
		return SQLQueryBuilder(c.(string))
	default:
		return SQLQueryBuilder(fmt.Sprintf("%v", v))
	}
}
