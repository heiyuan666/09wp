import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/api_exception.dart';
import '../models/user_profile.dart';
import '../state/app_state.dart';
import 'qr_scanner_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key, this.initialQrSid});

  /// 若已从扫码页带回会话 id（例如从「我的」进入扫码流程）
  final String? initialQrSid;

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _user = TextEditingController();
  final _pass = TextEditingController();
  String? _qrSid;
  bool _loading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _qrSid = widget.initialQrSid;
  }

  Future<void> _openScanner() async {
    final sid = await Navigator.of(context).push<String>(
      MaterialPageRoute(builder: (_) => const QrScannerScreen()),
    );
    if (!mounted || sid == null) return;
    setState(() {
      _qrSid = sid;
      _error = null;
    });
  }

  Future<void> _submit() async {
    final u = _user.text.trim();
    final p = _pass.text;
    if (u.isEmpty || p.isEmpty) {
      setState(() => _error = '请填写账号和密码');
      return;
    }

    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final app = context.read<AppState>();
      final LoginResult login;
      final sid = _qrSid?.trim();
      if (sid != null && sid.isNotEmpty) {
        login = await app.api.qrLoginConfirm(
          sid: sid,
          username: u,
          password: p,
        );
      } else {
        login = await app.api.login(u, p);
      }
      await app.setToken(login.token);
      if (!mounted) return;
      Navigator.of(context).pop(true);
    } on DioException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message ?? '网络错误';
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e is ApiException ? e.message : e.toString();
        _loading = false;
      });
    }
  }

  @override
  void dispose() {
    _user.dispose();
    _pass.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final hasQr = _qrSid != null && _qrSid!.isNotEmpty;

    return Scaffold(
      appBar: AppBar(title: const Text('用户登录')),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          OutlinedButton.icon(
            onPressed: _loading ? null : _openScanner,
            icon: const Icon(Icons.qr_code_scanner),
            label: Text(hasQr ? '重新扫描二维码' : '扫码登录（网页展示二维码）'),
          ),
          if (hasQr) ...[
            const SizedBox(height: 8),
            Align(
              alignment: Alignment.centerLeft,
              child: InputChip(
                label: const Text('已绑定扫码会话，登录将同步到网页端'),
                onDeleted: () => setState(() => _qrSid = null),
              ),
            ),
          ],
          const SizedBox(height: 20),
          TextField(
            controller: _user,
            decoration: const InputDecoration(
              labelText: '用户名或邮箱',
              border: OutlineInputBorder(),
            ),
            textInputAction: TextInputAction.next,
            autocorrect: false,
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _pass,
            decoration: const InputDecoration(
              labelText: '密码',
              border: OutlineInputBorder(),
            ),
            obscureText: true,
            onSubmitted: (_) => _submit(),
          ),
          if (_error != null) ...[
            const SizedBox(height: 12),
            Text(
              _error!,
              style: TextStyle(color: Theme.of(context).colorScheme.error),
            ),
          ],
          const SizedBox(height: 24),
          FilledButton(
            onPressed: _loading ? null : _submit,
            child: _loading
                ? const SizedBox(
                    height: 22,
                    width: 22,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('登录'),
          ),
        ],
      ),
    );
  }
}
