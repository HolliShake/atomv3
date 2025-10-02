@echo off
setlocal enabledelayedexpansion

:: Display help function
if "%1"=="--help" goto :show_help
if "%1"=="-h" goto :show_help
goto :main

:show_help
echo Atom Interpreter Build Script
echo =============================
echo.
echo USAGE:
echo     build.bat [OPTIONS]
echo.
echo DESCRIPTION:
echo     This script builds the Atom interpreter with various configuration options.
echo     In normal mode, it creates a development build for the current platform.
echo     In release mode, it creates optimized builds with packaging for distribution.
echo.
echo OPTIONS:
echo     --help, -h          Show this help message and exit
echo.
echo     --release           Enable release mode with optimized builds and packaging
echo                         Creates archives with lib/, test/ folders and atom.ico
echo.
echo RELEASE MODE OPTIONS:
echo     -t ^<bits^>           Target architecture bit width
echo                         32    Build for 32-bit systems (GOARCH=386)
echo                         64    Build for 64-bit systems (GOARCH=amd64)
echo.
echo     --arch ^<arch^>       Target processor architecture
echo                         amd   AMD/Intel x86 processors (default)
echo                         arm   ARM processors (GOARCH=arm)
echo                         riscv RISC-V processors (GOARCH=riscv64)
echo                         wasm  WebAssembly target (GOOS=js GOARCH=wasm)
echo.
echo     --os ^<os^>           Target operating system
echo                         linux Linux systems (default)
echo                         mac   macOS systems (GOOS=darwin)
echo                         win   Windows systems (GOOS=windows)
echo.
echo EXAMPLES:
echo     # Development build for current platform
echo     build.bat
echo.
echo     # Basic release build (64-bit Windows AMD/Intel)
echo     build.bat --release
echo.
echo     # 32-bit Windows release
echo     build.bat --release -t 32 --os win
echo.
echo     # ARM Linux release
echo     build.bat --release --arch arm --os linux
echo.
echo     # macOS release
echo     build.bat --release --os mac
echo.
echo     # WebAssembly build
echo     build.bat --release --arch wasm
echo.
echo     # RISC-V 64-bit Linux release
echo     build.bat --release --arch riscv --os linux
echo.
echo OUTPUT:
echo     Development Mode:
echo         - Creates executable: atom.exe, atom.linux, or atom.macos
echo         - Based on current platform detection
echo.
echo     Release Mode:
echo         - Creates timestamped archive: atom-release-[arch]-[os]-YYYYMMDD-HHMMSS.zip
echo         - Includes: executable, lib/ folder, test/ folder (*.atom files), atom.ico
echo         - Creates .zip archives for all platforms
echo         - Cleans up temporary release/ folder after packaging
echo.
echo NOTES:
echo     - Requires Go toolchain to be installed and in PATH
echo     - Cross-compilation requires appropriate Go cross-compile support
echo     - WebAssembly builds create .wasm files for browser/Node.js execution
echo     - Release archives are created in the current directory
echo     - Previous release archives are automatically cleaned up
echo.
exit /b 0

:main
echo Building Atom interpreter...

:: Initialize variables
set RELEASE_MODE=false
set ARCH_MODE=
set TARGET_ARCH=
set TARGET_OS=win
set EXECUTABLE=atom.exe

:: Parse all arguments
goto :parse_args

:parse_args
if "%1"=="" goto :build

