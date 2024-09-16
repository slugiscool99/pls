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
current_version=$(sed -n '170s/.*"version": "\([^"]*\)".*/\1/p' main.go)

# Increment version
new_version=$(increment_version "$current_version")

# Update version in main.go (line 170)
sed -i.bak '170s/"version": "[^"]*"/"version": "'"$new_version"'"/' main.go

echo "Updated version to $new_version"

# Build the project
go build -o "pls-$new_version"
tar -czf "pls-$new_version.tar.gz" "pls-$new_version"

cp "pls-$new_version" /Users/adam/Documents/GitHub/pls-site/public/pls
cp "pls-$new_version.tar.gz" /Users/adam/Documents/GitHub/pls-site/public

# Update install.sh with new version
install_sh_path="/Users/adam/Documents/GitHub/pls-site/public/install.sh"
sed -i.bak "s/DOWNLOAD_URL=\"https:\/\/pls\.mom\/pls-.*\.tar\.gz\"/DOWNLOAD_URL=\"https:\/\/pls\.mom\/pls-$new_version.tar.gz\"/" "$install_sh_path"
echo "Updated install.sh with new version: $new_version"

echo "Built and copied pls-$new_version and pls-$new_version.tar.gz to /Users/adam/Documents/GitHub/pls-site/public/pls"

# Remove the backup file created by sed
rm main.go.bak
rm "$install_sh_path.bak"