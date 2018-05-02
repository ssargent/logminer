package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

type logEntryRegexp struct {
	*regexp.Regexp
}

var logminer_regexp = logEntryRegexp{regexp.MustCompile(`((?P<datetime>[0-9\/:\sA-Z]+):\s(?P<msg>[A-Za-z]+): (?P<payload>.+))`)}
var logminer_order_regexp = logEntryRegexp{regexp.MustCompile(`(.*(?P<order>ORDER[0-9]+).*)`)}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("logminer /path/to/logfile ORDER1234")
		os.Exit(2)
	}

	file, err := os.Open(os.Args[1])
	//"C:/admin/data/logfile.log")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ordersMap := make(map[string][]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parsedEntry := logminer_regexp.FindStringSubmatchMap(scanner.Text())
		order := logminer_order_regexp.FindStringSubmatchMap(parsedEntry["payload"])

		if len(order) == 0 {
			continue
		}

		orderId := order["order"]

		ordersMap[orderId] = append(ordersMap[orderId], parsedEntry["payload"])
	}

	if len(os.Args) == 2 {
		for k := range ordersMap {
			fmt.Println(k)
		}
	} else {

		for _, oid := range os.Args[2:] {
			if ordersMap[oid] != nil {
				fmt.Println("Found Order")
				list := ordersMap[oid]

				i := 0
				l := len(list)

				for i < l {
					fmt.Println(list[i])
					fmt.Println(" ")
					i++
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (r *logEntryRegexp) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		// Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures //return this.  make a change.
}
