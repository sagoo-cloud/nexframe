package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/logs", handleLogs)

	port := "8089"
	fmt.Printf("Starting log receiver server on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// 打印接收到的日志
	fmt.Printf("Received log at %s:\n%s\n", time.Now().Format(time.RFC3339), string(body))

	// 将日志保存到文件
	filename := "received_logs.txt"
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening file: %v", err)
	} else {
		defer f.Close()
		if _, err := f.WriteString(fmt.Sprintf("--- Log received at %s ---\n%s\n\n", time.Now().Format(time.RFC3339), string(body))); err != nil {
			log.Printf("Error writing to file: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Log received successfully"))
}
