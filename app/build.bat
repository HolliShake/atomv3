@echo off
setlocal enabledelayedexpansion

REM Display help function
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
echo     In normal mode, it creates a development build for Windows.
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
echo                         win   Windows systems (default)
echo                         linux Linux systems (GOOS=linux)
echo                         mac   macOS systems (GOOS=darwin)
echo.
echo EXAMPLES:
echo     # Development build for current platform
echo     build.bat
echo.
echo     # Basic release build (64-bit Windows AMD/Intel)
echo     build.bat --release
echo.
echo     # 32-bit Windows release
echo     build.bat --release -t 32
echo.
echo     # ARM Windows release
echo     build.bat --release --arch arm
echo.
echo     # Linux release
echo     build.bat --release --os linux
echo.
echo     # WebAssembly build
echo     build.bat --release --arch wasm
echo.
echo     # RISC-V 64-bit Linux release
echo     build.bat --release --arch riscv --os linux
echo.
echo OUTPUT:
echo     Development Mode:
echo         - Creates executable: atom.exe
echo.
echo     Release Mode:
echo         - Creates timestamped archive: atom-release-[arch]-[os]-YYYYMMDD-HHMMSS.zip
echo         - Includes: executable, lib/ folder, test/ folder (*.atom files), atom.ico
echo         - Cleans up temporary release/ folder after packaging
echo.
echo NOTES:
echo     - Requires Go toolchain to be installed and in PATH
echo     - Cross-compilation requires appropriate Go cross-compile support
echo     - WebAssembly builds create .wasm files for browser/Node.js execution
echo     - Release archives are created in the current directory
echo     - Previous release archives are automatically cleaned up
echo.
goto :eof

:main
echo Building Atom interpreter...

REM Initialize variables
set RELEASE_MODE=false
set ARCH_MODE=
set TARGET_ARCH=
set TARGET_OS=win
set EXECUTABLE=atom.exe

REM Check for --release flag
if "%1"=="--release" (
    set RELEASE_MODE=true
    echo Release mode enabled
    shift
)

REM Parse additional flags
:parse_args
if "%1"=="" goto :end_parse
if "%1"=="-t" (
    if "%2"=="32" (
        set ARCH_MODE=32
        echo 32-bit build enabled
    ) else if "%2"=="64" (
        set ARCH_MODE=64
        echo 64-bit build enabled
    ) else (
        echo Invalid architecture specified. Use 32 or 64.
        exit /b 1
    )
    shift
    shift
    goto :parse_args
)
if "%1"=="--arch" (
    if "%2"=="arm" (
        set TARGET_ARCH=arm
        echo ARM architecture build enabled
    ) else if "%2"=="riscv" (
        set TARGET_ARCH=riscv
        echo RISC-V architecture build enabled
    ) else if "%2"=="wasm" (
        set TARGET_ARCH=wasm
        echo WebAssembly build enabled
    ) else (
        set TARGET_ARCH=amd
        echo Using default AMD/Intel architecture build
    )
    shift
    shift
    goto :parse_args
)
if "%1"=="--os" (
    if "%2"=="linux" (
        set TARGET_OS=linux
        echo Linux build enabled
    ) else if "%2"=="mac" (
        set TARGET_OS=mac
        echo macOS build enabled
    ) else if "%2"=="win" (
        set TARGET_OS=win
        echo Windows build enabled
    ) else (
        echo Invalid OS specified. Use win^|linux^|mac.
        exit /b 1
    )
    shift
    shift
    goto :parse_args
)
echo Unknown option: %1
echo Use --help for usage information.
exit /b 1

:end_parse
REM Set default architecture if not specified
if "%TARGET_ARCH%"=="" (
    set TARGET_ARCH=amd
    echo Using default AMD/Intel architecture build
)

