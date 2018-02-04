package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type task struct {
	url       string
	count     int
	errorText string
}

// Do http get and count "Go" substrings in response
func processTask(t *task) {
	resp, err := http.Get(t.url)
	if err != nil {
		t.errorText = err.Error()
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.errorText = err.Error()
		return
	}
	t.count = bytes.Count(body, []byte("Go"))
}

func main() {
	maxWorkers := 5
	workers := 0

	scanner := bufio.NewScanner(os.Stdin)
	taskCh := make(chan task)
	doneCh := make(chan bool)
	processedTaskCh := make(chan task)
	processedDoneCh := make(chan bool)

	// analyze processed tasks and print output
	go func() {
		totalCount := 0
		for pTask := range processedTaskCh {
			errText := ""
			if pTask.errorText != "" {
				errText = "Error: " + pTask.errorText
			}
			log.Printf("Count for %s: %d %s \n", pTask.url, pTask.count, errText)
			totalCount += pTask.count
		}
		log.Println("Total: ", totalCount)
		processedDoneCh <- true
	}()

	for scanner.Scan() {
		if workers < maxWorkers {
			// create a new worker if need
			workers++
			go func() {
				for task := range taskCh {
					processTask(&task)
					processedTaskCh <- task
				}
				doneCh <- true
			}()
		}
		url := scanner.Text()
		taskCh <- task{url: url}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
	close(taskCh) // notify to workers that there are no more tasks

	// wait workers until they finish process tasks
	for i := 0; i < workers; i++ {
		<-doneCh
	}
	close(processedTaskCh)
	<-processedDoneCh
}
