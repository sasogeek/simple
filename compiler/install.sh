#!/bin/bash

# Define the build output directory
OUTPUT_DIR=~/simple

# Get the directory where the script is located
SCRIPT_DIR=$(dirname "$(realpath "$0")")

# Check if Go is installed
if ! command -v go &> /dev/null
then
  echo "Go is not installed. Installing Go with Homebrew..."

  # Check if Homebrew is installed
  if ! command -v brew &> /dev/null
  then
    echo "Homebrew is not installed. Installing Homebrew..."
    # Install Homebrew
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

    # Add Homebrew to PATH if needed
    echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.bash_profile
    eval "$(/opt/homebrew/bin/brew shellenv)"
  fi

  # Install Go using Homebrew
  brew install go
else
  echo "Go is already installed."
fi

rm -r $OUTPUT_DIR
# Create the output directory if it doesn't exist
mkdir -p $OUTPUT_DIR

# Copy all files from the script's directory to the output directory
echo "Copying files from $SCRIPT_DIR to $OUTPUT_DIR..."
cp -R "$SCRIPT_DIR"/* $OUTPUT_DIR/

# Change directory to the output directory before building
cd $OUTPUT_DIR

# Build the binary
go build -o $OUTPUT_DIR/simple

# Check if the output directory is in the PATH
if [[ ":$PATH:" != *":$OUTPUT_DIR:"* ]]; then
  echo "Adding $OUTPUT_DIR to your PATH"
  # Add it to the current session's PATH
  export PATH=$PATH:$OUTPUT_DIR

  # Optionally, persist this by adding it to .bashrc or .zshrc
  if [[ -n "$BASH_VERSION" ]]; then
    echo "export PATH=\$PATH:$OUTPUT_DIR" >> ~/.bashrc  # for bash users
    echo "Added $OUTPUT_DIR to .bashrc"
  elif [[ -n "$ZSH_VERSION" ]]; then
    echo "export PATH=\$PATH:$OUTPUT_DIR" >> ~/.zshrc  # for zsh users
    echo "Added $OUTPUT_DIR to .zshrc"
  fi
fi

echo "The 'simple' compiler is now available in your PATH"



##!/bin/bash
#
## Check if Go is installed
#if ! command -v go &> /dev/null
#then
#  echo "Go is not installed. Installing Go with Homebrew..."
#
#  # Check if Homebrew is installed
#  if ! command -v brew &> /dev/null
#  then
#    echo "Homebrew is not installed. Installing Homebrew..."
#    # Install Homebrew
#    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
#
#    # Add Homebrew to PATH if needed
#    echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.bash_profile
#    eval "$(/opt/homebrew/bin/brew shellenv)"
#  fi
#
#  # Install Go using Homebrew
#  brew install go
#else
#  echo "Go is already installed."
#fi
#
## Build the binary
#go build -o ~/simple/simple
#
## Check if ~/simple is in the PATH
#if [[ ":$PATH:" != *":$HOME/simple:"* ]]; then
#  echo "Adding ~/simple to your PATH"
#  # Add it to the current session's PATH
#  export PATH=$PATH:~/simple
#
#  # Optionally, persist this by adding it to .bashrc or .zshrc
#  if [[ -n "$BASH_VERSION" ]]; then
#    echo 'export PATH=$PATH:~/simple' >> ~/.bashrc  # for bash users
#    echo "Added ~/simple to .bashrc"
#  elif [[ -n "$ZSH_VERSION" ]]; then
#    echo 'export PATH=$PATH:~/simple' >> ~/.zshrc  # for zsh users
#    echo "Added ~/simple to .zshrc"
#  fi
#fi
#
#echo "The 'simple' compiler is now available in your PATH"
