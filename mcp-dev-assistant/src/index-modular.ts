#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import path from 'path';

// Import modules
import { ModuleManager } from './ModuleManager.js';
import { CodeAnalysisModule } from './modules/CodeAnalysisModule.js';
import { AutomationModule } from './modules/AutomationModule.js';
import { ValidationModule } from './modules/ValidationModule.js';

class ModularGoDevAssistant {
  private server: Server;
  private moduleManager: ModuleManager;
  private projectRoot: string;

  constructor() {
    this.server = new Server(
      {
        name: 'go-dev-assistant-modular',
        version: '2.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.projectRoot = path.resolve('../');
    this.moduleManager = new ModuleManager();
    this.setupHandlers();
  }

  private setupHandlers() {
    // 列出所有可用工具
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      const tools = this.moduleManager.getAvailableTools();
      
      return {
        tools: tools,
      };
    });

    // 處理工具調用
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;
      const result = await this.moduleManager.handleToolCall(name, args as any);
      return {
        content: result.content,
        isError: false,
      };
    });
  }

  private async initializeModules() {
    console.error('🚀 Initializing modular Go Dev Assistant...');
    
    try {
      // 註冊核心模組
      await this.moduleManager.registerModule(
        new CodeAnalysisModule(this.projectRoot)
      );
      
      await this.moduleManager.registerModule(
        new AutomationModule(this.projectRoot)
      );
      
      await this.moduleManager.registerModule(
        new ValidationModule(this.projectRoot)
      );

      // 顯示模組統計
      const stats = this.moduleManager.getStats();
      console.error(`📊 Loaded ${stats.totalModules} modules with ${stats.totalTools} tools:`);
      
      for (const moduleStat of stats.moduleStats) {
        console.error(`   - ${moduleStat.name}: ${moduleStat.toolCount} tools`);
      }

      console.error('✅ All modules initialized successfully!');
      
    } catch (error) {
      console.error(`❌ Failed to initialize modules: ${error}`);
      throw error;
    }
  }

  async run() {
    try {
      // 初始化所有模組
      await this.initializeModules();

      // 設定優雅的關閉處理
      process.on('SIGINT', async () => {
        console.error('🛑 Received SIGINT, shutting down gracefully...');
        await this.shutdown();
        process.exit(0);
      });

      process.on('SIGTERM', async () => {
        console.error('🛑 Received SIGTERM, shutting down gracefully...');
        await this.shutdown();
        process.exit(0);
      });

      // 連接到 stdio transport
      const transport = new StdioServerTransport();
      await this.server.connect(transport);
      
      console.error('🎯 Modular Go Dev Assistant MCP server running on stdio');
      
    } catch (error) {
      console.error(`💥 Failed to start server: ${error}`);
      process.exit(1);
    }
  }

  private async shutdown() {
    console.error('🔄 Shutting down server...');
    try {
      await this.moduleManager.shutdown();
      console.error('✅ Server shutdown complete');
    } catch (error) {
      console.error(`❌ Error during shutdown: ${error}`);
    }
  }
}

// 啟動應用
const assistant = new ModularGoDevAssistant();
assistant.run().catch((error) => {
  console.error(`💥 Unhandled error: ${error}`);
  process.exit(1);
});