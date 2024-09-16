#!/bin/bash

# Function to increment version
increment_version() {
    local version=$1
    local major minor patch

    # Split version into components
    IFS='.' read -r major minor patch <<< "$version"

    # Increment patch version
    patch=$((patch + 1))

    # Return new version
    echo "$major.$minor.$patch"
}

# Get current version from main.go (line 170)
current_version=$(grep -o '"version": "[^"]*"' main.go | sed 's/"version": "//;s/"//')

if [ -z "$current_version" ]; then
    echo "Error: Unable to extract version from main.go"
    exit 1
fi

# Increment version
new_version=$(increment_version "$current_version")

# Update version in main.go (line 170)
sed -i.bak '170s/"version": "[^"]*"/"version": "'"$new_version"'"/' main.go

echo "Updated version to $new_version"

# Build the project
go build -o "pls"
tar -czf "pls-$new_version.tar.gz" "pls"

cp "pls" /Users/adam/Documents/GitHub/pls-site/public/pls
cp "pls-$new_version.tar.gz" /Users/adam/Documents/GitHub/pls-site/public

# Update install.sh with new version
install_sh_path="/Users/adam/Documents/GitHub/pls-site/public/install.sh"
sed -i.bak "s/VERSION=\".*\"/VERSION=\"$new_version\"/" "$install_sh_path"
echo "Updated install.sh with new version: $new_version"

# Update version in usage.js
usage_js_path="/Users/adam/Documents/GitHub/pls-site/functions/usage.js"
sed -i.bak "s/if (body\.version != \"[^\"]*\")/if (body.version != \"$new_version\")/" "$usage_js_path"
echo "Updated usage.js with new version: $new_version"

echo "Built and copied pls and pls-$new_version.tar.gz to /Users/adam/Documents/GitHub/pls-site/public/pls"

# Remove the backup file created by sed
rm main.go.bak
rm "$install_sh_path.bak"
rm "$usage_js_path.bak"