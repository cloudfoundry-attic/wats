@echo on

call "C:\Program Files (x86)\Microsoft Visual Studio\2017\Community\Common7\Tools\VsDevCmd.bat"

WHERE mycommand
IF %ERRORLEVEL% NEQ 0 ECHO nuget should be installed from (https://dist.nuget.org/win-x86-commandline/latest/nuget.exe)
nuget restore
msbuild .\nora.sln /t:Rebuild /p:Configuration=Release /p:PublishProfile=FolderProfile /p:DeployOnBuild=true
