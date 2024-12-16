package git

import (
	"fmt"

	"github.com/instructhub/backend/pkg/utils"
)

func GenerateCommmitEmail(id uint64) string {
	return fmt.Sprintf("%s@%s", utils.Uint64ToStr(id), utils.GiteaCommitEmail)
} 