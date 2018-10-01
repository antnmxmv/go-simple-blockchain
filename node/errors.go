package main

import "fmt"

var errorDesc = []string{
	"Port is not an integer.",
	"Port is not in right range. \n[port] must be int between 1000 and 9999.",
	"Port is not specified.",
}

func getErrorDesc(id int) string {
	if id > 0 && id <= len(errorDesc) {
		return fmt.Sprintf("Error %d: %v", id, errorDesc[id-1])
	}
	return "Undefined error"
}
