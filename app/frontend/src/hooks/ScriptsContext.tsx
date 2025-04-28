import React, { createContext, useContext, ReactNode } from "react";
import { Script, AppConfig, CategoryResponse } from "@/types/script";
import { useToast } from "@/hooks/use-toast";
import { useAppConfig, useScripts, useRecentScripts, useCategories, useReloadScripts } from "@/hooks/useApi";

interface ScriptsContextValue {
  scripts: Script[];
  loading: boolean;
  error: string | null;
  appConfig: AppConfig | null;
  categories: CategoryResponse[];
  recentScripts: Script[];
  refresh: () => Promise<void>;
}

const ScriptsContext = createContext<ScriptsContextValue | undefined>(undefined);

interface ScriptsProviderProps {
  children: ReactNode;
}

export function useScriptsContext() {
  const context = useContext(ScriptsContext);
  if (!context) {
    throw new Error("useScriptsContext must be used within a ScriptsProvider");
  }
  return context;
}

export function ScriptsProvider({ children }: ScriptsProviderProps) {
  const { toast } = useToast();
  
  const { data: scriptsData = [], isLoading: scriptsLoading, error: scriptsError } = useScripts();
  const { data: appConfigData, isLoading: configLoading } = useAppConfig();
  const { data: categoriesData = [], isLoading: categoriesLoading } = useCategories();
  const { data: recentScriptsData = [], isLoading: recentLoading } = useRecentScripts();
  const { mutateAsync: reloadScripts } = useReloadScripts();

  const loading = scriptsLoading || configLoading || categoriesLoading || recentLoading;
  const error = scriptsError ? "Failed to load scripts" : null;

  const refresh = async () => {
    try {
      await reloadScripts();
      toast({
        title: "Success",
        description: "Scripts reloaded successfully",
      });
    } catch (err) {
      toast({
        title: "Error",
        description: "Failed to reload scripts",
        variant: "destructive",
      });
    }
  };

  return (
    <ScriptsContext.Provider
      value={{
        scripts: scriptsData,
        loading,
        error,
        appConfig: appConfigData || null,
        categories: categoriesData,
        recentScripts: recentScriptsData,
        refresh,
      }}
    >
      {children}
    </ScriptsContext.Provider>
  );
}
