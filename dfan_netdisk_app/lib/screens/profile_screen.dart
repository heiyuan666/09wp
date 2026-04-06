import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/api_exception.dart';
import '../models/user_profile.dart';
import '../state/app_state.dart';
import 'login_screen.dart';
import 'qr_scanner_screen.dart';
import 'settings_screen.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  UserProfile? _profile;
  Object? _error;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      if (context.read<AppState>().isLoggedIn) {
        _loadProfile();
      }
    });
  }

  Future<void> _loadProfile() async {
    final app = context.read<AppState>();
    if (!app.isLoggedIn) return;
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final p = await app.api.profile();
      if (!mounted) return;
      setState(() {
        _profile = p;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e;
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final app = context.watch<AppState>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('我的'),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings),
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => const SettingsScreen(),
                ),
              );
            },
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          if (!app.isLoggedIn) ...[
            const Text('登录后可使用收藏、投稿等功能（接口已预留）。'),
            const SizedBox(height: 16),
            FilledButton(
              onPressed: () async {
                final ok = await Navigator.of(context).push<bool>(
                  MaterialPageRoute(builder: (_) => const LoginScreen()),
                );
                if (ok == true && context.mounted) {
                  setState(() {
                    _profile = null;
                    _error = null;
                  });
                  await _loadProfile();
                }
              },
              child: const Text('账号密码登录'),
            ),
            const SizedBox(height: 12),
            OutlinedButton.icon(
              onPressed: () async {
                final sid = await Navigator.of(context).push<String>(
                  MaterialPageRoute(builder: (_) => const QrScannerScreen()),
                );
                if (sid == null || !context.mounted) return;
                final ok = await Navigator.of(context).push<bool>(
                  MaterialPageRoute(
                    builder: (_) => LoginScreen(initialQrSid: sid),
                  ),
                );
                if (ok == true && context.mounted) {
                  setState(() {
                    _profile = null;
                    _error = null;
                  });
                  await _loadProfile();
                }
              },
              icon: const Icon(Icons.qr_code_2_outlined),
              label: const Text('扫码登录'),
            ),
          ] else ...[
            if (_loading)
              const Center(child: Padding(
                padding: EdgeInsets.all(24),
                child: CircularProgressIndicator(),
              ))
            else if (_error != null)
              Text(
                _error is ApiException
                    ? (_error! as ApiException).message
                    : _error is DioException
                        ? (_error! as DioException).message ?? '请求失败'
                        : _error.toString(),
              )
            else if (_profile != null) ...[
              ListTile(
                leading: const CircleAvatar(child: Icon(Icons.person)),
                title: Text(_profile!.username),
                subtitle: Text(_profile!.email),
              ),
              if (_profile!.name.isNotEmpty)
                ListTile(
                  title: const Text('显示名'),
                  subtitle: Text(_profile!.name),
                ),
            ],
            const SizedBox(height: 16),
            OutlinedButton(
              onPressed: () async {
                await app.logout();
                if (mounted) setState(() => _profile = null);
              },
              child: const Text('退出登录'),
            ),
          ],
        ],
      ),
    );
  }
}
