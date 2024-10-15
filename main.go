package main

import (
	"fmt"
	"strings"
	"time"
	"net/http"
	"encoding/json"
	"math"
	"os"
	"io/ioutil"
	"sync"
)

func greet() {
	fmt.Println("Hello, Simple World!")
}


func calculateSqrt(number float64) float64 {
	return math.Sqrt(number)
}

func displayResult() {
	num := 25.0
	result := calculateSqrt(num)
	fmt.Println("The square root of", num, "is", result)
}


func stringManipulation(original string) {
	upper := strings.ToUpper(original)
	contains := strings.Contains(original, "programming")
	fmt.Println("Original:", original)
	fmt.Println("Uppercase:", upper)
	fmt.Println("Contains 'programming':", contains)
}


func mainLogic() {
	text := "simple programming language"
	stringManipulation(text)
}


func countdownTimer(seconds interface{}) {
	for seconds.(int) > 0 {
		fmt.Println(seconds)
		time.Sleep(1 * time.Second)
		seconds = seconds.(int) - 1
	}
	fmt.Println("Liftoff!")
}


func startCountdown() {
	fmt.Println("Starting countdown...")
	countdownTimer(3)
}


func readEnv() {
	homeDir := os.Getenv("HOME")
	fmt.Println("Home Directory:", homeDir)
}


func createFile(filename string, content interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
	}
	file.WriteString(content.(string))
	file.Close()
	fmt.Println("File", filename, "created successfully.")
}


func mainLogic1() {
	readEnv()
	createFile("example.txt", "Hello from Simple!\n")
}


func fetchGitHubAPI(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return nil
	}
	defer	response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}
	return body
}

func displayResponse(body []byte) {
	if body != nil {
		fmt.Println("Response from GitHub API:")
		fmt.Println(string(body))
	} else {
		fmt.Println("No response to display.")
	}
}


func mainLogic2() {
	url := "https://api.github.com"
	responseBody := fetchGitHubAPI(url)
	displayResponse(responseBody)
}


func serializeData(data interface{}) []byte {
	jsonData, err := json.Marshal(data.(map[string]any))
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
	}
	return jsonData
}

func deserializeData(jsonStr []byte) map[string]any {
	decodedData := map[string]any{"placeholder1": "", "placeholder2": 0}
	json.Unmarshal(jsonStr, &decodedData)
	return decodedData
}

func processJSON() {
	data := map[string]any{"name": "Simple Language", "version": "1.0", "features": []string{"lexer", "parser", "semantic analyzer", "transformer", "code generator", }}
	jsonStr := serializeData(data)
	fmt.Println("Serialized JSON:")
	fmt.Println(string(jsonStr))
	decoded := deserializeData(jsonStr)
	fmt.Println("Deserialized Data:")
	fmt.Println(decoded)
}


func executeTask(id interface{}) {
	fmt.Println("Goroutine", id, "is running")
}


func startGoroutines() {
	wg := sync.WaitGroup{}
	numGoroutines := 3
	wg.Add(numGoroutines)
	work := func(id interface{}) {
		defer		wg.Done()
		executeTask(id.(int))
	}


	for i := range numGoroutines {
		go		work(i)
	}
	wg.Wait()
	fmt.Println("All goroutines have finished")
}


func worker(id int, jobs chan any, results chan any) {
	for j := range jobs {
		fmt.Println("Worker ", id, "started job ", j)
		time.Sleep(time.Second)
		fmt.Println("Worker ", id, "finished job ", j)
		results <- j.(int) * 2
	}
}


func main() {
	var a any
	a = 100
	fmt.Println(a)
	a = "hello"
	fmt.Println(a)
	greet()
	displayResult()
	mainLogic()
	startCountdown()
	mainLogic1()
	mainLogic2()
	processJSON()
	startGoroutines()
	numJobs := 5
	jobs := make(chan any, numJobs)
	results := make(chan any, numJobs)
	w := 1
	for w <= 3 {
		go		worker(w, jobs, results)
		w = w + 1
	}
	for jo := range numJobs {
		jobs <- jo
	}
	close(jobs)
	for _ = range numJobs {
		result := <- results
		fmt.Println("Result: ", result)
	}
}
