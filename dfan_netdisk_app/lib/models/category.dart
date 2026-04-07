class CategoryItem {
  CategoryItem({
    required this.id,
    required this.name,
    required this.slug,
    this.sortOrder = 0,
    this.status = 1,
  });

  final int id;
  final String name;
  final String slug;
  final int sortOrder;
  final int status;

  factory CategoryItem.fromJson(Map<String, dynamic> json) {
    return CategoryItem(
      id: (json['id'] as num).toInt(),
      name: json['name'] as String? ?? '',
      slug: json['slug'] as String? ?? '',
      sortOrder: (json['sort_order'] as num?)?.toInt() ?? 0,
      status: (json['status'] as num?)?.toInt() ?? 1,
    );
  }
}
