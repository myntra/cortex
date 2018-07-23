package js

import (
	"context"

	"github.com/golang/glog"
	"github.com/loadimpact/k6/js"
	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
	"github.com/spf13/afero"
)

// Execute js
func Execute(script []byte, data interface{}) interface{} {
	r, err := js.New(&lib.SourceData{
		Filename: "/script.js",
		Data:     script,
	}, afero.NewMemMapFs(), lib.RuntimeOptions{})

	if err != nil {
		glog.Fatal(err)
	}

	r.SetSetupData(data)

	vu, err := r.NewVU(make(chan stats.SampleContainer, 100))
	if err != nil {
		glog.Fatal(err)
	}

	vuc, ok := vu.(*js.VU)

	if !ok {
		glog.Fatal(err)
	}

	err = vu.RunOnce(context.Background())
	if err != nil {
		glog.Fatal(err)
	}

	result := vuc.Runtime.Get("result")

	if result == nil {
		return nil
	}

	return result.Export()
}
