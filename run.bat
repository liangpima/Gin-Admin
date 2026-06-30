@echo off
chcp 65001 >nul
powershell.exe -NoExit -ExecutionPolicy Bypass -File "%~dp0start-backend.ps1"
