import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ScriptExecution } from "@/types/script";
import { formatDistanceToNow, format } from "date-fns";
import { Terminal } from "@/components/ui/terminal";
import { AlertCircle, CheckCircle, Clock, PlayCircle, EyeOff, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { useDeleteHistory } from "@/hooks/useApi";
import { useToast } from "@/hooks/use-toast";

interface ScriptExecutionHistoryProps {
  executions: ScriptExecution[];
  isLoading?: boolean;
  onRerun: (execution: ScriptExecution) => void;
  scriptId?: string;
}

export function ScriptExecutionHistory({ executions, isLoading, onRerun, scriptId }: ScriptExecutionHistoryProps) {
  const { toast } = useToast();
  const { mutateAsync: deleteHistory } = useDeleteHistory(scriptId);

  const handleDeleteHistory = async (id: string) => {
    try {
      await deleteHistory(id);
      toast({
        title: 'History entry deleted successfully',
      });
    } catch (error) {
      toast({
        title: 'Failed to delete history entry',
        description: error instanceof Error ? error.message : 'An error occurred',
        variant: 'destructive',
      });
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Execution History</CardTitle>
          <CardDescription>Loading executions...</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (executions.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Execution History</CardTitle>
          <CardDescription>No executions yet</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center p-6 text-muted-foreground">
            <Clock className="h-12 w-12 mx-auto mb-2 opacity-30" />
            <p>This script hasn't been executed yet.</p>
            <p>Configure the parameters and run the script to see results here.</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Execution History</CardTitle>
        <CardDescription>Previous runs and their outputs</CardDescription>
      </CardHeader>
      <CardContent>
        <Accordion type="single" collapsible className="space-y-4">
          {executions.map((execution) => (
            <AccordionItem 
              value={execution.id} 
              key={execution.id} 
              className={cn(
                "border rounded-md overflow-hidden",
                execution.status === "success" && "border-terminal-success/30",
                execution.status === "error" && "border-terminal-error/30",
                execution.status === "running" && "border-terminal-warning/30"
              )}
            >
              <AccordionTrigger className="px-4 py-2 hover:no-underline">
                <div className="flex items-center w-full">
                  <div className="mr-3">
                    {execution.status === "success" && (
                      <CheckCircle className="h-5 w-5 text-terminal-success" />
                    )}
                    {execution.status === "error" && (
                      <AlertCircle className="h-5 w-5 text-terminal-error" />
                    )}
                    {execution.status === "running" && (
                      <div className="h-5 w-5 rounded-full border-2 border-terminal-warning border-t-transparent animate-spin" />
                    )}
                  </div>
                  <div className="flex-1 text-left">
                    <div className="font-medium flex items-center gap-2">
                      {format(new Date(execution.timestamp), "MMM d, yyyy 'at' h:mm a")}
                      {execution.incognito && (
                        <Badge variant="outline" size="sm" className="text-xs gap-1 text-muted-foreground">
                          <EyeOff className="h-3 w-3" />
                          Incognito
                        </Badge>
                      )}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {formatDistanceToNow(new Date(execution.timestamp))} ago
                      {execution.duration && ` · ${(execution.duration / 1000).toFixed(1)}s`}
                      {execution.inputs.length > 0 && ` · ${execution.inputs.length} parameters`}
                    </div>
                  </div>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <div className="space-y-4 p-4 pt-0">
                  <div>
                    <span className="font-mono text-muted-foreground mr-2">Command:</span>
                    <span className="font-medium truncate">{String(execution.command)}</span>
                  </div>

                  {execution.inputs.length > 0 && (
                    <div>
                      <h4 className="text-sm font-medium mb-2">Args:</h4>
                      {execution.incognito ? (
                        <div className="text-sm text-muted-foreground italic flex items-center">
                          <EyeOff className="h-3.5 w-3.5 mr-2" />
                          Args values hidden (incognito mode)
                        </div>
                      ) : (
                        <div className="grid grid-cols-1 gap-x-4 gap-y-1">
                          {execution.inputs.map((input, i) => (
                            <div key={input.name} className="flex text-sm">
                              <span className="font-mono text-muted-foreground mr-2">${i}:</span>
                              <span className="font-medium truncate">{String(input.value)}</span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}

                  {execution.env.length > 0 && (
                    <div>
                      <h4 className="text-sm font-medium mb-2">Environment Variables:</h4>
                      {execution.incognito ? (
                        <div className="text-sm text-muted-foreground italic flex items-center">
                          <EyeOff className="h-3.5 w-3.5 mr-2" />
                          Env values hidden (incognito mode)
                        </div>
                      ) : (
                        <div className="grid grid-cols-1 gap-x-4 gap-y-1">
                          {execution.env.map((input) => (
                            <div key={input.name} className="flex text-sm">
                              <span className="font-mono text-muted-foreground mr-2">{input.name}:</span>
                              <span className="font-medium truncate">{String(input.value)}</span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                  
                  <div>
                    <h4 className="text-sm font-medium mb-2">Output:</h4>
                    <Terminal>
                      {execution.incognito ? 
                        "Output hidden (incognito mode enabled)" : 
                        execution.output}
                    </Terminal>
                  </div>
                  
                  <div className="flex justify-end gap-2">
                    {!execution.incognito && (
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => onRerun(execution)}
                      >
                        <PlayCircle className="h-4 w-4 mr-1" />
                        Rerun with same parameters
                      </Button>
                    )}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleDeleteHistory(execution.id)}
                    >
                      <Trash2 className="h-4 w-4 mr-1" />
                      Delete
                    </Button>
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      </CardContent>
    </Card>
  );
}
