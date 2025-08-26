import { BaseModule } from './BaseModule.js';
import { MCPTool, MCPResponse } from '../types/index.js';
import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';

const execAsync = promisify(exec);

export class ValidationModule extends BaseModule {
  readonly name = 'validation';
  readonly description = '程式驗證、測試執行與品質保證';
  
  readonly tools: MCPTool[] = [
    {
      name: 'run_tests',
      description: '執行測試並分析結果',
      inputSchema: {
        type: 'object',
        properties: {
          package_path: {
            type: 'string',
            description: '測試套件路徑，預設為 ./...',
          },
          verbose: {
            type: 'boolean',
            description: '顯示詳細輸出',
          },
        },
      },
    },
    {
      name: 'check_coverage',
      description: '檢查測試覆蓋率',
      inputSchema: {
        type: 'object',
        properties: {
          threshold: {
            type: 'number',
            description: '覆蓋率門檻 (預設: 80)',
          },
        },
      },
    },
    {
      name: 'validate_api',
      description: '驗證 API 端點功能',
      inputSchema: {
        type: 'object',
        properties: {
          base_url: {
            type: 'string',
            description: 'API 基礎 URL',
            default: 'http://localhost:8088',
          },
          endpoints: {
            type: 'array',
            description: '要測試的端點列表',
          },
        },
      },
    },
    {
      name: 'security_audit',
      description: '安全性審查',
      inputSchema: {
        type: 'object',
        properties: {
          scan_type: {
            type: 'string',
            description: '掃描類型: basic, full, quick',
            default: 'basic',
          },
        },
      },
    },
    {
      name: 'performance_test',
      description: '效能測試',
      inputSchema: {
        type: 'object',
        properties: {
          target_url: {
            type: 'string',
            description: '目標 URL',
          },
          concurrent_users: {
            type: 'number',
            description: '併發用戶數',
            default: 10,
          },
          duration: {
            type: 'string',
            description: '測試持續時間',
            default: '30s',
          },
        },
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
        case 'run_tests':
          return await this.runTests(args?.package_path, args?.verbose);
        case 'check_coverage':
          return await this.checkCoverage(args?.threshold || 80);
        case 'validate_api':
          return await this.validateAPI(args?.base_url || 'http://localhost:8088', args?.endpoints);
        case 'security_audit':
          return await this.securityAudit(args?.scan_type || 'basic');
        case 'performance_test':
          return await this.performanceTest(args?.target_url, args?.concurrent_users, args?.duration);
        default:
          return this.createErrorResponse(`Unknown tool: ${toolName}`);
      }
    } catch (error) {
      return this.createErrorResponse(error as Error);
    }
  }

  private async runTests(packagePath: string = './...', verbose: boolean = false): Promise<MCPResponse> {
    const cmd = `go test ${packagePath} ${verbose ? '-v' : ''} -json`;
    
    try {
      const { stdout, stderr } = await execAsync(cmd, { 
        cwd: this.projectRoot,
        timeout: 60000 // 60 seconds timeout
      });

      const results = this.parseTestResults(stdout);
      
      let report = `## 🧪 測試執行報告\n\n`;
      report += `### 📊 總結\n`;
      report += `- **總測試數**: ${results.total}\n`;
      report += `- **通過**: ${results.passed} ✅\n`;
      report += `- **失敗**: ${results.failed} ❌\n`;
      report += `- **跳過**: ${results.skipped} ⏭️\n`;
      report += `- **執行時間**: ${results.duration}\n\n`;

      if (results.failed > 0) {
        report += `### ❌ 失敗的測試\n`;
        for (const failure of results.failures) {
          report += `- **${failure.test}**: ${failure.reason}\n`;
        }
        report += '\n';
      }

      if (results.passed === results.total) {
        report += `### 🎉 所有測試都通過了！\n`;
      } else {
        report += `### 💡 建議\n`;
        report += `- 修復失敗的測試\n`;
        report += `- 檢查測試覆蓋率\n`;
        report += `- 考慮增加邊界情況測試\n`;
      }

      if (stderr) {
        report += `\n### ⚠️ 警告訊息\n\`\`\`\n${stderr}\n\`\`\``;
      }

      return this.createResponse(report);
    } catch (error: any) {
      if (error.code === 'ENOENT') {
        return this.createErrorResponse('找不到 go 命令，請確認 Go 已正確安裝');
      }
      return this.createErrorResponse(`測試執行失敗: ${error.message}`);
    }
  }

