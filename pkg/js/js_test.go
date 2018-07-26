package js

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dop251/goja"
)

func TestSimple(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function() { result++; }`)

	result := Execute(&Script{ID: "myscript.js", Data: script}, 0)
	require.NotNil(t, result)
	require.Equal(t, int64(1), result.(int64))

}

func TestSimpleBad(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function() { result++; `)

	result := Execute(&Script{ID: "myscript.js", Data: script}, 0)
	require.NotNil(t, result)
	_, ok := result.(*goja.Exception)
	require.Equal(t, true, ok)
}

func TestData(t *testing.T) {
	script := []byte(`
	let result = 0;
	export default function(data) { result = result + data.key;}`)

	result := Execute(&Script{ID: "myscript.js", Data: script}, map[string]interface{}{"key": 5})
	require.NotNil(t, result)
	require.Equal(t, int64(5), result.(int64))
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

	_, ok := result.(*goja.Exception)
	require.Equal(t, true, ok)
	err, ok2 := result.(error)
	require.Equal(t, true, ok2)
	require.Error(t, err)
}
