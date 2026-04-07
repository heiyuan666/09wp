import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/api_exception.dart';
import '../models/category.dart';
import '../models/resource.dart';
import '../navigation/app_routes.dart';
import '../state/app_state.dart';
import '../widgets/resource_tile.dart';
import 'resource_detail_screen.dart';

class BrowseScreen extends StatefulWidget {
  const BrowseScreen({super.key, this.initialCategoryId});

  /// 从首页分类进入时可传入
  final int? initialCategoryId;

  @override
  State<BrowseScreen> createState() => _BrowseScreenState();
}

class _BrowseScreenState extends State<BrowseScreen> {
  List<CategoryItem> _cats = [];
  late int? _categoryId;
  String _sort = 'latest';
  int _page = 1;
  final List<NetdiskResource> _items = [];
  int _total = 0;
  bool _loading = false;
  bool _loadingMore = false;
  String? _error;
  final _scroll = ScrollController();

  @override
  void initState() {
    super.initState();
    _categoryId = widget.initialCategoryId;
    _scroll.addListener(_onScroll);
    _bootstrap();
  }

  @override
  void dispose() {
    _scroll.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_loadingMore || _loading) return;
    if (!_scroll.hasClients) return;
    final max = _scroll.position.maxScrollExtent;
    if (max <= 0) return;
    if (_scroll.position.pixels > max - 200) {
      _loadMore();
    }
  }

  Future<void> _bootstrap() async {
    setState(() {
      _error = null;
      _loading = true;
    });
    try {
      final app = context.read<AppState>();
      final cats = await app.api.categories();
      if (!mounted) return;
      setState(() {
        _cats = cats;
        _loading = false;
      });
      await _refresh();
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e is ApiException ? e.message : e.toString();
        _loading = false;
      });
    }
  }

  Future<void> _refresh() async {
    setState(() {
      _page = 1;
      _items.clear();
      _error = null;
      _loading = true;
    });
    try {
      final app = context.read<AppState>();
      final page = await app.api.resources(
        page: 1,
        pageSize: 20,
        sort: _sort,
        categoryId: _categoryId,
      );
      if (!mounted) return;
      setState(() {
        _items.addAll(page.list);
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
    if (_items.length >= _total) return;
    setState(() => _loadingMore = true);
    try {
      final app = context.read<AppState>();
      final next = _page + 1;
      final page = await app.api.resources(
        page: next,
        pageSize: 20,
        sort: _sort,
        categoryId: _categoryId,
      );
      if (!mounted) return;
      setState(() {
        _page = next;
        _items.addAll(page.list);
        _loadingMore = false;
      });
    } catch (_) {
      if (mounted) setState(() => _loadingMore = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('资源列表'),
        actions: [
          PopupMenuButton<String>(
            initialValue: _sort,
            onSelected: (v) {
              setState(() => _sort = v);
              _refresh();
            },
            itemBuilder: (context) => const [
              PopupMenuItem(value: 'latest', child: Text('最新')),
              PopupMenuItem(value: 'hot', child: Text('最热')),
            ],
          ),
        ],
      ),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
            child: DropdownButtonFormField<int?>(
              key: ValueKey(_categoryId),
              initialValue: _categoryId,
              decoration: const InputDecoration(
                labelText: '分类筛选',
                border: OutlineInputBorder(),
                isDense: true,
              ),
              items: [
                const DropdownMenuItem<int?>(value: null, child: Text('全部分类')),
                ..._cats.map(
                  (c) => DropdownMenuItem<int?>(
                    value: c.id,
                    child: Text(c.name),
                  ),
                ),
              ],
              onChanged: (v) {
                setState(() => _categoryId = v);
                _refresh();
              },
            ),
          ),
          const SizedBox(height: 8),
          Expanded(child: _buildList()),
        ],
      ),
    );
  }

  Widget _buildList() {
    if (_loading && _items.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null && _items.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_error!),
              const SizedBox(height: 12),
              FilledButton(onPressed: _refresh, child: const Text('重试')),
            ],
          ),
        ),
      );
    }
    if (_items.isEmpty) {
      return const Center(child: Text('暂无数据'));
    }
    return RefreshIndicator(
      onRefresh: _refresh,
      child: ListView.builder(
        controller: _scroll,
        itemCount: _items.length + (_loadingMore ? 1 : 0),
        itemBuilder: (context, i) {
          if (i >= _items.length) {
            return const Padding(
              padding: EdgeInsets.all(16),
              child: Center(child: CircularProgressIndicator()),
            );
          }
          final r = _items[i];
          return ResourceTile(
            resource: r,
            onTap: () {
              Navigator.of(context).push(
                fadeScaleRoute<void>(
                  ResourceDetailScreen(resourceId: r.id.toString()),
                ),
              );
            },
          );
        },
      ),
    );
  }
}
