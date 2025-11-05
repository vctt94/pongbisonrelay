//go:build darwin && arm64
// +build darwin,arm64

package golibbuilder

//go:generate go build -o ../build/macos/arm64/golib.dylib -buildmode=c-shared ../golib/sharedlib
//go:generate cp -r ../build/macos/arm64/golib.dylib ../flutterui/plugin/macos/Frameworks
//go:generate mkdir -p ../flutterui/pongui/build/macos/Build/Products/Debug/pongui.app/Contents/Frameworks
//go:generate cp -f ../build/macos/arm64/golib.dylib ../flutterui/pongui/build/macos/Build/Products/Debug/pongui.app/Contents/Frameworks/golib.dylib