if "%1"=="-t" (
    if "%2"=="32" (
        set ARCH_MODE=32
        echo 32-bit build enabled
:parse_args
if "%1"=="" goto :build

if "%1"=="--release" (
    set RELEASE_MODE=true
    echo Release mode enabled
    shift
    goto :parse_args
)

if "%1"=="-t" (
        set ARCH_MODE=64
        echo 64-bit build enabled
        shift
        shift
        goto :parse_args
    ) else (
        echo Invalid architecture specified. Use 32 or 64.
        exit /b 1
    )
)

if "%1"=="--arch" (
    if "%2"=="arm" (
        set TARGET_ARCH=arm
        echo ARM architecture build enabled
        shift
        shift
        goto :parse_args
    ) else if "%2"=="riscv" (
        set TARGET_ARCH=riscv
        echo RISC-V architecture build enabled
        shift
        shift
        goto :parse_args
    ) else if "%2"=="wasm" (
        set TARGET_ARCH=wasm
        echo WebAssembly build enabled
        shift
        shift
        goto :parse_args
    ) else (
        set TARGET_ARCH=amd
        echo Using default AMD/Intel architecture build
        shift
        shift
        goto :parse_args
    )
)

if "%1"=="--os" (
    if "%2"=="mac" (
        set TARGET_OS=mac
        echo macOS build enabled
        shift
        shift
        goto :parse_args
    ) else if "%2"=="win" (
        set TARGET_OS=win
        echo Windows build enabled
        shift
        shift
        goto :parse_args
    ) else if "%2"=="linux" (
        set TARGET_OS=linux
        echo Linux build enabled
        shift
        shift
        goto :parse_args
    ) else (
        echo Invalid OS specified. Use mac^|win^|linux.
        exit /b 1
    )
)

echo Unknown option: %1
echo Use --help for usage information.
exit /b 1

:build
:: Set default architecture if not specified in release mode
if "%RELEASE_MODE%"=="true" (
    if "%TARGET_ARCH%"=="" (
        set TARGET_ARCH=amd
        echo Using default AMD/Intel architecture build
    )
)

:: Set executable name based on target OS
if "%TARGET_OS%"=="linux" (
    set EXECUTABLE=atom.linux
    echo Building for Linux...
) else if "%TARGET_OS%"=="mac" (
    set EXECUTABLE=atom.macos
    echo Building for macOS...
) else (
    set EXECUTABLE=atom.exe
    echo Building for Windows...
)

if "%RELEASE_MODE%"=="true" goto :release_build

:: Normal build mode
echo Building for Windows...
if exist atom.ico (
    echo Using atom.ico for Windows executable...
    go build -ldflags="-H windowsgui" -o "%EXECUTABLE%" .
) else (
    go build -o "%EXECUTABLE%" .
)
echo Build complete! %EXECUTABLE% created.
exit /b 0

:release_build
echo Creating release build...

:: Delete previous release archives
echo Cleaning up previous releases...
if exist atom-release-*.zip del /q atom-release-*.zip

:: Create release folder if it doesn't exist
if not exist release mkdir release

:: Copy lib and test folders to release folder
if exist lib (
    echo Copying lib folder...
    xcopy /e /i lib release\lib > nul
)

if exist test (
    echo Copying test folder (only .atom files)...
    if not exist release\test mkdir release\test
    for /r test %%f in (*.atom) do copy "%%f" release\test\ > nul
)

:: Copy atom.ico file if it exists
if exist atom.ico (
    echo Copying atom.ico...
    copy atom.ico release\ > nul
)

:: Set build environment variables for architecture
set BUILD_CMD=go build
set ARCH_SUFFIX=
if "%ARCH_MODE%"=="32" (
    set GOARCH=386
    set BUILD_CMD=go build
    set ARCH_SUFFIX=-32bit
    echo Building for 32-bit architecture...
) else if "%ARCH_MODE%"=="64" (
    set GOARCH=amd64
    set BUILD_CMD=go build
    set ARCH_SUFFIX=-64bit
    echo Building for 64-bit architecture...
)

:: Set target OS environment variables
set OS_SUFFIX=
if "%TARGET_OS%"=="mac" (
    set GOOS=darwin
    set OS_SUFFIX=-mac
    set EXECUTABLE=atom.macos
    echo Building for macOS...
) else if "%TARGET_OS%"=="win" (
    set GOOS=windows
    set OS_SUFFIX=-win
    set EXECUTABLE=atom.exe
    echo Building for Windows...
) else if "%TARGET_OS%"=="linux" (
    set GOOS=linux
    set OS_SUFFIX=-linux
    set EXECUTABLE=atom.linux
    echo Building for Linux...
)

:: Set target architecture environment variables (only if ARCH_MODE is not set)
set TARGET_SUFFIX=
if "%TARGET_ARCH%"=="arm" (
    if "%ARCH_MODE%"=="" set GOARCH=arm
    set BUILD_CMD=go build
    set TARGET_SUFFIX=-arm
    if "%TARGET_OS%"=="win" set EXECUTABLE=atom.exe
    if "%TARGET_OS%"=="linux" set EXECUTABLE=atom.linux
    if "%TARGET_OS%"=="mac" set EXECUTABLE=atom.macos
    echo Building for ARM architecture...
) else if "%TARGET_ARCH%"=="riscv" (
    if "%ARCH_MODE%"=="" set GOARCH=riscv64
    set BUILD_CMD=go build
    set TARGET_SUFFIX=-riscv
    if "%TARGET_OS%"=="win" set EXECUTABLE=atom.exe
    if "%TARGET_OS%"=="linux" set EXECUTABLE=atom.linux
    if "%TARGET_OS%"=="mac" set EXECUTABLE=atom.macos
    echo Building for RISC-V architecture...
) else if "%TARGET_ARCH%"=="wasm" (
    set GOOS=js
    set GOARCH=wasm
    set BUILD_CMD=go build
    set TARGET_SUFFIX=-wasm
    set EXECUTABLE=atom.wasm
    echo Building for WebAssembly...
) else if "%TARGET_ARCH%"=="amd" (
    if "%ARCH_MODE%"=="" (
        if "%GOARCH%"=="" set GOARCH=amd64
    )
    set BUILD_CMD=go build
    set TARGET_SUFFIX=-amd
    echo Building for AMD/Intel architecture...
)

:: Build the Go application in release folder
if "%TARGET_OS%"=="win" (
    if exist atom.ico (
        echo Using atom.ico for Windows executable...
        %BUILD_CMD% -ldflags="-H windowsgui" -o release\%EXECUTABLE% .
    ) else (
        %BUILD_CMD% -o release\%EXECUTABLE% .
    )
) else (
    %BUILD_CMD% -o release\%EXECUTABLE% .
)

:: Create timestamp
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "YY=%dt:~2,2%" & set "YYYY=%dt:~0,4%" & set "MM=%dt:~4,2%" & set "DD=%dt:~6,2%"
set "HH=%dt:~8,2%" & set "Min=%dt:~10,2%" & set "Sec=%dt:~12,2%"
set "timestamp=%YYYY%%MM%%DD%-%HH%%Min%%Sec%"

:: Create zip archive
set ZIP_NAME=atom-release%ARCH_SUFFIX%%TARGET_SUFFIX%%OS_SUFFIX%-%timestamp%.zip
echo Creating zip file: %ZIP_NAME%

:: Use PowerShell to create zip (available on Windows 10+)
powershell -command "Compress-Archive -Path 'release\*' -DestinationPath '%ZIP_NAME%'"

:: Clean up release folder
echo Cleaning up release folder...
rmdir /s /q release

echo Release build complete! Created %ZIP_NAME%
exit /b 0
