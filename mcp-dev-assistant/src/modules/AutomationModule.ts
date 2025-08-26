import { BaseModule } from './BaseModule.js';
import { MCPTool, MCPResponse } from '../types/index.js';
import fs from 'fs/promises';
import path from 'path';

export class AutomationModule extends BaseModule {
  readonly name = 'automation';
  readonly description = '程式碼自動生成與開發自動化';
  
  readonly tools: MCPTool[] = [
    {
      name: 'generate_tests',
      description: '基於現有 handler 自動生成測試案例',
      inputSchema: {
        type: 'object',
        properties: {
          handler_file: {
            type: 'string',
            description: 'Handler 檔案路徑 (例如: handlers/auth.go)',
          },
        },
        required: ['handler_file'],
      },
    },
    {
      name: 'generate_handler',
      description: '根據規格自動生成 handler',
      inputSchema: {
        type: 'object',
        properties: {
          name: {
            type: 'string',
            description: 'Handler 名稱 (例如: user)',
          },
          operations: {
            type: 'array',
            description: '操作列表 (例如: ["create", "read", "update", "delete"])',
          },
        },
        required: ['name', 'operations'],
      },
    },
    {
      name: 'generate_model',
      description: '自動生成資料模型',
      inputSchema: {
        type: 'object',
        properties: {
          name: {
            type: 'string',
            description: '模型名稱 (例如: Product)',
          },
          fields: {
            type: 'array',
            description: '欄位定義 (例如: [{"name": "title", "type": "string"}, {"name": "price", "type": "float64"}])',
          },
        },
        required: ['name', 'fields'],
      },
    },
    {
      name: 'generate_migration',
      description: '自動生成資料庫 migration',
      inputSchema: {
        type: 'object',
        properties: {
          description: {
            type: 'string',
            description: 'Migration 描述 (例如: create_products_table)',
          },
          table_name: {
            type: 'string',
            description: '表格名稱',
          },
          columns: {
            type: 'array',
            description: '欄位定義',
          },
        },
        required: ['description', 'table_name'],
      },
    },
  ];

  private projectRoot: string;

  constructor(projectRoot: string) {
    super();
    this.projectRoot = projectRoot;
  }

  async handleTool(toolName: string, args: any): Promise<MCPResponse> {
    try {
      switch (toolName) {
        case 'generate_tests':
          return await this.generateTests(args?.handler_file);
        case 'generate_handler':
          return await this.generateHandler(args?.name, args?.operations);
        case 'generate_model':
          return await this.generateModel(args?.name, args?.fields);
        case 'generate_migration':
          return await this.generateMigration(args?.description, args?.table_name, args?.columns);
        default:
          return this.createErrorResponse(`Unknown tool: ${toolName}`);
      }
    } catch (error) {
      return this.createErrorResponse(error as Error);
    }
  }

  private async generateTests(handlerFile: string): Promise<MCPResponse> {
    const filePath = path.join(this.projectRoot, handlerFile);
    const content = await fs.readFile(filePath, 'utf-8');
    
    // 分析 handler 檔案
    const functions = this.extractFunctions(content);
    const packageName = this.extractPackageName(content);
    const testFileName = handlerFile.replace('.go', '_test.go');
    
    let testContent = this.generateTestFileHeader(packageName);
    
    for (const func of functions) {
      testContent += this.generateTestFunction(func, content);
    }
    
    // 寫入測試檔案
    const testFilePath = path.join(this.projectRoot, testFileName);
    await fs.writeFile(testFilePath, testContent);

    return this.createResponse(`✅ 成功生成測試檔案: ${testFileName}\n\n\`\`\`go\n${testContent}\n\`\`\``);
  }

  private async generateHandler(name: string, operations: string[]): Promise<MCPResponse> {
    const handlerContent = this.createHandlerTemplate(name, operations);
    const routeContent = this.createRouteTemplate(name, operations);
    
    const handlerFile = path.join(this.projectRoot, 'handlers', `${name}.go`);
    const routeFile = path.join(this.projectRoot, 'routes', `${name}.go`);
    
    await fs.writeFile(handlerFile, handlerContent);
    await fs.writeFile(routeFile, routeContent);
    
    let result = `🚀 成功生成 ${name} handler!\n\n`;
    result += `### 📄 handlers/${name}.go\n\`\`\`go\n${handlerContent}\n\`\`\`\n\n`;
    result += `### 📄 routes/${name}.go\n\`\`\`go\n${routeContent}\n\`\`\``;
    
    return this.createResponse(result);
  }

  private async generateModel(name: string, fields: any[]): Promise<MCPResponse> {
    const modelContent = this.createModelTemplate(name, fields);
    const modelFile = path.join(this.projectRoot, 'models', `${name.toLowerCase()}.go`);
    
    await fs.writeFile(modelFile, modelContent);
    
    return this.createResponse(`✅ 成功生成 ${name} 模型!\n\n\`\`\`go\n${modelContent}\n\`\`\``);
  }

