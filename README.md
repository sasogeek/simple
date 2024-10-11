# Simple Programming Language

Simple is an experimental programming language that compiles `.simple` files into Go binaries. It provides a Python-like syntax, allowing developers to write concise, readable code that gets converted into Go for performance.

This README will guide you through installing, using, and extending the Simple language.

Join the community on Slack for the latest updates:
[Join Slack](https://join.slack.com/t/simplecommunity/shared_invite/zt-2s32ctsx1-~kGgQvtwIlV39rCsR~ruug)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Syntax Guide](#syntax-guide)
  - [Variables](#variables)
  - [Control Flow](#control-flow)
  - [Data Types](#data-types)
  - [Printing](#printing)
  - [Functions](#functions)
    - [Example 1: `print` to Print Messages](#example-1-using-print-to-print-messages)
    - [Example 2: `math` for Math Operations](#example-2-using-math-for-mathematical-operations)
    - [Example 3: `strings` for String Manipulation](#example-3-using-strings-for-string-manipulation)
    - [Example 4: `time` for Time-Related Operations](#example-4-using-time-for-time-related-operations)
    - [Example 5: `os` for Operating System Interactions](#example-5-using-os-for-operating-system-interactions)
    - [Example 6: `net/http` to Make HTTP Requests](#example-6-using-nethttp-to-make-http-requests)
    - [Example 7: `encoding/json` for JSON Serialization and Deserialization](#example-7-using-encodingjson-for-json-serialization-and-deserialization)
    - [Example 8: `go routines` using goroutines](#example-8-goroutines-using-goroutines)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Python-like syntax**: Write familiar code, but get the performance benefits of Go.
- **Simple syntax**: Easier learning curve for beginners, with basic types like strings, arrays, and dictionaries.
- **Open Source**: Built by the community, for the community.

## Installation

### Prerequisites

To get started with Simple, you need the following installed:

1. **Go** (version 1.23+): Install from [here](https://golang.org/dl/).
2. **Git**: Used to clone the repository.

(The installation script will install Go for you if you're on macOS)

### Install Simple

1. Clone the Simple repository:

   ```bash
   git clone https://github.com/sasogeek/simple.git
   ```

2. Navigate to the project directory:

   ```bash
   cd simple
   ```

3. Run the installation script: (only tested on macOS)

   ```bash
   cd compiler
   sh install.sh
   ```

   This script will build the Simple compiler and move the binary to your system's `PATH`, making it available globally.

4. Verify the installation:

   ```bash
   simple --version
   ```

   This should display the version of the Simple compiler, confirming that it is installed correctly.

### VSCode Syntax Highlighting

A Visual Studio Code extension for Simple syntax highlighting is available to enhance the development experience.

#### Installing the VSCode Extension

1. Navigate to the `simple/simple-syntax/` folder in the repository.
2. Locate the `simple-syntax-0.1.0.vsix` file.
3. Open Visual Studio Code.
4. Go to the Extensions view by clicking the Extensions icon in the Activity Bar on the side of the window or pressing `Ctrl+Shift+X`.
5. Click on the `...` (ellipsis) in the top right corner of the Extensions view and select `Install from VSIX...`.
6. Browse to the `simple-syntax-0.1.0.vsix` file and click `Install`.

Once installed, VSCode will automatically apply syntax highlighting to `.simple` files.

## Quick Start

You can create a Simple program in a file with a `.simple` extension. Here's an example:

### hello_world.simple

```simple
print("Hello, World!")
```

Compile the Simple program into a Go binary and run it:

```bash
simple hello_world.simple
```

This will generate AND run the binary in a directory named after the file. To run the generated binary without recompiling it, run it like so:

```bash
./hello_world/hello_world
```
or 
```bash
cd hello_world
./hello_world
```
## Syntax Guide

### Variables

Variables in Simple are dynamically typed, meaning you can assign different types of values to variables without needing to declare the type,
and you may reassign a variable a value of a different type.
Rarely you may have to initialize a data structure if you want to use a go library function or method that expects
a particular format. [See example 7 in the functions section.(decodedData)](#example-7-using-encodingjson-for-json-serialization-and-deserialization)
```python
# String
name = "Simple"

# Integer
age = 5

# Array
arr = [1, 2, 3]

# Dictionary
person = {"name": "Alice", "age": 30}

my_variable = 100
print(my_variable)
my_variable = "hello"
print(my_variable)

```

### Control Flow

#### If Statements

```python
age = 25

if age >= 18:
    print("You are an adult")
else:
    print("You are not an adult")
```

#### While Loops

```python
counter = 0

while counter < 5:
    print(counter)
    counter = counter + 1
```

#### For Loops

```python
arr = [1, 2, 3]

for index in arr:
    print(arr[index])
```

### Data Types

- **String**: A sequence of characters, e.g., `"Hello"`.
- **Integer**: A whole number, e.g., `5`.
- **Float**: A floating-point number, e.g., `3.14`.
- **Array**: A collection of values, e.g., `[1, 2, 3]`.
- **Dictionary**: A key-value pair collection, e.g., `{"key": "value"}`.

### Printing

The `print()` function works similarly to Python, outputting to the console:

```python
print("Hello, World!")
```

You can print variables or expressions as well:

```python
name = "Simple"
print("Hello, " + name)
```


### Functions

Here are **7 examples** demonstrating the usage of different Go packages in Simple, written with Python-like syntax.

#### Example 1: Using `print` to Print Messages

```python
import "fmt"

def greet():
    print("Hello, Simple World!")

greet()
```

#### Example 2: Using `math` for Mathematical Operations

```python
import "math"

def calculateSqrt(number):
    return math.Sqrt(float64(number))

def displayResult():
    num = 25.0
    result = calculateSqrt(num)
    print("The square root of", num, "is", result)

displayResult()
```

#### Example 3: Using `strings` for String Manipulation

```python
import "strings"

def stringManipulation(original):
    upper = strings.ToUpper(original)
    contains = strings.Contains(original, "programming")
    print("Original:", original)
    print("Uppercase:", upper)
    print("Contains 'programming':", contains)

def mainLogic():
    text = "simple programming language"
    stringManipulation(text)

mainLogic()
```

#### Example 4: Using `time` for Time-Related Operations

```python
import "time"

def countdownTimer(seconds):
    while seconds > 0:
        print(seconds)
        time.Sleep(1 * time.Second)
        seconds = seconds - 1
    print("Liftoff!")

def startCountdown():
    print("Starting countdown...")
    countdownTimer(3)

startCountdown()
```

#### Example 5: Using `os` for Operating System Interactions

```python
import "os"

def readEnv():
    homeDir = os.Getenv("HOME")
    print("Home Directory:", homeDir)

def createFile(filename, content):
    file, err = os.Create(filename)
    if err != nil:
        print("Error creating file:", err)
        return
    file.WriteString(content)
    file.Close()
    print("File", filename, "created successfully.")

def mainLogic():
    readEnv()
    createFile("example.txt", "Hello from Simple!\n")

mainLogic()
```

#### Example 6: Using `net/http` to Make HTTP Requests

```python
import "net/http"
import "io/ioutil"

def fetchGitHubAPI(url):
    response, err = http.Get(url)
    if err != nil:
        print("Error making GET request:", err)
        return nil
    defer response.Body.Close()
    body, err = ioutil.ReadAll(response.Body)
    if err != nil:
        print("Error reading response body:", err)
        return nil
    return body

def displayResponse(body):
    if body != nil:
        print("Response from GitHub API:")
        print(string(body))
    else:
        print("No response to display.")

def mainLogic():
    url = "https://api.github.com"
    responseBody = fetchGitHubAPI(url)
    displayResponse(responseBody)

mainLogic()
```

#### Example 7: Using `encoding/json` for JSON Serialization and Deserialization

```python
import "encoding/json"

def serializeData(data):
    jsonData, err = json.Marshal(data)
    if err != nil:
        print("Error marshaling JSON:", err)
    return jsonData

def deserializeData(jsonStr):
    decodedData = {"placeholder": "", "dict": 0}
    json.Unmarshal(jsonStr, &decodedData)
    return decodedData

def processJSON():
    data = {"name": "Simple Language","version": "1.0","features": ["lexer", "parser", "semantic analyzer", "transformer", "code generator"]}
    jsonStr = serializeData(data)
    print("Serialized JSON:")
    print(jsonStr)

    decoded = deserializeData(jsonStr)
    print("Deserialized Data:")
    print(decoded)

processJSON()

```


#### Example 8: `goroutines!` Using goroutines

```python
import "fmt"
import "sync"

def executeTask(id):
    fmt.Println("Goroutine", id, "is running")

def startGoroutines():
    wg = sync.WaitGroup{}
    numGoroutines = 3
    wg.Add(numGoroutines)
    
    def work(id):
        defer wg.Done()
        executeTask(id)

    for i in numGoroutines:
        go work(i)

    wg.Wait()
    print("All goroutines have finished")

startGoroutines()


```


## Contributing

We welcome contributions from the community! To contribute:

1. Fork the repository on GitHub.
2. Create a new branch for your feature or bug fix.
3. Submit a pull request with a description of your changes.

Feel free to open an issue if you find a bug or have suggestions for new features.

## License

Simple is open-source software licensed under the MIT License. See the [LICENSE](https://opensource.org/license/mit) file for details.