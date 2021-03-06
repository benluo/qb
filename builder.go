package qb

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	LDefault = iota
	// log query flag
	LQuery
	// log bindings flag
	LBindings
)

// NewBuilder generates a new builder struct
func NewBuilder(driver string) *Builder {
	return &Builder{
		query:    NewQuery(),
		adapter:  NewAdapter(driver),
		logger:   log.New(os.Stdout, "", 0),
		logFlags: LDefault,
	}
}

// Builder is a struct that holds an active query that it is used for building common sql queries
// it has all the common functions except multiple statements & table crudders
type Builder struct {
	query    *Query
	adapter  Adapter
	logger   *log.Logger
	logFlags int
}

// SetLogFlags sets the builder log flags
func (b *Builder) SetLogFlags(logFlags int) {
	b.logFlags = logFlags
}

// LogFlags returns the log flags
func (b *Builder) LogFlags() int {
	return b.logFlags
}

// SetEscaping sets the escaping parameter of current adapter
func (b *Builder) SetEscaping(escaping bool) {
	b.adapter.SetEscaping(escaping)
}

// Escaping functions returns if escaping is available
func (b *Builder) Escaping() bool {
	return b.adapter.Escaping()
}

// Adapter returns the active adapter of builder
func (b *Builder) Adapter() Adapter {
	return b.adapter
}

// Reset clears query bindings and its errors
func (b *Builder) Reset() {
	b.query = NewQuery()
	b.adapter.Reset()
}

// Query returns the active query and resets the query.
// The query clauses and returns the sql and bindings
func (b *Builder) Query() *Query {
	query := b.query
	b.Reset()
	if b.logFlags == LQuery || b.logFlags == (LQuery|LBindings) {
		b.logger.Printf("%s", query.SQL())
	}
	if b.logFlags == LBindings || b.logFlags == (LQuery|LBindings) {
		b.logger.Printf("%s", query.Bindings())
	}
	if b.logFlags != LDefault {
		b.logger.Println()
	}
	return query
}

// Insert generates an "insert into %s(%s)" statement
func (b *Builder) Insert(table string) *Builder {
	clause := fmt.Sprintf("INSERT INTO %s", b.adapter.Escape(table))
	b.query.AddClause(clause)
	return b
}

// Values generates "values(%s)" statement and add bindings for each value
func (b *Builder) Values(m map[string]interface{}) *Builder {
	keys := []string{}
	values := []interface{}{}
	for k, v := range m {
		keys = append(keys, b.adapter.Escape(k))
		values = append(values, v)
		b.query.AddBinding(v)
	}

	b.query.AddClause(fmt.Sprintf("(%s)", strings.Join(keys, ", ")))

	placeholders := []string{}

	for range values {
		placeholders = append(placeholders, b.adapter.Placeholder())
	}
	clause := fmt.Sprintf("VALUES (%s)", strings.Join(placeholders, ", "))
	b.query.AddClause(clause)
	return b
}

// Returning generates RETURNING statement for postgres only
// NOTE: Do not use it with sqlite, mysql or other drivers
func (b *Builder) Returning(cols ...string) *Builder {
	cols = b.adapter.EscapeAll(cols)
	clause := fmt.Sprintf("RETURNING %s", strings.Join(cols, ", "))
	b.query.AddClause(clause)
	return b
}

// Update generates "update %s" statement
func (b *Builder) Update(table string) *Builder {
	clause := fmt.Sprintf("UPDATE %s", b.adapter.Escape(table))
	b.query.AddClause(clause)
	return b
}

// Set generates "set a = placeholder" statement for each key a and add bindings for map value
func (b *Builder) Set(m map[string]interface{}) *Builder {
	updates := []string{}
	for k, v := range m {
		// check if aliasing exists
		if strings.Contains(k, ".") {
			kpieces := strings.Split(k, ".")
			k = fmt.Sprintf("%s.%s", kpieces[0], b.adapter.Escape(kpieces[1]))
		} else {
			k = b.adapter.Escape(k)
		}
		updates = append(updates, fmt.Sprintf("%s = %s", k, b.adapter.Placeholder()))
		b.query.AddBinding(v)
	}
	clause := fmt.Sprintf("SET %s", strings.Join(updates, ", "))
	b.query.AddClause(clause)
	return b
}

// Delete generates "delete" statement
func (b *Builder) Delete(table string) *Builder {
	b.query.AddClause(fmt.Sprintf("DELETE FROM %s", b.adapter.Escape(table)))
	return b
}

// Select generates "select %s" statement
func (b *Builder) Select(columns ...string) *Builder {
	clause := fmt.Sprintf("SELECT %s", strings.Join(columns, ", "))
	b.query.AddClause(clause)
	return b
}

