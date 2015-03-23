:: Visual Studio must be in path

where devenv
if errorLevel 1 ( echo "devenv was not found on PATH" && exit /b 1 )
 
rmdir /S /Q packages
bin\nuget restore || exit /b 1
MSBuild Nora.sln /t:Rebuild /p:Configuration=Release || exit /b 1
packages\nspec.0.9.68\tools\NSpecRunner.exe Nora.Tests\bin\Release\Nora.Tests.dll || exit /b 1
