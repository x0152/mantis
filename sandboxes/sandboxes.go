package sandboxes

import "embed"

//go:embed base browser db ffmpeg netsec python runtimectl
var FS embed.FS
