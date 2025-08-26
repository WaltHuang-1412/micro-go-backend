import { BaseModule } from './BaseModule.js';
import { MCPTool, MCPResponse } from '../types/index.js';
import fs from 'fs/promises';
import path from 'path';

export class CodeAnalysisModule extends BaseModule {
  readonly name = 'code-analysis';
  readonly description = 'Go 程式碼分析與程式碼品質檢查';
  
  readonly tools: MCPTool[] = [
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
      name: 'review_code_quality',
      description: '深度程式碼品質審查',
      inputSchema: {
        type: 'object',
        properties: {
          file_path: {
            type: 'string',
            description: '要審查的檔案路徑',
          },
        },
        required: ['file_path'],
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
        case 'analyze_go_files':
          return await this.analyzeGoFiles(args?.directory || 'handlers');
        case 'review_code_quality':
          return await this.reviewCodeQuality(args?.file_path);
        default:
          return this.createErrorResponse(`Unknown tool: ${toolName}`);
      }
    } catch (error) {
      return this.createErrorResponse(error as Error);
    }
  }

  private async analyzeGoFiles(directory: string): Promise<MCPResponse> {
    const dirPath = path.join(this.projectRoot, directory);
    const files = await fs.readdir(dirPath);
    const goFiles = files.filter(file => file.endsWith('.go'));

    let analysis = `## 🔍 分析 ${directory} 目錄中的 Go 檔案\n\n`;
    
    for (const file of goFiles) {
      const filePath = path.join(dirPath, file);
      const content = await fs.readFile(filePath, 'utf-8');
      
      analysis += `### 📄 ${file}\n`;
      analysis += `- **行數**: ${content.split('\n').length}\n`;
      analysis += `- **函數數量**: ${(content.match(/func\s+\w+/g) || []).length}\n`;
      
      // 進階分析
      const issues = this.findCodeIssues(content);
      const suggestions = this.generateSuggestions(content, file);
      
      if (issues.length > 0) {
        analysis += `- ⚠️  **問題**: ${issues.join(', ')}\n`;
      }
      
      if (suggestions.length > 0) {
        analysis += `- 💡 **建議**: ${suggestions.join(', ')}\n`;
      }
      
      if (issues.length === 0 && suggestions.length === 0) {
        analysis += `- ✅ **狀態**: 程式碼品質良好\n`;
      }
      analysis += '\n';
    }

    // 整體評估
    analysis += this.generateOverallAssessment(goFiles.length);

    return this.createResponse(analysis);
  }

  private async reviewCodeQuality(filePath: string): Promise<MCPResponse> {
    if (!filePath) {
      return this.createErrorResponse('file_path is required');
    }
    
    const fullPath = path.join(this.projectRoot, filePath);
    const content = await fs.readFile(fullPath, 'utf-8');
    const fileName = path.basename(filePath);

    let review = `## 🔎 程式碼品質審查: ${fileName}\n\n`;

    // 詳細分析
    const metrics = this.calculateCodeMetrics(content);
    const patterns = this.analyzeDesignPatterns(content);
    const security = this.checkSecurityIssues(content);
    const performance = this.analyzePerformance(content);

    review += `### 📊 程式碼指標\n`;
    review += `- **複雜度**: ${metrics.complexity}\n`;
    review += `- **可讀性**: ${metrics.readability}\n`;
    review += `- **維護性**: ${metrics.maintainability}\n\n`;

    review += `### 🎨 設計模式分析\n${patterns}\n\n`;
    review += `### 🔒 安全性檢查\n${security}\n\n`;
    review += `### ⚡ 效能分析\n${performance}\n`;

    return this.createResponse(review);
  }

  private findCodeIssues(content: string): string[] {
    const issues = [];
    
    if (content.includes('panic(')) issues.push('使用了 panic');
    if (!content.includes('if err != nil')) issues.push('可能缺少錯誤處理');
    if (content.split('\n').length > 200) issues.push('檔案過長，建議拆分');
    if (content.includes('fmt.Print')) issues.push('包含 debug 輸出');
    const todoMatches = content.match(/\/\/ TODO/g);
    if (todoMatches && todoMatches.length > 3) issues.push('過多未完成項目');
    
    return issues;
  }

  private generateSuggestions(content: string, fileName: string): string[] {
    const suggestions = [];
    
    if (!content.includes('context.Context')) {
      suggestions.push('考慮加入 context 參數');
    }
    
    if (fileName.includes('handler') && !content.includes('gin.HandlerFunc')) {
      suggestions.push('建議使用 gin.HandlerFunc 類型');
    }
    
    if (content.includes('SELECT *')) {
      suggestions.push('避免使用 SELECT *，明確指定欄位');
    }
    
    if (!content.includes('log.')) {
      suggestions.push('考慮加入適當的日誌記錄');
    }

    return suggestions;
  }

  private generateOverallAssessment(fileCount: number): string {
    let assessment = `### 🎯 整體評估\n`;
    assessment += `- **總檔案數**: ${fileCount}\n`;
    
    if (fileCount > 10) {
      assessment += `- **建議**: 考慮將功能分組到不同套件中\n`;
    }
    
    assessment += `- **下一步**: 建議重點關注有問題的檔案，優先修復安全和效能問題\n`;
    
    return assessment;
  }

  private calculateCodeMetrics(content: string): any {
    const lines = content.split('\n');
    const functions = content.match(/func\s+\w+/g) || [];
    
    return {
      complexity: functions.length > 5 ? '高' : functions.length > 2 ? '中' : '低',
      readability: lines.length < 100 ? '良好' : '需要改善',
      maintainability: content.includes('// TODO') ? '需要關注' : '良好'
    };
  }

  private analyzeDesignPatterns(content: string): string {
    const patterns = [];
    
    if (content.includes('interface{') || content.includes('interface {')) {
      patterns.push('✅ 使用了介面抽象');
    }
    
    if (content.match(/func\s+\(\w+\s+\*\w+\)/)) {
      patterns.push('✅ 使用了方法接收者');
    }
    
    if (content.includes('sync.Mutex') || content.includes('sync.RWMutex')) {
      patterns.push('✅ 適當的併發控制');
    }

    return patterns.length > 0 ? patterns.join('\n') : '💡 可考慮使用更多設計模式提升程式碼結構';
  }

  private checkSecurityIssues(content: string): string {
    const issues = [];
    
    if (content.includes('sql.Query') && !content.includes('$')) {
      issues.push('⚠️ 可能存在 SQL 注入風險');
    }
    
    if (content.includes('os.Getenv') && !content.includes('default')) {
      issues.push('💡 建議為環境變數設定預設值');
    }
    
    if (content.includes('http.') && !content.includes('timeout')) {
      issues.push('💡 建議設定 HTTP 請求 timeout');
    }

    return issues.length > 0 ? issues.join('\n') : '✅ 未發現明顯安全問題';
  }

  private analyzePerformance(content: string): string {
    const suggestions = [];
    
    if (content.includes('append') && content.includes('for')) {
      suggestions.push('💡 考慮預分配 slice 容量');
    }
    
    if (content.includes('json.Marshal') && !content.includes('pool')) {
      suggestions.push('💡 考慮使用 JSON 物件池');
    }
    
    if (content.includes('strings.Split') && content.includes('for')) {
      suggestions.push('💡 考慮使用 strings.Fields 或正規表達式');
    }

    return suggestions.length > 0 ? suggestions.join('\n') : '✅ 效能看起來不錯';
  }
}