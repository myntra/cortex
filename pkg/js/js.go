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
	if len(script) == 0 {
		return nil
	}

	r, err := js.New(&lib.SourceData{
		Filename: "correlate.js",
		Data:     script,
	}, afero.NewMemMapFs(), lib.RuntimeOptions{})

	if err != nil {
		return err
	}
	glog.Infof("%v", data)
	r.SetSetupData(data)

	vu, err := r.NewVU(make(chan stats.SampleContainer, 100))
	if err != nil {
		return err
	}

	vuc, ok := vu.(*js.VU)

	if !ok {
		return err
	}

	err = vu.RunOnce(context.Background())
	if err != nil {
		return err
	}

	result := vuc.Runtime.Get("result")

	if result == nil {
		return nil
	}

	return result.Export()
}
