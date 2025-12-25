# 图床 iOS App

基于 Flutter 开发的图床客户端，支持 iOS 平台。

## 功能

- 📤 图片上传（相册选择 / 拍照）
- 📋 一键复制图片链接
- 🖼️ 图片管理（查看 / 删除）
- ⚙️ Token 鉴权配置

## 开发环境

- Flutter 3.x
- Dart 3.x
- Xcode 15+

## 快速开始

```bash
# 进入项目目录
cd ios-app

# 安装依赖
flutter pub get

# 运行 iOS 模拟器
flutter run -d ios

# 或构建 iOS 应用
flutter build ios
```

## 配置

### API 地址

编辑 `lib/config/api_config.dart`：

```dart
class ApiConfig {
  static const String baseUrl = 'https://img.leisureea.cn';
  // ...
}
```

### Token 设置

如果服务器开启了鉴权，在 App 的「设置」页面配置 Token。

## 项目结构

```
lib/
├── main.dart              # 入口文件
├── config/
│   └── api_config.dart    # API 配置
├── models/
│   └── image_model.dart   # 数据模型
├── services/
│   └── api_service.dart   # API 服务
└── pages/
    ├── upload_page.dart   # 上传页面
    ├── gallery_page.dart  # 图片管理
    └── settings_page.dart # 设置页面
```

## iOS 权限

App 需要以下权限（已在 Info.plist 中配置）：

- 相机访问权限
- 相册访问权限

## 构建发布

```bash
# 构建 Release 版本
flutter build ios --release

# 然后在 Xcode 中打开 ios/Runner.xcworkspace 进行签名和发布
```
