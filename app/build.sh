#!/bin/bash

# Display help function
show_help() {
    cat << EOF
Atom Interpreter Build Script
=============================

USAGE:
    ./build.sh [OPTIONS]

DESCRIPTION:
    This script builds the Atom interpreter with various configuration options.
    In normal mode, it creates a development build for the current platform.
    In release mode, it creates optimized builds with packaging for distribution.

OPTIONS:
    --help, -h          Show this help message and exit

    --release           Enable release mode with optimized builds and packaging
                        Creates archives with lib/, test/ folders and atom.ico

RELEASE MODE OPTIONS:
    -t <bits>           Target architecture bit width
                        32    Build for 32-bit systems (GOARCH=386)
                        64    Build for 64-bit systems (GOARCH=amd64)

    --arch <arch>       Target processor architecture
                        amd   AMD/Intel x86 processors (default)
                        arm   ARM processors (GOARCH=arm)
                        riscv RISC-V processors (GOARCH=riscv64)
                        wasm  WebAssembly target (GOOS=js GOARCH=wasm)

    --os <os>           Target operating system
                        linux Linux systems (default)
                        mac   macOS systems (GOOS=darwin)
                        win   Windows systems (GOOS=windows)

EXAMPLES:
    # Development build for current platform
    ./build.sh

    # Basic release build (64-bit Linux AMD/Intel)
    ./build.sh --release

    # 32-bit Windows release
    ./build.sh --release -t 32 --os win

    # ARM Linux release
    ./build.sh --release --arch arm --os linux

    # macOS release
    ./build.sh --release --os mac

    # WebAssembly build
    ./build.sh --release --arch wasm

    # RISC-V 64-bit Linux release
    ./build.sh --release --arch riscv --os linux

OUTPUT:
    Development Mode:
        - Creates executable: atom.linux, atom.macos, or atom.exe
        - Based on current platform detection

    Release Mode:
        - Creates timestamped archive: atom-release-[arch]-[os]-YYYYMMDD-HHMMSS.tar.gz/.zip
        - Includes: executable, lib/ folder, test/ folder (*.atom files), atom.ico
        - Linux targets create .tar.gz archives
        - Windows/macOS targets create .zip archives
        - Cleans up temporary release/ folder after packaging

NOTES:
    - Requires Go toolchain to be installed and in PATH
    - Cross-compilation requires appropriate Go cross-compile support
    - WebAssembly builds create .wasm files for browser/Node.js execution
    - Release archives are created in the current directory
    - Previous release archives are automatically cleaned up

EOF
}

# Check for help flag
if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    show_help
    exit 0
fi

echo "Building Atom interpreter..."

# Check for --release flag and architecture flag
RELEASE_MODE=false
ARCH_MODE=""
TARGET_ARCH=""
TARGET_OS="linux"
if [[ "$1" == "--release" ]]; then
    RELEASE_MODE=true
    echo "Release mode enabled"
    
    # Parse additional flags
    shift # Remove --release from arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t)
                if [[ "$2" == "32" ]]; then
                    ARCH_MODE="32"
                    echo "32-bit build enabled"
                elif [[ "$2" == "64" ]]; then
                    ARCH_MODE="64"
                    echo "64-bit build enabled"
                else
                    echo "Invalid architecture specified. Use 32 or 64."
                    exit 1
                fi
                shift 2
                ;;
            --arch)
                if [[ "$2" == "arm" ]]; then
                    TARGET_ARCH="arm"
                    echo "ARM architecture build enabled"
                elif [[ "$2" == "riscv" ]]; then
                    TARGET_ARCH="riscv"
                    echo "RISC-V architecture build enabled"
                elif [[ "$2" == "wasm" ]]; then
                    TARGET_ARCH="wasm"
                    echo "WebAssembly build enabled"
                else
                    # Default to amd/intel if not specified or invalid
                    TARGET_ARCH="amd"
                    echo "Using default AMD/Intel architecture build"
                fi
                shift 2
                ;;
            --os)
                if [[ "$2" == "mac" ]]; then
                    TARGET_OS="mac"
                    echo "macOS build enabled"
                elif [[ "$2" == "win" ]]; then
                    TARGET_OS="win"
                    echo "Windows build enabled"
                elif [[ "$2" == "linux" ]]; then
                    TARGET_OS="linux"
                    echo "Linux build enabled"
                else
                    echo "Invalid OS specified. Use mac|win|linux."
                    exit 1
                fi
                shift 2
                ;;
            *)
                echo "Unknown option: $1"
                echo "Use --help for usage information."
                exit 1
                ;;
        esac
    done
    
    # Set default architecture if not specified
    if [[ -z "$TARGET_ARCH" ]]; then
        TARGET_ARCH="amd"
        echo "Using default AMD/Intel architecture build"
    fi
fi

# Detect OS and set appropriate executable name
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    EXECUTABLE="atom.linux"
    echo "Building for Linux..."
elif [[ "$OSTYPE" == "darwin"* ]]; then
    EXECUTABLE="atom.macos"
    echo "Building for macOS..."
