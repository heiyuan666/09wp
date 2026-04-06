import 'dart:io' show Platform;

import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:url_launcher/url_launcher.dart';

/// 从主链接与附加链接推断该资源包含哪些网盘（去重、保序；无法识别时可为「网盘链接」）
List<String> netdiskLabelsForResource({
  required String link,
  List<String> extraLinks = const [],
}) {
  final seen = <String>{};
  final out = <String>[];
  void take(String? raw) {
    final t = raw?.trim() ?? '';
    if (t.isEmpty) return;
    final lab = netdiskLabelForUrl(t);
    if (seen.contains(lab)) return;
    seen.add(lab);
    out.add(lab);
  }

  take(link);
  for (final x in extraLinks) {
    take(x);
  }
  return out;
}

/// 根据分享 URL 推断展示名称
String netdiskLabelForUrl(String raw) {
  final u = raw.toLowerCase();
  if (u.contains('pan.baidu.com') || u.contains('yun.baidu.com')) {
    return '百度网盘';
  }
  if (u.contains('pan.quark.cn')) return '夸克网盘';
  if (u.contains('aliyundrive.com') || u.contains('alipan.com')) {
    return '阿里云盘';
  }
  if (u.contains('pan.xunlei.com')) return '迅雷云盘';
  if (u.contains('drive.uc.cn') || u.contains('drive-h.uc.cn')) {
    return 'UC网盘';
  }
  if (u.contains('cloud.189.cn') || u.contains('caiyun.189')) {
    return '天翼云盘';
  }
  if (u.contains('yun.139.com') || u.contains('caiyun.139.com')) {
    return '移动云盘';
  }
  if (u.contains('115.com') || u.contains('115cdn.com')) return '115网盘';
  if (u.contains('123pan') ||
      u.contains('123684') ||
      u.contains('123685') ||
      u.contains('123912') ||
      u.contains('123592') ||
      u.contains('123865') ||
      u.contains('123.net')) {
    return '123网盘';
  }
  return '网盘链接';
}

/// 生成依次尝试打开的 URI（优先常见 App URL Scheme，最后回退 https）
List<Uri> netdiskLaunchCandidates(String rawUrl) {
  final trimmed = rawUrl.trim();
  final httpsUri = Uri.tryParse(trimmed);
  if (httpsUri == null || !httpsUri.hasScheme) {
    return [];
  }

  final scheme = httpsUri.scheme.toLowerCase();
  if (scheme != 'http' && scheme != 'https') {
    return [httpsUri];
  }

  final host = httpsUri.host.toLowerCase();
  final path = httpsUri.path;
  final query = httpsUri.hasQuery ? '?${httpsUri.query}' : '';
  final pathQ = '$path$query';
  final full = httpsUri.toString();

  final native = !kIsWeb && (Platform.isAndroid || Platform.isIOS);

  final candidates = <Uri>[];

  void add(Uri u) {
    if (!candidates.any((e) => e.toString() == u.toString())) {
      candidates.add(u);
    }
  }

  if (host.contains('pan.baidu.com') || host.contains('yun.baidu.com')) {
    if (native) {
      add(Uri.parse('baiduwangpan://$host$pathQ'));
      add(Uri.parse(
        'baiduwangpan://webview/link?url=${Uri.encodeComponent(full)}',
      ));
    }
    add(httpsUri);
  } else if (host.contains('quark.cn')) {
    if (native) {
      add(Uri.parse('quark://$host$pathQ'));
    }
    add(httpsUri);
  } else if (host.contains('aliyundrive.com') || host.contains('alipan.com')) {
    if (native) {
      add(Uri.parse(
        'alipans://open?action=openurl&url=${Uri.encodeComponent(full)}',
      ));
      add(Uri.parse(
        'smartdrive://router/open?url=${Uri.encodeComponent(full)}',
      ));
    }
    add(httpsUri);
  } else if (host.contains('pan.xunlei.com')) {
    if (native) {
      add(Uri.parse(
        'xunlei://share?url=${Uri.encodeComponent(full)}',
      ));
    }
    add(httpsUri);
  } else if (host.contains('drive.uc.cn') || host.contains('drive-h.uc.cn')) {
    if (native) {
      add(Uri.parse('uclink://open?url=${Uri.encodeComponent(full)}'));
    }
    add(httpsUri);
  } else if (host.contains('115.com') || host.contains('115cdn.com')) {
    if (native) {
      add(Uri.parse('url115://$host$pathQ'));
    }
    add(httpsUri);
  } else if (host.contains('123pan') ||
      host.contains('123684') ||
      host.contains('123685') ||
      host.contains('123912') ||
      host.contains('123592') ||
      host.contains('123865') ||
      host.contains('123.net')) {
    if (native) {
      add(Uri.parse('pan123://$host$pathQ'));
    }
    add(httpsUri);
  } else if (host.contains('cloud.189.cn') || host.contains('caiyun.189')) {
    if (native) {
      add(Uri.parse('ctcloud://$host$pathQ'));
    }
    add(httpsUri);
  } else if (host.contains('yun.139.com') || host.contains('caiyun.139.com')) {
    if (native) {
      add(Uri.parse('mcloud://$host$pathQ'));
    }
    add(httpsUri);
  } else {
    add(httpsUri);
  }

  return candidates;
}

/// 尝试唤起对应网盘 App；若均失败返回 false（可再提示用户复制链接用浏览器打开）
Future<bool> launchNetdiskInApp(String rawUrl) async {
  final candidates = netdiskLaunchCandidates(rawUrl);
  if (candidates.isEmpty) return false;

  for (final uri in candidates) {
    try {
      final ok = await launchUrl(uri, mode: LaunchMode.externalApplication);
      if (ok) return true;
    } catch (_) {
      continue;
    }
  }
  return false;
}
