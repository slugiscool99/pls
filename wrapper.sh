#!/bin/bash

# Define the log file path
LOG_FILE="$HOME/terminal_output.log"

# Run the original command and capture the output
"$@" &> >(tee -a "$LOG_FILE")