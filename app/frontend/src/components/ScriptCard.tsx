import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Script } from "@/types/script";
import { PlayCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Link, useNavigate } from "react-router-dom";

interface ScriptCardProps {
  script: Script;
}

export function ScriptCard({ script }: ScriptCardProps) {
  const navigate = useNavigate();
  if (!script) return null;
  const extension = script.path.split(".").pop();
  const fileExtensionColors: Record<string, string> = {
    py: "bg-blue-600",
    js: "bg-yellow-500",
    ts: "bg-blue-400",
    sh: "bg-green-600",
    bash: "bg-green-700",
    rb: "bg-red-600",
    pl: "bg-purple-600",
    php: "bg-indigo-500"
  };
  const extensionColor = fileExtensionColors[extension] || "bg-gray-600";

  const handleCardClick = () => {
    navigate(`/script/${script.id}`);
  };

  return (
    <Card 
      className="overflow-hidden transition-all hover:shadow-md cursor-pointer flex flex-col h-full" 
      onClick={handleCardClick}
    >
      <CardHeader className="pb-3">
        <div className="flex justify-between items-start">
          <div className="flex items-center gap-2">
            <div className={cn("w-6 h-6 rounded flex items-center justify-center text-xs text-white font-mono", extensionColor)}>
              {extension}
            </div>
            <CardTitle className="text-lg">{script.name}</CardTitle>
          </div>
        </div>
        <CardDescription className="line-clamp-2">
          {script.description}
        </CardDescription>
      </CardHeader>
      <CardContent className="pb-2 flex-1">
        <div className="space-y-2">
          <div className="flex flex-wrap gap-1">
            {script.tags && script.tags.map((tag) => (
              <Badge key={tag} variant="secondary" className="text-xs">
                {tag}
              </Badge>
            ))}
          </div>
          <div className="text-xs text-muted-foreground space-y-1">
            <span className="block truncate">Path: {script.path}</span>
            {script.author && (
              <span className="block">Author: {script.author}</span>
            )}
            {script.version && (
              <span className="block">Version: {script.version}</span>
            )}
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex justify-between pt-2 mt-auto">
        <div className="text-xs text-muted-foreground flex items-center">
          {script.lastExecuted && (
            <>
              <span>Last run: {new Date(script.lastExecuted).toLocaleString()}</span>
            </>
          )}
        </div>
        <Button asChild size="sm" onClick={(e) => e.stopPropagation()}>
          <Link to={`/script/${script.id}`}>
            <PlayCircle className="h-4 w-4 mr-1" />
            Run
          </Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
