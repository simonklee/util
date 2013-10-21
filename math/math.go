package math

func IntMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func IntMax(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func UintMin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func UintMax(a, b uint) uint {
	if a < b {
		return b
	}
	return a
}
