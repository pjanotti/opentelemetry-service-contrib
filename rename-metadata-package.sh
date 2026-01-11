#!/bin/bash

set -e

# Usage: ./rename-metadata-package.sh <path_to_metadata_package>
# Example: ./rename-metadata-package.sh receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata

if [ -z "$1" ]; then
    echo "Usage: $0 <path_to_metadata_package>"
    echo "Example: $0 receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata"
    exit 1
fi

METADATA_PATH="$1"

# Validate the path exists
if [ ! -d "$METADATA_PATH" ]; then
    echo "Error: Directory $METADATA_PATH does not exist"
    exit 1
fi

# Remove trailing slash if present
METADATA_PATH="${METADATA_PATH%/}"

# Validate the path ends with /internal/metadata
if [[ ! "$METADATA_PATH" =~ /internal/metadata$ ]]; then
    echo "Error: Path must end with /internal/metadata"
    exit 1
fi

# Extract the scraper directory (e.g., receiver/hostmetricsreceiver/internal/scraper/cpuscraper)
SCRAPER_DIR="${METADATA_PATH%/internal/metadata}"

# Extract the scraper name (e.g., cpuscraper)
SCRAPER_NAME=$(basename "$SCRAPER_DIR")

# Generate the new package name by removing "scraper" suffix and adding "metadata"
# e.g., cpuscraper -> cpumetadata
NEW_PACKAGE_NAME="${SCRAPER_NAME%scraper}metadata"

# Define the new path
NEW_METADATA_PATH="${SCRAPER_DIR}/internal/${NEW_PACKAGE_NAME}"

# Define the metadata.yaml path
METADATA_YAML="${SCRAPER_DIR}/metadata.yaml"

# Extract component directory (first two path components)
COMPONENT_DIR=$(echo "$METADATA_PATH" | cut -d'/' -f1-2)

echo "=========================================="
echo "Metadata Package Rename Configuration"
echo "=========================================="
echo "Scraper directory: $SCRAPER_DIR"
echo "Scraper name: $SCRAPER_NAME"
echo "New package name: $NEW_PACKAGE_NAME"
echo "Old path: $METADATA_PATH"
echo "New path: $NEW_METADATA_PATH"
echo "Metadata YAML: $METADATA_YAML"
echo "Component directory: $COMPONENT_DIR"
echo "=========================================="
echo ""

# Confirm before proceeding
read -p "Proceed with renaming? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

# Step 1: Git mv the directory
echo ""
echo "Step 1: Renaming directory with git mv..."
git mv "$METADATA_PATH" "$NEW_METADATA_PATH"
echo "✓ Directory renamed"

# Step 2: Update package declarations in the moved files
echo ""
echo "Step 2: Updating package declarations in moved files..."
GO_FILES_IN_NEW_DIR=$(find "$NEW_METADATA_PATH" -name "*.go" -type f)
if [ -z "$GO_FILES_IN_NEW_DIR" ]; then
    echo "Warning: No .go files found in $NEW_METADATA_PATH"
else
    for file in $GO_FILES_IN_NEW_DIR; do
        if grep -q "^package metadata$" "$file"; then
            echo "  Updating package declaration in: $file"
            sed -i.bak "s/^package metadata$/package ${NEW_PACKAGE_NAME}/" "$file"
            rm "${file}.bak"
        fi
    done
    echo "✓ Package declarations updated"
fi

# Step 3: Add generated_package_name to metadata.yaml
echo ""
echo "Step 3: Updating metadata.yaml..."
if [ ! -f "$METADATA_YAML" ]; then
    echo "Warning: $METADATA_YAML not found, skipping..."
else
    # Check if generated_package_name already exists
    if grep -q "generated_package_name:" "$METADATA_YAML"; then
        echo "Warning: generated_package_name already exists in $METADATA_YAML"
        echo "  Current value: $(grep "generated_package_name:" "$METADATA_YAML")"
        read -p "  Update it to ${NEW_PACKAGE_NAME}? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            sed -i.bak "s/generated_package_name:.*/generated_package_name: ${NEW_PACKAGE_NAME}/" "$METADATA_YAML"
            rm "${METADATA_YAML}.bak"
            echo "✓ Updated generated_package_name in $METADATA_YAML"
        fi
    else
        # Find the line with "type:" and add generated_package_name after it
        if grep -q "^type:" "$METADATA_YAML"; then
            sed -i.bak "/^type:/a\\
generated_package_name: ${NEW_PACKAGE_NAME}
" "$METADATA_YAML"
            rm "${METADATA_YAML}.bak"
            echo "✓ Added generated_package_name to $METADATA_YAML"
        else
            echo "Warning: Could not find 'type:' in $METADATA_YAML"
            echo "  Please add manually: generated_package_name: ${NEW_PACKAGE_NAME}"
        fi
    fi
fi

# Step 4: Run make to regenerate code
echo ""
echo "Step 4: Running make -C $COMPONENT_DIR generate to regenerate code..."
if make -C "$COMPONENT_DIR" generate; then
    echo "✓ Code regenerated successfully"
else
    echo "⚠ Make command failed. You may need to run it manually:"
    echo "  make -C $COMPONENT_DIR generate"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted. Please fix the make errors and rerun the script."
        exit 1
    fi
fi

# Step 5: Update all package references
echo ""
echo "Step 5: Updating package references (imports and code)..."

# Find all .go files under the scraper directory
GO_FILES=$(find "$SCRAPER_DIR" -name "*.go" -type f)

if [ -z "$GO_FILES" ]; then
    echo "Warning: No .go files found under $SCRAPER_DIR"
else
    echo "  Scanning files for references to update..."
    FILES_UPDATED=0
    for file in $GO_FILES; do
        NEEDS_UPDATE=false
        
        # Check if file contains references to the old package
        if grep -q "internal/metadata" "$file" 2>/dev/null; then
            NEEDS_UPDATE=true
        fi
        if grep -q "\bmetadata\." "$file" 2>/dev/null; then
            NEEDS_UPDATE=true
        fi
        
        if [ "$NEEDS_UPDATE" = true ]; then
            echo "  Updating: $file"
            # Replace import path
            sed -i.bak "s|${SCRAPER_DIR}/internal/metadata|${SCRAPER_DIR}/internal/${NEW_PACKAGE_NAME}|g" "$file"
            # Replace package references (metadata. -> newpackagename.)
            # Using word boundary to avoid replacing things like "mymetadata."
            sed -i.bak "s|\([^a-zA-Z0-9_]\)metadata\.\([A-Z]\)|\1${NEW_PACKAGE_NAME}.\2|g" "$file"
            rm "${file}.bak"
            FILES_UPDATED=$((FILES_UPDATED + 1))
        fi
    done
    echo "✓ Updated $FILES_UPDATED file(s)"
fi

make -C "$COMPONENT_DIR" generate
make -C "$COMPONENT_DIR"

echo ""
echo "=========================================="
echo "✓ Done! Package renamed from 'metadata' to '${NEW_PACKAGE_NAME}'"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Review the changes:"
echo "     git diff"
echo ""
echo "  2. Test the changes:"
echo "     make -C $COMPONENT_DIR test"
echo ""
echo "  3. If everything looks good, stage and commit:"
echo "     git add ."
echo "     git commit -m \"[${COMPONENT_DIR}] Rename ${SCRAPER_NAME} metadata package to ${NEW_PACKAGE_NAME}\""
echo ""
