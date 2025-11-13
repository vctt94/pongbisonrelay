#define AppVersion "0.0.1"

[Setup]
AppName=Bison Pong
AppVersion={#AppVersion}
DefaultDirName={pf}\Bison Pong
DefaultGroupName=Bison Pong
OutputBaseFilename=bisonpong-windows-amd64-{#AppVersion}
Compression=lzma
SolidCompression=yes

[Files]
Source: "C:\Users\vctt\projects\pong-bisonrelay\pongui\flutterui\pongui\build\windows\x64\runner\Release\*"; DestDir: "{app}"; Flags: recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Bison Pong"; Filename: "{app}\bisonpong-{#AppVersion}.exe"
Name: "{group}\Uninstall Bison Pong"; Filename: "{uninstallexe}"

[Run]
Filename: "{app}\bisonpong.exe"; Description: "Launch Bison Pong"; Flags: nowait postinstall skipifsilent