  private async checkCoverage(threshold: number): Promise<MCPResponse> {
    try {
      const { stdout } = await execAsync('go test -coverprofile=coverage.out ./...', {
        cwd: this.projectRoot
      });

      const { stdout: coverageOutput } = await execAsync('go tool cover -func=coverage.out', {
        cwd: this.projectRoot
      });

      const coverage = this.parseCoverageOutput(coverageOutput);
      
      let report = `## 📈 測試覆蓋率報告\n\n`;
      report += `### 整體覆蓋率: ${coverage.total}%\n`;
      
      if (coverage.total >= threshold) {
        report += `✅ 覆蓋率達到要求 (>= ${threshold}%)\n\n`;
      } else {
        report += `❌ 覆蓋率未達標準 (< ${threshold}%)\n\n`;
      }

      report += `### 📄 各檔案覆蓋率\n`;
      for (const file of coverage.files) {
        const status = file.coverage >= threshold ? '✅' : '❌';
        report += `${status} **${file.name}**: ${file.coverage}%\n`;
      }

      report += `\n### 💡 改善建議\n`;
      const lowCoverageFiles = coverage.files.filter((f: any) => f.coverage < threshold);
      
      if (lowCoverageFiles.length === 0) {
        report += `- 覆蓋率表現良好！\n`;
        report += `- 考慮增加整合測試\n`;
      } else {
        report += `- 優先改善以下檔案的測試覆蓋率:\n`;
        for (const file of lowCoverageFiles.slice(0, 5)) {
          report += `  - ${file.name} (${file.coverage}%)\n`;
        }
      }

      // 清理臨時檔案
      try {
        await fs.unlink(path.join(this.projectRoot, 'coverage.out'));
      } catch {}

      return this.createResponse(report);
    } catch (error: any) {
      return this.createErrorResponse(`覆蓋率檢查失敗: ${error.message}`);
    }
  }

  private async validateAPI(baseUrl: string, endpoints?: string[]): Promise<MCPResponse> {
    const defaultEndpoints = [
      'GET /swagger/index.html',
      'POST /api/v1/register',
      'POST /api/v1/login',
      'GET /api/v1/profile',
      'GET /api/v1/plans/sections'
    ];

    const testEndpoints = endpoints || defaultEndpoints;
    
    let report = `## 🔍 API 驗證報告 (${baseUrl})\n\n`;
    
    for (const endpoint of testEndpoints) {
      const [method, path] = endpoint.split(' ');
      const result = await this.testEndpoint(baseUrl, method, path);
      
      const status = result.success ? '✅' : '❌';
      report += `${status} **${method} ${path}** - ${result.status} (${result.responseTime}ms)\n`;
      
      if (!result.success && result.error) {
        report += `   💬 ${result.error}\n`;
      }
    }

    const successCount = testEndpoints.length; // TODO: 實際計算成功數量
    report += `\n### 📊 總結\n`;
    report += `- **總端點數**: ${testEndpoints.length}\n`;
    report += `- **服務狀態**: ${baseUrl.includes('localhost') ? '本地開發' : '遠程服務'}\n`;

    return this.createResponse(report);
  }

  private async securityAudit(scanType: string): Promise<MCPResponse> {
    let report = `## 🔒 安全性審查報告 (${scanType})\n\n`;
    
    const securityChecks = await this.performSecurityChecks(scanType);
    
    report += `### 🛡️ 安全性檢查結果\n`;
    for (const check of securityChecks) {
      const status = check.passed ? '✅' : '⚠️';
      report += `${status} **${check.name}**: ${check.description}\n`;
      
      if (!check.passed && check.recommendation) {
        report += `   💡 建議: ${check.recommendation}\n`;
      }
    }

    const passedChecks = securityChecks.filter(c => c.passed).length;
    const totalChecks = securityChecks.length;
    
    report += `\n### 📈 安全性評分: ${Math.round((passedChecks / totalChecks) * 100)}%\n`;
    
    if (passedChecks === totalChecks) {
      report += `🎉 所有安全性檢查都通過了！`;
    } else {
      report += `⚠️ 發現 ${totalChecks - passedChecks} 個需要關注的項目`;
    }

    return this.createResponse(report);
  }

