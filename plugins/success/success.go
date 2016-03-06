// Package success implements a very simple plugin that looks that the
// ResultSet.Success() value to determine if the process from the sandbox
// exited successfully.
//
// Most engines implements ResultSet.Success() to mean the sub-process exited
// non-zero. In this plugin we use this in the Stopped() hook to ensure that
// tasks are declared "failed" if they had a non-zero exit code.
//
// The attentive reader might think this is remarkably simple and stupid plugin.
// This is true, but it does display the concept of plugins and more importantly
// removes a special case that we would otherwise have to take into
// consideration in the runtime.
package success

import (
	"github.com/taskcluster/taskcluster-worker/engines"
	"github.com/taskcluster/taskcluster-worker/plugins"
	"github.com/taskcluster/taskcluster-worker/plugins/extpoints"
	"github.com/taskcluster/taskcluster-worker/runtime"
)

type pluginProvider struct {
}

func (pluginProvider) NewPlugin(extpoints.PluginOptions) (plugins.Plugin, error) {
	return plugin{}, nil
}

func (pluginProvider) ConfigSchema() runtime.CompositeSchema {
	return runtime.NewEmptyCompositeSchema()
}

type plugin struct {
	plugins.PluginBase
}

type taskPlugin struct {
}

func init() {
	extpoints.PluginProviders.Register(new(pluginProvider), "success")
}

func (plugin) NewTaskPlugin(plugins.TaskPluginOptions) (plugins.TaskPlugin, error) {
	return new(taskPlugin), nil
}

// Prepare ignores the sandbox preparation stage.
func (t *taskPlugin) Prepare(*runtime.TaskContext) error {
	return nil
}

// BuildSandbox ignores the sandbox building stage.
func (t *taskPlugin) BuildSandbox(engines.SandboxBuilder) error {
	return nil
}

// Started ignores the stage where the sandbox has started
func (t *taskPlugin) Started(engines.Sandbox) error {
	return nil
}

func (t *taskPlugin) Stopped(result engines.ResultSet) (bool, error) {
	return result.Success(), nil
}

// Finished ignores the stage where a task has been finished
func (t *taskPlugin) Finished(success bool) error {
	return nil
}

// Exception ignores the stage where a task is resolved exception
func (t *taskPlugin) Exception(reason runtime.ExceptionReason) error {
	return nil
}

// Dispose ignores the stage where resources are disposed.
func (t *taskPlugin) Dispose() error {
	return nil
}
