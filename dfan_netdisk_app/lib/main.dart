import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:provider/provider.dart';

import 'screens/shell_screen.dart';
import 'state/app_state.dart';
import 'theme/app_theme.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const DfanNetdiskApp());
}

class DfanNetdiskApp extends StatelessWidget {
  const DfanNetdiskApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => AppState()..load(),
      child: Consumer<AppState>(
        builder: (context, app, _) {
          if (!app.isLoaded) {
            final bootTheme = buildAppTheme();
            return MaterialApp(
              theme: bootTheme,
              home: Scaffold(
                body: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const CircularProgressIndicator()
                          .animate(
                            onPlay: (c) => c.repeat(reverse: true),
                          )
                          .scale(
                            duration: 1200.ms,
                            begin: const Offset(0.92, 0.92),
                            end: const Offset(1.04, 1.04),
                            curve: Curves.easeInOut,
                          ),
                      const SizedBox(height: 20),
                      Text(
                        '加载配置…',
                        style: TextStyle(
                          color: bootTheme.colorScheme.onSurfaceVariant,
                        ),
                      )
                          .animate(onPlay: (c) => c.repeat(reverse: true))
                          .fade(
                            begin: 0.45,
                            end: 1,
                            duration: 900.ms,
                            curve: Curves.easeInOut,
                          ),
                    ],
                  ),
                ),
              ),
            );
          }
          return MaterialApp(
            title: '网盘导航',
            theme: buildAppTheme(),
            home: const ShellScreen(),
          );
        },
      ),
    );
  }
}
