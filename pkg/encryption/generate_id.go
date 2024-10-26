package encryption

import (
	"time"

	"github.com/godruoyi/go-snowflake"
)

func InitSnowflake() {
	snowflake.SetMachineID(1)
	snowflake.SetStartTime(time.Date(2024, 10, 24, 0, 0, 0, 0, time.UTC))
}

func GenerateID() uint64 {
	return snowflake.ID()
}
