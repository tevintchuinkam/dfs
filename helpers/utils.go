package helpers

import "bytes"

func CountWordOccurrences(data []byte, word string) int {
	// Split the text into words
	words := bytes.Fields(data)
	// Define the word we are looking for
	road := []byte(word)
	count := 0
	for _, word := range words {
		if bytes.Equal(word, road) {
			count++
		}
	}

	return count
}
