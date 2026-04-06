class NetdiskResource {
  NetdiskResource({
    required this.id,
    required this.title,
    required this.link,
    required this.categoryId,
    this.description = '',
    this.extractCode = '',
    this.cover = '',
    this.source = '',
    this.externalId = '',
    this.tags = '',
    this.viewCount = 0,
    this.linkValid = true,
    this.transferStatus = '',
    this.extraLinks = const [],
    this.createdAt,
  });

  final int id;
  final String title;
  final String link;
  final int categoryId;
  final String description;
  final String extractCode;
  final String cover;
  /// 来源：manual / telegram 等（与后端 resources.source 一致）
  final String source;
  final String externalId;
  final String tags;
  final int viewCount;
  final bool linkValid;
  final String transferStatus;
  final List<String> extraLinks;
  final DateTime? createdAt;

  factory NetdiskResource.fromJson(Map<String, dynamic> json) {
    final extra = json['extra_links'];
    List<String> extras = [];
    if (extra is List) {
      extras = extra.map((e) => e.toString()).toList();
    }

    DateTime? created;
    final ca = json['created_at'];
    if (ca is String && ca.isNotEmpty) {
      created = DateTime.tryParse(ca);
    }

    return NetdiskResource(
      id: (json['id'] as num).toInt(),
      title: json['title'] as String? ?? '',
      link: json['link'] as String? ?? '',
      categoryId: (json['category_id'] as num?)?.toInt() ?? 0,
      description: json['description'] as String? ?? '',
      extractCode: json['extract_code'] as String? ?? '',
      cover: json['cover'] as String? ?? '',
      source: json['source'] as String? ?? '',
      externalId: json['external_id'] as String? ?? '',
      tags: json['tags'] as String? ?? '',
      viewCount: (json['view_count'] as num?)?.toInt() ?? 0,
      linkValid: json['link_valid'] as bool? ?? true,
      transferStatus: json['transfer_status'] as String? ?? '',
      extraLinks: extras,
      createdAt: created,
    );
  }
}

class PageResult<T> {
  PageResult({required this.list, required this.total});

  final List<T> list;
  final int total;
}
