@echo off
echo Building Atom interpreter...

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
pause
