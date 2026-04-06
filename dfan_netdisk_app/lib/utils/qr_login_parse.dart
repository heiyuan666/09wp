import 'dart:convert';

/// 从二维码原始内容解析扫码登录会话 `sid`。
/// 支持：`dfannetdisk://qr-login?sid=...`、含 `sid` 的 https URL、JSON `{"type":"dfan_qr_login","sid":"..."}`、纯 32 位 hex。
String? parseQrLoginSid(String raw) {
  var s = raw.trim();
  if (s.isEmpty) return null;

  try {
    final dynamic j = jsonDecode(s);
    if (j is Map) {
      final t = j['type']?.toString();
      if ((t == 'dfan_qr_login' || t == 'dfan_qr_admin_login') && j['sid'] != null) {
        return j['sid'].toString().trim();
      }
    }
  } catch (_) {}

  final uri = Uri.tryParse(s);
  if (uri != null && uri.hasQuery) {
    final sid = uri.queryParameters['sid']?.trim();
    if (sid != null && sid.isNotEmpty) return sid;
  }

  if (RegExp(r'^[a-fA-F0-9]{32}$').hasMatch(s)) {
    return s.toLowerCase();
  }

  return null;
}
