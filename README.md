# Simple Programming Language

Simple is an experimental programming language that compiles `.simple` files into Go binaries. It provides a Python-like syntax, allowing developers to write concise, readable code that gets converted into Go for performance.

This README will guide you through installing, using, and extending the Simple language.

Join the community on Slack for the latest updates;
[Join Slack](https://join.slack.com/t/simplecommunity/shared_invite/zt-2s32ctsx1-~kGgQvtwIlV39rCsR~ruug)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Syntax Guide](#syntax-guide)
  - [Variables](#variables)
  - [Control Flow](#control-flow)
  - [Functions](#functions)
  - [Data Types](#data-types)
  - [Built-in Methods](#built-in-methods)
    - [String Methods](#string-methods)
    - [Array Methods](#array-methods)
    - [Dictionary Methods](#dictionary-methods)
  - [Printing](#printing)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Python-like syntax**: Write familiar code, but get the performance benefits of Go.
- **Simple syntax**: Easier learning curve for beginners, with basic types like strings, arrays, and dictionaries.
- **Method support**: In-place and non-in-place methods for strings, arrays, and dictionaries.
- **Open Source**: Built by the community, for the community.

## Installation

### Prerequisites

To get started with Simple, you need the following installed:

1. **Go** (version 1.23+): Install from [here](https://golang.org/dl/).
2. **Git**: Used to clone the repository.

(The installation script will installs go for you, if you're on macos)
### Install Simple

1. Clone the Simple repository:

   ```bash
   git clone https://github.com/sasogeek/simple.git
   ```

2. Navigate to the project directory:

   ```bash
   cd simple
   ```

3. Run the installation script: (only tested on macos)

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

```python
print("Hello, World!")
```

Compile the Simple program into a Go binary and run it:

```bash
simple hello_world.simple
```
This will generate AND run the binary. To run the generated binary without recompiling it,
run it like so;

```bash
./hello_world
```

## Syntax Guide

### Variables

Variables in Simple are dynamically typed, meaning you can assign different types of values to variables without needing to declare the type.

```python
# String
name = "Simple"

# Integer
age = 5

# Array
arr = [1, 2, 3]

# Dictionary
person = {"name": "Alice", "age": 30}
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

for num in arr:
    print(num)
```

### Data Types

- **String**: A sequence of characters, e.g., `"Hello"`.
- **Integer**: A whole number, e.g., `5`.
- **Float**: A floating-point number, e.g., `3.14`.
- **Array**: A collection of values, e.g., `[1, 2, 3]`.
- **Dictionary**: A key-value pair collection, e.g., `{"key": "value"}`.

### Built-in Methods

#### String Methods

- `upper()`: Converts the string to uppercase.
- `lower()`: Converts the string to lowercase.
- `replace(old, new)`: Replaces occurrences of `old` with `new`.
- `split(separator)`: Splits the string into an array based on the separator.
- `find(substring)`: Returns the index of the first occurrence of `substring`, or `-1` if not found.
- `strip()`: Removes leading and trailing whitespace.
- `startswith(prefix)`: Returns `True` if the string starts with `prefix`.
- `endswith(suffix)`: Returns `True` if the string ends with `suffix`.

#### Array Methods

- `append(item)`: Adds `item` to the end of the array.
- `extend(another_array)`: Extends the array by appending all elements from `another_array`.

#### Dictionary Methods

- `update(another_dict)`: Updates the dictionary with key-value pairs from `another_dict`.

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

## Contributing

We welcome contributions from the community! To contribute:

1. Fork the repository on GitHub.
2. Create a new branch for your feature or bug fix.
3. Submit a pull request with a description of your changes.

Feel free to open an issue if you find a bug or have suggestions for new features.

## License

Simple is open-source software licensed under the MIT License. See the [LICENSE](https://opensource.org/license/mit) file for details.
