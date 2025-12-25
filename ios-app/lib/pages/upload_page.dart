import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:image_picker/image_picker.dart';
import '../services/api_service.dart';
import '../models/image_model.dart';

/// 上传任务
class UploadTask {
  final String id;
  final File file;
  final String name;
  int progress;
  String status; // pending, uploading, success, error
  ImageModel? result;
  String? error;

  UploadTask({
    required this.id,
    required this.file,
    required this.name,
    this.progress = 0,
    this.status = 'pending',
    this.result,
    this.error,
  });
}

/// 上传页面
class UploadPage extends StatefulWidget {
  const UploadPage({super.key});

  @override
  State<UploadPage> createState() => _UploadPageState();
}

class _UploadPageState extends State<UploadPage> {
  final ApiService _api = ApiService();
  final ImagePicker _picker = ImagePicker();
  
  List<UploadTask> _tasks = [];
  bool _uploading = false;

  /// 选择图片来源
  Future<void> _showImageSourceDialog() async {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.photo_library),
              title: const Text('从相册选择（可多选）'),
              onTap: () {
                Navigator.pop(context);
                _pickMultipleImages();
              },
            ),
            ListTile(
              leading: const Icon(Icons.camera_alt),
              title: const Text('拍照'),
              onTap: () {
                Navigator.pop(context);
                _pickSingleImage(ImageSource.camera);
              },
            ),
          ],
        ),
      ),
    );
  }

  /// 多选图片
  Future<void> _pickMultipleImages() async {
    try {
      final List<XFile> images = await _picker.pickMultiImage(
        imageQuality: 100,
      );
      
      if (images.isEmpty) return;
      
      // 创建上传任务
      final newTasks = images.map((image) => UploadTask(
        id: '${DateTime.now().millisecondsSinceEpoch}_${image.name}',
        file: File(image.path),
        name: image.name,
      )).toList();
      
      setState(() {
        _tasks.insertAll(0, newTasks);
      });
      
      // 开始上传
      _startUpload();
    } catch (e) {
      _showError('选择图片失败: $e');
    }
  }

  /// 单选图片（拍照）
  Future<void> _pickSingleImage(ImageSource source) async {
    try {
      final XFile? image = await _picker.pickImage(
        source: source,
        imageQuality: 100,
      );
      
      if (image == null) return;
      
      final task = UploadTask(
        id: '${DateTime.now().millisecondsSinceEpoch}_${image.name}',
        file: File(image.path),
        name: image.name,
      );
      
      setState(() {
        _tasks.insert(0, task);
      });
      
      _startUpload();
    } catch (e) {
      _showError('选择图片失败: $e');
    }
  }

  /// 开始上传队列
  Future<void> _startUpload() async {
    if (_uploading) return;
    
    setState(() => _uploading = true);
    
    // 找到所有待上传的任务
    final pendingTasks = _tasks.where((t) => t.status == 'pending').toList();
    
    for (final task in pendingTasks) {
      setState(() => task.status = 'uploading');
      
      try {
        final result = await _api.uploadImage(
          task.file,
          onProgress: (progress) {
            setState(() => task.progress = progress);
          },
        );
        
        setState(() {
          task.status = 'success';
          task.result = result;
        });
      } catch (e) {
        setState(() {
          task.status = 'error';
          task.error = e.toString();
        });
      }
    }
    
    setState(() => _uploading = false);
  }

  /// 显示错误
  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: Colors.red),
    );
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
                _copyToClipboard(image.fullUrl, 'URL');
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
                _copyToClipboard(image.markdownUrl, 'Markdown');
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
                _copyToClipboard(image.htmlUrl, 'HTML');
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
                _copyToClipboard(image.bbcodeUrl, 'BBCode');
              },
            ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  /// 复制到剪贴板
  Future<void> _copyToClipboard(String text, String label) async {
    await Clipboard.setData(ClipboardData(text: text));
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('$label 已复制到剪贴板')),
      );
    }
  }

  /// 清除已完成的任务
  void _clearCompleted() {
    setState(() {
      _tasks.removeWhere((t) => t.status == 'success' || t.status == 'error');
    });
  }

  /// 格式化文件大小
  String _formatSize(int bytes) {
    if (bytes < 1024) return '$bytes B';
    if (bytes < 1024 * 1024) return '${(bytes / 1024).toStringAsFixed(1)} KB';
    return '${(bytes / (1024 * 1024)).toStringAsFixed(2)} MB';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('上传图片'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
        actions: [
          if (_tasks.isNotEmpty)
            TextButton(
              onPressed: _clearCompleted,
              child: const Text('清除'),
            ),
        ],
      ),
      body: Column(
        children: [
          // 上传区域
          GestureDetector(
            onTap: _uploading ? null : _showImageSourceDialog,
            child: Container(
              margin: const EdgeInsets.all(20),
              height: 160,
              decoration: BoxDecoration(
                border: Border.all(
                  color: Colors.grey.shade300,
                  width: 2,
                  style: BorderStyle.solid,
                ),
                borderRadius: BorderRadius.circular(12),
                color: Colors.grey.shade50,
              ),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.cloud_upload_outlined,
                    size: 48,
                    color: Colors.grey.shade400,
                  ),
                  const SizedBox(height: 12),
                  Text(
                    '点击选择图片（支持多选）',
                    style: TextStyle(
                      fontSize: 16,
                      color: Colors.grey.shade600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '支持 JPG、PNG、WebP',
                    style: TextStyle(
                      fontSize: 13,
                      color: Colors.grey.shade400,
                    ),
                  ),
                ],
              ),
            ),
          ),
          
          // 任务列表
          Expanded(
            child: _tasks.isEmpty
                ? Center(
                    child: Text(
                      '暂无上传任务',
                      style: TextStyle(color: Colors.grey.shade400),
                    ),
                  )
                : ListView.builder(
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    itemCount: _tasks.length,
                    itemBuilder: (context, index) {
                      final task = _tasks[index];
                      return _buildTaskItem(task);
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildTaskItem(UploadTask task) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                // 缩略图
                if (task.status == 'success' && task.result != null)
                  ClipRRect(
                    borderRadius: BorderRadius.circular(4),
                    child: Image.network(
                      task.result!.fullUrl,
                      width: 50,
                      height: 50,
                      fit: BoxFit.cover,
                    ),
                  )
                else
                  ClipRRect(
                    borderRadius: BorderRadius.circular(4),
                    child: Image.file(
                      task.file,
                      width: 50,
                      height: 50,
                      fit: BoxFit.cover,
                    ),
                  ),
                const SizedBox(width: 12),
                
                // 信息
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        task.name,
                        style: const TextStyle(fontWeight: FontWeight.w500),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 4),
                      if (task.status == 'uploading')
                        Text(
                          '上传中 ${task.progress}%',
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.grey.shade600,
                          ),
                        )
                      else if (task.status == 'success' && task.result != null)
                        Text(
                          '${task.result!.originalFormat.toUpperCase()} → WebP  ${_formatSize(task.result!.processedSize)}',
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.green.shade600,
                          ),
                        )
                      else if (task.status == 'error')
                        Text(
                          task.error ?? '上传失败',
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.red.shade600,
                          ),
                        )
                      else
                        Text(
                          '等待上传',
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.grey.shade400,
                          ),
                        ),
                    ],
                  ),
                ),
                
                // 操作按钮
                if (task.status == 'success' && task.result != null)
                  IconButton(
                    icon: const Icon(Icons.copy),
                    onPressed: () => _showCopyOptions(task.result!),
                    tooltip: '复制链接',
                  ),
              ],
            ),
            
            // 进度条
            if (task.status == 'uploading') ...[
              const SizedBox(height: 8),
              LinearProgressIndicator(value: task.progress / 100),
            ],
          ],
        ),
      ),
    );
  }
}
