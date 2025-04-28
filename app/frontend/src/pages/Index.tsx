import { useState } from "react";
import { AppLayout } from "@/components/AppLayout";
import { ScriptCard } from "@/components/ScriptCard";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import { useParams, useSearchParams, useLocation } from "react-router-dom";
import { useScripts, useRecentScripts } from "@/hooks/useApi";
import { Script } from "@/types/script";

const Index = () => {
  const { category } = useParams<{ category: string }>();
  const [searchParams] = useSearchParams();
  const [searchQuery, setSearchQuery] = useState(searchParams.get("search") || "");
  const location = useLocation();
  const isRecentRoute = location.pathname === "/recent";

  const { data: recentScripts = [], isLoading: isLoadingRecent } = useRecentScripts();
  const { data: scripts = [], isLoading: isLoadingScripts } = useScripts({
    search: searchQuery,
    category,
    tag: searchParams.get("tag") || undefined,
    limit: 50,
    page: 1
  });

  const displayScripts = isRecentRoute ? recentScripts : scripts;
  const isLoading = isRecentRoute ? isLoadingRecent : isLoadingScripts;

  return (
    <AppLayout>
      <div className="container py-6 space-y-6">
        <div className="flex flex-col md:flex-row gap-4 md:items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {isRecentRoute 
                ? 'Recently Executed Scripts'
                : category 
                  ? `Category: ${category}` 
                  : 'Developer Loop'}
            </h1>
            <p className="text-muted-foreground mt-1">
              {isRecentRoute
                ? 'View and manage your recently executed scripts'
                : category 
                  ? `Scripts in the ${category} category`
                  : 'Run, Track, Repeat.'}
            </p>
          </div>
          
          {!isRecentRoute && (
            <div className="relative w-full md:w-64">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input 
                type="search" 
                placeholder="Search scripts..." 
                className="pl-8"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
          )}
        </div>
        
        {isLoading ? (
          <div className="text-center py-12">
            <h2 className="text-2xl font-bold">Loading Scripts...</h2>
          </div>
        ) : displayScripts.length === 0 ? (
          <div className="text-center py-12">
            <h2 className="text-2xl font-bold">No Scripts Found</h2>
            <p className="mt-2 text-muted-foreground">
              {isRecentRoute
                ? "You haven't executed any scripts yet."
                : searchQuery
                  ? "No scripts match your search criteria."
                  : "There are no scripts available."}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {displayScripts.map((script: Script) => (
              <ScriptCard 
                key={script.id} 
                script={script}
              />
            ))}
          </div>
        )}
      </div>
    </AppLayout>
  );
};

export default Index;