elif [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
    EXECUTABLE="atom.exe"
    echo "Building for Windows..."
else
    EXECUTABLE="atom.unknown"
    echo "Building for unknown OS, using default executable name..."
fi

if [[ "$RELEASE_MODE" == true ]]; then
    # Release mode
    echo "Creating release build..."
    
    # Delete previous release archives
    echo "Cleaning up previous releases..."
    rm -f atom-release-*.tar.gz atom-release-*.zip
    
    # Create release folder if it doesn't exist
    mkdir -p release
    
    # Copy lib and test folders to release folder
    if [[ -d "lib" ]]; then
        echo "Copying lib folder..."
        cp -r lib release/
    fi
    
    if [[ -d "test" ]]; then
        echo "Copying test folder (only .atom files)..."
        mkdir -p release/test
        find test -name "*.atom" -exec cp {} release/test/ \;
    fi
    
    # Copy atom.ico file if it exists
    if [[ -f "atom.ico" ]]; then
        echo "Copying atom.ico..."
        cp atom.ico release/
    fi
    
    # Set build environment variables for architecture
    BUILD_CMD="go build"
    ARCH_SUFFIX=""
    if [[ "$ARCH_MODE" == "32" ]]; then
        export GOARCH=386
        BUILD_CMD="GOARCH=386 go build"
        ARCH_SUFFIX="-32bit"
        echo "Building for 32-bit architecture..."
    elif [[ "$ARCH_MODE" == "64" ]]; then
        export GOARCH=amd64
        BUILD_CMD="GOARCH=amd64 go build"
        ARCH_SUFFIX="-64bit"
        echo "Building for 64-bit architecture..."
    fi
    
    # Set target OS environment variables
    OS_SUFFIX=""
    if [[ "$TARGET_OS" == "mac" ]]; then
        export GOOS=darwin
        OS_SUFFIX="-mac"
        EXECUTABLE="atom.macos"
        echo "Building for macOS..."
    elif [[ "$TARGET_OS" == "win" ]]; then
        export GOOS=windows
        OS_SUFFIX="-win"
        EXECUTABLE="atom.exe"
        echo "Building for Windows..."
    elif [[ "$TARGET_OS" == "linux" ]]; then
        export GOOS=linux
        OS_SUFFIX="-linux"
        EXECUTABLE="atom.linux"
        echo "Building for Linux..."
    fi
    
    # Set target architecture environment variables
    TARGET_SUFFIX=""
    if [[ "$TARGET_ARCH" == "arm" ]]; then
        export GOARCH=arm
        BUILD_CMD="GOOS=$GOOS GOARCH=arm go build"
        TARGET_SUFFIX="-arm"
        EXECUTABLE="atom.arm"
        echo "Building for ARM architecture..."
    elif [[ "$TARGET_ARCH" == "riscv" ]]; then
        export GOARCH=riscv64
        BUILD_CMD="GOOS=$GOOS GOARCH=riscv64 go build"
        TARGET_SUFFIX="-riscv"
        EXECUTABLE="atom.riscv"
        echo "Building for RISC-V architecture..."
    elif [[ "$TARGET_ARCH" == "wasm" ]]; then
        export GOOS=js
        export GOARCH=wasm
        BUILD_CMD="GOOS=js GOARCH=wasm go build"
        TARGET_SUFFIX="-wasm"
        EXECUTABLE="atom.wasm"
        echo "Building for WebAssembly..."
    elif [[ "$TARGET_ARCH" == "amd" ]]; then
        # Default AMD/Intel - use existing GOARCH if set by ARCH_MODE, otherwise default to amd64
        if [[ -z "$GOARCH" ]]; then
            export GOARCH=amd64
        fi
        BUILD_CMD="GOOS=$GOOS GOARCH=$GOARCH go build"
        TARGET_SUFFIX="-amd"
        echo "Building for AMD/Intel architecture..."
    fi
    
    # Build the Go application in release folder
    eval "$BUILD_CMD -o release/$EXECUTABLE ."
    
    # Create archive file based on target OS with architecture suffix
    if [[ "$TARGET_OS" == "win" ]]; then
        ZIP_NAME="atom-release${ARCH_SUFFIX}${TARGET_SUFFIX}${OS_SUFFIX}-$(date +%Y%m%d-%H%M%S).zip"
        echo "Creating zip file: $ZIP_NAME"
        cd release && zip -r "../$ZIP_NAME" . && cd ..
        ARCHIVE_NAME="$ZIP_NAME"
    else
        ARCHIVE_NAME="atom-release${ARCH_SUFFIX}${TARGET_SUFFIX}${OS_SUFFIX}-$(date +%Y%m%d-%H%M%S).tar.gz"
        echo "Creating tar archive: $ARCHIVE_NAME"
        tar -czf "$ARCHIVE_NAME" -C release .
    fi
    
    # Clean up release folder
    echo "Cleaning up release folder..."
    rm -rf release
    
    echo "Release build complete! Created $ARCHIVE_NAME"
else
    # Normal build mode
    go build -o "$EXECUTABLE" .
    echo "Build complete! $EXECUTABLE created."
fi