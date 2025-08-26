import { IModule, MCPTool, MCPResponse } from '../types/index.js';

export abstract class BaseModule implements IModule {
  abstract readonly name: string;
  abstract readonly description: string;
  abstract readonly tools: MCPTool[];

  async initialize(): Promise<void> {
    console.error(`📦 Initializing module: ${this.name}`);
  }

  abstract handleTool(toolName: string, args: any): Promise<MCPResponse>;

  async cleanup(): Promise<void> {
    console.error(`🧹 Cleaning up module: ${this.name}`);
  }

  protected createResponse(text: string): MCPResponse {
    return {
      content: [
        {
          type: 'text',
          text,
        },
      ],
    };
  }

  protected createErrorResponse(error: string | Error): MCPResponse {
    const message = error instanceof Error ? error.message : error;
    return this.createResponse(`❌ Error: ${message}`);
  }
}