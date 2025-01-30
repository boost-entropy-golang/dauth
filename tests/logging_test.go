package dauth

import "github.com/streamingfast/logging"

var zlog, _ = logging.PackageLogger("dauth", "github.com/streamingfast/dauth/tests")

func init() {
	logging.InstantiateLoggers()
}
