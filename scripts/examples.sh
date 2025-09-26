#!/bin/bash

# This script is used to generate an SVG animation of magi-cli usage.

# The -n flag is for no-execute, just to show the commands

# Show the version
echo "magi --version"
./magi --version

# Show the help
echo "magi --help"
./magi --help

# Run the setup non-interactively
# For the animation, we just show the command.
echo "magi setup --api-key <YOUR_API_KEY> --model gpt-4 --format text"
./magi setup --api-key YOUR_API_KEY --model gpt-4 --format text
