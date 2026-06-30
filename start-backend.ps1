chcp 65001 > $null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Set-Location 'I:\phpstudy_pro\WWW\go'

# Kill old process
$procs = Get-Process -Name "server.exe" -ErrorAction SilentlyContinue
if ($procs) {
    $procs | Stop-Process -Force
    Write-Host "[stop] server stopped" -ForegroundColor Yellow
    Start-Sleep -Seconds 1
}

# Build
Write-Host "[build] compiling..." -ForegroundColor Cyan
& 'C:\Go\bin\go.exe' build -o server.exe ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "[error] build failed" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

# Start
Write-Host "[start] running server..." -ForegroundColor Green
& .\server.exe

Write-Host ""
Write-Host "[exit] server stopped" -ForegroundColor Yellow
Read-Host "Press Enter to exit"
