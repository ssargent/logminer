package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ssargent/logminer/model"
)

type logEntryRegexp struct {
	*regexp.Regexp
}

func main() {
	defer timeTrack(time.Now(), "main")
	if len(os.Args) < 2 {
		fmt.Println("logminer /path/to/logfile ORDER1234")
		os.Exit(2)
	}

	ordersMap := parseFile(os.Args[1])

	if len(os.Args) == 2 {
		for k := range ordersMap {
			fmt.Println(k)
		}
	} else {

		for _, oid := range os.Args[2:] {
			if ordersMap[oid] != nil {
				list := ordersMap[oid]

				i := 0
				l := len(list)

				for i < l {
					fmt.Println(list[i])
					fmt.Println(" ")
					i++
				}
			} else {
				for k := range ordersMap {
					if strings.Contains(k, oid) {
						partialMatch := fmt.Sprintf("%s - (%d)", k, len(ordersMap[k]))
						fmt.Println(partialMatch)
					}
				}
			}
		}
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func parseFile(fileName string) map[string][]string {

	var parseLogEntryRegexp = logEntryRegexp{regexp.MustCompile(`((?P<datetime>[0-9\/:\sA-Z]+):\s(?P<msg>[A-Za-z]+): (?P<payload>.+))`)}
	var parseEntryPayloadRegexp = logEntryRegexp{regexp.MustCompile(`(.*(?P<order>[A-Z]{5,}[0-9]+).*)`)}

	defer timeTrack(time.Now(), "parseFile")

	jobs := make(chan string)
	results := make(chan model.OrderEntry)

	wg := new(sync.WaitGroup)

	for w := 1; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go parseLogEntries(jobs, results, wg, parseLogEntryRegexp, parseEntryPayloadRegexp)
	}

	file, err := os.Open(fileName)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ordersMap := make(map[string][]string)

	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			jobs <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for v := range results {
		ordersMap[v.OrderID] = append(ordersMap[v.OrderID], v.Entry)
	}

	return ordersMap
}

func parseLogEntries(jobs <-chan string, results chan<- model.OrderEntry, wg *sync.WaitGroup, parseLogEntryRegexp logEntryRegexp, parseEntryPayloadRegexp logEntryRegexp) {
	defer wg.Done()

	for j := range jobs {
		parsedEntry := parseLogEntryRegexp.FindStringSubmatchMap(j)
		order := parseEntryPayloadRegexp.FindStringSubmatchMap(parsedEntry["payload"])

		if len(order) == 0 {
			continue
		}

		orderID := order["order"]

		results <- model.OrderEntry{OrderID: orderID, Entry: parsedEntry["payload"]}
	}
}

func (r *logEntryRegexp) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}
