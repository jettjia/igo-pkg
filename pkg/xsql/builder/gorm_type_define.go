package builder

// PageData pagination
type PageData struct {
	PageNum     int   `json:"page_num"`     // page number
	PageSize    int   `json:"page_size"`    // the number of rows displayed per page
	TotalNumber int64 `json:"total_number"` // how many in total
	TotalPage   int64 `json:"total_page"`   // how many pages in total
}

// SortData sort
type SortData struct {
	Sort      string `json:"sort"`      // sort fields
	Direction string `json:"direction"` // asc;desc
}

// Query search
type Query struct {
	Key      string      `json:"key"`      // the key to search for a keyword
	Value    interface{} `json:"value"`    // search for the value of the keyword
	Operator Operator    `json:"operator"` // judging conditions
}

type Operator int

const (
	Operator_opEq          Operator = 0  // =
	Operator_opNe1         Operator = 1  // !=
	Operator_opNe2         Operator = 2  // <>
	Operator_opIn          Operator = 3  // in
	Operator_opNotIn       Operator = 4  // not in
	Operator_opGt          Operator = 5  // >
	Operator_opGte         Operator = 6  // >=
	Operator_opLt          Operator = 7  // <
	Operator_opLte         Operator = 8  // <=
	Operator_opLike        Operator = 9  // like, it needs to be passed on yourself %
	Operator_opLikePercent Operator = 92 // like，automatically bring on both sides %
	Operator_opNotLike     Operator = 10 // not like
	Operator_opBetween     Operator = 11 // between
	Operator_opNotBetween  Operator = 12 // not between
	Operator_opNull        Operator = 13 // null
)

var OperatorMap = map[Operator]string{
	Operator_opEq:          OpEq,
	Operator_opNe1:         OpNe1,
	Operator_opNe2:         OpNe2,
	Operator_opIn:          OpIn,
	Operator_opNotIn:       OpNotIn,
	Operator_opGt:          OpGt,
	Operator_opGte:         OpGte,
	Operator_opLt:          OpLt,
	Operator_opLte:         OpLte,
	Operator_opLike:        OpLike,
	Operator_opLikePercent: OpLikePercent,
	Operator_opNotLike:     OpNotLike,
	Operator_opBetween:     OpBetween,
	Operator_opNotBetween:  OpNotBetween,
	Operator_opNull:        OpNull,
}