// From generates "from %s" statement for each table name
func (b *Builder) From(tables ...string) *Builder {
	tbls := []string{}
	for _, v := range tables {
		tablePieces := strings.Split(v, " ")
		v = b.adapter.Escape(tablePieces[0])
		if len(tablePieces) > 1 {
			v = fmt.Sprintf("%s %s", v, tablePieces[1])
		}
		tbls = append(tbls, v)
	}
	b.query.AddClause(fmt.Sprintf("FROM %s", strings.Join(tbls, ", ")))
	return b
}

// InnerJoin generates "inner join %s on %s" statement for each expression
func (b *Builder) InnerJoin(table string, expressions ...string) *Builder {
	tablePieces := strings.Split(table, " ")

	v := b.adapter.Escape(tablePieces[0])
	if len(tablePieces) > 1 {
		v = fmt.Sprintf("%s %s", v, tablePieces[1])
	}
	b.query.AddClause(fmt.Sprintf("INNER JOIN %s ON %s", v, strings.Join(expressions, " ")))
	return b
}

// CrossJoin generates "cross join %s" statement for table
func (b *Builder) CrossJoin(table string) *Builder {
	tablePieces := strings.Split(table, " ")

	v := b.adapter.Escape(tablePieces[0])
	if len(tablePieces) > 1 {
		v = fmt.Sprintf("%s %s", v, tablePieces[1])
	}

	b.query.AddClause(fmt.Sprintf("CROSS JOIN %s", v))
	return b
}

// LeftOuterJoin generates "left outer join %s on %s" statement for each expression
func (b *Builder) LeftOuterJoin(table string, expressions ...string) *Builder {
	tablePieces := strings.Split(table, " ")

	v := b.adapter.Escape(tablePieces[0])
	if len(tablePieces) > 1 {
		v = fmt.Sprintf("%s %s", v, tablePieces[1])
	}

	b.query.AddClause(fmt.Sprintf("LEFT OUTER JOIN %s ON %s", v, strings.Join(expressions, " ")))
	return b
}

// RightOuterJoin generates "right outer join %s on %s" statement for each expression
func (b *Builder) RightOuterJoin(table string, expressions ...string) *Builder {
	tablePieces := strings.Split(table, " ")

	v := b.adapter.Escape(tablePieces[0])
	if len(tablePieces) > 1 {
		v = fmt.Sprintf("%s %s", v, tablePieces[1])
	}
	b.query.AddClause(fmt.Sprintf("RIGHT OUTER JOIN %s ON %s", v, strings.Join(expressions, " ")))
	return b
}

// FullOuterJoin generates "full outer join %s on %s" for each expression
func (b *Builder) FullOuterJoin(table string, expressions ...string) *Builder {
	tablePieces := strings.Split(table, " ")

	v := b.adapter.Escape(tablePieces[0])
	if len(tablePieces) > 1 {
		v = fmt.Sprintf("%s %s", v, tablePieces[1])
	}
	b.query.AddClause(fmt.Sprintf("FULL OUTER JOIN %s ON %s", v, strings.Join(expressions, " ")))
	return b
}

// Where generates "where %s" for the expression and adds bindings for each value
func (b *Builder) Where(expression string, bindings ...interface{}) *Builder {
	if expression == "" {
		return b
	}
	b.query.AddClause(fmt.Sprintf("WHERE %s", expression))
	b.query.AddBinding(bindings...)
	return b
}

// OrderBy generates "order by %s" for each expression
func (b *Builder) OrderBy(expressions ...string) *Builder {
	b.query.AddClause(fmt.Sprintf("ORDER BY %s", strings.Join(expressions, ", ")))
	return b
}

// GroupBy generates "group by %s" for each column
func (b *Builder) GroupBy(columns ...string) *Builder {
	b.query.AddClause(fmt.Sprintf("GROUP BY %s", strings.Join(columns, ", ")))
	return b
}

// Having generates "having %s" for each expression
func (b *Builder) Having(expressions ...string) *Builder {
	b.query.AddClause(fmt.Sprintf("HAVING %s", strings.Join(expressions, ", ")))
	return b
}

// Limit generates limit %d offset %d for offset and count
func (b *Builder) Limit(offset int, count int) *Builder {
	b.query.AddClause(fmt.Sprintf("LIMIT %d OFFSET %d", count, offset))
	return b
}

// aggregates

// Avg function generates "avg(%s)" statement for column
func (b *Builder) Avg(column string) string {
	return fmt.Sprintf("AVG(%s)", b.adapter.Escape(column))
}

// Count function generates "count(%s)" statement for column
func (b *Builder) Count(column string) string {
	return fmt.Sprintf("COUNT(%s)", b.adapter.Escape(column))
}

// Sum function generates "sum(%s)" statement for column
func (b *Builder) Sum(column string) string {
	return fmt.Sprintf("SUM(%s)", b.adapter.Escape(column))
}

