@echo off

if "%~1" == "" (
	echo no version specified
	goto :eof
)

set GOARCH=amd64
go build -buildmode=exe
upx --best mgcrl.exe
rename mgcrl.exe mgcrl_v%~1_windows_amd64.exe

set GOARCH=386
go build -buildmode=exe
upx --best mgcrl.exe
rename mgcrl.exe mgcrl_v%~1_windows_386.exe

if "%~2" == "-gui" (
	cd gui
	set GOARCH=amd64
	go build -buildmode=exe -o=repl.exe
	upx --best repl.exe
	pyinstaller --onefile --windowed --icon=icon.ico --add-data=icon.ico;. --add-binary=repl.exe;. main.py
	del repl.exe
)
