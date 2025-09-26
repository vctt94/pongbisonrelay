[Setup]
AppName=BR Pong
AppVersion=0.0.1-rc2
DefaultDirName={pf}\BR Pong
DefaultGroupName=BR Pong
OutputBaseFilename=BR_Pong_Installer
Compression=lzma
SolidCompression=yes

[Files]
Source: "C:\Users\vctt\projects\pong-bisonrelay\pongui\flutterui\pongui\build\windows\x64\runner\Release\*"; DestDir: "{app}"; Flags: recursesubdirs createallsubdirs

[Icons]
Name: "{group}\BR Pong"; Filename: "{app}\pongui-0.0.1-rc2.exe"
Name: "{group}\Uninstall BR Pong"; Filename: "{uninstallexe}"

[Run]
Filename: "{app}\pongui.exe"; Description: "Launch BR Pong"; Flags: nowait postinstall skipifsilent