// Min function generates "min(%s)" statement for column
func (b *Builder) Min(column string) string {
	return fmt.Sprintf("MIN(%s)", b.adapter.Escape(column))
}

// Max function generates "max(%s)" statement for column
func (b *Builder) Max(column string) string {
	return fmt.Sprintf("MAX(%s)", b.adapter.Escape(column))
}

// expressions

// NotIn function generates "%s not in (%s)" for key and adds bindings for each value
func (b *Builder) NotIn(key string, values ...interface{}) string {
	b.query.AddBinding(values...)
	return fmt.Sprintf("%s NOT IN (%s)", b.adapter.Escape(key), strings.Join(b.adapter.Placeholders(values...), ","))
}

// In function generates "%s in (%s)" for key and adds bindings for each value
func (b *Builder) In(key string, values ...interface{}) string {
	b.query.AddBinding(values...)
	return fmt.Sprintf("%s IN (%s)", b.adapter.Escape(key), strings.Join(b.adapter.Placeholders(values...), ","))
}

// NotEq function generates "%s != placeholder" for key and adds binding for value
func (b *Builder) NotEq(key string, value interface{}) string {
	b.query.AddBinding(value)
	return fmt.Sprintf("%s != %s", b.adapter.Escape(key), b.adapter.Placeholder())
}

// Eq function generates "%s = placeholder" for key and adds binding for value
func (b *Builder) Eq(key string, value interface{}) string {
	b.query.AddBinding(value)
	return fmt.Sprintf("%s = %s", b.adapter.Escape(key), b.adapter.Placeholder())
}

// Gt function generates "%s > placeholder" for key and adds binding for value
func (b *Builder) Gt(key string, value interface{}) string {
	b.query.AddBinding(value)
	return fmt.Sprintf("%s > %s", b.adapter.Escape(key), b.adapter.Placeholder())
}

// Gte function generates "%s >= placeholder" for key and adds binding for value
func (b *Builder) Gte(key string, value interface{}) string {
	b.query.AddBinding(value)
	return fmt.Sprintf("%s >= %s", b.adapter.Escape(key), b.adapter.Placeholder())
}

// St function generates "%s < placeholder" for key and adds binding for value
func (b *Builder) St(key string, value interface{}) string {
	b.query.AddBinding(value)
	return fmt.Sprintf("%s < %s", b.adapter.Escape(key), b.adapter.Placeholder())
}

// Ste function generates "%s <= placeholder" for key and adds binding for value
func (b *Builder) Ste(key string, value interface{}) string {
	b.query.AddBinding(value)
	return fmt.Sprintf("%s <= %s", b.adapter.Escape(key), b.adapter.Placeholder())
}

// And function generates " AND " between any number of expressions
func (b *Builder) And(expressions ...string) string {
	if len(expressions) == 0 {
		return ""
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " AND "))
}

// Or function generates " OR " between any number of expressions
func (b *Builder) Or(expressions ...string) string {
	return strings.Join(expressions, " OR ")
}

// CreateTable generates generic CREATE TABLE statement
func (b *Builder) CreateTable(table string, fields []string, constraints []string) *Builder {
	b.query.AddClause(fmt.Sprintf("CREATE TABLE %s(", b.adapter.Escape(table)))

	for k, f := range fields {
		clause := fmt.Sprintf("\t%s", f)
		if len(fields)-1 > k || len(constraints) > 0 {
			clause += ","
		}
		b.query.AddClause(clause)
	}

	for k, c := range constraints {
		constraint := fmt.Sprintf("\t%s", c)
		if len(constraints)-1 > k {
			constraint += ","
		}
		b.query.AddClause(fmt.Sprintf("%s", constraint))
	}

	b.query.AddClause(")")
	return b
}

// AlterTable generates generic ALTER TABLE statement
func (b *Builder) AlterTable(table string) *Builder {
	b.query.AddClause(fmt.Sprintf("ALTER TABLE %s", table))
	return b
}

// DropTable generates generic DROP TABLE statement
func (b *Builder) DropTable(table string) *Builder {
	b.query.AddClause(fmt.Sprintf("DROP TABLE %s", b.adapter.Escape(table)))
	return b
}

// Add generates generic ADD COLUMN statement
func (b *Builder) Add(colName string, colType string) *Builder {
	b.query.AddClause(fmt.Sprintf("ADD %s %s", colName, colType))
	return b
}

// Drop generates generic DROP COLUMN statement
func (b *Builder) Drop(colName string) *Builder {
	b.query.AddClause(fmt.Sprintf("DROP %s", colName))
	return b
}

// CreateIndex generates an index on columns
func (b *Builder) CreateIndex(indexName string, tableName string, columns ...string) *Builder {
	b.query.AddClause(fmt.Sprintf("CREATE INDEX %s ON %s(%s)", indexName, tableName, strings.Join(b.adapter.EscapeAll(columns), ",")))
	return b
}
