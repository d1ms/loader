package misc

//SplitLinks - split array for manager threads
func SplitLinks(links []string, workersNumber int) ([][]string, error) {

	workersNumber = max(1, workersNumber)
	result := [][]string{}
	for i := 0; i < workersNumber; i++ {
		result = append(result, []string{})
	}
	counter := 0
	for _, link := range links {
		result[counter] = append(result[counter], link)
		counter++
		if counter == workersNumber {
			counter = 0
		}
	}
	return result, nil
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
