@echo off
chcp 65001 >nul
cd /d "%~dp0"

echo.
echo ╔══════════════════════════════════════════╗
echo ║        AI-Forge 启动脚本                 ║
echo ╚══════════════════════════════════════════╝
echo.

REM 第一步：清理占用 8080 端口的旧进程
echo [1/3] 检查端口 8080 是否被占用...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr "LISTENING" 2^>nul') do (
    echo   发现旧进程 PID=%%a，正在结束...
    taskkill /PID %%a /F >nul 2>&1
)
echo   端口已释放

REM 第二步：清理旧编译产物
echo [2/3] 编译项目...
go build -o app.exe main.go 2>&1
if %errorlevel% neq 0 (
    echo   编译失败！
    pause
    exit /b 1
)
echo   编译成功

REM 第三步：启动服务
echo [3/3] 启动服务...
echo.
echo   访问地址:
echo   API        http://localhost:8080
echo   Dashboard  http://localhost:8080/dashboard
echo   Swagger    http://localhost:8080/docs
echo.
echo   按 Ctrl+C 停止服务
echo ══════════════════════════════════════════
echo.

app.exe
