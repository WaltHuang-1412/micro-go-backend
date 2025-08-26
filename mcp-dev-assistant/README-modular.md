# 🏗️ 模組化 Go 開發助手 MCP

重新設計的模組化架構，支援可擴展的功能分工。

## 🎯 **模組化架構優勢**

### ✅ **分工明確**
```bash
📦 CodeAnalysisModule    # 程式碼分析與品質檢查
📦 AutomationModule      # 程式碼自動生成
📦 ValidationModule      # 驗證與測試
```

### ✅ **易於擴展**
```bash
新增模組只需要：
1. 繼承 BaseModule
2. 實作 handleTool 方法  
3. 在 index-modular.ts 中註冊
```

### ✅ **獨立維護**
```bash
每個模組都可以：
- 獨立開發和測試
- 單獨啟用/停用
- 個別配置設定
```

## 🛠️ **可用工具列表**

### 📊 **程式碼分析模組**
| 工具 | 功能 | 使用範例 |
|------|------|----------|
| `analyze_go_files` | 分析 Go 檔案品質 | "分析 handlers 目錄" |
| `review_code_quality` | 深度程式碼審查 | "審查 profile.go" |

### 🤖 **自動化模組**
| 工具 | 功能 | 使用範例 |
|------|------|----------|
| `generate_tests` | 生成測試案例 | "為 auth.go 生成測試" |
| `generate_handler` | 生成 handler | "生成 user CRUD handler" |
| `generate_model` | 生成資料模型 | "生成 Product 模型" |
| `generate_migration` | 生成 migration | "生成 products 表格" |

### ✅ **驗證模組**
| 工具 | 功能 | 使用範例 |
|------|------|----------|
| `run_tests` | 執行測試 | "執行所有測試" |
| `check_coverage` | 檢查覆蓋率 | "檢查測試覆蓋率" |
| `validate_api` | API 驗證 | "驗證 API 端點" |
| `security_audit` | 安全性檢查 | "執行安全審查" |
| `performance_test` | 效能測試 | "測試 API 效能" |

## 🚀 **使用方法**

### **1. 使用模組化版本**
```bash
# 編譯
npm run build

# 在 Claude Desktop 設定中指向新版本
"command": "node",
"args": ["./mcp-dev-assistant/dist/index-modular.js"]
```

### **2. 範例對話**
```bash
你：「分析我的程式碼品質」
→ 使用 analyze_go_files

你：「為 auth.go 生成完整測試」  
→ 使用 generate_tests

你：「執行測試並檢查覆蓋率」
→ 使用 run_tests + check_coverage

你：「生成一個 Product 的 CRUD handler」
→ 使用 generate_handler + generate_model
```

## 🔧 **模組配置**

### **啟用/停用模組**
```typescript
// src/config.ts
modules: {
  codeAnalysis: { enabled: true, priority: 1 },
  automation: { enabled: true, priority: 2 },
  validation: { enabled: false, priority: 3 }, // 停用驗證模組
}
```

### **自訂設定**
```typescript
settings: {
  maxFileSize: 1000000,      // 分析檔案大小限制
  generateBackup: true,      // 生成程式碼時備份
  coverageThreshold: 80,     // 覆蓋率門檻
}
```

## 🆕 **新增模組步驟**

### **1. 建立模組檔案**
```typescript
// src/modules/YourModule.ts
export class YourModule extends BaseModule {
  readonly name = 'your-module';
  readonly description = '你的模組描述';
  readonly tools = [/* 工具定義 */];
  
  async handleTool(toolName: string, args: any): Promise<MCPResponse> {
    // 實作邏輯
  }
}
```

### **2. 註冊模組**
```typescript
// src/index-modular.ts
await this.moduleManager.registerModule(
  new YourModule(this.projectRoot)
);
```

### **3. 更新配置**
```typescript
// src/config.ts
modules: {
  yourModule: { enabled: true, priority: 4 }
}
```

## 📈 **效能監控**

每個模組的執行狀況都會記錄：
```bash
📊 Loaded 3 modules with 11 tools:
   - code-analysis: 2 tools
   - automation: 4 tools  
   - validation: 5 tools
```

## 🔄 **未來擴展計劃**

1. **資料庫模組** - MySQL 操作與優化
2. **部署模組** - Docker 和 CI/CD 管理
3. **監控模組** - 效能和錯誤監控
4. **文件模組** - 自動生成文件

讓你的開發流程更加模組化和高效！