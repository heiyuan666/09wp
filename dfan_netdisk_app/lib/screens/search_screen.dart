import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/api_exception.dart';
import '../models/resource.dart';
import '../navigation/app_routes.dart';
import '../state/app_state.dart';
import '../widgets/resource_tile.dart';
import 'resource_detail_screen.dart';

class SearchScreen extends StatefulWidget {
  const SearchScreen({super.key, this.initialQuery});

  final String? initialQuery;

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  late final TextEditingController _q;
  final List<NetdiskResource> _items = [];
  int _total = 0;
  int _page = 1;
  bool _loading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _q = TextEditingController(text: widget.initialQuery ?? '');
    if ((widget.initialQuery ?? '').trim().isNotEmpty) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _search(reset: true));
    }
  }

  @override
  void dispose() {
    _q.dispose();
    super.dispose();
  }

  Future<void> _search({bool reset = false}) async {
    final keyword = _q.text.trim();
    if (keyword.isEmpty) {
      setState(() {
        _error = '请输入关键词';
        _items.clear();
        _total = 0;
      });
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
      if (reset) {
        _page = 1;
        _items.clear();
      }
    });
    try {
      final app = context.read<AppState>();
      final page = await app.api.search(
        q: keyword,
        page: reset ? 1 : _page,
        pageSize: 20,
        sort: 'relevance',
      );
      if (!mounted) return;
      setState(() {
        if (reset) {
          _items
            ..clear()
            ..addAll(page.list);
          _page = 1;
        } else {
          _items.addAll(page.list);
        }
        _total = page.total;
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

  Future<void> _loadMore() async {
    if (_loading || _items.length >= _total) return;
    setState(() => _page += 1);
    final keyword = _q.text.trim();
    setState(() => _loading = true);
    try {
      final app = context.read<AppState>();
      final page = await app.api.search(
        q: keyword,
        page: _page,
        pageSize: 20,
        sort: 'relevance',
      );
      if (!mounted) return;
      setState(() {
        _items.addAll(page.list);
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _page -= 1;
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final canPop = Navigator.canPop(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('搜索'),
        automaticallyImplyLeading: canPop,
      ),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: TextField(
              controller: _q,
              decoration: InputDecoration(
                hintText: '输入关键词，搜索网盘资源',
                border: const OutlineInputBorder(),
                isDense: true,
                suffixIcon: IconButton(
                  icon: const Icon(Icons.search),
                  onPressed: () => _search(reset: true),
                ),
              ),
              textInputAction: TextInputAction.search,
              onSubmitted: (_) => _search(reset: true),
            ),
          ),
          if (_error != null)
            MaterialBanner(
              content: Text(_error!),
              actions: [
                TextButton(
                  onPressed: () => setState(() => _error = null),
                  child: const Text('关闭'),
                ),
              ],
            ),
          Expanded(
            child: _items.isEmpty && !_loading
                ? Center(
                    child: Text(
                      '输入关键词后点搜索',
                      style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                            color: Theme.of(context).colorScheme.onSurfaceVariant,
                          ),
                    ),
                  )
                : ListView.builder(
                    // 最后一格：loading /「加载更多」/ 占位，不要与 _items 重复计数
                    itemCount: _items.length + 1,
                    itemBuilder: (context, i) {
                      if (i == _items.length) {
                        if (_loading) {
                          return const Padding(
                            padding: EdgeInsets.all(16),
                            child: Center(child: CircularProgressIndicator()),
                          );
                        }
                        if (_items.length < _total) {
                          return TextButton(
                            onPressed: _loadMore,
                            child: const Text('加载更多'),
                          );
                        }
                        return const SizedBox.shrink();
                      }
                      final r = _items[i];
                      return ResourceTile(
                        resource: r,
                        onTap: () {
                          Navigator.of(context).push(
                            fadeScaleRoute<void>(
                              ResourceDetailScreen(
                                resourceId: r.id.toString(),
                              ),
                            ),
                          );
                        },
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}
