import 'dart:io';
import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../config/api_config.dart';
import '../models/image_model.dart';

/// API 服务
/// 封装所有与后端的通信
class ApiService {
  static final ApiService _instance = ApiService._internal();
  factory ApiService() => _instance;
  
  late Dio _dio;
  String? _token;
  
  ApiService._internal() {
    _dio = Dio(BaseOptions(
      baseUrl: ApiConfig.apiUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
    ));
    
    // 添加拦截器处理 Token
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        if (_token != null && _token!.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $_token';
        }
        return handler.next(options);
      },
      onError: (error, handler) {
        // 统一错误处理
        return handler.next(error);
      },
    ));
  }
  
  /// 初始化，从本地加载 Token
  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('api_token');
  }
  
  /// 设置 Token
  Future<void> setToken(String token) async {
    _token = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('api_token', token);
  }
  
  /// 获取当前 Token
  String? get token => _token;
  
  /// 清除 Token
  Future<void> clearToken() async {
    _token = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('api_token');
  }
  
  /// 上传图片
  /// [file] 图片文件
  /// [onProgress] 进度回调 (0-100)
  Future<ImageModel> uploadImage(
    File file, {
    Function(int)? onProgress,
  }) async {
    final formData = FormData.fromMap({
      'file': await MultipartFile.fromFile(
        file.path,
        filename: file.path.split('/').last,
      ),
    });
    
    final response = await _dio.post(
      '/upload',
      data: formData,
      onSendProgress: (sent, total) {
        if (onProgress != null && total > 0) {
          onProgress((sent / total * 100).round());
        }
      },
    );
    
    final data = response.data;
    if (data['code'] != 0) {
      throw ApiException(data['message'] ?? '上传失败');
    }
    
    return ImageModel.fromJson(data['data']);
  }
  
  /// 获取图片列表
  Future<PaginatedImages> getImages({int page = 1, int pageSize = 20}) async {
    final response = await _dio.get(
      '/images',
      queryParameters: {'page': page, 'page_size': pageSize},
    );
    
    final data = response.data;
    if (data['code'] != 0) {
      throw ApiException(data['message'] ?? '获取列表失败');
    }
    
    return PaginatedImages.fromJson(data['data']);
  }
  
  /// 获取单张图片信息
  Future<ImageModel> getImage(String id) async {
    final response = await _dio.get('/image/$id');
    
    final data = response.data;
    if (data['code'] != 0) {
      throw ApiException(data['message'] ?? '获取图片失败');
    }
    
    return ImageModel.fromJson(data['data']);
  }
  
  /// 删除图片
  Future<void> deleteImage(String id) async {
    final response = await _dio.delete('/image/$id');
    
    final data = response.data;
    if (data['code'] != 0) {
      throw ApiException(data['message'] ?? '删除失败');
    }
  }
}

/// API 异常
class ApiException implements Exception {
  final String message;
  ApiException(this.message);
  
  @override
  String toString() => message;
}
