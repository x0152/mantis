package sandboxes

import "embed"

//go:embed base browser ffmpeg netsec runtimectl
var FS embed.FS
