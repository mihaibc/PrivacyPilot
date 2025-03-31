#!/bin/bash

# Script to re-initialize Go modules for PrivacyPilot services/adapters
# WARNING: This will delete existing go.mod and go.sum files in the target directories.

# Define Go module directories and their desired module names
declare -A GO_MODULES
GO_MODULES=(
    ["./services/api-gateway"]="privacypilot-api-gateway"
    ["./services/anonymizer-service"]="privacypilot-anonymizer-service"
    ["./services/ai-coordinator"]="privacypilot-ai-coordinator"
    ["./ai-adapters/ollama-adapter"]="privacypilot-ollama-adapter"
    # Add other Go services/adapters here if created later
    # ["./ai-adapters/some-other-go-adapter"]="privacypilot-other-adapter"
)

# Define common dependencies to fetch (adjust versions if needed)
COMMON_DEPS=(
    "github.com/gin-gonic/gin@latest"
    "github.com/stretchr/testify/assert@latest"
)

# Specific dependencies per module (add more as needed)
declare -A SPECIFIC_DEPS
SPECIFIC_DEPS["./ai-adapters/ollama-adapter"]="github.com/ollama/ollama/api@latest"

# Get the absolute path of the script's directory to ensure correct relative paths
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT="$SCRIPT_DIR/.." # Assumes script is in 'scripts' dir at project root

# Function to initialize a module
init_module() {
    local mod_dir=$1
    local mod_name=$2
    local full_path="$PROJECT_ROOT/$mod_dir"

    echo "--------------------------------------------------"
    echo "Processing Module: $mod_name in $mod_dir"
    echo "--------------------------------------------------"

    if [ ! -d "$full_path" ]; then
        echo "Directory $full_path does not exist. Skipping."
        return
    fi

    # Navigate to the directory
    pushd "$full_path" > /dev/null || { echo "Failed to enter directory $full_path. Skipping."; return; }

    # Remove existing mod/sum files
    echo "Removing old go.mod and go.sum (if they exist)..."
    rm -f go.mod go.sum

    # Initialize the module
    echo "Running 'go mod init $mod_name'..."
    go mod init "$mod_name"
    if [ $? -ne 0 ]; then
        echo "ERROR: Failed to initialize module $mod_name."
        popd > /dev/null
        return
    fi

    # Get common dependencies
    echo "Getting common dependencies..."
    for dep in "${COMMON_DEPS[@]}"; do
        echo "  go get $dep"
        go get "$dep"
        if [ $? -ne 0 ]; then
            echo "WARN: Failed to get common dependency $dep for $mod_name."
            # Continue anyway
        fi
    done

    # Get specific dependencies for this module
    if [ -n "${SPECIFIC_DEPS[$mod_dir]}" ]; then
        local specific_dep="${SPECIFIC_DEPS[$mod_dir]}"
        echo "Getting specific dependency: $specific_dep..."
        echo "  go get $specific_dep"
        go get "$specific_dep"
         if [ $? -ne 0 ]; then
            echo "WARN: Failed to get specific dependency $specific_dep for $mod_name."
            # Continue anyway
        fi
    fi


    # Tidy up the module
    echo "Running 'go mod tidy'..."
    go mod tidy
    if [ $? -ne 0 ]; then
        echo "WARN: 'go mod tidy' failed or had issues for $mod_name."
    fi

    echo "Module $mod_name initialization complete."
    echo ""

    # Return to the original directory
    popd > /dev/null
}

# Iterate over the defined modules and initialize them
for dir in "${!GO_MODULES[@]}"; do
    init_module "$dir" "${GO_MODULES[$dir]}"
done

echo "=================================================="
echo "Go module re-initialization process finished."
echo "=================================================="