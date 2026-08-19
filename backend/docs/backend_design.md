# 后端架构设计

## 技术栈选择

  Web框架: Gin
  ORM: GORM,开发环境使用Auto Migration
  配置: Viper
  日志: logrus (封装为logger包)
  认证: JWT (支持管理员和C端用户双JWT体系)
  缓存: Redis/内存缓存 (支持切换)
  数据库: MySQL (推荐)/PostgreSQL
  加密算法: RSA
  国际化: 支持多语言错误信息和枚举显示
  

## 目录结构


基于Gin + GORM技术栈，采用**Clean Architecture**风格的目录结构：

```
backend/
├── cmd/
│   ├── main.go                  # 程序入口
│   ├── license_manager.sh       # 启动脚本
│   ├── client-demo/             # C端客户端演示程序
│   │   ├── main.go
│   │   ├── rsa.go
│   │   ├── client_config.json
│   │   ├── README.md
│   │   └── license_code/
│   └── gen-rsa-keys/            # RSA密钥生成工具
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP处理器
│   │   │   ├── auth_handler.go
│   │   │   ├── authorization_code_handler.go
│   │   │   ├── customer_handler.go
│   │   │   ├── dashboard_handler.go
│   │   │   ├── enum_handler.go
│   │   │   ├── license_handler.go
│   │   │   ├── system_handler.go
│   │   ├── middleware/          # 中间件
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── i18n.go
│   │   │   └── logging.go
│   │   └── routes/              # 路由定义
│   │       └── router.go
│   ├── service/                 # 业务逻辑层
│   │   ├── auth_service.go
│   │   ├── authorization_code_service.go
│   │   ├── customer_service.go
│   │   ├── dashboard_service.go
│   │   ├── enum_service.go
│   │   ├── license_service.go
│   │   ├── system_service.go
│   │   └── interfaces.go        # 接口定义
│   ├── repository/              # 数据访问层
│   │   ├── authorization_code_repository.go
│   │   ├── customer_repository.go
│   │   ├── dashboard_repository.go
│   │   ├── license_repository.go
│   │   ├── user_repository.go
│   │   ├── gorm/                 # GORM相关扩展
│   │   ├── errors.go
│   │   └── interfaces.go
│   ├── models/                  # 数据模型
│   │   ├── auth.go
│   │   ├── common.go
│   │   ├── customer.go
│   │   ├── dashboard.go
│   │   ├── license.go
│   │   ├── system.go
│   │   └── user.go
│   ├── config/                  # 配置管理
│   │   └── config.go
│   └── database/                # 数据库连接和迁移
│       ├── connection.go
│       └── migration.go
├── pkg/
│   ├── cache/                   # 缓存系统
│   │   ├── cache.go
│   │   ├── factory.go
│   │   ├── memory.go
│   │   ├── redis.go
│   │   └── key_builder.go
│   ├── context/                 # 上下文封装
│   │   └── context.go
│   ├── i18n/                    # 国际化
│   │   ├── errors.go
│   │   ├── manager.go
│   │   └── messages/
│   ├── utils/                   # 工具函数
│   │   ├── authorization_code.go # 授权码生成
│   │   ├── crypto.go            # 加密工具
│   │   ├── jwt.go               # JWT工具
│   └── logger/                  # 日志封装
│       └── logger.go
├── configs/
│   ├── config.dev.yaml          # 开发环境配置
│   ├── config.prod.yaml         # 生产环境配置
│   ├── config.yaml              # 默认配置
│   ├── config.example.yaml      # 配置示例
│   └── i18n/                    # 国际化配置
│       └── errors/
│           ├── en-US.yaml
│           ├── ja-JP.yaml
│           └── zh-CN.yaml
├── docs/                        # 文档目录
│   ├── api_development_guide.md
│   ├── backend_design.md
│   ├── design_philosophy.md
│   ├── error_i18n_design.md
│   ├── install/
│   │   ├── db_install.md
│   │   └── swagger_install.md
│   ├── modules_design/
│   │   ├── auth_api.md
│   │   ├── customer-api.md
│   │   ├── customer.md
│   │   ├── login&auth.md
│   │   ├── user.md
│   │   ├── user_story.md
│   │   ├── 产品原型设计方案.md
│   │   └── 授权模块属性设计.md
│   └── swagger/                 # Swagger文档
│       ├── docs.go
│       ├── swagger.json
│       └── swagger.yaml
├── migrations/                  # 数据库迁移
│   ├── 001_create_customers_table.sql
│   ├── 002_insert_sample_data.sql
│   ├── 003_create_authorization_codes_table.sql
│   ├── 004_create_licenses_table.sql
│   ├── 005_create_authorization_changes_table.sql
│   ├── 006_create_users_table.sql
│   ├── 009_update_authorization_codes_code_length.sql
│   └── README.md
├── go.mod
└── go.sum
```

**核心设计原则：**
- **分层架构**：Handler → Service → Repository
- **依赖注入**：通过接口解耦各层
- **配置驱动**：Viper管理所有配置
- **缓存抽象**：支持内存/Redis切换，带过期时间和键构建器
- **双JWT体系**：支持管理员和C端用户独立认证体系
- **国际化支持**：完整的多语言错误信息和枚举显示系统
- **软删除支持**：所有数据表支持gorm.DeletedAt软删除
- **SQL迁移**：复杂的索引和约束通过SQL文件管理，保证生产环境一致性

