package dynamodb

import (
	"math"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type mySimpleStruct struct {
	String  string
	Int     int
	Uint    uint
	Float32 float32
	Float64 float64
	Bool    bool
	Null    *interface{}
}

type myComplexStruct struct {
	Simple []mySimpleStruct
}

type converterTestInput struct {
	input     interface{}
	expected  interface{}
	err       awserr.Error
	inputType string // "enum" of types
}

var trueValue = true
var falseValue = false

var converterScalarInputs = []converterTestInput{
	converterTestInput{
		input:    nil,
		expected: &dynamodb.AttributeValue{NULL: &trueValue},
	},
	converterTestInput{
		input:    "some string",
		expected: &dynamodb.AttributeValue{S: aws.String("some string")},
	},
	converterTestInput{
		input:    true,
		expected: &dynamodb.AttributeValue{BOOL: &trueValue},
	},
	converterTestInput{
		input:    false,
		expected: &dynamodb.AttributeValue{BOOL: &falseValue},
	},
	converterTestInput{
		input:    3.14,
		expected: &dynamodb.AttributeValue{N: aws.String("3.14")},
	},
	converterTestInput{
		input:    math.MaxFloat32,
		expected: &dynamodb.AttributeValue{N: aws.String("340282346638528860000000000000000000000")},
	},
	converterTestInput{
		input:    math.MaxFloat64,
		expected: &dynamodb.AttributeValue{N: aws.String("179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")},
	},
	converterTestInput{
		input:    12,
		expected: &dynamodb.AttributeValue{N: aws.String("12")},
	},
	converterTestInput{
		input: mySimpleStruct{},
		expected: &dynamodb.AttributeValue{
			M: map[string]*dynamodb.AttributeValue{
				"Bool":    &dynamodb.AttributeValue{BOOL: &falseValue},
				"Float32": &dynamodb.AttributeValue{N: aws.String("0")},
				"Float64": &dynamodb.AttributeValue{N: aws.String("0")},
				"Int":     &dynamodb.AttributeValue{N: aws.String("0")},
				"Null":    &dynamodb.AttributeValue{NULL: &trueValue},
				"String":  &dynamodb.AttributeValue{S: aws.String("")},
				"Uint":    &dynamodb.AttributeValue{N: aws.String("0")},
			},
		},
		inputType: "mySimpleStruct",
	},
}

var converterMapTestInputs = []converterTestInput{
	// Scalar tests
	converterTestInput{
		input: nil,
		err:   awserr.New("SerializationError", "in must be a map[string]interface{} or struct, got <nil>", nil),
	},
	converterTestInput{
		input:    map[string]interface{}{"string": "some string"},
		expected: map[string]*dynamodb.AttributeValue{"string": &dynamodb.AttributeValue{S: aws.String("some string")}},
	},
	converterTestInput{
		input:    map[string]interface{}{"bool": true},
		expected: map[string]*dynamodb.AttributeValue{"bool": &dynamodb.AttributeValue{BOOL: &trueValue}},
	},
	converterTestInput{
		input:    map[string]interface{}{"bool": false},
		expected: map[string]*dynamodb.AttributeValue{"bool": &dynamodb.AttributeValue{BOOL: &falseValue}},
	},
	converterTestInput{
		input:    map[string]interface{}{"null": nil},
		expected: map[string]*dynamodb.AttributeValue{"null": &dynamodb.AttributeValue{NULL: &trueValue}},
	},
	converterTestInput{
		input:    map[string]interface{}{"float": 3.14},
		expected: map[string]*dynamodb.AttributeValue{"float": &dynamodb.AttributeValue{N: aws.String("3.14")}},
	},
	converterTestInput{
		input:    map[string]interface{}{"float": math.MaxFloat32},
		expected: map[string]*dynamodb.AttributeValue{"float": &dynamodb.AttributeValue{N: aws.String("340282346638528860000000000000000000000")}},
	},
	converterTestInput{
		input:    map[string]interface{}{"float": math.MaxFloat64},
		expected: map[string]*dynamodb.AttributeValue{"float": &dynamodb.AttributeValue{N: aws.String("179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")}},
	},
	converterTestInput{
		input:    map[string]interface{}{"int": int(12)},
		expected: map[string]*dynamodb.AttributeValue{"int": &dynamodb.AttributeValue{N: aws.String("12")}},
	},
	// List
	converterTestInput{
		input: map[string]interface{}{"list": []interface{}{"a string", 12, 3.14, true, nil, false}},
		expected: map[string]*dynamodb.AttributeValue{
			"list": &dynamodb.AttributeValue{
				L: []*dynamodb.AttributeValue{
					&dynamodb.AttributeValue{S: aws.String("a string")},
					&dynamodb.AttributeValue{N: aws.String("12")},
					&dynamodb.AttributeValue{N: aws.String("3.14")},
					&dynamodb.AttributeValue{BOOL: &trueValue},
					&dynamodb.AttributeValue{NULL: &trueValue},
					&dynamodb.AttributeValue{BOOL: &falseValue},
				},
			},
		},
	},
	// Map
	converterTestInput{
		input: map[string]interface{}{"map": map[string]interface{}{"nestedint": 12}},
		expected: map[string]*dynamodb.AttributeValue{
			"map": &dynamodb.AttributeValue{
				M: map[string]*dynamodb.AttributeValue{
					"nestedint": &dynamodb.AttributeValue{
						N: aws.String("12"),
					},
				},
			},
		},
	},
	// Structs
	converterTestInput{
		input: mySimpleStruct{},
		expected: map[string]*dynamodb.AttributeValue{
			"Bool":    &dynamodb.AttributeValue{BOOL: &falseValue},
			"Float32": &dynamodb.AttributeValue{N: aws.String("0")},
			"Float64": &dynamodb.AttributeValue{N: aws.String("0")},
			"Int":     &dynamodb.AttributeValue{N: aws.String("0")},
			"Null":    &dynamodb.AttributeValue{NULL: &trueValue},
			"String":  &dynamodb.AttributeValue{S: aws.String("")},
			"Uint":    &dynamodb.AttributeValue{N: aws.String("0")},
		},
		inputType: "mySimpleStruct",
	},
	converterTestInput{
		input: myComplexStruct{},
		expected: map[string]*dynamodb.AttributeValue{
			"Simple": &dynamodb.AttributeValue{NULL: &trueValue},
		},
		inputType: "myComplexStruct",
	},
	converterTestInput{
		input: myComplexStruct{Simple: []mySimpleStruct{mySimpleStruct{Int: -2}, mySimpleStruct{Uint: 5}}},
		expected: map[string]*dynamodb.AttributeValue{
			"Simple": &dynamodb.AttributeValue{
				L: []*dynamodb.AttributeValue{
					&dynamodb.AttributeValue{
						M: map[string]*dynamodb.AttributeValue{
							"Bool":    &dynamodb.AttributeValue{BOOL: &falseValue},
							"Float32": &dynamodb.AttributeValue{N: aws.String("0")},
							"Float64": &dynamodb.AttributeValue{N: aws.String("0")},
							"Int":     &dynamodb.AttributeValue{N: aws.String("-2")},
							"Null":    &dynamodb.AttributeValue{NULL: &trueValue},
							"String":  &dynamodb.AttributeValue{S: aws.String("")},
							"Uint":    &dynamodb.AttributeValue{N: aws.String("0")},
						},
					},
					&dynamodb.AttributeValue{
						M: map[string]*dynamodb.AttributeValue{
							"Bool":    &dynamodb.AttributeValue{BOOL: &falseValue},
							"Float32": &dynamodb.AttributeValue{N: aws.String("0")},
							"Float64": &dynamodb.AttributeValue{N: aws.String("0")},
							"Int":     &dynamodb.AttributeValue{N: aws.String("0")},
							"Null":    &dynamodb.AttributeValue{NULL: &trueValue},
							"String":  &dynamodb.AttributeValue{S: aws.String("")},
							"Uint":    &dynamodb.AttributeValue{N: aws.String("5")},
						},
					},
				},
			},
		},
		inputType: "myComplexStruct",
	},
}

var converterListTestInputs = []converterTestInput{
	converterTestInput{
		input: nil,
		err:   awserr.New("SerializationError", "in must be an array or slice, got <nil>", nil),
	},
	converterTestInput{
		input:    []interface{}{},
		expected: []*dynamodb.AttributeValue{},
	},
	converterTestInput{
		input: []interface{}{"a string", 12, 3.14, true, nil, false},
		expected: []*dynamodb.AttributeValue{
			&dynamodb.AttributeValue{S: aws.String("a string")},
			&dynamodb.AttributeValue{N: aws.String("12")},
			&dynamodb.AttributeValue{N: aws.String("3.14")},
			&dynamodb.AttributeValue{BOOL: &trueValue},
			&dynamodb.AttributeValue{NULL: &trueValue},
			&dynamodb.AttributeValue{BOOL: &falseValue},
		},
	},
	converterTestInput{
		input: []mySimpleStruct{mySimpleStruct{}},
		expected: []*dynamodb.AttributeValue{
			&dynamodb.AttributeValue{
				M: map[string]*dynamodb.AttributeValue{
					"Bool":    &dynamodb.AttributeValue{BOOL: &falseValue},
					"Float32": &dynamodb.AttributeValue{N: aws.String("0")},
					"Float64": &dynamodb.AttributeValue{N: aws.String("0")},
					"Int":     &dynamodb.AttributeValue{N: aws.String("0")},
					"Null":    &dynamodb.AttributeValue{NULL: &trueValue},
					"String":  &dynamodb.AttributeValue{S: aws.String("")},
					"Uint":    &dynamodb.AttributeValue{N: aws.String("0")},
				},
			},
		},
		inputType: "mySimpleStruct",
	},
}

func TestConvertTo(t *testing.T) {
	for _, test := range converterScalarInputs {
		testConvertTo(t, test)
	}
}

func testConvertTo(t *testing.T, test converterTestInput) {
	actual, err := ConvertTo(test.input)
	if test.err != nil {
		if err == nil {
			t.Errorf("ConvertTo with input %#v retured %#v, expected error `%s`", test.input, actual, test.err)
		} else if err.Error() != test.err.Error() {
			t.Errorf("ConvertTo with input %#v retured error `%s`, expected error `%s`", test.input, err, test.err)
		}
	} else {
		if err != nil {
			t.Errorf("ConvertTo with input %#v retured error `%s`", test.input, err)
		}
		compareObjects(t, test.expected, actual)
	}
}

func TestConvertFrom(t *testing.T) {
	// Using the same inputs from TestConvertTo, test the reverse mapping.
	for _, test := range converterScalarInputs {
		if test.expected != nil {
			testConvertFrom(t, test)
		}
	}
}

func testConvertFrom(t *testing.T, test converterTestInput) {
	switch test.inputType {
	case "mySimpleStruct":
		var actual mySimpleStruct
		if err := ConvertFrom(test.expected.(*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFrom with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	case "myComplexStruct":
		var actual myComplexStruct
		if err := ConvertFrom(test.expected.(*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFrom with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	default:
		var actual interface{}
		if err := ConvertFrom(test.expected.(*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFrom with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	}
}

func TestConvertFromError(t *testing.T) {
	// Test that we get an error using ConvertFrom to convert to a map.
	var actual map[string]interface{}
	expected := awserr.New("SerializationError", `v must be a non-nil pointer to an interface{} or struct, got *map[string]interface {}`, nil).Error()
	if err := ConvertFrom(nil, &actual); err == nil {
		t.Errorf("ConvertFrom with input %#v returned no error, expected error `%s`", nil, expected)
	} else if err.Error() != expected {
		t.Errorf("ConvertFrom with input %#v returned error `%s`, expected error `%s`", nil, err, expected)
	}

	// Test that we get an error using ConvertFrom to convert to a list.
	var actual2 []interface{}
	expected = awserr.New("SerializationError", `v must be a non-nil pointer to an interface{} or struct, got *[]interface {}`, nil).Error()
	if err := ConvertFrom(nil, &actual2); err == nil {
		t.Errorf("ConvertFrom with input %#v returned no error, expected error `%s`", nil, expected)
	} else if err.Error() != expected {
		t.Errorf("ConvertFrom with input %#v returned error `%s`, expected error `%s`", nil, err, expected)
	}
}

func TestConvertToMap(t *testing.T) {
	for _, test := range converterMapTestInputs {
		testConvertToMap(t, test)
	}
}

func testConvertToMap(t *testing.T, test converterTestInput) {
	actual, err := ConvertToMap(test.input)
	if test.err != nil {
		if err == nil {
			t.Errorf("ConvertToMap with input %#v retured %#v, expected error `%s`", test.input, actual, test.err)
		} else if err.Error() != test.err.Error() {
			t.Errorf("ConvertToMap with input %#v retured error `%s`, expected error `%s`", test.input, err, test.err)
		}
	} else {
		if err != nil {
			t.Errorf("ConvertToMap with input %#v retured error `%s`", test.input, err)
		}
		compareObjects(t, test.expected, actual)
	}
}

func TestConvertFromMap(t *testing.T) {
	// Using the same inputs from TestConvertToMap, test the reverse mapping.
	for _, test := range converterMapTestInputs {
		if test.expected != nil {
			testConvertFromMap(t, test)
		}
	}
}

func testConvertFromMap(t *testing.T, test converterTestInput) {
	switch test.inputType {
	case "mySimpleStruct":
		var actual mySimpleStruct
		if err := ConvertFromMap(test.expected.(map[string]*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFromMap with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	case "myComplexStruct":
		var actual myComplexStruct
		if err := ConvertFromMap(test.expected.(map[string]*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFromMap with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	default:
		var actual map[string]interface{}
		if err := ConvertFromMap(test.expected.(map[string]*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFromMap with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	}
}

func TestConvertFromMapError(t *testing.T) {
	// Test that we get an error using ConvertFromMap to convert to an interface{}.
	var actual interface{}
	expected := awserr.New("SerializationError", `v must be a non-nil pointer to a map[string]interface{} or struct, got *interface {}`, nil).Error()
	if err := ConvertFromMap(nil, &actual); err == nil {
		t.Errorf("ConvertFromMap with input %#v returned no error, expected error `%s`", nil, expected)
	} else if err.Error() != expected {
		t.Errorf("ConvertFromMap with input %#v returned error `%s`, expected error `%s`", nil, err, expected)
	}

	// Test that we get an error using ConvertFromMap to convert to a slice.
	var actual2 []interface{}
	expected = awserr.New("SerializationError", `v must be a non-nil pointer to a map[string]interface{} or struct, got *[]interface {}`, nil).Error()
	if err := ConvertFromMap(nil, &actual2); err == nil {
		t.Errorf("ConvertFromMap with input %#v returned no error, expected error `%s`", nil, expected)
	} else if err.Error() != expected {
		t.Errorf("ConvertFromMap with input %#v returned error `%s`, expected error `%s`", nil, err, expected)
	}
}

func TestConvertToList(t *testing.T) {
	for _, test := range converterListTestInputs {
		testConvertToList(t, test)
	}
}

func testConvertToList(t *testing.T, test converterTestInput) {
	actual, err := ConvertToList(test.input)
	if test.err != nil {
		if err == nil {
			t.Errorf("ConvertToList with input %#v retured %#v, expected error `%s`", test.input, actual, test.err)
		} else if err.Error() != test.err.Error() {
			t.Errorf("ConvertToList with input %#v retured error `%s`, expected error `%s`", test.input, err, test.err)
		}
	} else {
		if err != nil {
			t.Errorf("ConvertToList with input %#v retured error `%s`", test.input, err)
		}
		compareObjects(t, test.expected, actual)
	}
}

func TestConvertFromList(t *testing.T) {
	// Using the same inputs from TestConvertToList, test the reverse mapping.
	for _, test := range converterListTestInputs {
		if test.expected != nil {
			testConvertFromList(t, test)
		}
	}
}

func testConvertFromList(t *testing.T, test converterTestInput) {
	switch test.inputType {
	case "mySimpleStruct":
		var actual []mySimpleStruct
		if err := ConvertFromList(test.expected.([]*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFromList with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	case "myComplexStruct":
		var actual []myComplexStruct
		if err := ConvertFromList(test.expected.([]*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFromList with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	default:
		var actual []interface{}
		if err := ConvertFromList(test.expected.([]*dynamodb.AttributeValue), &actual); err != nil {
			t.Errorf("ConvertFromList with input %#v retured error `%s`", test.expected, err)
		}
		compareObjects(t, test.input, actual)
	}
}

func TestConvertFromListError(t *testing.T) {
	// Test that we get an error using ConvertFromList to convert to a map.
	var actual map[string]interface{}
	expected := awserr.New("SerializationError", `v must be a non-nil pointer to an array or slice, got *map[string]interface {}`, nil).Error()
	if err := ConvertFromList(nil, &actual); err == nil {
		t.Errorf("ConvertFromList with input %#v returned no error, expected error `%s`", nil, expected)
	} else if err.Error() != expected {
		t.Errorf("ConvertFromList with input %#v returned error `%s`, expected error `%s`", nil, err, expected)
	}

	// Test that we get an error using ConvertFromList to convert to a struct.
	var actual2 myComplexStruct
	expected = awserr.New("SerializationError", `v must be a non-nil pointer to an array or slice, got *dynamodb.myComplexStruct`, nil).Error()
	if err := ConvertFromList(nil, &actual2); err == nil {
		t.Errorf("ConvertFromList with input %#v returned no error, expected error `%s`", nil, expected)
	} else if err.Error() != expected {
		t.Errorf("ConvertFromList with input %#v returned error `%s`, expected error `%s`", nil, err, expected)
	}

	// Test that we get an error using ConvertFromList to convert to an interface{}.
	var actual3 interface{}
	expected = awserr.New("SerializationError", `v must be a non-nil pointer to an array or slice, got *interface {}`, nil).Error()
	if err := ConvertFromList(nil, &actual3); err == nil {
		t.Errorf("ConvertFromList with input %#v returned no error, expected error `%s`", nil, expected)
	} else if err.Error() != expected {
		t.Errorf("ConvertFromList with input %#v returned error `%s`, expected error `%s`", nil, err, expected)
	}
}

func compareObjects(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("\nExpected %s:\n%s\nActual %s:\n%s\n",
			reflect.ValueOf(expected).Kind(),
			awsutil.StringValue(expected),
			reflect.ValueOf(actual).Kind(),
			awsutil.StringValue(actual))
	}
}
