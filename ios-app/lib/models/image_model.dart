/// 图片数据模型
class ImageModel {
  final String id;
  final String url;
  final String originalFormat;
  final int originalSize;
  final int processedSize;
  final int width;
  final int height;
  final DateTime createdAt;

  ImageModel({
    required this.id,
    required this.url,
    required this.originalFormat,
    required this.originalSize,
    required this.processedSize,
    required this.width,
    required this.height,
    required this.createdAt,
  });

  factory ImageModel.fromJson(Map<String, dynamic> json) {
    return ImageModel(
      id: json['id'] ?? '',
      url: json['url'] ?? '',
      originalFormat: json['original_format'] ?? '',
      originalSize: json['original_size'] ?? 0,
      processedSize: json['processed_size'] ?? 0,
      width: json['width'] ?? 0,
      height: json['height'] ?? 0,
      createdAt: DateTime.tryParse(json['created_at'] ?? '') ?? DateTime.now(),
    );
  }

  /// 获取完整图片 URL
  String get fullUrl => 'https://img.leisureea.cn$url';

  /// Markdown 格式
  String get markdownUrl => '![]($fullUrl)';

  /// HTML 格式
  String get htmlUrl => '<img src="$fullUrl" alt="image" />';

  /// BBCode 格式
  String get bbcodeUrl => '[img]$fullUrl[/img]';

  /// 格式化文件大小
  String get formattedSize {
    if (processedSize < 1024) return '$processedSize B';
    if (processedSize < 1024 * 1024) {
      return '${(processedSize / 1024).toStringAsFixed(1)} KB';
    }
    return '${(processedSize / (1024 * 1024)).toStringAsFixed(2)} MB';
  }

  /// 格式化原始大小
  String get formattedOriginalSize {
    if (originalSize < 1024) return '$originalSize B';
    if (originalSize < 1024 * 1024) {
      return '${(originalSize / 1024).toStringAsFixed(1)} KB';
    }
    return '${(originalSize / (1024 * 1024)).toStringAsFixed(2)} MB';
  }
}

/// 分页列表响应
class PaginatedImages {
  final List<ImageModel> items;
  final int total;
  final int page;
  final int pageSize;
  final int totalPages;

  PaginatedImages({
    required this.items,
    required this.total,
    required this.page,
    required this.pageSize,
    required this.totalPages,
  });

  factory PaginatedImages.fromJson(Map<String, dynamic> json) {
    final itemsList = (json['items'] as List?) ?? [];
    return PaginatedImages(
      items: itemsList.map((e) => ImageModel.fromJson(e)).toList(),
      total: json['total'] ?? 0,
      page: json['page'] ?? 1,
      pageSize: json['page_size'] ?? 20,
      totalPages: json['total_pages'] ?? 0,
    );
  }
}
