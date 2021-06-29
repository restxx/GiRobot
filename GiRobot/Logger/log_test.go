package log

import (
	"testing"
)

func TestInit(t *testing.T) {
	Init(BuildLogger("./", "CaseA", "guest", 2, "debug"))
	defer Flush()

	Debug("asdfadsfadfads")
	Debug("asdfadsfadfads")
	Debug("asdfadsfadfads")
	Debug("asdfadsfadfads")
	Debug("asdfadsfadfads")

}
