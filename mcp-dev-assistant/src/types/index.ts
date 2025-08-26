// 通用類型定義
export interface MCPTool {
  name: string;
  description: string;
  inputSchema: {
    type: 'object';
    properties: Record<string, any>;
    required?: string[];
  };
}

export interface MCPResponse {
  content: Array<{
    type: 'text';
    text: string;
  }>;
}

export interface ModuleConfig {
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
}

// 基礎模組介面
export interface IModule {
  readonly name: string;
  readonly description: string;
  readonly tools: MCPTool[];
  
  initialize(): Promise<void>;
  handleTool(toolName: string, args: any): Promise<MCPResponse>;
  cleanup?(): Promise<void>;
}