chcp 65001 > $null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Set-Location 'I:\phpstudy_pro\WWW\go'

# Kill process on port 8080
$port = 8080
$conn = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
if ($conn) {
    $pid = $conn.OwningProcess | Select-Object -First 1
    Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
    Write-Host "[stop] killed process on port $port" -ForegroundColor Yellow
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
Write-Host "[start] running server on port $port ..." -ForegroundColor Green
& .\server.exe

Write-Host ""
Write-Host "[exit] server stopped" -ForegroundColor Yellow
Read-Host "Press Enter to exit"
