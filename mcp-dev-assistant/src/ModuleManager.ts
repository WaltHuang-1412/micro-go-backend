import { IModule, MCPTool, MCPResponse } from './types/index.js';

export class ModuleManager {
  private modules: Map<string, IModule> = new Map();
  private toolToModuleMap: Map<string, string> = new Map();

  async registerModule(module: IModule): Promise<void> {
    console.error(`🔧 Registering module: ${module.name}`);
    
    // 初始化模組
    await module.initialize();
    
    // 註冊模組
    this.modules.set(module.name, module);
    
    // 建立工具到模組的對應
    for (const tool of module.tools) {
      this.toolToModuleMap.set(tool.name, module.name);
    }
    
    console.error(`✅ Module registered: ${module.name} (${module.tools.length} tools)`);
  }

  async unregisterModule(moduleName: string): Promise<void> {
    const module = this.modules.get(moduleName);
    if (!module) {
      throw new Error(`Module not found: ${moduleName}`);
    }

    // 清理模組
    if (module.cleanup) {
      await module.cleanup();
    }

    // 移除工具對應
    for (const tool of module.tools) {
      this.toolToModuleMap.delete(tool.name);
    }

    // 移除模組
    this.modules.delete(moduleName);
    
    console.error(`🗑️ Module unregistered: ${moduleName}`);
  }

  getAvailableTools(): MCPTool[] {
    const tools: MCPTool[] = [];
    
    for (const module of this.modules.values()) {
      tools.push(...module.tools);
    }
    
    return tools;
  }

  async handleToolCall(toolName: string, args: any): Promise<MCPResponse> {
    const moduleName = this.toolToModuleMap.get(toolName);
    
    if (!moduleName) {
      return {
        content: [
          {
            type: 'text',
            text: `❌ Error: Unknown tool '${toolName}'`,
          },
        ],
      };
    }

    const module = this.modules.get(moduleName);
    if (!module) {
      return {
        content: [
          {
            type: 'text',
            text: `❌ Error: Module '${moduleName}' not found`,
          },
        ],
      };
    }

    try {
      console.error(`🔨 Executing tool: ${toolName} (module: ${moduleName})`);
      const result = await module.handleTool(toolName, args);
      console.error(`✅ Tool executed successfully: ${toolName}`);
      return result;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      console.error(`❌ Tool execution failed: ${toolName} - ${errorMessage}`);
      
      return {
        content: [
          {
            type: 'text',
            text: `❌ Error executing ${toolName}: ${errorMessage}`,
          },
        ],
      };
    }
  }

  getModuleInfo(): Array<{ name: string; description: string; tools: string[] }> {
    return Array.from(this.modules.values()).map(module => ({
      name: module.name,
      description: module.description,
      tools: module.tools.map(tool => tool.name),
    }));
  }

  getToolInfo(toolName: string): MCPTool | null {
    const moduleName = this.toolToModuleMap.get(toolName);
    if (!moduleName) return null;

    const module = this.modules.get(moduleName);
    if (!module) return null;

    return module.tools.find(tool => tool.name === toolName) || null;
  }

  async shutdown(): Promise<void> {
    console.error('🔄 Shutting down module manager...');
    
    for (const [name, module] of this.modules.entries()) {
      try {
        if (module.cleanup) {
          await module.cleanup();
        }
        console.error(`✅ Module ${name} cleaned up`);
      } catch (error) {
        console.error(`❌ Error cleaning up module ${name}: ${error}`);
      }
    }
    
    this.modules.clear();
    this.toolToModuleMap.clear();
    
    console.error('🏁 Module manager shutdown complete');
  }

  // 動態載入/卸載模組
  async enableModule(moduleName: string): Promise<void> {
    const module = this.modules.get(moduleName);
    if (module) {
      console.error(`⚠️ Module ${moduleName} is already enabled`);
      return;
    }
    
    // 這裡可以實作動態載入邏輯
    console.error(`🔄 Loading module: ${moduleName}`);
    // const ModuleClass = await import(`./modules/${moduleName}Module.js`);
    // const moduleInstance = new ModuleClass.default();
    // await this.registerModule(moduleInstance);
  }

  async disableModule(moduleName: string): Promise<void> {
    await this.unregisterModule(moduleName);
  }

  // 獲取模組統計信息
  getStats(): {
    totalModules: number;
    totalTools: number;
    moduleStats: Array<{ name: string; toolCount: number }>;
  } {
    const moduleStats = Array.from(this.modules.values()).map(module => ({
      name: module.name,
      toolCount: module.tools.length,
    }));

    return {
      totalModules: this.modules.size,
      totalTools: this.toolToModuleMap.size,
      moduleStats,
    };
  }
}