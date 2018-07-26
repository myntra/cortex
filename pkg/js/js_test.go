package js

import (
	"fmt"
	"testing"

	"github.com/dop251/goja"
)

func TestSimple(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function() { result++; }`)

	result := Execute(&Script{ID: "myscript.js", Data: script}, 0)

	if result == nil {
		t.Fatal("result is nil")
	}
	if result.(int64) != 1 {
		t.Fatalf("unexpected result %v", result)
	}
}

func TestSimpleBad(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function() { result++; `)

	result := Execute(&Script{ID: "myscript.js", Data: script}, 0)

	if result == nil {
		t.Fatal("result is nil")
	}
	ex, ok := result.(*goja.Exception)
	fmt.Println(ex, ok)

	if !ok {
		t.Fatalf("unexpected result %v", result)
	}
}

func TestData(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function(data) { result = result + data.key;}`)

	result := Execute(&Script{ID: "myscript.js", Data: script}, map[string]interface{}{"key": 5})

	if result == nil {
		t.Fatal("result is nil")
	}
	if result.(int64) != 5 {
		t.Fatalf("unexpected result %v", result)
	}
}

func TestException(t *testing.T) {
	script := []byte(`
	import http from "k6/http";
	import moment from "cdnjs.com/libraries/moment.js/2.18.1";
	
	export default function() {
		http.get("http://test.loadimpact.com/");
		console.log(moment().format());
		throw "execption"
	}`)

	result := Execute(&Script{ID: "myscript.js", Data: script}, nil)

	ex, ok := result.(*goja.Exception)
	err, ok2 := result.(error)
	fmt.Println(ex, ok, err, ok2)
}
