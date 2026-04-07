import 'package:dio/dio.dart';

import '../models/api_exception.dart';
import '../models/category.dart';
import '../models/home_data.dart';
import '../models/resource.dart';
import '../models/user_profile.dart';

/// 解析统一响应：{ code, message, data }
dynamic unwrapData(dynamic responseData) {
  if (responseData is! Map) {
    throw ApiException(0, '响应格式错误');
  }
  final map = Map<String, dynamic>.from(responseData);
  final code = (map['code'] as num?)?.toInt() ?? 0;
  if (code != 200) {
    throw ApiException(code, map['message']?.toString() ?? '请求失败');
  }
  return map['data'];
}

class NetdiskApi {
  NetdiskApi({
    required String Function() baseUrl,
    required String? Function() token,
  })  : _baseUrl = baseUrl,
        _token = token {
    _dio = Dio(
      BaseOptions(
        connectTimeout: const Duration(seconds: 20),
        receiveTimeout: const Duration(seconds: 45),
        headers: {'Accept': 'application/json'},
      ),
    );
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          options.baseUrl = _baseUrl();
          final t = _token();
          if (t != null && t.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $t';
          }
          handler.next(options);
        },
      ),
    );
  }

  final String Function() _baseUrl;
  final String? Function() _token;
  late final Dio _dio;

  Future<HomeData> home() async {
    final res = await _dio.get<dynamic>('/home');
    final data = unwrapData(res.data);
    if (data is! Map<String, dynamic>) {
      throw ApiException(0, '首页数据格式错误');
    }
    return HomeData.fromJson(data);
  }

  Future<List<CategoryItem>> categories() async {
    final res = await _dio.get<dynamic>('/categories');
    final data = unwrapData(res.data);
    if (data is! List) return [];
    return data
        .whereType<Map>()
        .map((e) => CategoryItem.fromJson(Map<String, dynamic>.from(e)))
        .toList();
  }

  Future<PageResult<NetdiskResource>> resources({
    int page = 1,
    int pageSize = 20,
    String sort = 'latest',
    int? categoryId,
    String? platform,
  }) async {
    final res = await _dio.get<dynamic>(
      '/resources',
      queryParameters: <String, dynamic>{
        'page': page,
        'page_size': pageSize,
        'sort': sort,
        if (categoryId != null) 'category_id': categoryId,
        if (platform != null && platform.isNotEmpty) 'platform': platform,
      },
    );
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '列表格式错误');
    }
    final m = Map<String, dynamic>.from(data);
    final listRaw = m['list'];
    final total = (m['total'] as num?)?.toInt() ?? 0;
    final list = <NetdiskResource>[];
    if (listRaw is List) {
      for (final item in listRaw) {
        if (item is Map) {
          list.add(
            NetdiskResource.fromJson(Map<String, dynamic>.from(item)),
          );
        }
      }
    }
    return PageResult(list: list, total: total);
  }

  Future<PageResult<NetdiskResource>> search({
    required String q,
    int page = 1,
    int pageSize = 20,
    String sort = 'relevance',
    int? categoryId,
    String? platform,
  }) async {
    final res = await _dio.get<dynamic>(
      '/search',
      queryParameters: <String, dynamic>{
        'q': q,
        'page': page,
        'page_size': pageSize,
        'sort': sort,
        if (categoryId != null) 'category_id': categoryId,
        if (platform != null && platform.isNotEmpty) 'platform': platform,
      },
    );
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '搜索格式错误');
    }
    final m = Map<String, dynamic>.from(data);
    final listRaw = m['list'];
    final total = (m['total'] as num?)?.toInt() ?? 0;
    final list = <NetdiskResource>[];
    if (listRaw is List) {
      for (final item in listRaw) {
        if (item is Map) {
          list.add(
            NetdiskResource.fromJson(Map<String, dynamic>.from(item)),
          );
        }
      }
    }
    return PageResult(list: list, total: total);
  }

  Future<NetdiskResource> resourceDetail(String id) async {
    final res = await _dio.get<dynamic>('/resources/$id');
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '详情格式错误');
    }
    return NetdiskResource.fromJson(Map<String, dynamic>.from(data));
  }

  /// 获取可用分享链接（可能触发转存逻辑，与 Web 一致）
  Future<AccessLinkResult> accessLink(String id) async {
    final res = await _dio.post<dynamic>('/resources/$id/access-link');
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '链接响应格式错误');
    }
    final m = Map<String, dynamic>.from(data);
    final linksRaw = m['links'];
    final links = linksRaw is List
        ? linksRaw.map((e) => e.toString()).toList()
        : const <String>[];

    return AccessLinkResult(
      status: m['status'] as String? ?? '',
      link: m['link'] as String?,
      message: m['message'] as String?,
      extraLinks: (m['extra_links'] is List)
          ? (m['extra_links'] as List).map((e) => e.toString()).toList()
          : const [],
      links: links,
    );
  }

  Future<LoginResult> login(String username, String password) async {
    final res = await _dio.post<dynamic>(
      '/auth/login',
      data: {'username': username, 'password': password},
    );
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '登录响应格式错误');
    }
    return LoginResult.fromJson(Map<String, dynamic>.from(data));
  }

  /// 扫码登录确认（与 [login] 返回结构一致）
  Future<LoginResult> qrLoginConfirm({
    required String sid,
    required String username,
    required String password,
  }) async {
    final res = await _dio.post<dynamic>(
      '/auth/qr/confirm',
      data: {
        'sid': sid,
        'username': username,
        'password': password,
      },
    );
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '扫码登录响应格式错误');
    }
    return LoginResult.fromJson(Map<String, dynamic>.from(data));
  }

  /// Web 端创建会话并展示二维码；App 端一般不需要调用
  Future<Map<String, dynamic>> qrLoginCreate() async {
    final res = await _dio.post<dynamic>('/auth/qr/create');
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '创建扫码会话失败');
    }
    return Map<String, dynamic>.from(data);
  }

  /// Web 端轮询：`status` 为 pending | confirmed | expired
  Future<Map<String, dynamic>> qrLoginStatus(String sid) async {
    final res = await _dio.get<dynamic>('/auth/qr/status/$sid');
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '查询扫码状态失败');
    }
    return Map<String, dynamic>.from(data);
  }

  Future<UserProfile> profile() async {
    final res = await _dio.get<dynamic>('/user/profile');
    final data = unwrapData(res.data);
    if (data is! Map) {
      throw ApiException(0, '用户信息格式错误');
    }
    return UserProfile.fromJson(Map<String, dynamic>.from(data));
  }

  Future<Map<String, dynamic>> publicConfig() async {
    final res = await _dio.get<dynamic>('/public/config');
    final data = unwrapData(res.data);
    if (data is! Map) {
      return {};
    }
    return Map<String, dynamic>.from(data);
  }
}

class AccessLinkResult {
  AccessLinkResult({
    required this.status,
    this.link,
    this.message,
    this.extraLinks = const [],
    this.links = const [],
  });

  final String status;
  final String? link;
  final String? message;
  final List<String> extraLinks;
  /// 后端合并后的全部分享链接（若有则优先使用）
  final List<String> links;

  /// 去重后的可用 URL 列表
  List<String> get allUrls {
    final out = <String>[];
    void add(String? s) {
      final t = s?.trim();
      if (t == null || t.isEmpty) return;
      if (!out.contains(t)) out.add(t);
    }

    for (final x in links) {
      add(x);
    }
    add(link);
    for (final x in extraLinks) {
      add(x);
    }
    return out;
  }
}
