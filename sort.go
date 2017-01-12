package main

type byIndex []result

func (results byIndex) Len() int {
	return len(results)
}

func (results byIndex) Swap(i, j int) {
	temp := results[i]
	results[i] = results[j]
	results[j] = temp
}

func (results byIndex) Less(i, j int) bool {
	return results[i].index < results[j].index
}
