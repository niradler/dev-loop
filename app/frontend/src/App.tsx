import React from "react";
import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ApiKeyModal } from "@/components/Auth";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { HashRouter , Routes, Route, useNavigate } from "react-router-dom";
import { ScriptsProvider } from "@/hooks/ScriptsContext";
import { setNavigate } from "@/services/navigationService";
import Index from "./pages/Index";
import ScriptDetail from "./pages/ScriptDetail";
import Settings from "./pages/Settings";
import NotFound from "./pages/NotFound";

// Create a client with default options
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      refetchOnWindowFocus: false,
    },
  },
});

function NavigationInitializer() {
  const navigate = useNavigate();
  React.useEffect(() => {
    setNavigate(navigate);
  }, [navigate]);
  return null;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Index />} />
      <Route path="/auth" element={<ApiKeyModal isOpen={true} />} />
      <Route path="/script/:id" element={<ScriptDetail />} />
      <Route path="/settings" element={<Settings />} />
      <Route path="/category/:category" element={<Index />} />
      <Route path="/recent" element={<Index />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ScriptsProvider>
        <HashRouter >
          <TooltipProvider>
            <NavigationInitializer />
            <AppRoutes />
            <Toaster />
            <Sonner />
          </TooltipProvider>
        </HashRouter >
      </ScriptsProvider>
    </QueryClientProvider>
  );
}

export default App;
