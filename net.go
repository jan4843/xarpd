package main

import (
	"bufio"
	"os"
	"strings"
)

func defaultIfaceName() string {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 8 {
			continue
		}
		name, destination, mask := fields[0], fields[1], fields[7]
		if destination == "00000000" && mask == "00000000" {
			return name
		}
	}
	return ""
}
