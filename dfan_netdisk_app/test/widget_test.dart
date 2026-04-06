// Basic Flutter widget smoke test for dfan_netdisk_app.

import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:dfan_netdisk_app/main.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  testWidgets('DfanNetdiskApp loads and shows shell', (WidgetTester tester) async {
    await tester.pumpWidget(const DfanNetdiskApp());

    // Allow SharedPreferences + AppState.load() to finish; loading UI uses repeating animations,
    // so avoid pumpAndSettle until the main shell is shown.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));

    expect(find.text('首页'), findsOneWidget);
  });
}
