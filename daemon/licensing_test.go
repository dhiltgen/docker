package daemon // import "github.com/docker/docker/daemon"

import (
	"testing"

	"github.com/docker/docker/api/types"
	"gotest.tools/assert"
)

func TestfillLicense(t *testing.T) {
	v := &types.Info{}
	d := &Daemon{
		root: "/var/lib/docker/",
	}
	d.fillLicense(v)
	assert.Assert(t, v.ProductLicense == "Community Engine")
}
