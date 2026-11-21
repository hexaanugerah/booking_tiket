@echo off
REM Deploy Kawah Putih dari Windows dalam satu langkah.
REM Jalankan: scripts\deploy-now.cmd
REM Opsional: scripts\deploy-now.cmd -Server 103.x.x.x -User root -Port 22

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0deploy-now.ps1" %*
