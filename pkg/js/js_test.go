package js

import "testing"

func TestSimple(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function() { result++; }`)

	result := Execute(script, 0)

	if result == nil {
		t.Fatal("result is nil")
	}
	if result.(int64) != 1 {
		t.Fatalf("unexpected result %v", result)
	}
}

func TestData(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function(data) { result = result + data.key;}`)

	result := Execute(script, map[string]interface{}{"key": 5})

	if result == nil {
		t.Fatal("result is nil")
	}
	if result.(int64) != 5 {
		t.Fatalf("unexpected result %v", result)
	}
}
