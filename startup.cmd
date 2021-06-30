@echo off
PowerShell -Command "Start-Process -FilePath 'F:\manga\mgcrl.exe' -ArgumentList 'serve','1232','-o=F:\manga','-c=F:\manga\manga.csv','-l=F:\manga\mgcrl.log' -WindowStyle Hidden"
