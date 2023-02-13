package main

const MAX_ATTEMPTS = 10

func MutateLittleFloat64(current float64, minMax MinMaxFloat64) float64 {
	//if !shouldMutate() {
	//	return current
	//}

	attemptsCount := 0
	result := current

	for {
		value := MutatePercentFloat(current)
		result = current - value

		if minMax.min > result || result > minMax.max {
			result = current
		}

		attemptsCount++
		if result != current || attemptsCount > MAX_ATTEMPTS {
			break
		}
	}

	return result
}

func MutateLittleInt(current int, minMax MinMaxInt) int {
	//if !shouldMutate() {
	//	return current
	//}

	result := current
	attemptsCount := 0
	for {
		value := MutatePercentInt(current)
		result = current - value

		if minMax.min > result || result > minMax.max {
			result = current
		}

		attemptsCount++
		if result != current || attemptsCount > MAX_ATTEMPTS {
			break
		}
	}

	return result
}

func MutatePercentFloat(current float64) float64 {
	dir := 1.0
	mutatePercent := GetRandFloat64(0, 100)
	mutateValue := (current * mutatePercent) / 100

	if GetRandInt(0, 1) == 1 {
		dir = -1.0
	}

	return dir * mutateValue
}

func MutatePercentInt(current int) int {
	dir := 1
	mutatePercent := int(GetRandInt(0, 100))
	mutateValue := int((current * mutatePercent) / 100)

	if GetRandInt(0, 1) == 1 {
		dir = -1
	}

	return dir * mutateValue
}

func shouldMutate() bool {
	return GetRandInt(0, 5) == 1
}
