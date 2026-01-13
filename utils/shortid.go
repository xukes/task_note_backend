package utils

import (
	"math/rand"
	"time"
)

const (
	// Base62 字符集：0-9, a-z, A-Z
	charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 基准时间：2024-01-01 00:00:00 UTC
	baseTime = 1704067200
)

// GenerateShortID 生成8位短ID
// 格式：6位时间戳(Base62) + 2位随机字符(Base62)
// 可表示约178年的时间范围，每秒最多3844个唯一ID
func GenerateShortID() string {
	now := time.Now()

	// 1. 计算与基准时间的秒数差（6位 Base62，可表示约178年）
	seconds := now.Unix() - baseTime

	// 2. 生成2位随机字符防止冲突
	randomPart := rand.Intn(62 * 62)

	// 3. 组合并转换为 Base62
	id := ""

	// 时间部分（6位，从低位到高位）
	for i := 0; i < 6; i++ {
		id = string(charset[seconds%62]) + id
		seconds /= 62
	}

	// 随机部分（2位）
	for i := 0; i < 2; i++ {
		id = string(charset[randomPart%62]) + id
		randomPart /= 62
	}

	return id
}
