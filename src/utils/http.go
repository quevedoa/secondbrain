package utils

import (
	"fmt"
	"net/http"
	"strings"
)

func SendSSE(w http.ResponseWriter, event string, data string) {
	if event != "" {
		fmt.Fprintf(w, "event: %s\n", event)
	}
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
}
