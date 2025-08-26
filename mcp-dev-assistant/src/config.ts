// 模組配置
export interface ModuleConfig {
  enabled: boolean;
  priority: number;
  settings?: Record<string, any>;
}

export interface AppConfig {
  modules: {
    codeAnalysis: ModuleConfig;
    automation: ModuleConfig;
    validation: ModuleConfig;
  };
  general: {
    projectRoot: string;
    defaultTimeout: number;
    logLevel: 'debug' | 'info' | 'warn' | 'error';
  };
}

export const defaultConfig: AppConfig = {
  modules: {
    codeAnalysis: {
      enabled: true,
      priority: 1,
      settings: {
        maxFileSize: 1000000, // 1MB
        analysisTimeout: 30000, // 30 seconds
      },
    },
    automation: {
      enabled: true,
      priority: 2,
      settings: {
        generateBackup: true,
        templateStyle: 'standard',
      },
    },
    validation: {
      enabled: true,
      priority: 3,
      settings: {
        testTimeout: 60000, // 60 seconds
        coverageThreshold: 80,
      },
    },
  },
  general: {
    projectRoot: '../',
    defaultTimeout: 120000, // 2 minutes
    logLevel: 'info',
  },
};

export function loadConfig(): AppConfig {
  // 這裡可以從環境變數或設定檔案載入配置
  return { ...defaultConfig };
}