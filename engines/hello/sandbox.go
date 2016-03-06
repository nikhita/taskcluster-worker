package helloengine

import (
	"fmt"

	"github.com/CrowdSurge/banner"
	"github.com/taskcluster/taskcluster-worker/engines"
	"github.com/taskcluster/taskcluster-worker/runtime"
)

// In this example it is easier to just implement with one object.
// This way we won't have to pass data between different instances.
// In larger more complex engines that downloads stuff, etc. it's probably not
// a good idea to implement everything in one structure.
type sandbox struct {
	engines.SandboxBuilderBase
	engines.SandboxBase
	engines.ResultSetBase
	payload *payload
	context *runtime.TaskContext
	result  bool
}

///////////////////////////// Implementation of SandboxBuilder interface

func (s *sandbox) StartSandbox() (engines.Sandbox, error) {
	msg := banner.PrintS(fmt.Sprintf("hello %s", s.payload.Name))
	fmt.Fprintln(s.context.LogDrain(), msg)
	s.result = true
	return s, nil
}

///////////////////////////// Implementation of Sandbox interface

func (s *sandbox) WaitForResult() (engines.ResultSet, error) {
	return s, nil
}

///////////////////////////// Implementation of ResultSet interface

func (s *sandbox) Success() bool {
	return s.result
}

func (s *sandbox) SetResult(result bool) {
	s.result = result
}