if "%RELEASE_MODE%"=="true" (
    REM Release mode
    echo Creating release build...
    
    REM Delete previous release archives
    echo Cleaning up previous releases...
    if exist atom-release-*.zip del atom-release-*.zip
    
    REM Create release folder if it doesn't exist
    if not exist release mkdir release
    
    REM Copy lib and test folders to release folder
    if exist lib (
        echo Copying lib folder...
        xcopy lib release\lib\ /E /I /Q
    )
    
    if exist test (
        echo Copying test folder (only .atom files)...
        if not exist release\test mkdir release\test
        for %%f in (test\*.atom) do copy "%%f" release\test\ >nul
    )
    
    REM Copy atom.ico file if it exists
    if exist atom.ico (
        echo Copying atom.ico...
        copy atom.ico release\ >nul
    )
    
    REM Set build environment variables for architecture
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
    
    REM Set target OS environment variables
    set OS_SUFFIX=
    if "%TARGET_OS%"=="mac" (
        set GOOS=darwin
        set OS_SUFFIX=-mac
        set EXECUTABLE=atom.macos
        echo Building for macOS...
    ) else if "%TARGET_OS%"=="linux" (
        set GOOS=linux
        set OS_SUFFIX=-linux
        set EXECUTABLE=atom.linux
        echo Building for Linux...
    ) else if "%TARGET_OS%"=="win" (
        set GOOS=windows
        set OS_SUFFIX=-win
        set EXECUTABLE=atom.exe
        echo Building for Windows...
    )
    
    REM Set target architecture environment variables
    set TARGET_SUFFIX=
    if "%TARGET_ARCH%"=="arm" (
        set GOARCH=arm
        set BUILD_CMD=go build
        set TARGET_SUFFIX=-arm
        set EXECUTABLE=atom.arm
        echo Building for ARM architecture...
    ) else if "%TARGET_ARCH%"=="riscv" (
        set GOARCH=riscv64
        set BUILD_CMD=go build
        set TARGET_SUFFIX=-riscv
        set EXECUTABLE=atom.riscv
        echo Building for RISC-V architecture...
    ) else if "%TARGET_ARCH%"=="wasm" (
        set GOOS=js
        set GOARCH=wasm
        set BUILD_CMD=go build
        set TARGET_SUFFIX=-wasm
        set EXECUTABLE=atom.wasm
        echo Building for WebAssembly...
    ) else if "%TARGET_ARCH%"=="amd" (
        REM Default AMD/Intel - use existing GOARCH if set by ARCH_MODE, otherwise default to amd64
        if "%GOARCH%"=="" set GOARCH=amd64
        set BUILD_CMD=go build
        set TARGET_SUFFIX=-amd
        echo Building for AMD/Intel architecture...
    )
    
    REM Create resource file for icon if building for Windows
    if "%TARGET_OS%"=="win" (
        echo Creating resource file...
        echo 1 ICON "atom.ico" > app.rc
        windres -o app.syso app.rc
    )
    
    REM Build the Go application in release folder
    %BUILD_CMD% -o release\!EXECUTABLE! .
    
    REM Clean up temporary files if Windows build
    if "%TARGET_OS%"=="win" (
        if exist app.rc del app.rc
        if exist app.syso del app.syso
    )
    
    REM Create archive file with architecture suffix
    for /f "tokens=1-4 delims=/ " %%a in ('date /t') do set DATE=%%d%%b%%c
    for /f "tokens=1-2 delims=: " %%a in ('time /t') do set TIME=%%a%%b
    set TIMESTAMP=%DATE:~-4%%DATE:~3,2%%DATE:~0,2%-%TIME::=%
    set ZIP_NAME=atom-release!ARCH_SUFFIX!!TARGET_SUFFIX!!OS_SUFFIX!-!TIMESTAMP!.zip
    echo Creating zip file: !ZIP_NAME!
    
    REM Use PowerShell to create zip file
    powershell -command "Compress-Archive -Path 'release\*' -DestinationPath '!ZIP_NAME!' -Force"
    
    REM Clean up release folder
    echo Cleaning up release folder...
    rmdir /s /q release
    
    echo Release build complete! Created !ZIP_NAME!
) else (
    REM Normal build mode
    REM Create resource file for icon
    echo Creating resource file...
    echo 1 ICON "atom.ico" > app.rc
    
    REM Compile resource file (requires windres from MinGW or similar)
    windres -o app.syso app.rc
    
    REM Build the Go application with icon
    go build -o atom.exe .
    
    REM Clean up temporary files
    del app.rc
    del app.syso
    
    echo Build complete! atom.exe created.
)

pause
