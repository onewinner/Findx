@echo off
SETLOCAL

SET CGO_ENABLED=0

REM 获取当前目录的上级目录名
FOR %%F IN (%CD%) DO SET PROJECT_NAME=%%~NXF

REM 如果需要从上级目录获取项目名，可以使用以下行
REM FOR %%F IN (%CD%\..) DO SET PROJECT_NAME=%%~NXF

REM Build for Windows 64-bit
SET GOOS=windows
SET GOARCH=amd64
go build -trimpath "-ldflags=-s -w" -o %PROJECT_NAME%_windows_amd64.exe

IF ERRORLEVEL 1 (
    echo 64-bit Windows build failed.
    EXIT /B 1
)

REM Compress the Windows 64-bit executable with UPX at level 9
upx -9 %PROJECT_NAME%_windows_amd64.exe
IF ERRORLEVEL 1 (
    echo UPX compression failed for %PROJECT_NAME%_windows_amd64.exe.
    EXIT /B 1
)

REM Build for Windows 32-bit
SET GOARCH=386
go build -trimpath "-ldflags=-s -w" -o %PROJECT_NAME%_windows_386.exe

IF ERRORLEVEL 1 (
    echo 32-bit Windows build failed.
    EXIT /B 1
)

REM Compress the Windows 32-bit executable with UPX at level 9
upx -9 %PROJECT_NAME%_windows_386.exe
IF ERRORLEVEL 1 (
    echo UPX compression failed for %PROJECT_NAME%_windows_386.exe.
    EXIT /B 1
)

REM Build for Linux 64-bit
SET GOOS=linux
SET GOARCH=amd64
go build -trimpath "-ldflags=-s -w" -o %PROJECT_NAME%_linux_amd64

IF ERRORLEVEL 1 (
    echo 64-bit Linux build failed.
    EXIT /B 1
)

REM Compress the Linux 64-bit executable with UPX at level 9
upx -9 %PROJECT_NAME%_linux_amd64
IF ERRORLEVEL 1 (
    echo UPX compression failed for %PROJECT_NAME%_linux_amd64.
    EXIT /B 1
)

REM Build for Linux 32-bit
SET GOARCH=386
go build -trimpath "-ldflags=-s -w" -o %PROJECT_NAME%_linux_386

IF ERRORLEVEL 1 (
    echo 32-bit Linux build failed.
    EXIT /B 1
)

REM Compress the Linux 32-bit executable with UPX at level 9
upx -9 %PROJECT_NAME%_linux_386
IF ERRORLEVEL 1 (
    echo UPX compression failed for %PROJECT_NAME%_linux_386.
    EXIT /B 1
)

echo Build and compression completed successfully for Windows and Linux.
exit
ENDLOCAL

CD /D %~dp0
dir
pause
