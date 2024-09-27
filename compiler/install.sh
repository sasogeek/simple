#!/bin/bash

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

# Build the binary
go build -o ~/simple/simple

# Check if ~/simple is in the PATH
if [[ ":$PATH:" != *":$HOME/simple:"* ]]; then
  echo "Adding ~/simple to your PATH"
  # Add it to the current session's PATH
  export PATH=$PATH:~/simple

  # Optionally, persist this by adding it to .bashrc or .zshrc
  if [[ -n "$BASH_VERSION" ]]; then
    echo 'export PATH=$PATH:~/simple' >> ~/.bashrc  # for bash users
    echo "Added ~/simple to .bashrc"
  elif [[ -n "$ZSH_VERSION" ]]; then
    echo 'export PATH=$PATH:~/simple' >> ~/.zshrc  # for zsh users
    echo "Added ~/simple to .zshrc"
  fi
fi

echo "The 'simple' compiler is now available in your PATH"