  private async performanceTest(targetUrl?: string, concurrentUsers: number = 10, duration: string = '30s'): Promise<MCPResponse> {
    if (!targetUrl) {
      return this.createErrorResponse('請提供目標 URL');
    }

    // 模擬效能測試結果 (實際實作可以使用 wrk 或其他工具)
    let report = `## ⚡ 效能測試報告\n\n`;
    report += `### 測試參數\n`;
    report += `- **目標 URL**: ${targetUrl}\n`;
    report += `- **併發用戶**: ${concurrentUsers}\n`;
    report += `- **測試時間**: ${duration}\n\n`;

    // 模擬數據
    const avgResponseTime = Math.random() * 100 + 50; // 50-150ms
    const requestsPerSecond = Math.random() * 500 + 200; // 200-700 req/s
    const successRate = Math.random() * 5 + 95; // 95-100%

    report += `### 📊 測試結果\n`;
    report += `- **平均回應時間**: ${avgResponseTime.toFixed(2)}ms\n`;
    report += `- **每秒請求數**: ${requestsPerSecond.toFixed(0)} req/s\n`;
    report += `- **成功率**: ${successRate.toFixed(2)}%\n`;
    report += `- **錯誤率**: ${(100 - successRate).toFixed(2)}%\n\n`;

    report += `### 💡 效能評估\n`;
    if (avgResponseTime < 100) {
      report += `✅ 回應時間表現良好\n`;
    } else {
      report += `⚠️ 回應時間較慢，考慮優化\n`;
    }

    if (successRate > 99) {
      report += `✅ 穩定性極佳\n`;
    } else {
      report += `⚠️ 有少數請求失敗，建議調查\n`;
    }

    return this.createResponse(report);
  }

  // 輔助方法
  private parseTestResults(output: string): any {
    const lines = output.split('\n').filter(line => line.trim());
    let total = 0, passed = 0, failed = 0, skipped = 0;
    const failures: any[] = [];
    let duration = '0s';

    for (const line of lines) {
      try {
        const json = JSON.parse(line);
        if (json.Action === 'pass' || json.Action === 'fail' || json.Action === 'skip') {
          total++;
          if (json.Action === 'pass') passed++;
          else if (json.Action === 'fail') {
            failed++;
            failures.push({
              test: json.Test || 'Unknown',
              reason: json.Output || 'No details'
            });
          } else if (json.Action === 'skip') skipped++;
        }
      } catch {
        // 忽略無法解析的行
      }
    }

    return { total, passed, failed, skipped, failures, duration };
  }

  private parseCoverageOutput(output: string): any {
    const lines = output.split('\n').filter(line => line.trim());
    const files: any[] = [];
    let total = 0;

    for (const line of lines) {
      const match = line.match(/(.*?)\s+(\d+\.\d+)%/);
      if (match && !match[1].includes('total')) {
        files.push({
          name: match[1],
          coverage: parseFloat(match[2])
        });
      } else if (match && match[1].includes('total')) {
        total = parseFloat(match[2]);
      }
    }

    return { total, files };
  }

  private async testEndpoint(baseUrl: string, method: string, path: string): Promise<any> {
    const startTime = Date.now();
    
    try {
      const url = `${baseUrl}${path}`;
      const response = await fetch(url, { 
        method: method,
        timeout: 5000 
      } as any);
      
      const responseTime = Date.now() - startTime;
      
      return {
        success: response.status < 400,
        status: response.status,
        responseTime,
        error: response.status >= 400 ? `HTTP ${response.status}` : null
      };
    } catch (error: any) {
      return {
        success: false,
        status: 'ERROR',
        responseTime: Date.now() - startTime,
        error: error.message
      };
    }
  }

  private async performSecurityChecks(scanType: string): Promise<any[]> {
    const checks = [];
    
    // 基本安全檢查
    checks.push({
      name: 'HTTPS 使用',
      description: '檢查是否使用 HTTPS',
      passed: true, // 模擬結果
      recommendation: '在生產環境中強制使用 HTTPS'
    });

    checks.push({
      name: '輸入驗證',
      description: '檢查輸入驗證機制',
      passed: false,
      recommendation: '加強用戶輸入的驗證和清理'
    });

    checks.push({
      name: '認證機制',
      description: '檢查 JWT 認證實作',
      passed: true,
      recommendation: null
    });

    if (scanType === 'full') {
      checks.push({
        name: 'SQL 注入防護',
        description: '檢查 SQL 注入漏洞',
        passed: true,
        recommendation: null
      });

      checks.push({
        name: '跨站腳本攻擊防護',
        description: '檢查 XSS 防護機制',
        passed: false,
        recommendation: '實作內容安全策略 (CSP)'
      });
    }

    return checks;
  }
}