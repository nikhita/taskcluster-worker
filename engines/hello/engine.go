//go:generate go-composite-schema --unexported --required start payload-schema.yml generated_payloadschema.go

// Package mockengine implements a MockEngine that doesn't really do anything,
// but allows us to test plugins without having to run a real engine.
package helloengine

import (
	"github.com/Sirupsen/logrus"
	"github.com/taskcluster/taskcluster-worker/engines"
	"github.com/taskcluster/taskcluster-worker/engines/extpoints"
	"github.com/taskcluster/taskcluster-worker/runtime"
)

type engine struct {
	engines.EngineBase
	Log *logrus.Entry
}

type engineProvider struct{}

func (e engineProvider) NewEngine(options extpoints.EngineOptions) (engines.Engine, error) {
	return engine{Log: options.Log}, nil
}

func init() {
	extpoints.EngineProviders.Register(new(engineProvider), "hello")
}

//  hello config contains no fields
func (e engineProvider) ConfigSchema() runtime.CompositeSchema {
	return runtime.NewEmptyCompositeSchema()
}

func (e engine) PayloadSchema() runtime.CompositeSchema {
	return payloadSchema
}

func (e engine) NewSandboxBuilder(options engines.SandboxOptions) (engines.SandboxBuilder, error) {
	e.Log.Debug("Building Sandbox")
	p, valid := options.Payload.(*payload)
	if !valid {
		// TODO: Write to some sort of log if the type assertion fails
		return nil, engines.ErrContractViolation
	}
	return &sandbox{
		payload: p,
		context: options.TaskContext,
	}, nil
}
