import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiService } from '@/services/apiService';
import { Script, ScriptExecution, ScriptExecutionInput, AppConfig, CategoryResponse } from '@/types/script';

// App Config
export const useAppConfig = () => {
    return useQuery({
        queryKey: ['appConfig'],
        queryFn: () => apiService.getAppConfig(),
        staleTime: 5 * 60 * 1000, // 5 minutes
    });
};

export const useUpdateAppConfig = () => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (config: AppConfig) => apiService.updateAppConfig(config),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['appConfig'] });
        },
    });
};

// Scripts
export const useScripts = (params?: {
    search?: string;
    limit?: number;
    page?: number;
    category?: string;
    tag?: string;
}) => {
    return useQuery({
        queryKey: ['scripts', params],
        queryFn: () => apiService.getScripts(params),
        staleTime: 5 * 60 * 1000, // 5 minutes
    });
};

export const useScript = (id: string) => {
    return useQuery({
        queryKey: ['script', id],
        queryFn: () => apiService.getScriptById(id),
        enabled: !!id,
        staleTime: 5 * 60 * 1000, // 5 minutes
    });
};

export const useRecentScripts = () => {
    return useQuery({
        queryKey: ['recentScripts'],
        queryFn: () => apiService.getRecentScripts(),
        staleTime: 5 * 60 * 1000, // 5 minutes
    });
};

export const useReloadScripts = () => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: () => apiService.reloadScripts(),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['scripts'] });
            queryClient.invalidateQueries({ queryKey: ['recentScripts'] });
            queryClient.invalidateQueries({ queryKey: ['categories'] });
        },
    });
};

// Script Executions
export const useScriptExecutions = (scriptId: string) => {
    return useQuery({
        queryKey: ['scriptExecutions', scriptId],
        queryFn: () => apiService.getExecutionsByScriptId(scriptId),
        enabled: !!scriptId,
        staleTime: 5 * 60 * 1000, // 5 minutes
    });
};

export const useExecuteScript = (scriptId: string) => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ inputs, isIncognito }: { inputs: ScriptExecutionInput[]; isIncognito: boolean }) =>
            apiService.executeScript(scriptId, inputs, isIncognito),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['scriptExecutions', scriptId] });
        },
    });
};

export const useDeleteScript = () => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ id, removeFile }: { id: string; removeFile?: boolean }) =>
            apiService.deleteScript(id, removeFile),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['scripts'] });
        },
    });
};

export const useDeleteHistory = (scriptId?: string) => {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => apiService.deleteHistory(id),
        onSuccess: () => {
            if (scriptId) {
                queryClient.invalidateQueries({ queryKey: ['scriptExecutions', scriptId] });
            }
            queryClient.invalidateQueries({ queryKey: ['recentScripts'] });
        },
    });
};

// Categories
export const useCategories = () => {
    return useQuery({
        queryKey: ['categories'],
        queryFn: () => apiService.getCategories(),
        staleTime: 5 * 60 * 1000, // 5 minutes
    });
}; 