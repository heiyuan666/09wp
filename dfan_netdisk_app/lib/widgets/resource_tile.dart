import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../models/resource.dart';
import '../state/app_state.dart';
import '../utils/cover_proxy.dart';
import '../utils/netdisk_launch.dart';

class ResourceTile extends StatelessWidget {
  const ResourceTile({
    super.key,
    required this.resource,
    this.onTap,
  });

  final NetdiskResource resource;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final variant = theme.colorScheme.onSurfaceVariant;
    final small = theme.textTheme.bodySmall?.copyWith(color: variant);
    final labels = netdiskLabelsForResource(
      link: resource.link,
      extraLinks: resource.extraLinks,
    );

    final hasSubtitle = labels.isNotEmpty || resource.tags.isNotEmpty;
    final cardColor = theme.colorScheme.surfaceContainerLow;

    final leading = Consumer<AppState>(
      builder: (context, app, _) {
        if (resource.cover.trim().isEmpty) {
          return const SizedBox.shrink();
        }
        final url = resolveResourceCoverDisplayUrl(
          cover: resource.cover,
          source: resource.source,
          externalId: resource.externalId,
          proxyTemplate: app.tgImageProxyUrl,
          apiBaseUrl: app.baseUrl,
        );
        if (url.isEmpty) return const SizedBox.shrink();
        return ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: SizedBox(
            width: 48,
            height: 48,
            child: Image.network(
              url,
              fit: BoxFit.cover,
              errorBuilder: (_, __, ___) => const SizedBox.shrink(),
            ),
          ),
        );
      },
    );

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 5),
      child: Material(
        color: cardColor,
        elevation: 0,
        borderRadius: BorderRadius.circular(14),
        clipBehavior: Clip.antiAlias,
        child: InkWell(
          onTap: onTap,
          child: ListTile(
            leading: leading,
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
            isThreeLine: hasSubtitle,
            title: Text(
              resource.title,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: theme.textTheme.titleSmall?.copyWith(
                fontWeight: FontWeight.w600,
                letterSpacing: -0.2,
              ),
            ),
            subtitle: hasSubtitle
                ? Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (labels.isNotEmpty)
                        Text(
                          '网盘 ${labels.join(' · ')}',
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: small,
                        ),
                      if (labels.isNotEmpty && resource.tags.isNotEmpty)
                        const SizedBox(height: 4),
                      if (resource.tags.isNotEmpty)
                        Text(
                          resource.tags,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: small,
                        ),
                    ],
                  )
                : null,
            trailing: resource.linkValid
                ? null
                : Icon(Icons.link_off, color: theme.colorScheme.error, size: 20),
          ),
        ),
      ),
    );
  }
}
