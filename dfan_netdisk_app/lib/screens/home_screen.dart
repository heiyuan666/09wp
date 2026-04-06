import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:provider/provider.dart';

import '../models/api_exception.dart';
import '../models/home_data.dart';
import '../models/resource.dart';
import '../navigation/app_routes.dart';
import '../state/app_state.dart';
import '../widgets/resource_tile.dart';
import 'browse_screen.dart';
import 'resource_detail_screen.dart';
import 'search_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  HomeData? _data;
  String? _siteTitle;
  Object? _error;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    final app = context.read<AppState>();
    try {
      final cfg = await app.api.publicConfig();
      app.applyPublicConfigSnapshot(cfg);
      final title = cfg['site_title'] as String?;
      final home = await app.api.home();
      if (!mounted) return;
      setState(() {
        _siteTitle = title;
        _data = home;
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
    return Scaffold(
      appBar: AppBar(
        title: Text(_siteTitle ?? '网盘导航'),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {
              Navigator.of(context).push(
                fadeScaleRoute<void>(const SearchScreen()),
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _load,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _load,
        child: _buildBody(context),
      ),
    );
  }

  Widget _buildBody(BuildContext context) {
    if (_loading) {
      return ListView(
        children: [
          const SizedBox(height: 120),
          Center(
            child: const CircularProgressIndicator()
                .animate(onPlay: (c) => c.repeat(reverse: true))
                .scale(
                  duration: 1000.ms,
                  begin: const Offset(0.9, 0.9),
                  end: const Offset(1.05, 1.05),
                  curve: Curves.easeInOut,
                ),
          ),
        ],
      );
    }
    if (_error != null) {
      final msg = _error is ApiException
          ? (_error! as ApiException).message
          : _error.toString();
      return ListView(
        padding: const EdgeInsets.all(24),
        children: [
          Text('加载失败', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Text(msg),
          const SizedBox(height: 16),
          FilledButton(onPressed: _load, child: const Text('重试')),
        ],
      );
    }
    final d = _data!;
    return ListView(
      padding: const EdgeInsets.only(bottom: 24),
      children: [
        if (d.hotSearches.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
            child: Text(
              '热搜',
              style: Theme.of(context).textTheme.titleSmall,
            )
                .animate()
                .fadeIn(duration: 400.ms, curve: Curves.easeOut)
                .slideX(begin: -0.03, duration: 400.ms, curve: Curves.easeOutCubic),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Wrap(
              spacing: 8,
              runSpacing: 8,
              children: d.hotSearches.asMap().entries.map((e) {
                final h = e.value;
                final i = e.key;
                return ActionChip(
                  label: Text(h.keyword),
                  onPressed: () {
                    Navigator.of(context).push(
                      fadeScaleRoute<void>(
                        SearchScreen(initialQuery: h.keyword),
                      ),
                    );
                  },
                )
                    .animate(delay: (35 * i).ms)
                    .fadeIn(duration: 320.ms, curve: Curves.easeOutCubic)
                    .scale(
                      begin: const Offset(0.92, 0.92),
                      duration: 320.ms,
                      curve: Curves.easeOutCubic,
                    );
              }).toList(),
            ),
          ),
          const Divider(height: 24),
        ],
        if (d.categories.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 0, 16, 4),
            child: Text(
              '分类',
              style: Theme.of(context).textTheme.titleSmall,
            )
                .animate()
                .fadeIn(duration: 400.ms, curve: Curves.easeOut)
                .slideX(begin: -0.03, duration: 400.ms, curve: Curves.easeOutCubic),
          ),
          SizedBox(
            height: 40,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              itemCount: d.categories.length,
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemBuilder: (context, i) {
                final c = d.categories[i];
                return ActionChip(
                  label: Text(c.name),
                  onPressed: () {
                    Navigator.of(context).push(
                      fadeScaleRoute<void>(
                        BrowseScreen(initialCategoryId: c.id),
                      ),
                    );
                  },
                )
                    .animate(delay: (40 * i).ms)
                    .fadeIn(duration: 320.ms, curve: Curves.easeOutCubic)
                    .slideX(
                      begin: 0.04,
                      duration: 320.ms,
                      curve: Curves.easeOutCubic,
                    );
              },
            ),
          ),
          const SizedBox(height: 8),
        ],
        _sectionTitle(context, '最新'),
        ...d.latest.asMap().entries.map(
              (e) => _resourceRow(context, e.value, e.key),
            ),
        _sectionTitle(context, '热门'),
        ...d.hot.asMap().entries.map(
              (e) => _resourceRow(
                context,
                e.value,
                e.key + d.latest.length,
              ),
            ),
      ],
    );
  }

  Widget _sectionTitle(BuildContext context, String t) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Text(
        t,
        style: Theme.of(context).textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
              letterSpacing: -0.3,
            ),
      )
          .animate()
          .fadeIn(duration: 380.ms, curve: Curves.easeOut)
          .slideY(begin: 0.08, duration: 380.ms, curve: Curves.easeOutCubic),
    );
  }

  Widget _resourceRow(BuildContext context, NetdiskResource r, int listIndex) {
    final stagger = (48 * (listIndex > 14 ? 14 : listIndex)).ms;
    return ResourceTile(
      resource: r,
      onTap: () {
        Navigator.of(context).push(
          fadeScaleRoute<void>(
            ResourceDetailScreen(resourceId: r.id.toString()),
          ),
        );
      },
    )
        .animate(delay: stagger)
        .fadeIn(duration: 420.ms, curve: Curves.easeOutCubic)
        .slideY(
          begin: 0.05,
          duration: 420.ms,
          curve: Curves.easeOutCubic,
        );
  }
}
