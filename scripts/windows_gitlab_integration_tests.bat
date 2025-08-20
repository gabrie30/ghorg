@echo off
title Windows GitLab Integration Test
echo Starting Windows GitLab Integration Tests

REM Set GitLab test group (same as Linux version)
set GITLAB_GROUP=gitlab-examples
set GITLAB_SUB_GROUP=wayne-enterprises

echo.
echo ========================================
echo Testing Basic GitLab Clone
echo ========================================
REM Test basic clone functionality
ghorg.exe clone %GITLAB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab

IF NOT EXIST "%USERPROFILE%\ghorg\%GITLAB_GROUP%\microservice" (
    echo FAIL: Basic GitLab clone test failed
    exit /b 1
)
echo PASS: Basic GitLab clone test

REM Clean up for next test
rmdir /s /q "%USERPROFILE%\ghorg\%GITLAB_GROUP%" 2>nul

echo.
echo ========================================
echo Testing Clone with Output Directory
echo ========================================
REM Test clone with custom output directory
ghorg.exe clone %GITLAB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab --output-dir=gitlab-examples-output

IF NOT EXIST "%USERPROFILE%\ghorg\gitlab-examples-output\microservice" (
    echo FAIL: GitLab clone with output directory test failed
    exit /b 1
)
echo PASS: GitLab clone with output directory test

REM Clean up for next test
rmdir /s /q "%USERPROFILE%\ghorg\gitlab-examples-output" 2>nul

echo.
echo ========================================
echo Testing Clone with Preserve Directory
echo ========================================
REM Test clone with directory structure preservation
ghorg.exe clone %GITLAB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab --preserve-dir

IF NOT EXIST "%USERPROFILE%\ghorg\%GITLAB_GROUP%\%GITLAB_SUB_GROUP%\wayne-industries\microservice" (
    echo FAIL: GitLab clone with preserve directory test failed
    exit /b 1
)
echo PASS: GitLab clone with preserve directory test

REM Clean up for next test
rmdir /s /q "%USERPROFILE%\ghorg\%GITLAB_GROUP%" 2>nul

echo.
echo ========================================
echo Testing Clone with Preserve SCM Hostname
echo ========================================
REM Test clone with SCM hostname preservation
ghorg.exe clone %GITLAB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab --preserve-scm-hostname

IF NOT EXIST "%USERPROFILE%\ghorg\gitlab.com\%GITLAB_GROUP%\microservice" (
    echo FAIL: GitLab clone with preserve SCM hostname test failed
    exit /b 1
)
echo PASS: GitLab clone with preserve SCM hostname test

REM Clean up for next test
rmdir /s /q "%USERPROFILE%\ghorg\gitlab.com" 2>nul

echo.
echo ========================================
echo Testing Prune with Preserve Directory
echo ========================================
REM This is the main test requested - test prune functionality
REM First, do an initial clone with preserve-dir, prune, and prune-no-confirm
ghorg.exe clone %GITLAB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab --preserve-dir --prune --prune-no-confirm

REM Create a fake git repository that should be pruned in the next run
mkdir "%USERPROFILE%\ghorg\%GITLAB_GROUP%\%GITLAB_SUB_GROUP%\wayne-industries\fake-repo-to-prune"
cd "%USERPROFILE%\ghorg\%GITLAB_GROUP%\%GITLAB_SUB_GROUP%\wayne-industries\fake-repo-to-prune"
git init > nul 2>&1
cd /d "%~dp0\.."

REM Run clone again with prune - this should remove the fake repo we just created
ghorg.exe clone %GITLAB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab --preserve-dir --prune --prune-no-confirm

REM Check that legitimate repository still exists
IF NOT EXIST "%USERPROFILE%\ghorg\%GITLAB_GROUP%\%GITLAB_SUB_GROUP%\wayne-industries\microservice" (
    echo FAIL: Prune test failed - legitimate repository was removed
    exit /b 1
)

REM Check that fake repository was pruned (removed)
IF EXIST "%USERPROFILE%\ghorg\%GITLAB_GROUP%\%GITLAB_SUB_GROUP%\wayne-industries\fake-repo-to-prune" (
    echo FAIL: Prune test failed - fake repository was not removed
    exit /b 1
)
echo PASS: GitLab prune with preserve directory test

REM Clean up for next test
rmdir /s /q "%USERPROFILE%\ghorg\%GITLAB_GROUP%" 2>nul

echo.
echo ========================================
echo Testing Subgroup Clone
echo ========================================
REM Test cloning a GitLab subgroup
ghorg.exe clone %GITLAB_GROUP%/%GITLAB_SUB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab

IF NOT EXIST "%USERPROFILE%\ghorg\%GITLAB_GROUP%\%GITLAB_SUB_GROUP%\microservice" (
    echo FAIL: GitLab subgroup clone test failed
    exit /b 1
)
echo PASS: GitLab subgroup clone test

REM Clean up for next test
rmdir /s /q "%USERPROFILE%\ghorg\%GITLAB_GROUP%" 2>nul

echo.
echo ========================================
echo Testing Subgroup Clone with Preserve Directory
echo ========================================
REM Test subgroup clone with directory preservation
ghorg.exe clone %GITLAB_GROUP%/%GITLAB_SUB_GROUP% --token=%GITLAB_TOKEN% --scm=gitlab --preserve-dir

IF NOT EXIST "%USERPROFILE%\ghorg\%GITLAB_GROUP%\%GITLAB_SUB_GROUP%\wayne-industries\microservice" (
    echo FAIL: GitLab subgroup clone with preserve directory test failed
    exit /b 1
)
echo PASS: GitLab subgroup clone with preserve directory test

REM Final cleanup
rmdir /s /q "%USERPROFILE%\ghorg\%GITLAB_GROUP%" 2>nul

echo.
echo ========================================
echo All GitLab Integration Tests Completed Successfully
echo ========================================

EXIT /B 0
