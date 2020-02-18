package toggle

import (
	"encoding/json"
	"fmt"
)

type ValueType int
type ConditionOp int
type FieldOp int

const (
	IntValue ValueType = iota
	FloatValue
	StringValue
)

const (
	AndOp ConditionOp = iota
	OrOp
)

const (
	EqOp FieldOp = iota
	NeOp
	LtOp
	GtOp

	ServiceNameValue = "serviceName"
)

type ConditionValue struct {
	Name  string      `json:"name"`
	Type  ValueType   `json:"type"`
	Value interface{} `json:"value"`
}

type ConditionField struct {
	ConditionValue
	Op FieldOp `json:"op"`
}

type Condition struct {
	Op         ConditionOp      `json:"op,omitempty"`
	Conditions []Condition      `json:"conditions,omitempty"`
	Fields     []ConditionField `json:"fields,omitempty"`
}

type matcher interface {
	match(values []ConditionValue) bool
}

func (v *ConditionField) UnmarshalJSON(b []byte) error {
	type value ConditionField
	var val value
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}

	if val.Type == IntValue {
		val.Value = int64(val.Value.(float64))
	}

	*v = ConditionField(val)

	return nil
}

// Validate checks if the value type and its underlying type are consistent
func (v ConditionValue) Validate() error {
	switch v.Type {
	case IntValue:
		if _, ok := v.Value.(int64); !ok {
			return fmt.Errorf("invalid int type for value %T", v.Value)
		}
	case FloatValue:
		if _, ok := v.Value.(float64); !ok {
			return fmt.Errorf("invalid float64 type for value %T", v.Value)
		}
	case StringValue:
		if _, ok := v.Value.(string); !ok {
			return fmt.Errorf("invalid string type for value %T", v.Value)
		}
	default:
		return fmt.Errorf("invalid type %v", v.Type)

	}
	return nil
}

// Match checks if the given condition values match the condition logic
func (c Condition) Match(values []ConditionValue) bool {
	return c.match(values)
}

func (c Condition) hasMatchers() bool {
	return len(c.Conditions) > 0 || len(c.Fields) > 0
}

func (c Condition) match(values []ConditionValue) bool {
	if !c.hasMatchers() {
		return true
	}

	matchers := make([]matcher, 0, len(c.Conditions)+len(c.Fields))
	for _, m := range c.Conditions {
		matchers = append(matchers, m)
	}
	for _, m := range c.Fields {
		matchers = append(matchers, m)
	}

	if len(values) == 0 {
		return false
	}

	var match bool
	for i, m := range matchers {
		res := m.match(values)

		switch c.Op {
		case OrOp:
			match = match || res
		default:
			if i == 0 {
				match = true
			}
			match = match && res

			if !match {
				break
			}
		}
	}

	return match
}

func (f ConditionField) match(values []ConditionValue) bool {
	for _, v := range values {
		if v.Name != f.Name || v.Type != f.Type {
			continue
		}

		switch f.Op {
		case NeOp:
			return f.ne(v)
		case LtOp:
			return f.lt(v)
		case GtOp:
			return f.gt(v)
		default:
			return f.eq(v)
		}
	}

	return false
}

func (f ConditionField) eq(v ConditionValue) bool {
	return f.Value == v.Value
}

func (f ConditionField) ne(v ConditionValue) bool {
	return f.Value != v.Value
}

func (f ConditionField) lt(v ConditionValue) bool {
	switch f.Type {
	case IntValue:
		return f.Value.(int64) < v.Value.(int64)
	case FloatValue:
		return f.Value.(float64) < v.Value.(float64)
	case StringValue:
		return f.Value.(string) < v.Value.(string)
	}
	return false
}

func (f ConditionField) gt(v ConditionValue) bool {
	switch f.Type {
	case IntValue:
		return f.Value.(int64) > v.Value.(int64)
	case FloatValue:
		return f.Value.(float64) > v.Value.(float64)
	case StringValue:
		return f.Value.(string) > v.Value.(string)
	}
	return false
}
