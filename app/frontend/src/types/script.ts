export interface ScriptInput {
  name: string;
  description?: string;
  type: 'string' | 'number' | 'boolean' | 'file';
  required?: boolean;
  default?: string | number | boolean;
}

export interface Script {
  id: string;
  name: string;
  description: string;
  author?: string;
  version?: string;
  category?: string;
  tags?: string[];
  inputs: ScriptInput[];
  path: string;
  lastExecuted?: string;
  content?: string;
}

export interface ExecutionConfig {
  python?: string;
  node?: string;
  bash?: string;
  sh?: string;
  [key: string]: string | undefined;
}

export interface ScriptExecutionInput {
  name: string;
  value: string | number | boolean;
}

export interface ScriptExecution {
  id: string;
  scriptId: string;
  timestamp: string;
  command: string;
  env: { name: string; value: string }[];
  status: 'success' | 'error' | 'running';
  inputs: ScriptExecutionInput[];
  output?: string;
  duration?: number;
  incognito?: boolean;
}

export interface AppConfig {
  scriptFolders: string[];
  extensionCommands: {
    [extension: string]: string;
  };
  environmentVariables?: {
    [key: string]: string;
  };
  features?: {
    showCategories?: boolean;
    showRecent?: boolean;
  };
  editor?: string;
}

export interface CategoryResponse {
  category: string;
  count: number;
}
