import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';

import '../models/api_exception.dart';
import '../models/resource.dart';
import '../state/app_state.dart';
import '../utils/cover_proxy.dart';
import '../utils/netdisk_launch.dart';

class ResourceDetailScreen extends StatefulWidget {
  const ResourceDetailScreen({super.key, required this.resourceId});

  final String resourceId;

  @override
  State<ResourceDetailScreen> createState() => _ResourceDetailScreenState();
}

class _ResourceDetailScreenState extends State<ResourceDetailScreen> {
  NetdiskResource? _res;
  Object? _error;
  bool _loading = true;

  /// 最近一次 access-link 解析出的链接（含后端 `links` 合并）
  List<String>? _accessUrls;
  String? _accessMessage;
  String? _accessStatus;
  bool _accessLoading = false;
  String? _launchHint;

  @override
  void initState() {
    super.initState();
    _load();
  }

  List<String> _displayUrls() {
    if (_accessUrls != null && _accessUrls!.isNotEmpty) {
      return _accessUrls!;
    }
    final r = _res;
    if (r == null) return [];
    final out = <String>[];
    final main = r.link.trim();
    if (main.isNotEmpty) out.add(main);
    for (final x in r.extraLinks) {
      final t = x.trim();
      if (t.isNotEmpty && !out.contains(t)) out.add(t);
    }
    return out;
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
      _accessUrls = null;
      _accessMessage = null;
      _accessStatus = null;
    });
    try {
      final app = context.read<AppState>();
      final r = await app.api.resourceDetail(widget.resourceId);
      if (!mounted) return;
      setState(() {
        _res = r;
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

  Future<void> _refreshAccessLinks() async {
    setState(() {
      _accessLoading = true;
      _launchHint = null;
    });
    try {
      final app = context.read<AppState>();
      final result = await app.api.accessLink(widget.resourceId);
      if (!mounted) return;
      setState(() {
        _accessStatus = result.status;
        _accessMessage = result.message;
        final urls = result.allUrls;
        _accessUrls = urls.isNotEmpty ? urls : null;
        _accessLoading = false;
      });
      if (result.message != null && result.message!.isNotEmpty && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(result.message!)),
        );
      }
    } on DioException catch (e) {
      if (!mounted) return;
      setState(() {
        _accessLoading = false;
        _launchHint = e.message ?? '网络错误';
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _accessLoading = false;
        _launchHint = e is ApiException ? e.message : e.toString();
      });
    }
  }

  Future<void> _openNetdiskUrl(String url) async {
    setState(() => _launchHint = null);
    final ok = await launchNetdiskInApp(url);
    if (!mounted) return;
    if (!ok) {
      setState(() {
        _launchHint = '未能唤起网盘 App，可复制链接到浏览器或已安装的网盘中打开';
      });
    }
  }

  Future<void> _copyUrl(String url) async {
    await Clipboard.setData(ClipboardData(text: url));
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('链接已复制')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('资源详情'),
        actions: [
          if (_res != null)
            IconButton(
              tooltip: '同步分享链接',
              onPressed: _accessLoading ? null : _refreshAccessLinks,
              icon: _accessLoading
                  ? const SizedBox(
                      width: 22,
                      height: 22,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.sync),
            ),
        ],
      ),
      body: _buildBody(context),
    );
  }

  Widget _buildBody(BuildContext context) {
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null) {
      final msg = _error is ApiException
          ? (_error! as ApiException).message
          : _error.toString();
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(msg),
              const SizedBox(height: 12),
              FilledButton(onPressed: _load, child: const Text('重试')),
            ],
          ),
        ),
      );
    }
    final r = _res!;
    final urls = _displayUrls();
    final app = context.watch<AppState>();
    final coverUrl = resolveResourceCoverDisplayUrl(
      cover: r.cover,
      source: r.source,
      externalId: r.externalId,
      proxyTemplate: app.tgImageProxyUrl,
      apiBaseUrl: app.baseUrl,
    );

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(r.title, style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 12),
          if (coverUrl.isNotEmpty)
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: AspectRatio(
                aspectRatio: 16 / 9,
                child: Image.network(
                  coverUrl,
                  fit: BoxFit.cover,
                  errorBuilder: (_, __, ___) => const SizedBox.shrink(),
                ),
              ),
            ),
          if (r.description.isNotEmpty) ...[
            const SizedBox(height: 12),
            Text(r.description, style: Theme.of(context).textTheme.bodyMedium),
          ],
          if (r.extractCode.isNotEmpty) ...[
            const SizedBox(height: 12),
            SelectableText('提取码：${r.extractCode}'),
          ],
          if (r.tags.isNotEmpty) ...[
            const SizedBox(height: 8),
            Text('标签：${r.tags}', style: Theme.of(context).textTheme.bodySmall),
          ],
          const SizedBox(height: 8),
          Text(
            '浏览 ${r.viewCount} · 链接${r.linkValid ? "有效" : "失效"}',
            style: Theme.of(context).textTheme.labelMedium,
          ),
          if (_accessStatus == 'pending') ...[
            const SizedBox(height: 12),
            Card(
              color: Theme.of(context).colorScheme.surfaceContainerHighest,
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Text(
                  _accessMessage ?? '正在准备可用链接，请稍后点击右上角同步重试',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ),
            ),
          ],
          const SizedBox(height: 20),
          Row(
            children: [
              Text(
                '网盘链接',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const Spacer(),
              TextButton.icon(
                onPressed: _accessLoading ? null : _refreshAccessLinks,
                icon: const Icon(Icons.cloud_download_outlined, size: 18),
                label: const Text('获取最新'),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            '点击条目将优先尝试打开对应网盘 App，失败时请复制链接。',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
                ),
          ),
          const SizedBox(height: 8),
          if (urls.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 24),
              child: Center(
                child: Text(
                  '暂无链接，请点击「获取最新」',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ),
            )
          else
            Card(
              clipBehavior: Clip.antiAlias,
              child: Column(
                children: [
                  for (var i = 0; i < urls.length; i++) ...[
                    if (i > 0) const Divider(height: 1),
                    ListTile(
                      leading: Icon(
                        Icons.folder_shared_outlined,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      title: Text(netdiskLabelForUrl(urls[i])),
                      subtitle: Text(
                        urls[i],
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                      trailing: IconButton(
                        icon: const Icon(Icons.copy_outlined),
                        tooltip: '复制',
                        onPressed: () => _copyUrl(urls[i]),
                      ),
                      onTap: () => _openNetdiskUrl(urls[i]),
                    ),
                  ],
                ],
              ),
            ),
          if (_launchHint != null) ...[
            const SizedBox(height: 12),
            Text(
              _launchHint!,
              style: TextStyle(
                color: Theme.of(context).colorScheme.error,
                fontSize: 13,
              ),
            ),
          ],
          const SizedBox(height: 24),
        ],
      ),
    );
  }
}
