package js

import (
	"context"

	"github.com/golang/glog"
	"github.com/loadimpact/k6/js"
	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
	"github.com/spf13/afero"
)

//go:generate msgp

// Script contains the javascript code
type Script struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}

// Execute js
func Execute(script *Script, data interface{}) interface{} {
	if script == nil || len(script.ID) == 0 {
		return nil
	}

	r, err := js.New(&lib.SourceData{
		Filename: script.ID,
		Data:     script.Data,
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
