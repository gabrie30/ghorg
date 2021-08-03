@echo off
title Windows GitHub Integration Test
echo Starting Windows Integration Tests

ghorg.exe clone underdeveloped --token=%GITHUB_TOKEN%

IF %ERRORLEVEL% NEQ 0 Echo Tests Failed
IF %ERRORLEVEL% EQU 0 Echo Tests Passed

EXIT /B %ERRORLEVEL%
