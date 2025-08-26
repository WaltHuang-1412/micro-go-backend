#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import fs from 'fs/promises';
import path from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

class GoDevAssistant {
  private server: Server;
  private projectRoot: string;

  constructor() {
    this.server = new Server(
      {
        name: 'go-dev-assistant',
        version: '1.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.projectRoot = path.resolve('../');
    this.setupHandlers();
  }

  private setupHandlers() {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: [
          {
            name: 'analyze_go_files',
            description: '分析 Go 檔案結構並找出可以改善的地方',
            inputSchema: {
              type: 'object',
              properties: {
                directory: {
                  type: 'string',
                  description: '要分析的目錄路徑，預設為 handlers/',
                },
              },
            },
          },
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
            name: 'analyze_docker_logs',
            description: '分析 Docker 容器日誌找出錯誤和效能問題',
            inputSchema: {
              type: 'object',
              properties: {
                container_name: {
                  type: 'string',
                  description: '容器名稱，預設為 go-app',
                  default: 'go-app',
                },
                lines: {
                  type: 'number',
                  description: '要分析的日誌行數',
                  default: 100,
                },
              },
            },
          },
          {
            name: 'check_api_health',
            description: '檢查 API 端點健康狀況',
            inputSchema: {
              type: 'object',
              properties: {
                base_url: {
                  type: 'string',
                  description: 'API 基礎 URL',
                  default: 'http://localhost:8088',
                },
              },
            },
          },
        ],
      };
    });

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;

      try {
        switch (name) {
          case 'analyze_go_files':
            return await this.analyzeGoFiles((args as any)?.directory || 'handlers');
          case 'generate_tests':
            return await this.generateTests((args as any)?.handler_file);
          case 'analyze_docker_logs':
            return await this.analyzeDockerLogs((args as any)?.container_name || 'go-app', (args as any)?.lines || 100);
          case 'check_api_health':
            return await this.checkApiHealth((args as any)?.base_url || 'http://localhost:8088');
          default:
            throw new Error(`Unknown tool: ${name}`);
        }
      } catch (error) {
        return {
          content: [
            {
              type: 'text',
              text: `Error: ${error instanceof Error ? error.message : String(error)}`,
            },
          ],
        };
      }
    });
  }

  private async analyzeGoFiles(directory: string) {
    try {
      const dirPath = path.join(this.projectRoot, directory);
      const files = await fs.readdir(dirPath);
      const goFiles = files.filter(file => file.endsWith('.go'));

      let analysis = `## 分析 ${directory} 目錄中的 Go 檔案\n\n`;
      
      for (const file of goFiles) {
        const filePath = path.join(dirPath, file);
        const content = await fs.readFile(filePath, 'utf-8');
        
        analysis += `### ${file}\n`;
        analysis += `- 行數: ${content.split('\n').length}\n`;
        analysis += `- 函數數量: ${(content.match(/func\s+\w+/g) || []).length}\n`;
        
        // 檢查常見問題
        const issues = [];
        if (content.includes('panic(')) issues.push('使用了 panic');
        if (!content.includes('if err != nil')) issues.push('可能缺少錯誤處理');
        if (content.split('\n').length > 200) issues.push('檔案過長，建議拆分');
        
        if (issues.length > 0) {
          analysis += `- ⚠️  問題: ${issues.join(', ')}\n`;
        } else {
          analysis += `- ✅ 看起來不錯\n`;
        }
        analysis += '\n';
      }

      return {
        content: [
          {
            type: 'text',
            text: analysis,
          },
        ],
      };
    } catch (error) {
      throw new Error(`無法分析檔案: ${error}`);
    }
  }

  private async generateTests(handlerFile: string) {
    try {
      const filePath = path.join(this.projectRoot, handlerFile);
      const content = await fs.readFile(filePath, 'utf-8');
      
      // 找出所有的函數
      const functions = content.match(/func\s+(\w+)\s*\(/g) || [];
      const testFileName = handlerFile.replace('.go', '_test.go');
      
      let testContent = `package ${path.basename(path.dirname(filePath))}\n\n`;
      testContent += `import (\n`;
      testContent += `\t"testing"\n`;
      testContent += `\t"net/http/httptest"\n`;
      testContent += `\t"strings"\n`;
      testContent += `\t"github.com/gin-gonic/gin"\n`;
      testContent += `)\n\n`;
      
      for (const func of functions) {
        const funcName = func.match(/func\s+(\w+)/)?.[1];
        if (funcName && !funcName.startsWith('_')) {
          testContent += `func Test${funcName}(t *testing.T) {\n`;
          testContent += `\t// TODO: 實作 ${funcName} 的測試\n`;
          testContent += `\tgin.SetMode(gin.TestMode)\n`;
          testContent += `\trouter := gin.New()\n`;
          testContent += `\t\n`;
          testContent += `\t// 設置測試數據\n`;
          testContent += `\t// req := httptest.NewRequest("POST", "/api/v1/test", strings.NewReader(\`{}\`))\n`;
          testContent += `\t// w := httptest.NewRecorder()\n`;
          testContent += `\t// router.ServeHTTP(w, req)\n`;
          testContent += `\t\n`;
          testContent += `\t// 驗證結果\n`;
          testContent += `\t// if w.Code != http.StatusOK {\n`;
          testContent += `\t//     t.Errorf("Expected status 200, got %d", w.Code)\n`;
          testContent += `\t// }\n`;
          testContent += `}\n\n`;
        }
      }

      // 寫入測試檔案
      const testFilePath = path.join(this.projectRoot, testFileName);
      await fs.writeFile(testFilePath, testContent);

      return {
        content: [
          {
            type: 'text',
            text: `✅ 成功生成測試檔案: ${testFileName}\n\n${testContent}`,
          },
        ],
      };
    } catch (error) {
      throw new Error(`無法生成測試: ${error}`);
    }
  }

  private async analyzeDockerLogs(containerName: string, lines: number) {
    try {
      const { stdout } = await execAsync(`docker logs --tail ${lines} ${containerName}`);
      
      let analysis = `## Docker 日誌分析 (${containerName})\n\n`;
      
      // 分析錯誤和警告
      const errors = stdout.split('\n').filter(line => 
        line.toLowerCase().includes('error') || 
        line.toLowerCase().includes('panic') ||
        line.toLowerCase().includes('fatal')
      );
      
      const warnings = stdout.split('\n').filter(line => 
        line.toLowerCase().includes('warning') ||
        line.toLowerCase().includes('warn')
      );

      if (errors.length > 0) {
        analysis += `### ❌ 錯誤 (${errors.length})\n`;
        errors.slice(0, 5).forEach(error => {
          analysis += `- ${error}\n`;
        });
        analysis += '\n';
      }

      if (warnings.length > 0) {
        analysis += `### ⚠️  警告 (${warnings.length})\n`;
        warnings.slice(0, 3).forEach(warning => {
          analysis += `- ${warning}\n`;
        });
        analysis += '\n';
      }

      if (errors.length === 0 && warnings.length === 0) {
        analysis += '✅ 沒有發現明顯的錯誤或警告\n\n';
      }

      // 最近的日誌
      analysis += '### 📝 最近的日誌\n';
      analysis += '```\n';
      analysis += stdout.split('\n').slice(-10).join('\n');
      analysis += '\n```\n';

      return {
        content: [
          {
            type: 'text',
            text: analysis,
          },
        ],
      };
    } catch (error) {
      throw new Error(`無法分析 Docker 日誌: ${error}`);
    }
  }

  private async checkApiHealth(baseUrl: string) {
    try {
      const endpoints = [
        '/swagger/index.html',
        '/api/v1/profile',
        '/api/v1/plans/sections'
      ];

      let healthReport = `## API 健康檢查 (${baseUrl})\n\n`;
      
      for (const endpoint of endpoints) {
        try {
          const response = await fetch(`${baseUrl}${endpoint}`);
          const status = response.status;
          
          if (status === 200) {
            healthReport += `✅ ${endpoint} - OK (${status})\n`;
          } else if (status === 401) {
            healthReport += `🔒 ${endpoint} - 需要認證 (${status})\n`;
          } else {
            healthReport += `❌ ${endpoint} - 錯誤 (${status})\n`;
          }
        } catch (error) {
          healthReport += `💥 ${endpoint} - 連線失敗\n`;
        }
      }

      return {
        content: [
          {
            type: 'text',
            text: healthReport,
          },
        ],
      };
    } catch (error) {
      throw new Error(`API 健康檢查失敗: ${error}`);
    }
  }

  async run() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('Go Dev Assistant MCP server running on stdio');
  }
}

const assistant = new GoDevAssistant();
assistant.run().catch(console.error);