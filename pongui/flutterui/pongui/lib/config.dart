import 'dart:io';
import 'package:args/args.dart';
import 'package:ini/ini.dart' as ini;
import 'package:path_provider/path_provider.dart';
import 'package:path/path.dart' as path;

const APPNAME = "pongui";
const BRUIGNAME = "bruig";
String mainConfigFilename = "";

class Config {
  late final String serverAddr;
  late final String grpcCertPath;
  late final String rpcCertPath;
  late final String rpcClientCertPath;
  late final String rpcClientKeyPath;
  late final String rpcWebsocketURL;
  late final String debugLevel;
  late final String rpcUser;
  late final String rpcPass;
  late final bool wantsLogNtfns;
  late final String dataDir;
  late final String address;

  Config();

  Config.filled({
    this.serverAddr = "",
    this.grpcCertPath = "",
    this.rpcCertPath = "",
    this.rpcClientCertPath = "",
    this.rpcClientKeyPath = "",
    this.debugLevel = "info",
    this.rpcWebsocketURL = "",
    this.rpcUser = "",
    this.rpcPass = "",
    this.wantsLogNtfns = false,
    this.dataDir = "",
    this.address = "",
  });

  // Save a new config from scratch
  Future<void> saveNewConfig(String filepath) async {
    var f = ini.Config.fromString("[clientrpc]\n[log]\n");
    set(String section, String opt, String val) =>
        val != "" ? f.set(section, opt, val) : null;

    set("default", "server", serverAddr);
    set("default", "grpccertpath", grpcCertPath);
    set("default", "address", address);
    set("clientrpc", "rpccertpath", rpcCertPath);
    set("log", "debug", debugLevel);
    set("clientrpc", "rpcwebsocketurl", rpcWebsocketURL);
    set("clientrpc", "rpcclientcertpath", rpcClientCertPath);
    set("clientrpc", "rpcclientkeypath", rpcClientKeyPath);
    set("clientrpc", "rpccertpath", rpcCertPath);
    set("clientrpc", "rpcuser", rpcUser);
    set("clientrpc", "rpcpass", rpcPass);
    set("clientrpc", "wantsLogNtfns", wantsLogNtfns ? "1" : "0");

    // Write the config file
    await File(filepath).parent.create(recursive: true);
    await File(filepath).writeAsString(f.toString());
  }

  // Load existing config
  static Future<Config> loadConfig(String filepath) async {
    var f = ini.Config.fromStrings(File(filepath).readAsLinesSync());

    return Config.filled(
      serverAddr: f.get("default", "server") ?? "localhost:443",
      grpcCertPath: f.get("default", "grpccertpath") ?? "",
      address: f.get("default", "address") ?? "",
      rpcCertPath: f.get("clientrpc", "rpccertpath") ?? "",
      rpcClientCertPath: f.get("clientrpc", "rpcclientcertpath") ?? "",
      rpcClientKeyPath: f.get("clientrpc", "rpcclientkeypath") ?? "",
      debugLevel: f.get("log", "debug") ?? "info",
      rpcUser: f.get("clientrpc", "rpcuser") ?? "",
      rpcPass: f.get("clientrpc", "rpcpass") ?? "",
      rpcWebsocketURL: f.get("clientrpc", "rpcwebsocketurl") ?? "",
      wantsLogNtfns: f.get("clientrpc", "wantsLogNtfns") == "1",
      dataDir: await defaultAppDataDir(),
    );
  }
}

// Function to get the default app data directory based on the platform
Future<String> defaultAppDataDir() async {
  if (Platform.isLinux) {
    final home = Platform.environment["HOME"];
    if (home != null && home != "") {
      return path.join(home, ".$APPNAME");
    }
  } else if (Platform.isWindows &&
      Platform.environment.containsKey("LOCALAPPDATA")) {
    return path.join(Platform.environment["LOCALAPPDATA"]!, APPNAME);
  } else if (Platform.isMacOS) {
    final baseDir = (await getApplicationSupportDirectory()).parent.path;
    return path.join(baseDir, APPNAME);
  }

  // For other platforms, get the parent directory to avoid bundle identifier paths
  final dir = (await getApplicationSupportDirectory()).parent;
  return path.join(dir.path, APPNAME);
}

