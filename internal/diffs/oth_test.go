package diffs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOtherFilesDiffSummary_String(t *testing.T) {
	summary := &OtherFilesDiffSummary{
		Diffs: []OtherFileDiff{
			{
				Ext:      ".sh",
				Added:    []string{"scripts/setup.sh"},
				Modified: []string{"scripts/deploy.sh"},
				Removed:  []string{},
				Other:    []string{"scripts/misc.sh"},
			},
			{
				Ext:      ".sql",
				Added:    []string{"migrations/001_init.sql"},
				Modified: []string{},
				Removed:  []string{"migrations/000_old.sql"},
				Other:    []string{},
			},
		},
	}

	output := summary.String()

	assert.Contains(t, output, "## Other Files Changes")
	assert.Contains(t, output, "### `.sh`")
	assert.Contains(t, output, "- Added:")
	assert.Contains(t, output, "scripts/setup.sh")
	assert.Contains(t, output, "- Modified:")
	assert.Contains(t, output, "scripts/deploy.sh")
	assert.Contains(t, output, "- Other:")
	assert.Contains(t, output, "scripts/misc.sh")

	assert.Contains(t, output, "### `.sql`")
	assert.Contains(t, output, "- Added:")
	assert.Contains(t, output, "migrations/001_init.sql")
	assert.Contains(t, output, "- Removed:")
	assert.Contains(t, output, "migrations/000_old.sql")
}
