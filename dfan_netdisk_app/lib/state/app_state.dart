import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../services/netdisk_api.dart';

const _kPrefsBaseUrl = 'api_base_url';
const _kPrefsToken = 'jwt_token';

/// 默认需包含 /api/v1；Android 模拟器访问本机用 10.0.2.2
const kDefaultApiBase = 'http://127.0.0.1:8080/api/v1';

class AppState extends ChangeNotifier {
  AppState() {
    _api = NetdiskApi(
      baseUrl: () => _baseUrl,
      token: () => _token,
    );
  }

  String _baseUrl = kDefaultApiBase;
  String? _token;
  bool _loaded = false;
  /// 与后台系统配置「TG 资源图片返代」一致，来自 `/public/config` 的 `tg_image_proxy_url`
  String _tgImageProxyUrl = '';
  late final NetdiskApi _api;

  String get baseUrl => _baseUrl;
  String get tgImageProxyUrl => _tgImageProxyUrl;
  String? get token => _token;
  bool get isLoaded => _loaded;
  NetdiskApi get api => _api;
  bool get isLoggedIn => _token != null && _token!.isNotEmpty;

  Future<void> load() async {
    final p = await SharedPreferences.getInstance();
    _baseUrl = p.getString(_kPrefsBaseUrl) ?? kDefaultApiBase;
    _token = p.getString(_kPrefsToken);
    _loaded = true;
    notifyListeners();
    unawaited(refreshPublicConfig());
  }

  /// 由已获取的 `/public/config` 更新 TG 图片返代等（例如首页与 load 并行时先拿到配置）。
  void applyPublicConfigSnapshot(Map<String, dynamic> cfg) {
    final v = cfg['tg_image_proxy_url'];
    final s = v is String ? v.trim() : '';
    if (_tgImageProxyUrl != s) {
      _tgImageProxyUrl = s;
      notifyListeners();
    }
  }

  /// 拉取公开配置（含 TG 图片返代地址），失败时静默忽略。
  Future<void> refreshPublicConfig() async {
    try {
      final cfg = await _api.publicConfig();
      applyPublicConfigSnapshot(cfg);
    } catch (_) {
      // 离线或接口异常时不阻断客户端
    }
  }

  Future<void> setBaseUrl(String url) async {
    var u = url.trim();
    if (u.endsWith('/')) {
      u = u.substring(0, u.length - 1);
    }
    _baseUrl = u.isEmpty ? kDefaultApiBase : u;
    final p = await SharedPreferences.getInstance();
    await p.setString(_kPrefsBaseUrl, _baseUrl);
    notifyListeners();
  }

  Future<void> setToken(String? t) async {
    _token = t;
    final p = await SharedPreferences.getInstance();
    if (t == null || t.isEmpty) {
      await p.remove(_kPrefsToken);
    } else {
      await p.setString(_kPrefsToken, t);
    }
    notifyListeners();
  }

  Future<void> logout() => setToken(null);
}
