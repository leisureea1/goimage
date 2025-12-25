import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../services/api_service.dart';
import '../models/image_model.dart';

/// 图片管理页面
class GalleryPage extends StatefulWidget {
  const GalleryPage({super.key});

  @override
  State<GalleryPage> createState() => _GalleryPageState();
}

class _GalleryPageState extends State<GalleryPage> {
  final ApiService _api = ApiService();
  
  List<ImageModel> _images = [];
  bool _loading = true;
  String? _error;
  int _page = 1;
  int _totalPages = 1;
  bool _loadingMore = false;

  @override
  void initState() {
    super.initState();
    _loadImages();
  }

  /// 加载图片列表
  Future<void> _loadImages({bool refresh = false}) async {
    if (refresh) {
      setState(() {
        _page = 1;
        _loading = true;
        _error = null;
      });
    }
    
    try {
      final result = await _api.getImages(page: _page);
      setState(() {
        if (refresh || _page == 1) {
          _images = result.items;
        } else {
          _images.addAll(result.items);
        }
        _totalPages = result.totalPages;
        _loading = false;
        _loadingMore = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
        _loadingMore = false;
      });
    }
  }

  /// 加载更多
  Future<void> _loadMore() async {
    if (_loadingMore || _page >= _totalPages) return;
    
    setState(() {
      _loadingMore = true;
      _page++;
    });
    
    await _loadImages();
  }

  /// 删除图片
  Future<void> _deleteImage(ImageModel image) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('确认删除'),
        content: const Text('确定要删除这张图片吗？此操作不可恢复。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('删除'),
          ),
        ],
      ),
    );
    
    if (confirm != true) return;
    
    try {
      await _api.deleteImage(image.id);
      setState(() {
        _images.removeWhere((i) => i.id == image.id);
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('删除成功')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('删除失败: $e')),
        );
      }
    }
  }

  /// 复制链接
  Future<void> _copyUrl(String text, String label) async {
    await Clipboard.setData(ClipboardData(text: text));
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('$label 已复制到剪贴板')),
      );
    }
  }

  /// 显示复制选项
  void _showCopyOptions(ImageModel image) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.link),
              title: const Text('URL'),
              subtitle: Text(image.fullUrl, 
                style: const TextStyle(fontSize: 12),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              onTap: () {
                Navigator.pop(context);
                _copyUrl(image.fullUrl, 'URL');
              },
            ),
            ListTile(
              leading: const Icon(Icons.code),
              title: const Text('Markdown'),
              subtitle: Text(image.markdownUrl,
                style: const TextStyle(fontSize: 12),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              onTap: () {
                Navigator.pop(context);
                _copyUrl(image.markdownUrl, 'Markdown');
              },
            ),
            ListTile(
              leading: const Icon(Icons.html),
              title: const Text('HTML'),
              subtitle: Text(image.htmlUrl,
                style: const TextStyle(fontSize: 12),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              onTap: () {
                Navigator.pop(context);
                _copyUrl(image.htmlUrl, 'HTML');
              },
            ),
            ListTile(
              leading: const Icon(Icons.format_quote),
              title: const Text('BBCode'),
              subtitle: Text(image.bbcodeUrl,
                style: const TextStyle(fontSize: 12),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              onTap: () {
                Navigator.pop(context);
                _copyUrl(image.bbcodeUrl, 'BBCode');
              },
            ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  /// 显示图片详情
  void _showImageDetail(ImageModel image) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.7,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        expand: false,
        builder: (context, scrollController) => Column(
          children: [
            // 拖动指示器
            Container(
              margin: const EdgeInsets.symmetric(vertical: 12),
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            
            Expanded(
              child: SingleChildScrollView(
                controller: scrollController,
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // 图片预览
                    ClipRRect(
                      borderRadius: BorderRadius.circular(12),
                      child: CachedNetworkImage(
                        imageUrl: image.fullUrl,
                        fit: BoxFit.contain,
                      ),
                    ),
                    const SizedBox(height: 20),
                    
                    // 信息卡片
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          children: [
                            _buildDetailRow('ID', image.id),
                            _buildDetailRow('格式', image.originalFormat.toUpperCase()),
                            _buildDetailRow('大小', image.formattedSize),
                            _buildDetailRow('尺寸', '${image.width} × ${image.height}'),
                            _buildDetailRow('上传时间', _formatDate(image.createdAt)),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    
                    // 链接
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.grey.shade100,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: SelectableText(
                        image.fullUrl,
                        style: const TextStyle(
                          fontSize: 13,
                          fontFamily: 'monospace',
                        ),
                      ),
                    ),
                    const SizedBox(height: 16),
                    
                    // 操作按钮
                    Row(
                      children: [
                        Expanded(
                          child: ElevatedButton.icon(
                            onPressed: () {
                              Navigator.pop(context);
                              _showCopyOptions(image);
                            },
                            icon: const Icon(Icons.copy),
                            label: const Text('复制链接'),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: ElevatedButton.icon(
                            onPressed: () {
                              Navigator.pop(context);
                              _deleteImage(image);
                            },
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.red,
                              foregroundColor: Colors.white,
                            ),
                            icon: const Icon(Icons.delete),
                            label: const Text('删除'),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(color: Colors.grey.shade600)),
          Flexible(
            child: Text(
              value,
              textAlign: TextAlign.right,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')} '
        '${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('图片管理'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => _loadImages(refresh: true),
          ),
        ],
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(_error!, style: const TextStyle(color: Colors.red)),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => _loadImages(refresh: true),
              child: const Text('重试'),
            ),
          ],
        ),
      );
    }
    
    if (_images.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.image_not_supported, size: 64, color: Colors.grey),
            SizedBox(height: 16),
            Text('还没有上传任何图片', style: TextStyle(color: Colors.grey)),
          ],
        ),
      );
    }
    
    return NotificationListener<ScrollNotification>(
      onNotification: (notification) {
        if (notification is ScrollEndNotification &&
            notification.metrics.extentAfter < 200) {
          _loadMore();
        }
        return false;
      },
      child: RefreshIndicator(
        onRefresh: () => _loadImages(refresh: true),
        child: GridView.builder(
          padding: const EdgeInsets.all(8),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 3,
            crossAxisSpacing: 8,
            mainAxisSpacing: 8,
          ),
          itemCount: _images.length + (_loadingMore ? 1 : 0),
          itemBuilder: (context, index) {
            if (index >= _images.length) {
              return const Center(child: CircularProgressIndicator());
            }
            
            final image = _images[index];
            return GestureDetector(
              onTap: () => _showImageDetail(image),
              onLongPress: () => _showCopyOptions(image),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(8),
                child: CachedNetworkImage(
                  imageUrl: image.fullUrl,
                  fit: BoxFit.cover,
                  placeholder: (context, url) => Container(
                    color: Colors.grey.shade200,
                    child: const Center(
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  ),
                  errorWidget: (context, url, error) => Container(
                    color: Colors.grey.shade200,
                    child: const Icon(Icons.error),
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }
}
