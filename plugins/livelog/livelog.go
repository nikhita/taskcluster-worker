//go:generate go-composite-schema --unexported --required config config-schema.yml generated_configschema.go

package livelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/taskcluster/taskcluster-client-go/queue"
	"github.com/taskcluster/taskcluster-client-go/tcclient"
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
	return configSchema
}

type plugin struct {
	plugins.PluginBase
}

type taskPlugin struct {
	context  *runtime.TaskContext
	expires  tcclient.Time
	deadline tcclient.Time
}

func init() {
	extpoints.PluginProviders.Register(new(pluginProvider), "livelog")
}

func (plugin) NewTaskPlugin(plugins.TaskPluginOptions) (plugins.TaskPlugin, error) {
	return new(taskPlugin), nil
}

func (p *taskPlugin) Prepare(context *runtime.TaskContext) error {
	p.context = context
	p.deadline = context.Definition().Deadline
	p.expires = context.Definition().Expires
	return nil
}

func (p taskPlugin) BuildSandbox(engines.SandboxBuilder) error {
	return nil
}

func (p taskPlugin) Started(engines.Sandbox) error {
	return nil
}

func (p taskPlugin) Stopped(engines.ResultSet) (bool, error) {
	return true, nil
}

func (p taskPlugin) Finished(success bool) error {
	client := queue.New(p.context.Credentials())

	msg := "unsuccessfully"
	if success {
		msg = "successfully"
	}
	p.context.Log(fmt.Sprintf("Task completed %s", msg))

	logReader, err := p.context.NewLogReader()
	if err != nil {
		return err
	}
	s3Req := queue.S3ArtifactRequest{
		ContentType: "text/plain; charset=utf-8",
		Expires:     p.expires,
		StorageType: "s3",
	}

	payload, err := json.Marshal(s3Req)
	if err != nil {
		return err
	}

	par := queue.PostArtifactRequest(json.RawMessage(payload))
	if err != nil {
		return err
	}

	parsp, _, err := client.CreateArtifact(
		p.context.TaskId,
		strconv.Itoa(int(p.context.TaskInfo.RunId)),
		"public/live_backing.log",
		&par,
	)
	if err != nil {
		panic(err)
	}

	resp := new(queue.S3ArtifactResponse)
	err = json.Unmarshal(json.RawMessage(*parsp), resp)
	if err != nil {
		return err
	}

	l, _ := ioutil.ReadAll(logReader)
	req, err := http.NewRequest("PUT", resp.PutUrl, bytes.NewReader(l))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	httpClient := &http.Client{}
	res, err := httpClient.Do(req)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (p taskPlugin) Exception(reason runtime.ExceptionReason) error {
	return nil
}

func (p taskPlugin) Dispose() error {
	return nil
}
