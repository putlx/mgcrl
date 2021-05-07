@echo off
PowerShell -Command "Start-Process -FilePath 'F:\manga\mgcrl.exe' -ArgumentList '-f','F:\manga\config.json' -WindowStyle Hidden -RedirectStandardError 'F:\manga\mgcrl.log'"
