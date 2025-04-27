import { Script, ScriptExecution, ScriptExecutionInput, AppConfig, ScriptInput } from '@/types/script';
import axios, { AxiosError } from 'axios';
import { navigationService } from './navigationService';

interface ApiScriptResponse {
  id: string;
  name: string;
  description: string;
  author: string;
  version?: string;
  category: string;
  tags: string[];
  inputs: ScriptInput[];
  path: string;
  content?: string;
}

interface ApiExecutionResponse {
  executed_at: string;
  script_id: string;
  finished_at?: string;
  exitcode: number;
  execute_request: {
    command: string;
    args: string[];
    env?: Record<string, string>;
  };
  output: string;
  incognito: boolean;
}

interface CategoryResponse {
  category: string;
  count: number;
}

const API_BASE = `http://localhost:${import.meta.env.VITE_DEV_LOOP_PORT || '8997'}/api`;

// Create axios instance with default config
const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to handle authentication
api.interceptors.request.use((config) => {
  const apiKey = localStorage.getItem('dev-loop-api-key');
  if (apiKey) {
    config.headers.Authorization = `Bearer ${apiKey}`;
  }
  return config;
});

// Add response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    console.log('api error:', error);
    if (error?.status === 401) {
      navigationService.toAuth();
    }
    return Promise.reject(error);
  }
);

export const apiService = {
  getAppConfig: async (): Promise<AppConfig> => {
    const { data } = await api.get<AppConfig>('/config');
    return data;
  },

  updateAppConfig: async (config: AppConfig): Promise<AppConfig> => {
    const { data } = await api.post<AppConfig>('/config', config);
    return data;
  },

  getScripts: async (params?: {
    search?: string;
    limit?: number;
    page?: number;
    category?: string;
    tag?: string;
  }): Promise<Script[]> => {
    const { data } = await api.get<ApiScriptResponse[]>('/scripts', { params });
    return data.map((s) => ({
      id: s.id,
      name: s.name,
      description: s.description,
      author: s.author,
      version: s.version,
      category: s.category,
      tags: s.tags,
      inputs: s.inputs,
      path: s.path,
      lastExecuted: undefined
    }));
  },

  getRecentScripts: async (): Promise<Script[]> => {
    const { data } = await api.get<ApiScriptResponse[]>('/history/scripts/recent');
    return data.map((s) => ({
      id: s.id,
      name: s.name,
      description: s.description,
      author: s.author,
      version: s.version,
      category: s.category,
      tags: s.tags,
      inputs: s.inputs,
      path: s.path,
      lastExecuted: undefined
    }));
  },

  reloadScripts: async (): Promise<void> => {
    const config = await apiService.getAppConfig();
    await api.post('/actions/scripts/load', { folders: config.scriptFolders });
  },

  getScriptById: async (id: string): Promise<Script | undefined> => {
    try {
      const { data: s } = await api.get<ApiScriptResponse>(`/scripts/${id}`);
      return {
        id: s.id,
        name: s.name,
        description: s.description,
        author: s.author,
        version: s.version,
        category: s.category,
        tags: s.tags,
        inputs: s.inputs,
        path: s.path,
        content: s.content,
        lastExecuted: undefined
      };
    } catch (error) {
      return undefined;
    }
  },

  toggleFavorite: async (id: string): Promise<Script | undefined> => {
    return apiService.getScriptById(id);
  },

  getExecutions: async (): Promise<ScriptExecution[]> => {
    return [];
  },

  getExecutionsByScriptId: async (scriptId: string): Promise<ScriptExecution[]> => {
    const { data } = await api.get<ApiExecutionResponse[]>(`/history/scripts/${scriptId}`);
    return data.map((e) => ({
      id: e.executed_at,
      scriptId: e.script_id,
      timestamp: e.executed_at,
      command: e.execute_request.command,
      env: Object.keys(e.execute_request.env ?? {}).map(key => ({
        name: key,
        value: e.execute_request.env[key]
      })),
      status: e.exitcode === 0 ? 'success' : 'error',
      inputs: (e.execute_request?.args || []).map((arg: string, idx: number) => ({
        name: `arg${idx + 1}`,
        value: arg
      })),
      output: e.output,
      duration: e.finished_at && e.executed_at
        ? new Date(e.finished_at).getTime() - new Date(e.executed_at).getTime()
        : 0,
      incognito: e.incognito
    }));
  },

  executeScript: async (scriptId: string, inputs: ScriptExecutionInput[], isIncognito: boolean = false): Promise<string> => {
    const args = inputs.map(i => i.value.toString());
    const env: Record<string, string> = {};
    inputs.forEach(i => {
      if (i.name && i.name.startsWith("env:")) {
        env[i.name.slice(4)] = String(i.value);
      }
    });
    const command = "";
    const payload = {
      args,
      env,
      command,
    };
    const { data } = await api.post<string>(`/actions/exec/scripts/${scriptId}?incognito=${isIncognito ? 'true' : 'false'}`, payload);
    return data;
  },

  editScript: async (id: string): Promise<void> => {
    await api.patch(`/scripts/${id}`);
  },

  rerunExecution: async (executionId: string): Promise<ScriptExecution> => {
    throw new Error('Rerun execution via API not implemented');
  },

  getCategories: async (): Promise<CategoryResponse[]> => {
    const { data } = await api.get<CategoryResponse[]>('/categories');
    return data;
  }
};
