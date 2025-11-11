//go:build darwin && amd64

package golibbuilder

//go:generate go build -o ../build/macos/amd64/golib.dylib -buildmode=c-shared ../golib/sharedlib
//go:generate cp -f ../build/macos/amd64/golib.dylib ../flutterui/plugin/macos/Frameworks
//go:generate mkdir -p ../flutterui/pongui/build/macos/Build/Products/Release/pongui.app/Contents/Frameworks
//go:generate cp -f ../build/macos/amd64/golib.dylib ../flutterui/pongui/build/macos/Build/Products/Release/pongui.app/Contents/Frameworks/golib.dylib
