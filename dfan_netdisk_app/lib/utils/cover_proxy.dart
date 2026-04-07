/// 是否与 Web 端 `shouldApplyTgCoverProxy` 一致
bool shouldApplyTgCoverProxy({
  required String source,
  required String externalId,
  required String coverRaw,
}) {
  if (source.trim().toLowerCase() == 'telegram') return true;
  final ext = externalId.trim().toLowerCase();
  if (ext.startsWith('tg:')) return true;
  final u = coverRaw.trim().toLowerCase();
  if (u.isEmpty) return false;
  if (u.contains('telesco.pe')) return true;
  if (u.contains('telegram-cdn.org')) return true;
  if (u.contains('cdn.telegram.org')) return true;
  if (u.contains('telegram.org/file')) return true;
  return false;
}

/// 与后台 / Web 端一致的图片返代拼接（如 `https://wsrv.nl/?url=`）。
String buildProxiedImageSrc(String cover, String template) {
  final c = cover.trim();
  if (c.isEmpty) return '';

  final tmpl = template.trim();
  if (tmpl.isEmpty) return c;

  final enc = Uri.encodeComponent(c);

  if (tmpl.contains('{url}')) {
    return tmpl.replaceAll('{url}', enc);
  }
  if (tmpl.endsWith('url=')) {
    return tmpl + enc;
  }
  final idx = tmpl.indexOf('url=');
  if (idx >= 0) {
    final prefix = tmpl.substring(0, idx + 'url='.length);
    return '$prefix$enc';
  }
  if (tmpl.contains('?')) {
    if (tmpl.endsWith('?') || tmpl.endsWith('&')) {
      return '${tmpl}url=$enc';
    }
    return '$tmpl&url=$enc';
  }
  var base = tmpl;
  if (base.endsWith('/')) {
    base = base.substring(0, base.length - 1);
  }
  return '$base?url=$enc';
}

/// 从 API Base（如 `http://host:8080/api/v1`）得到站点 origin，用于拼相对封面路径。
String apiOriginFromBase(String apiBaseUrl) {
  final u = Uri.parse(apiBaseUrl.trim());
  if (!u.hasScheme || u.host.isEmpty) return '';
  final port = u.hasPort ? ':${u.port}' : '';
  return '${u.scheme}://${u.host}$port';
}

/// 将封面字段解析为可请求的绝对 URL（不套返代）。
String resolveCoverToAbsoluteUrl(String cover, String apiBaseUrl) {
  final t = cover.trim();
  if (t.isEmpty) return '';

  final origin = apiOriginFromBase(apiBaseUrl);
  final scheme = Uri.tryParse(apiBaseUrl.trim())?.scheme ?? 'https';

  if (t.startsWith('//')) {
    return '$scheme:$t';
  }
  if (t.toLowerCase().startsWith('http://') || t.toLowerCase().startsWith('https://')) {
    return t;
  }
  if (origin.isEmpty) return t;
  return '$origin${t.startsWith('/') ? '' : '/'}$t';
}

/// TG 外链封面在配置返代后用于展示的最终 URL。
String resolveResourceCoverDisplayUrl({
  required String cover,
  required String source,
  String externalId = '',
  required String proxyTemplate,
  required String apiBaseUrl,
}) {
  final resolved = resolveCoverToAbsoluteUrl(cover, apiBaseUrl);
  if (resolved.isEmpty) return '';

  final raw = cover.trim();
  final isRemote = RegExp(r'^https?://', caseSensitive: false).hasMatch(raw) ||
      raw.startsWith('//');
  final tmpl = proxyTemplate.trim();
  if (tmpl.isNotEmpty &&
      isRemote &&
      shouldApplyTgCoverProxy(
        source: source,
        externalId: externalId,
        coverRaw: raw,
      )) {
    return buildProxiedImageSrc(resolved, tmpl);
  }
  return resolved;
}