// Function to get the default app data directory based on the platform
Future<String> defaultAppDataBRUIGDir() async {
  if (Platform.isLinux) {
    final home = Platform.environment["HOME"];
    if (home != null && home != "") {
      return path.join(home, ".$BRUIGNAME");
    }
  } else if (Platform.isWindows &&
      Platform.environment.containsKey("LOCALAPPDATA")) {
    return path.join(Platform.environment["LOCALAPPDATA"]!, BRUIGNAME);
  } else if (Platform.isMacOS) {
    final baseDir = (await getApplicationSupportDirectory()).parent.path;
    return path.join(baseDir, BRUIGNAME);
  }

  final dir = await getApplicationSupportDirectory();
  return path.join(dir.path, BRUIGNAME);
}

final usageException = Exception("Usage Displayed");
final newConfigNeededException = Exception("Config needed");

Future<Config> loadConfig(String filepath) async {
  late ini.Config f;

  var configDir = cleanAndExpandPath(filepath);

  String getPath(String section, String option, String def) {
    var iniVal = f.get(section, option);
    if (iniVal == null || iniVal == "") {
      return def;
    }
    return cleanAndExpandPath(iniVal);
  }

  bool getBool(String section, String opt) {
    var v = f.get(section, opt);
    return v == "yes" || v == "true" || v == "1";
  }

  try {
    f = ini.Config.fromStrings(File(configDir).readAsLinesSync());
  } catch (e) {
    print('loadConfig: Error parsing config file: $e');
    // Create empty config with default sections
    f = ini.Config.fromString("[clientrpc]\n[log]\n");
  }

  // Accept alt names used by legacy configs in [default] section.
  String pick(List<String?> xs, String fallback) {
    for (final v in xs) {
      if (v != null && v.trim().isNotEmpty) return v;
    }
    return fallback;
  }

  final serverAddr = pick([
    f.get("default", "server"),
    f.get("default", "serveraddr"),
  ], "localhost:50051");
  final grpcCertPath = pick([
    f.get("default", "grpccertpath"),
    f.get("default", "grpcservercert"),
  ], "");
  final rpcWebsocket = pick([
    f.get("clientrpc", "rpcwebsocketurl"),
    f.get("default", "brrpcurl"),
  ], "");
  final rpccert = pick([
    getPath("clientrpc", "rpccertpath", ""),
    f.get("default", "brclientcert"),
  ], "");
  final rpcclientcert = pick([
    getPath("clientrpc", "rpcclientcertpath", ""),
    f.get("default", "brclientrpccert"),
  ], "");
  final rpcclientkey = pick([
    getPath("clientrpc", "rpcclientkeypath", ""),
    f.get("default", "brclientrpckey"),
  ], "");

  var c = Config.filled(
      serverAddr: serverAddr,
      grpcCertPath: grpcCertPath,
      address: f.get("default", "address") ?? "",
      debugLevel: f.get("log", "debuglevel") ?? "info",
      rpcWebsocketURL: rpcWebsocket,
      rpcCertPath: rpccert,
      rpcClientCertPath: rpcclientcert,
      rpcClientKeyPath: rpcclientkey,
      rpcUser: f.get("clientrpc", "rpcuser") ?? "",
      rpcPass: f.get("clientrpc", "rpcpass") ?? "",
      wantsLogNtfns: getBool("clientrpc", "wantsLogNtfns"),
      dataDir: await defaultAppDataDir());

  return c;
}

String homeDir() {
  var env = Platform.environment;
  if (Platform.isWindows) {
    return env['UserProfile'] ?? "";
  } else {
    return env['HOME'] ?? "";
  }
}

String cleanAndExpandPath(String p) {
  if (p == "") {
    return p;
  }

  if (p.startsWith("~")) {
    p = homeDir() + p.substring(1);
  }

  return path.canonicalize(p);
}

Future<ArgParser> appArgParser() async {
  var defaultCfgFile = path.join(await defaultAppDataDir(), "$APPNAME.conf");
  var p = ArgParser();
  p.addFlag("help", abbr: "h", help: "Display usage info", negatable: false);
  p.addOption("configfile",
      abbr: "c", defaultsTo: defaultCfgFile, help: "Path to config file");
  return p;
}

Future<String> configFileName(List<String> args) async {
  var p = await appArgParser();
  var res = p.parse(args);
  return res["configfile"];
}

// Function to load config using arguments
Future<Config> configFromArgs(List<String> args) async {
  var p = await appArgParser();
  var res = p.parse(args);

  if (res["help"]) {
    // ignore: avoid_print
    print(p.usage);
    throw usageException;
  }

  var cfgFilePath = res["configfile"];
  if (!File(cfgFilePath).existsSync()) {
    throw newConfigNeededException;
  }

  return loadConfig(cfgFilePath);
}
