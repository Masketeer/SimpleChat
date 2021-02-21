package conf

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

var DirtyWords map[string]interface{}

func loadWords() {
	file, err := os.Open("list.txt")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if file != nil {
		fmt.Println("file is nil")
		return
	}

	defer file.Close()

	DirtyWords = map[string]interface{}{}
	bReader := bufio.NewReader(file)
	loopCount := 0
	for {
		if loopCount > 100000 {
			break
		}
		loopCount++

		word, _, err := bReader.ReadLine()
		if err == io.EOF {
			break
		}

		DirtyWords[string(word)] = nil
	}
}

func init() {
	loadWords()
}