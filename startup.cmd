@echo off
PowerShell -Command "Start-Process -FilePath 'F:\manga\mgcrl.exe' -ArgumentList 'serve','1232','-f=F:\manga\config.json','-l=F:\manga\mgcrl.log' -WindowStyle Hidden"