  private async generateMigration(description: string, tableName: string, columns: any[] = []): Promise<MCPResponse> {
    const timestamp = new Date().toISOString().replace(/[-:.]/g, '').slice(0, 14);
    const migrationNumber = await this.getNextMigrationNumber();
    
    const upContent = this.createMigrationUpTemplate(tableName, columns);
    const downContent = this.createMigrationDownTemplate(tableName);
    
    const upFile = path.join(this.projectRoot, 'migrations', `${migrationNumber}_${description}.up.sql`);
    const downFile = path.join(this.projectRoot, 'migrations', `${migrationNumber}_${description}.down.sql`);
    
    await fs.writeFile(upFile, upContent);
    await fs.writeFile(downFile, downContent);
    
    let result = `🗄️ 成功生成 migration 檔案!\n\n`;
    result += `### ⬆️ ${migrationNumber}_${description}.up.sql\n\`\`\`sql\n${upContent}\n\`\`\`\n\n`;
    result += `### ⬇️ ${migrationNumber}_${description}.down.sql\n\`\`\`sql\n${downContent}\n\`\`\``;
    
    return this.createResponse(result);
  }

  // 輔助方法
  private extractFunctions(content: string): string[] {
    const matches = content.match(/func\s+(\w+)/g) || [];
    return matches.map(match => match.replace('func ', '')).filter(name => !name.startsWith('_'));
  }

  private extractPackageName(content: string): string {
    const match = content.match(/package\s+(\w+)/);
    return match ? match[1] : 'handlers';
  }

  private generateTestFileHeader(packageName: string): string {
    return `package ${packageName}

import (
\t"testing"
\t"net/http/httptest"
\t"strings"
\t"github.com/gin-gonic/gin"
\t"github.com/stretchr/testify/assert"
)

`;
  }

  private generateTestFunction(funcName: string, originalContent: string): string {
    const isHandler = originalContent.includes('gin.Context');
    
    if (isHandler) {
      return this.generateHandlerTest(funcName);
    } else {
      return this.generateRegularTest(funcName);
    }
  }

  private generateHandlerTest(funcName: string): string {
    return `func Test${funcName}(t *testing.T) {
\tgin.SetMode(gin.TestMode)
\trouter := gin.New()
\t
\t// 設置路由
\trouter.GET("/test", ${funcName}())
\t
\tt.Run("成功案例", func(t *testing.T) {
\t\treq := httptest.NewRequest("GET", "/test", nil)
\t\tw := httptest.NewRecorder()
\t\trouter.ServeHTTP(w, req)
\t\t
\t\tassert.Equal(t, http.StatusOK, w.Code)
\t})
\t
\tt.Run("錯誤案例", func(t *testing.T) {
\t\t// TODO: 實作錯誤案例測試
\t})
}

`;
  }

  private generateRegularTest(funcName: string): string {
    return `func Test${funcName}(t *testing.T) {
\t// TODO: 實作 ${funcName} 的測試
\tt.Run("基本測試", func(t *testing.T) {
\t\t// 準備測試資料
\t\t
\t\t// 執行函數
\t\t
\t\t// 驗證結果
\t\tassert.NotNil(t, nil, "請實作測試邏輯")
\t})
}

`;
  }

  private createHandlerTemplate(name: string, operations: string[]): string {
    const capitalizedName = name.charAt(0).toUpperCase() + name.slice(1);
    
    let content = `package handlers

import (
\t"net/http"
\t"github.com/gin-gonic/gin"
\t"github.com/Walter1412/micro-backend/models"
)

`;

    for (const operation of operations) {
      content += this.generateHandlerFunction(capitalizedName, operation);
    }

    return content;
  }

  private generateHandlerFunction(name: string, operation: string): string {
    const funcName = operation.charAt(0).toUpperCase() + operation.slice(1) + name;
    
    switch (operation.toLowerCase()) {
      case 'create':
        return `// @Summary Create ${name}
// @Description Create a new ${name}
// @Tags ${name}
// @Accept json
// @Produce json
// @Param ${name} body models.${name} true "${name} object"
// @Success 201 {object} models.${name}
// @Router /${name.toLowerCase()} [post]
func ${funcName}() gin.HandlerFunc {
\treturn func(c *gin.Context) {
\t\tvar ${name.toLowerCase()} models.${name}
\t\tif err := c.ShouldBindJSON(&${name.toLowerCase()}); err != nil {
\t\t\tc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
\t\t\treturn
\t\t}
\t\t
\t\t// TODO: 實作建立邏輯
\t\tc.JSON(http.StatusCreated, ${name.toLowerCase()})
\t}
}

`;
      
      case 'read':
        return `// @Summary Get ${name}
// @Description Get ${name} by ID
// @Tags ${name}
// @Produce json
// @Param id path int true "${name} ID"
// @Success 200 {object} models.${name}
// @Router /${name.toLowerCase()}/{id} [get]
func Get${name}() gin.HandlerFunc {
\treturn func(c *gin.Context) {
\t\tid := c.Param("id")
\t\t
\t\t// TODO: 實作查詢邏輯
\t\tc.JSON(http.StatusOK, gin.H{"id": id})
\t}
}

`;
      
      default:
        return `func ${funcName}() gin.HandlerFunc {
\treturn func(c *gin.Context) {
\t\t// TODO: 實作 ${operation} 邏輯
\t\tc.JSON(http.StatusOK, gin.H{"message": "${operation} ${name}"})
\t}
}

`;
    }
  }

