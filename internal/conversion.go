package internal

func Bool2Int(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
