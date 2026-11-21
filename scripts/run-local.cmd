@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0run-local.ps1"
if errorlevel 1 (
    echo.
    echo Gagal menjalankan aplikasi. Periksa pesan di atas.
    pause
)
endlocal