  private createRouteTemplate(name: string, operations: string[]): string {
    let content = `package routes

import (
\t"github.com/gin-gonic/gin"
\t"github.com/Walter1412/micro-backend/handlers"
\t"github.com/Walter1412/micro-backend/middlewares"
)

func Register${name.charAt(0).toUpperCase() + name.slice(1)}Routes(router *gin.Engine) {
\tv1 := router.Group("/api/v1")
\tv1.Use(middlewares.JWTMiddleware())
\t{
`;

    for (const operation of operations) {
      const method = this.getHTTPMethod(operation);
      const path = this.getRoutePath(name, operation);
      const handler = this.getHandlerName(name, operation);
      
      content += `\t\tv1.${method}("${path}", handlers.${handler}())\n`;
    }

    content += `\t}
}
`;

    return content;
  }

  private createModelTemplate(name: string, fields: any[]): string {
    let content = `package models

import (
\t"time"
\t"gorm.io/gorm"
)

type ${name} struct {
\tID        uint           \`json:"id" gorm:"primaryKey"\`
\tCreatedAt time.Time      \`json:"created_at"\`
\tUpdatedAt time.Time      \`json:"updated_at"\`
\tDeletedAt gorm.DeletedAt \`json:"deleted_at,omitempty" gorm:"index"\`
\t
`;

    for (const field of fields) {
      const goType = this.convertToGoType(field.type);
      const jsonTag = field.name.toLowerCase();
      content += `\t${field.name.charAt(0).toUpperCase() + field.name.slice(1)} ${goType} \`json:"${jsonTag}"\`\n`;
    }

    content += `}
`;

    return content;
  }

  private createMigrationUpTemplate(tableName: string, columns: any[]): string {
    let content = `CREATE TABLE ${tableName} (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL`;

    for (const column of columns) {
      const sqlType = this.convertToSQLType(column.type);
      content += `,\n    ${column.name} ${sqlType}`;
    }

    content += `
);

CREATE INDEX idx_${tableName}_deleted_at ON ${tableName}(deleted_at);
`;

    return content;
  }

  private createMigrationDownTemplate(tableName: string): string {
    return `DROP TABLE IF EXISTS ${tableName};
`;
  }

  private async getNextMigrationNumber(): Promise<string> {
    try {
      const migrationsDir = path.join(this.projectRoot, 'migrations');
      const files = await fs.readdir(migrationsDir);
      const numbers = files
        .filter(f => f.match(/^\d{6}_/))
        .map(f => parseInt(f.substring(0, 6)))
        .filter(n => !isNaN(n));
      
      const nextNumber = numbers.length > 0 ? Math.max(...numbers) + 1 : 1;
      return nextNumber.toString().padStart(6, '0');
    } catch {
      return '000001';
    }
  }

  private getHTTPMethod(operation: string): string {
    const methodMap: Record<string, string> = {
      'create': 'POST',
      'read': 'GET',
      'update': 'PUT',
      'delete': 'DELETE',
      'list': 'GET'
    };
    return methodMap[operation.toLowerCase()] || 'GET';
  }

  private getRoutePath(name: string, operation: string): string {
    const basePath = `/${name.toLowerCase()}`;
    
    if (operation === 'create' || operation === 'list') {
      return basePath;
    } else {
      return `${basePath}/:id`;
    }
  }

  private getHandlerName(name: string, operation: string): string {
    const capitalizedName = name.charAt(0).toUpperCase() + name.slice(1);
    const capitalizedOperation = operation.charAt(0).toUpperCase() + operation.slice(1);
    
    if (operation === 'read') {
      return `Get${capitalizedName}`;
    } else {
      return `${capitalizedOperation}${capitalizedName}`;
    }
  }

  private convertToGoType(type: string): string {
    const typeMap: Record<string, string> = {
      'string': 'string',
      'int': 'int',
      'float': 'float64',
      'bool': 'bool',
      'time': 'time.Time',
      'json': 'interface{}'
    };
    return typeMap[type.toLowerCase()] || 'string';
  }

  private convertToSQLType(type: string): string {
    const typeMap: Record<string, string> = {
      'string': 'VARCHAR(255)',
      'int': 'INT',
      'float': 'DECIMAL(10,2)',
      'bool': 'BOOLEAN',
      'time': 'TIMESTAMP',
      'text': 'TEXT'
    };
    return typeMap[type.toLowerCase()] || 'VARCHAR(255)';
  }
}