package runtime

import "testing"

func TestInspectOutputs(t *testing.T) {
	enum := &EnumType{Name: "Result", Cases: map[string][]string{"Err": []string{"error"}, "Ok": []string{"value"}}}
	module := NewModule("math", map[string]Value{"Pi": NewNumber(3.14)})
	structType := &StructType{Name: "Point"}
	classType := &ClassType{Name: "User"}

	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{"array", &Array{Elements: []Value{NewNumber(1), NewString("two")}}, "[1, two]"},
		{"array empty", &Array{}, "[]"},
		{"object", &Object{Properties: map[string]Value{"b": TrueValue, "a": NewNumber(2)}}, "{a: 2, b: true}"},
		{"object empty", &Object{}, "{}"},
		{"module", module, "<module math [Pi]>"},
		{"struct type", structType, "<struct Point>"},
		{"struct instance", &StructInstance{Definition: structType, Fields: map[string]Value{"y": NewNumber(2), "x": NewNumber(1)}}, "Point{x: 1, y: 2}"},
		{"struct empty", &StructInstance{Definition: structType}, "Point{}"},
		{"class type", classType, "<class User>"},
		{"class instance", &ClassInstance{Definition: classType, Fields: map[string]Value{"name": NewString("selene"), "id": NewNumber(42)}}, "User{id: 42, name: selene}"},
		{"enum type", enum, "<enum Result [Err, Ok]>"},
		{"enum instance", &EnumInstance{Enum: enum, Case: "Ok", Fields: map[string]Value{"value": NewNumber(1)}, Order: []string{"value"}}, "Result.Ok(1)"},
		{"enum bare", &EnumInstance{Enum: enum, Case: "Err"}, "Result.Err"},
		{"channel open", &ChannelValue{ch: make(chan Value), capacity: 0}, "<channel open capacity=0>"},
		{"channel closed", &ChannelValue{closed: true, capacity: 2}, "<channel closed capacity=2>"},
		{"error simple", &ErrorValue{Message: "boom"}, "<error boom>"},
		{"error cause", &ErrorValue{Message: "boom", Cause: NewString("details")}, "<error boom : details>"},
		{"contract", &Contract{Name: "Foo", Exports: map[string]Value{"a": TrueValue, "b": FalseValue}}, "<contract Foo [a, b]>"},
		{"interface", &InterfaceType{Name: "Thing", Methods: map[string]int{"Alpha": 1, "Beta": 2}}, "<interface Thing [Alpha, Beta]>"},
		{"function named", &Function{Name: "do"}, "<fn do>"},
		{"function anonymous", &Function{}, "<fn>"},
		{"pointer", &Pointer{name: "value"}, "&value"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.Inspect(); got != tt.want {
				t.Fatalf("Inspect() = %q, want %q", got, tt.want)
			}
		})
	}
}
