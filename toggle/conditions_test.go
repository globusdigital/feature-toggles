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
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}},
		}}},

		{name: "mismatched types", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatValue, Value: int64(10)},
		}},

		{name: "svc1 and svc2", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringValue, Value: "svc1"}},
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue + "2", Type: toggle.StringValue, Value: "svc2"}},
		}}, values: []toggle.ConditionValue{
			{Name: toggle.ServiceNameValue, Type: toggle.StringValue, Value: "svc1"},
			{Name: toggle.ServiceNameValue + "2", Type: toggle.StringValue, Value: "svc3"},
		}},

		{name: "svc1 and svc3", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringValue, Value: "svc1"}},
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue + "2", Type: toggle.StringValue, Value: "svc3"}},
		}}, values: []toggle.ConditionValue{
			{Name: toggle.ServiceNameValue, Type: toggle.StringValue, Value: "svc1"},
			{Name: toggle.ServiceNameValue + "2", Type: toggle.StringValue, Value: "svc3"},
		}, want: true},

		{name: "svc1 or svc2", c: toggle.Condition{Op: toggle.OrOp, Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue, Type: toggle.StringValue, Value: "svc1"}},
			{ConditionValue: toggle.ConditionValue{Name: toggle.ServiceNameValue + "2", Type: toggle.StringValue, Value: "svc2"}},
		}}, values: []toggle.ConditionValue{
			{Name: toggle.ServiceNameValue, Type: toggle.StringValue, Value: "svc1"},
			{Name: toggle.ServiceNameValue + "2", Type: toggle.StringValue, Value: "svc3"},
		}, want: true},

		{name: "10 < int - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntValue, Value: int64(20)},
		}, want: true},

		{name: "10 < int", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntValue, Value: int64(2)},
		}},

		{name: "10 < float64 - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatValue, Value: float64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatValue, Value: float64(20)},
		}, want: true},

		{name: "10 < float64", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatValue, Value: float64(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatValue, Value: float64(2)},
		}},

		{name: "10 < string - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringValue, Value: string(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringValue, Value: string(20)},
		}, want: true},

		{name: "10 < string", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringValue, Value: string(10)}, Op: toggle.LtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringValue, Value: string(2)},
		}},

		{name: "10 > int - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntValue, Value: int64(2)},
		}, want: true},

		{name: "10 > int", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntValue, Value: int64(20)},
		}},

		{name: "10 > float64 - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatValue, Value: float64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatValue, Value: float64(2)},
		}, want: true},

		{name: "10 > float64", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatValue, Value: float64(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatValue, Value: float64(20)},
		}},

		{name: "10 > string - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringValue, Value: string(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringValue, Value: string(2)},
		}, want: true},

		{name: "10 > string", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringValue, Value: string(10)}, Op: toggle.GtOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringValue, Value: string(20)},
		}},

		{name: "10 != int - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntValue, Value: int64(2)},
		}, want: true},

		{name: "10 != int", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.IntValue, Value: int64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.IntValue, Value: int64(10)},
		}},

		{name: "10 != float64 - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatValue, Value: float64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatValue, Value: float64(2)},
		}, want: true},

		{name: "10 != float64", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.FloatValue, Value: float64(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.FloatValue, Value: float64(10)},
		}},

		{name: "10 != string - true", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringValue, Value: string(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringValue, Value: string(2)},
		}, want: true},

		{name: "10 != string", c: toggle.Condition{Fields: []toggle.ConditionField{
			{ConditionValue: toggle.ConditionValue{Name: "field", Type: toggle.StringValue, Value: string(10)}, Op: toggle.NeOp},
		}}, values: []toggle.ConditionValue{
			{Name: "field", Type: toggle.StringValue, Value: string(10)},
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
		{name: "invalid int type", fields: fields{Type: toggle.IntValue, Value: 4.13}, wantErr: true},
		{name: "valid int type", fields: fields{Type: toggle.IntValue, Value: int64(43)}},
		{name: "invalid float type", fields: fields{Type: toggle.FloatValue, Value: 43}, wantErr: true},
		{name: "valid float type", fields: fields{Type: toggle.FloatValue, Value: 43.5}},
		{name: "invalid string type", fields: fields{Type: toggle.StringValue, Value: 43}, wantErr: true},
		{name: "valid string type", fields: fields{Type: toggle.StringValue, Value: "43.5"}},
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
