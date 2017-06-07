package version

import (
	"fmt"
)

const VERSION_MAJOR = 0
const VERSION_MINOR = 2
const VERSION_PATCH = 0

func Version() string {
	return fmt.Sprintf("v%d.%d.%d", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH)
}
