package toggle_test

import (
	"testing"

	"github.com/globusdigital/feature-toggles/toggle"
)

func TestCondition_Match(t *testing.T) {
	tests := []struct {
		name   string
		c      toggle.Condition
		values []toggle.ConditionValue
		want   bool
	}{
		{name: "empty", want: true},

		{name: "no values", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}},
		}}},

		{name: "mismatched types", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatType, Value: int64(10)},
		}},

		{name: "svc1 and svc2", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "svc1"}},
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue + "2", Type: toggle.StringType, Value: "svc2"}},
		}}, values: []toggle.ConditionValue{
			{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "svc1"},
			{Name: toggle.ServiceNameValue + "2", Type: toggle.StringType, Value: "svc3"},
		}},

		{name: "svc1 and svc3", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "svc1"}},
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue + "2", Type: toggle.StringType, Value: "svc3"}},
		}}, values: []toggle.ConditionValue{
			{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "svc1"},
			{Name: toggle.ServiceNameValue + "2", Type: toggle.StringType, Value: "svc3"},
		}, want: true},

		{name: "svc1 or svc2", c: toggle.Condition{Op: toggle.OrOp, Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "svc1"}},
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue + "2", Type: toggle.StringType, Value: "svc2"}},
		}}, values: []toggle.ConditionValue{
			{Name: toggle.ServiceNameValue, Type: toggle.StringType, Value: "svc1"},
			{Name: toggle.ServiceNameValue + "2", Type: toggle.StringType, Value: "svc3"},
		}, want: true},

		{name: "10 < int - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntType, Value: int64(20)},
		}, want: true},

		{name: "10 < int", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntType, Value: int64(2)},
		}},

		{name: "10 < float64 - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatType, Value: float64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatType, Value: float64(20)},
		}, want: true},

		{name: "10 < float64", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatType, Value: float64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatType, Value: float64(2)},
		}},

		{name: "10 < bool - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.BoolType, Value: true}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.BoolType, Value: true},
		}},

		{name: "10 < bool", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.BoolType, Value: true}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.BoolType, Value: false},
		}},

		{name: "10 < string - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringType, Value: "10"}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringType, Value: "20"},
		}, want: true},

		{name: "10 < string", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringType, Value: "10"}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringType, Value: "1"},
		}},

		{name: "10 > int - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntType, Value: int64(2)},
		}, want: true},

		{name: "10 > int", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntType, Value: int64(20)},
		}},

		{name: "10 > bool - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.BoolType, Value: true}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.BoolType, Value: true},
		}},

		{name: "10 > bool", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.BoolType, Value: true}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.BoolType, Value: false},
		}},

		{name: "10 > float64 - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatType, Value: float64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatType, Value: float64(2)},
		}, want: true},

		{name: "10 > float64", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatType, Value: float64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatType, Value: float64(20)},
		}},

		{name: "10 > string - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringType, Value: "10"}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringType, Value: "1"},
		}, want: true},

		{name: "10 > string", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringType, Value: "10"}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringType, Value: "20"},
		}},

		{name: "10 != int - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntType, Value: int64(2)},
		}, want: true},

		{name: "10 != int", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntType, Value: int64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntType, Value: int64(10)},
		}},

		{name: "10 != bool - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.BoolType, Value: true}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.BoolType, Value: false},
		}, want: true},

		{name: "10 != bool", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.BoolType, Value: true}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.BoolType, Value: true},
		}},

		{name: "10 != float64 - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatType, Value: float64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatType, Value: float64(2)},
		}, want: true},

		{name: "10 != float64", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatType, Value: float64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatType, Value: float64(10)},
		}},

		{name: "10 != string - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringType, Value: "10"}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringType, Value: "2"},
		}, want: true},

		{name: "10 != string", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringType, Value: "10"}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringType, Value: "10"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Match(tt.values); got != tt.want {
				t.Errorf("Condition.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionValue_Validate(t *testing.T) {
	type fields struct {
		Name  string
		Type  toggle.ValueType
		Value interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "empty data", wantErr: true},
		{name: "invalid type", fields: fields{Type: 55050}, wantErr: true},
		{name: "invalid int type", fields: fields{Type: toggle.IntType, Value: 4.13}, wantErr: true},
		{name: "valid int type", fields: fields{Type: toggle.IntType, Value: int64(43)}},
		{name: "invalid float type", fields: fields{Type: toggle.FloatType, Value: 43}, wantErr: true},
		{name: "valid float type", fields: fields{Type: toggle.FloatType, Value: 43.5}},
		{name: "invalid bool type", fields: fields{Type: toggle.BoolType, Value: 4.13}, wantErr: true},
		{name: "valid bool type", fields: fields{Type: toggle.BoolType, Value: true}},
		{name: "invalid string type", fields: fields{Type: toggle.StringType, Value: 43}, wantErr: true},
		{name: "valid string type", fields: fields{Type: toggle.StringType, Value: "43.5"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := toggle.ConditionValue{
				Name:  tt.fields.Name,
				Type:  tt.fields.Type,
				Value: tt.fields.Value,
			}
			if err := v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ConditionValue.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCondition_Validate(t *testing.T) {
	type fields struct {
		Op         toggle.ConditionOp
		Conditions []toggle.Condition
		Fields     []toggle.ConditionField
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "empty data"},
		{name: "invalid field", fields: fields{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Type: toggle.IntType, Value: int64(42)}},
			{ConditionValue: toggle.ConditionValue{Type: toggle.IntType, Value: float64(42)}},
		}}, wantErr: true},
		{name: "invalid condition", fields: fields{Conditions: []toggle.Condition{
			{Fields: []toggle.ConditionField{{ConditionValue: toggle.ConditionValue{Type: toggle.IntType, Value: int64(50)}}}},
			{Fields: []toggle.ConditionField{{ConditionValue: toggle.ConditionValue{Type: toggle.IntType, Value: float64(50)}}}},
		}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := toggle.Condition{
				Op:         tt.fields.Op,
				Conditions: tt.fields.Conditions,
				Fields:     tt.fields.Fields,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Condition.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCondition_String(t *testing.T) {
	type fields struct {
		Op         toggle.ConditionOp
		Conditions []toggle.Condition
		Fields     []toggle.ConditionField
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "complete", fields: fields{Op: toggle.AndOp, Conditions: []toggle.Condition{
			{Fields: []toggle.ConditionField{
				{ConditionValue: toggle.ConditionValue{Name: "userID", Type: toggle.IntType, Value: int64(50)}, Op: toggle.NeOp},
				{ConditionValue: toggle.ConditionValue{Name: "userGroup", Type: toggle.StringType, Value: "some value"}},
			}, Op: toggle.AndOp},
			{Fields: []toggle.ConditionField{
				{ConditionValue: toggle.ConditionValue{Name: "accountLimit", Type: toggle.FloatType, Value: 20.0}, Op: toggle.LtOp},
				{ConditionValue: toggle.ConditionValue{Name: "purchases", Type: toggle.IntType, Value: int64(10)}},
			}, Op: toggle.OrOp},
		}, Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "time", Type: toggle.FloatType, Value: 52.0}, Op: toggle.GtOp},
		}}, want: "((userID != int(50) && userGroup = string(some value)) && (accountLimit < float(20) || purchases = int(10)) && time > float(52))"},

		{name: "complete", fields: fields{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "invalid", Type: toggle.IntType, Value: "some value"}, Op: toggle.GtOp},
		}}, want: "(invalid > int(some value))"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := toggle.Condition{
				Op:         tt.fields.Op,
				Conditions: tt.fields.Conditions,
				Fields:     tt.fields.Fields,
			}
			if got := c.String(); got != tt.want {
				t.Errorf("Condition.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
