/// API 配置
class ApiConfig {
  // 服务器地址
  static const String baseUrl = 'https://img.leisureea.cn';
  
  // API 路径前缀
  static const String apiPrefix = '/api/v1';
  
  // 完整 API 地址
  static String get apiUrl => '$baseUrl$apiPrefix';
  
  // 图片基础 URL
  static String get imageBaseUrl => '$baseUrl/images';
}
