@echo off
setlocal enabledelayedexpansion

echo Running Sourcehut Integration Tests (Windows)

REM Note: Username can be with or without ~ prefix (e.g., "gabrie30" or "~gabrie30")
REM Check if SOURCEHUT_TEST_USER is set
if "%SOURCEHUT_TEST_USER%"=="" (
    echo SOURCEHUT_TEST_USER environment variable is not set
    echo Skipping sourcehut integration tests
    exit /b 0
)

REM Check if SOURCEHUT_TOKEN is set
if "%SOURCEHUT_TOKEN%"=="" (
    echo SOURCEHUT_TOKEN environment variable is not set
    echo Skipping sourcehut integration tests
    exit /b 0
)

echo Testing sourcehut user: %SOURCEHUT_TEST_USER%

REM Test 1: Clone a user's repos
ghorg clone %SOURCEHUT_TEST_USER% --scm=sourcehut --token=%SOURCEHUT_TOKEN%

if exist "%USERPROFILE%\ghorg\%SOURCEHUT_TEST_USER%" (
    echo Pass: sourcehut user clone
) else (
    echo Fail: sourcehut user clone
    exit /b 1
)

REM Test 2: Clone with preserve-scm-hostname
ghorg clone %SOURCEHUT_TEST_USER% --scm=sourcehut --token=%SOURCEHUT_TOKEN% --preserve-scm-hostname

if exist "%USERPROFILE%\ghorg\git.sr.ht\%SOURCEHUT_TEST_USER%" (
    echo Pass: sourcehut user clone preserving scm hostname
) else (
    echo Fail: sourcehut user clone preserving scm hostname
    exit /b 1
)

REM Test 3: Clone with SSH protocol
ghorg clone %SOURCEHUT_TEST_USER% --scm=sourcehut --token=%SOURCEHUT_TOKEN% --protocol=ssh --path=%TEMP% --output-dir=testing_sourcehut_ssh

if exist "%TEMP%\testing_sourcehut_ssh" (
    echo Pass: sourcehut user clone with SSH protocol
) else (
    echo Fail: sourcehut user clone with SSH protocol
    exit /b 1
)

REM Test 4: Clone with HTTPS protocol
ghorg clone %SOURCEHUT_TEST_USER% --scm=sourcehut --token=%SOURCEHUT_TOKEN% --protocol=https --path=%TEMP% --output-dir=testing_sourcehut_https

if exist "%TEMP%\testing_sourcehut_https" (
    echo Pass: sourcehut user clone with HTTPS protocol
) else (
    echo Fail: sourcehut user clone with HTTPS protocol
    exit /b 1
)

echo.
echo ==========================================
echo All sourcehut integration tests passed!
echo ==========================================

exit /b 0

