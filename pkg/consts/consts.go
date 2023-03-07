package consts

//http api path
const (
	Ping   = "/ping"
	Root   = "/"
	View   = "/view"
	Fetch  = "/fetch"
	Upload = "/upload"
	Static = "/static"
)

const CfgKeyPort = "port"
const CfgKeyDir = "dir"

const MaxUploadSize = 32 << 20 // 32 MiB
const FileFormID = "file"

var ProfileTypes = map[string]bool{
	"profile":      true,
	"allocs":       true,
	"block":        true,
	"cmdline":      true,
	"goroutine":    true,
	"heap":         true,
	"mutex":        true,
	"threadcreate": true,
}
