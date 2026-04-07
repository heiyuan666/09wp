class UserProfile {
  UserProfile({
    required this.id,
    required this.username,
    required this.email,
    this.name = '',
    this.avatar,
  });

  final int id;
  final String username;
  final String email;
  final String name;
  final String? avatar;

  factory UserProfile.fromJson(Map<String, dynamic> json) {
    final idVal = json['id'];
    final id = idVal is int
        ? idVal
        : idVal is num
            ? idVal.toInt()
            : int.tryParse(idVal?.toString() ?? '') ?? 0;
    final av = json['avatar'];
    String? avatarStr;
    if (av is String) {
      avatarStr = av;
    }
    return UserProfile(
      id: id,
      username: json['username'] as String? ?? '',
      email: json['email'] as String? ?? '',
      name: json['name'] as String? ?? '',
      avatar: avatarStr,
    );
  }
}

class LoginResult {
  LoginResult({required this.token, required this.user});

  final String token;
  final UserProfile user;

  factory LoginResult.fromJson(Map<String, dynamic> json) {
    final u = json['user'];
    if (u is! Map<String, dynamic>) {
      throw const FormatException('login: missing user');
    }
    return LoginResult(
      token: json['token'] as String? ?? '',
      user: UserProfile.fromJson(u),
    );
  }
}
