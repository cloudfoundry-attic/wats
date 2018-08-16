@echo on

call "C:\Program Files (x86)\Microsoft Visual Studio\2017\Community\Common7\Tools\VsDevCmd.bat"
nuget restore
msbuild .\nora.sln /t:Rebuild /p:Configuration=Release /p:PublishProfile=FolderProfile /p:DeployOnBuild=true
