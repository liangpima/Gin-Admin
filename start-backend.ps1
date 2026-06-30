cd 'I:\phpstudy_pro\WWW\go'

# 杀掉旧进程
$procs = Get-Process -Name "server.exe" -ErrorAction SilentlyContinue
if ($procs) {
    $procs | Stop-Process -Force
    Write-Host "[stop] 已停止旧进程" -ForegroundColor Yellow
    Start-Sleep -Seconds 1
}

# 编译
Write-Host "[build] 正在编译..." -ForegroundColor Cyan
& 'C:\Go\bin\go.exe' build -o server.exe ./cmd/server
if ($LASTEXITCODE -ne 0) {
    Write-Host "[error] 编译失败" -ForegroundColor Red
    exit 1
}

# 启动
Write-Host "[start] 启动后端服务..." -ForegroundColor Green
& .\server.exe
