// Code generated by "stringer -type ValueType,ConditionOp,FieldOp -linecomment ./toggle"; DO NOT EDIT.

package toggle

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[IntType-0]
	_ = x[FloatType-1]
	_ = x[StringType-2]
}

const _ValueType_name = "intfloatstring"

var _ValueType_index = [...]uint8{0, 3, 8, 14}

func (i ValueType) String() string {
	if i < 0 || i >= ValueType(len(_ValueType_index)-1) {
		return "ValueType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ValueType_name[_ValueType_index[i]:_ValueType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AndOp-0]
	_ = x[OrOp-1]
}

const _ConditionOp_name = "&&||"

var _ConditionOp_index = [...]uint8{0, 2, 4}

func (i ConditionOp) String() string {
	if i < 0 || i >= ConditionOp(len(_ConditionOp_index)-1) {
		return "ConditionOp(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ConditionOp_name[_ConditionOp_index[i]:_ConditionOp_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EqOp-0]
	_ = x[NeOp-1]
	_ = x[LtOp-2]
	_ = x[GtOp-3]
}

const _FieldOp_name = "=!=<>"

var _FieldOp_index = [...]uint8{0, 1, 3, 4, 5}

func (i FieldOp) String() string {
	if i < 0 || i >= FieldOp(len(_FieldOp_index)-1) {
		return "FieldOp(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FieldOp_name[_FieldOp_index[i]:_FieldOp_index[i+1]]
}
