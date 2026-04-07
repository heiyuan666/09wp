import 'category.dart';
import 'resource.dart';

class HotSearchItem {
  HotSearchItem({required this.keyword, required this.searchCount});

  final String keyword;
  final int searchCount;

  factory HotSearchItem.fromJson(Map<String, dynamic> json) {
    return HotSearchItem(
      keyword: json['keyword'] as String? ?? '',
      searchCount: (json['search_count'] as num?)?.toInt() ?? 0,
    );
  }
}

class HomeData {
  HomeData({
    required this.latest,
    required this.hot,
    required this.categories,
    required this.hotSearches,
  });

  final List<NetdiskResource> latest;
  final List<NetdiskResource> hot;
  final List<CategoryItem> categories;
  final List<HotSearchItem> hotSearches;

  factory HomeData.fromJson(Map<String, dynamic> json) {
    List<NetdiskResource> mapRes(String key) {
      final v = json[key];
      if (v is! List) return [];
      return v
          .whereType<Map<String, dynamic>>()
          .map(NetdiskResource.fromJson)
          .toList();
    }

    List<CategoryItem> cats() {
      final v = json['categories'];
      if (v is! List) return [];
      return v
          .whereType<Map<String, dynamic>>()
          .map(CategoryItem.fromJson)
          .toList();
    }

    List<HotSearchItem> hs() {
      final v = json['hot_searches'];
      if (v is! List) return [];
      return v
          .whereType<Map<String, dynamic>>()
          .map(HotSearchItem.fromJson)
          .toList();
    }

    return HomeData(
      latest: mapRes('latest'),
      hot: mapRes('hot'),
      categories: cats(),
      hotSearches: hs(),
    );
  }
}
