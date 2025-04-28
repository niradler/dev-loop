import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Terminal } from "@/components/ui/terminal";
import { Script } from "@/types/script";
import { Edit } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { apiService } from "@/services/apiService";

interface ScriptCodeViewerProps {
  script: Script;
}

export function ScriptCodeViewer({ script }: ScriptCodeViewerProps) {
  const { toast } = useToast();

  const handleEdit = async () => {
    try {
      await apiService.editScript(script.id);
      toast({
        title: "Success",
        description: "Script opened for editing",
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to open script for editing",
        variant: "destructive",
      });
    }
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Script Code</CardTitle>
            <CardDescription>
              View and edit the script source code
            </CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={handleEdit}>
            <Edit className="h-4 w-4 mr-2" />
            Edit
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <Terminal className="h-[500px]">
          <pre className="whitespace-pre-wrap font-mono text-sm">
            {script.content || "No content available"}
          </pre>
        </Terminal>
      </CardContent>
    </Card>
  );
}
