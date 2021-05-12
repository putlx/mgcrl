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
